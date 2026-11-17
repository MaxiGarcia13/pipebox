package jenkins

import (
  "path/filepath"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestPresentAndListJenkinsfile(t *testing.T) {
  repo := filepath.Join("..", "..", "..", "testdata", "repos", "jenkins-only")
  if !Present(repo) {
    t.Fatal("expected Present true")
  }

  list, err := List(repo)
  if err != nil {
    t.Fatalf("List: %v", err)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.ID != "Jenkinsfile" || workflow.Path != "Jenkinsfile" {
    t.Fatalf("workflow: %#v", workflow)
  }
  if workflow.Provider != event.ProviderJenkins {
    t.Fatalf("provider: %q", workflow.Provider)
  }
  if len(workflow.Events) != 1 || workflow.Events[0] != "pipeline" {
    t.Fatalf("events: %#v", workflow.Events)
  }
}

func TestPresentFalseWithoutJenkinsfile(t *testing.T) {
  repo := filepath.Join("..", "..", "..", "testdata", "repos", "github-only")
  if Present(repo) {
    t.Fatal("expected Present false")
  }
}

func TestIsJenkinsfileName(t *testing.T) {
  cases := []struct {
    name string
    want bool
  }{
    {"Jenkinsfile", true},
    {"Jenkinsfile.ci", true},
    {"Jenkinsfile.prod", true},
    {"jenkinsfile", false},
    {"Dockerfile", false},
    {"Jenkinsfile.bak.txt", true},
  }
  for _, testCase := range cases {
    if got := isJenkinsfileName(testCase.name); got != testCase.want {
      t.Fatalf("%s: got %v want %v", testCase.name, got, testCase.want)
    }
  }
}

func TestIsJenkinsfilePath(t *testing.T) {
  if !IsJenkinsfilePath("Jenkinsfile") {
    t.Fatal("Jenkinsfile")
  }
  if !IsJenkinsfilePath("/tmp/repo/Jenkinsfile.ci") {
    t.Fatal("absolute Jenkinsfile.ci")
  }
  if IsJenkinsfilePath(".github/workflows/ci.yml") {
    t.Fatal("github path must not match")
  }
}
