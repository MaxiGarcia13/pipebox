package e2e_test

import (
  "bytes"
  "encoding/json"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  pipeboxBin string
  reposDir   string
)

func TestMain(m *testing.M) {
  moduleRoot, err := moduleRoot()
  if err != nil {
    panic(err)
  }
  reposDir = filepath.Join(moduleRoot, "testdata", "repos")

  buildDir, err := os.MkdirTemp("", "pipebox-e2e-")
  if err != nil {
    panic(err)
  }

  pipeboxBin = filepath.Join(buildDir, "pipebox")
  if runtime.GOOS == "windows" {
    pipeboxBin += ".exe"
  }

  build := exec.Command("go", "build", "-o", pipeboxBin, "./cmd/pipebox")
  build.Dir = moduleRoot
  build.Stdout = os.Stdout
  build.Stderr = os.Stderr
  if err := build.Run(); err != nil {
    _ = os.RemoveAll(buildDir)
    panic(err)
  }

  code := m.Run()
  _ = os.RemoveAll(buildDir)
  os.Exit(code)
}

func moduleRoot() (string, error) {
  _, file, _, ok := runtime.Caller(0)
  if !ok {
    return "", os.ErrNotExist
  }
  // packages/runner/internal/e2e → packages/runner
  return filepath.Abs(filepath.Join(filepath.Dir(file), "../.."))
}

func repoPath(t *testing.T, name string) string {
  t.Helper()
  path := filepath.Join(reposDir, name)
  if _, err := os.Stat(path); err != nil {
    t.Fatalf("fixture %s: %v", name, err)
  }
  return path
}

func runPipebox(t *testing.T, args ...string) (stdout string, stderr string, exitCode int) {
  t.Helper()
  return runPipeboxDir(t, "", args...)
}

func runPipeboxDir(t *testing.T, dir string, args ...string) (stdout string, stderr string, exitCode int) {
  t.Helper()
  cmd := exec.Command(pipeboxBin, args...)
  cmd.Dir = dir
  var outBuf, errBuf bytes.Buffer
  cmd.Stdout = &outBuf
  cmd.Stderr = &errBuf
  err := cmd.Run()
  exitCode = 0
  if err != nil {
    exitError, ok := err.(*exec.ExitError)
    if !ok {
      t.Fatalf("run %v: %v", args, err)
    }
    exitCode = exitError.ExitCode()
  }
  return outBuf.String(), errBuf.String(), exitCode
}

func decodeLine[T any](t *testing.T, stdout string) T {
  t.Helper()
  line := strings.TrimSpace(stdout)
  if line == "" {
    t.Fatal("expected NDJSON on stdout")
  }
  // CLI emits one JSON object per command for doctor/list/graph
  firstLine, _, _ := strings.Cut(line, "\n")
  var value T
  if err := json.Unmarshal([]byte(firstLine), &value); err != nil {
    t.Fatalf("decode %q: %v", firstLine, err)
  }
  return value
}

func requireHostDockerE2E(t *testing.T) {
  t.Helper()
  // Nested act (DooD) resolves -v paths on the host daemon; job-container
  // fixture paths like /workspace/... are not valid host mounts.
  if os.Getenv("ACT") != "" {
    t.Skip("cache/artifact e2e needs host docker; nested act is unsupported")
  }
  cmd := exec.Command("docker", "info")
  if err := cmd.Run(); err != nil {
    t.Skip("docker daemon not reachable")
  }
}

func ndjsonLines(stdout string) []string {
  raw := strings.Split(stdout, "\n")
  lines := make([]string, 0, len(raw))
  for _, line := range raw {
    line = strings.TrimSpace(line)
    if line != "" {
      lines = append(lines, line)
    }
  }
  return lines
}

func logContains(lines []string, needle string) bool {
  for _, line := range lines {
    var entry event.Log
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
      continue
    }
    if entry.Type == event.TypeLog && strings.Contains(entry.Line, needle) {
      return true
    }
  }
  return false
}

func assertJobMarkerLogs(t *testing.T, lines []string, jobID, markerPrefix string) {
  t.Helper()
  matched := 0
  for _, line := range lines {
    var entry event.Log
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
      continue
    }
    if entry.Type != event.TypeLog || !strings.Contains(entry.Line, markerPrefix) {
      continue
    }
    matched++
    if entry.Job != jobID {
      t.Fatalf("marker %q tagged job %q, want %q (line=%q)", markerPrefix, entry.Job, jobID, entry.Line)
    }
  }
  if matched == 0 {
    t.Fatalf("missing logs with marker prefix %q", markerPrefix)
  }
}

func lastArtifacts(t *testing.T, lines []string) event.Artifacts {
  t.Helper()
  var last event.Artifacts
  found := false
  for _, line := range lines {
    var entry event.Artifacts
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
      continue
    }
    if entry.Type != event.TypeArtifacts {
      continue
    }
    last = entry
    found = true
  }
  if !found {
    t.Fatal("missing artifacts event")
  }
  return last
}

func lastRunFinished(t *testing.T, lines []string) event.RunFinished {
  t.Helper()
  var last event.RunFinished
  found := false
  for _, line := range lines {
    var entry event.RunFinished
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
      continue
    }
    if entry.Type != event.TypeRunFinished {
      continue
    }
    last = entry
    found = true
  }
  if !found {
    t.Fatal("missing run_finished event")
  }
  return last
}
