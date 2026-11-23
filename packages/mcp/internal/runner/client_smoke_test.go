package runner

import (
  "context"
  "encoding/json"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runcache"
)

func TestRunSkinnySummaryWithFakePipebox(t *testing.T) {
  bin := buildFakePipebox(t)
  client := &Client{Bin: bin}
  cache := runcache.New()

  runID := cache.Begin("/repo", "ci.yml", "")
  eventCount := 0
  exitCode, err := client.Run(context.Background(), RunOptions{
    Repo:     "/repo",
    Workflow: "ci.yml",
  }, func(raw json.RawMessage) error {
    eventCount++
    return cache.Ingest(runID, raw)
  })
  if err != nil {
    t.Fatalf("run: %v", err)
  }
  if exitCode != 1 {
    t.Fatalf("exit code: %d", exitCode)
  }
  if eventCount < 200 {
    t.Fatalf("expected noisy NDJSON stream, got %d events", eventCount)
  }

  summary := cache.Finish(runID, exitCode)
  raw, err := json.Marshal(summary)
  if err != nil {
    t.Fatal(err)
  }
  if strings.Contains(string(raw), "noise-line-") {
    t.Fatalf("summary leaked log lines: %s", raw)
  }
  if len(raw) > 512 {
    t.Fatalf("summary too large (%d bytes): %s", len(raw), raw)
  }
  if summary.Status != "failed" || len(summary.FailedSteps) != 1 {
    t.Fatalf("unexpected summary: %#v", summary)
  }

  lines, err := cache.StepLogs(runID, "build", "test", 80)
  if err != nil {
    t.Fatal(err)
  }
  if len(lines) != 80 {
    t.Fatalf("expected tailed 80 lines, got %d", len(lines))
  }
  if lines[0] != "noise-line-120" || lines[79] != "noise-line-199" {
    t.Fatalf("unexpected tail window: first=%q last=%q", lines[0], lines[79])
  }
}

func TestListWithFakePipebox(t *testing.T) {
  client := &Client{Bin: buildFakePipebox(t)}
  out, err := client.List(context.Background(), ListOptions{Repo: "/repo"})
  if err != nil {
    t.Fatal(err)
  }
  if len(out.Workflows) != 1 || out.Workflows[0].Path != ".github/workflows/ci.yml" {
    t.Fatalf("unexpected list: %#v", out)
  }
}

func buildFakePipebox(t *testing.T) string {
  t.Helper()
  dir := t.TempDir()
  out := filepath.Join(dir, "fake-pipebox")
  cmd := exec.Command("go", "build", "-o", out, "./testdata/fake-pipebox")
  cmd.Dir = filepath.Join(moduleDir(t), "internal", "runner")
  cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
  if output, err := cmd.CombinedOutput(); err != nil {
    t.Fatalf("build fake-pipebox: %v\n%s", err, output)
  }
  return out
}

func moduleDir(t *testing.T) string {
  t.Helper()
  dir, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  for {
    if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
      return dir
    }
    parent := filepath.Dir(dir)
    if parent == dir {
      t.Fatal("go.mod not found")
    }
    dir = parent
  }
}
