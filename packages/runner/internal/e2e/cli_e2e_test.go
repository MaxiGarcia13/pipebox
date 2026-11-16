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
  cmd := exec.Command(pipeboxBin, args...)
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

func TestDoctorEmitsDoctorEvent(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "doctor")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  doctor := decodeLine[event.Doctor](t, stdout)
  if doctor.Type != event.TypeDoctor {
    t.Fatalf("type: got %q", doctor.Type)
  }
  if doctor.Version == "" {
    t.Fatal("expected version")
  }
  if len(doctor.Checks) == 0 {
    t.Fatal("expected checks")
  }
  foundDocker := false
  for _, check := range doctor.Checks {
    if check.Name == "docker" {
      foundDocker = true
      break
    }
  }
  if !foundDocker {
    t.Fatalf("expected docker check, got %#v", doctor.Checks)
  }
}

func TestListGitHubOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "github-only"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Providers) != 1 || list.Providers[0] != event.ProviderGitHub {
    t.Fatalf("providers: %#v", list.Providers)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.ID != "ci" || workflow.Name != "CI" {
    t.Fatalf("workflow identity: %#v", workflow)
  }
  wantPath := filepath.ToSlash(filepath.Join(".github", "workflows", "ci.yml"))
  if filepath.ToSlash(workflow.Path) != wantPath {
    t.Fatalf("path: got %q want %q", workflow.Path, wantPath)
  }
  if workflow.Provider != event.ProviderGitHub {
    t.Fatalf("provider: got %q", workflow.Provider)
  }
  if !workflow.Callable {
    t.Fatal("expected callable")
  }
  if len(workflow.Inputs) < 2 {
    t.Fatalf("inputs: %#v", workflow.Inputs)
  }
}

func TestListEmpty(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "empty"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Workflows) != 0 {
    t.Fatalf("expected no workflows, got %#v", list.Workflows)
  }
}

func TestListJenkinsOnlyDocumentsStub(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "jenkins-only"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  // jenkins.Present is stubbed false — Jenkinsfile alone yields no workflows.
  if len(list.Workflows) != 0 {
    t.Fatalf("expected no workflows until jenkins Present is implemented, got %#v", list.Workflows)
  }
  for _, provider := range list.Providers {
    if provider == event.ProviderJenkins {
      t.Fatalf("unexpected jenkins provider: %#v", list.Providers)
    }
  }
}

func TestListMixedGitHubOnlyUntilJenkinsPresent(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "mixed"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.Name != "Mixed CI" || workflow.Provider != event.ProviderGitHub {
    t.Fatalf("workflow: %#v", workflow)
  }
  for _, provider := range list.Providers {
    if provider == event.ProviderJenkins {
      t.Fatalf("unexpected jenkins provider: %#v", list.Providers)
    }
  }
}

func TestGraphGitHubOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(
    t,
    "graph",
    "--repo", repoPath(t, "github-only"),
    "--workflow", ".github/workflows/ci.yml",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  graph := decodeLine[event.JobGraph](t, stdout)
  if graph.Type != event.TypeJobGraph {
    t.Fatalf("type: got %q", graph.Type)
  }
  if len(graph.Jobs) != 1 {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
  job := graph.Jobs[0]
  if job.ID != "lint" || job.Name != "Lint" {
    t.Fatalf("job: %#v", job)
  }
  if len(job.Steps) != 1 || job.Steps[0] != "echo ok" {
    t.Fatalf("steps: %#v", job.Steps)
  }
}
