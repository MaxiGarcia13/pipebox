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
  daemonOK := false
  foundJenkinsImage := false
  for _, check := range doctor.Checks {
    switch check.Name {
    case "docker":
      foundDocker = true
    case "docker_daemon":
      daemonOK = check.OK
    case "jenkins_image":
      foundJenkinsImage = true
    }
  }
  if !foundDocker {
    t.Fatalf("expected docker check, got %#v", doctor.Checks)
  }
  if daemonOK && !foundJenkinsImage {
    t.Fatalf("expected jenkins_image check when daemon is up, got %#v", doctor.Checks)
  }
  anyFailed := false
  for _, check := range doctor.Checks {
    if !check.OK {
      anyFailed = true
    }
  }
  if doctor.OK == anyFailed {
    t.Fatalf("ok=%v does not match checks %#v", doctor.OK, doctor.Checks)
  }
  wantCode := 0
  if !doctor.OK {
    wantCode = 1
  }
  if code != wantCode {
    t.Fatalf("exit %d want %d stderr=%q stdout=%q", code, wantCode, stderr, stdout)
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

func TestListJenkinsOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "jenkins-only"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.Path != "Jenkinsfile" || workflow.Provider != event.ProviderJenkins {
    t.Fatalf("workflow: %#v", workflow)
  }
  if len(list.Providers) != 1 || list.Providers[0] != event.ProviderJenkins {
    t.Fatalf("providers: %#v", list.Providers)
  }
}

func TestListMixedGitHubAndJenkins(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "mixed"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if len(list.Workflows) != 2 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }

  byProvider := map[string]event.Workflow{}
  for _, workflow := range list.Workflows {
    byProvider[workflow.Provider] = workflow
  }
  githubWorkflow, ok := byProvider[event.ProviderGitHub]
  if !ok || githubWorkflow.Name != "Mixed CI" {
    t.Fatalf("github workflow: %#v", list.Workflows)
  }
  jenkinsWorkflow, ok := byProvider[event.ProviderJenkins]
  if !ok || jenkinsWorkflow.Path != "Jenkinsfile" {
    t.Fatalf("jenkins workflow: %#v", list.Workflows)
  }
  if len(list.Providers) != 2 {
    t.Fatalf("providers: %#v", list.Providers)
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

func TestGraphJenkinsOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(
    t,
    "graph",
    "--repo", repoPath(t, "jenkins-only"),
    "--workflow", "Jenkinsfile",
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
  if job.ID != "build" || job.Name != "Build" {
    t.Fatalf("job: %#v", job)
  }
  if len(job.Steps) != 1 || job.Steps[0] != "echo" {
    t.Fatalf("steps: %#v", job.Steps)
  }
}

func TestCompletionScript(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "completion", "zsh")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  if !strings.Contains(stdout, "pipebox __complete") {
    t.Fatalf("expected __complete hook, got %q", stdout)
  }
}

func TestCompleteRootCommands(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "__complete", "--", "")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  lines := strings.Split(strings.TrimSpace(stdout), "\n")
  found := map[string]bool{}
  for _, line := range lines {
    found[line] = true
  }
  for _, want := range []string{"doctor", "list", "run", "completion"} {
    if !found[want] {
      t.Fatalf("missing %q in %#v", want, lines)
    }
  }
  if found["__complete"] {
    t.Fatal("hidden __complete should not be suggested")
  }
}

func TestCompleteWorkflows(t *testing.T) {
  repo := repoPath(t, "github-only")
  stdout, stderr, code := runPipebox(
    t,
    "__complete",
    "--",
    "run",
    "--repo",
    repo,
    "--workflow",
    "",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  want := filepath.ToSlash(filepath.Join(".github", "workflows", "ci.yml"))
  if !strings.Contains(filepath.ToSlash(stdout), want) {
    t.Fatalf("expected workflow %q in %q", want, stdout)
  }
}
