package runner

import (
  "reflect"
  "testing"
)

func TestListArgs(t *testing.T) {
  got := listArgs(ListOptions{Repo: "/tmp/repo"})
  want := []string{"list", "--repo", "/tmp/repo"}
  if !reflect.DeepEqual(got, want) {
    t.Fatalf("got %#v want %#v", got, want)
  }
}

func TestGraphArgs(t *testing.T) {
  got := graphArgs(GraphOptions{
    Repo:     "/tmp/repo",
    Workflow: ".github/workflows/ci.yml",
    Provider: "github",
  })
  want := []string{
    "graph",
    "--repo", "/tmp/repo",
    "--workflow", ".github/workflows/ci.yml",
    "--provider", "github",
  }
  if !reflect.DeepEqual(got, want) {
    t.Fatalf("got %#v want %#v", got, want)
  }

  got = graphArgs(GraphOptions{Repo: "/r", Workflow: "Jenkinsfile"})
  want = []string{"graph", "--repo", "/r", "--workflow", "Jenkinsfile"}
  if !reflect.DeepEqual(got, want) {
    t.Fatalf("got %#v want %#v", got, want)
  }
}

func TestRunArgs(t *testing.T) {
  got := runArgs(RunOptions{
    Repo:     "/tmp/repo",
    Workflow: "ci.yml",
    Provider: "jenkins",
    Job:      "build",
    Inputs:   []string{"FOO=1", "BAR=2"},
  })
  want := []string{
    "run",
    "--repo", "/tmp/repo",
    "--workflow", "ci.yml",
    "--provider", "jenkins",
    "--job", "build",
    "--input", "FOO=1",
    "--input", "BAR=2",
  }
  if !reflect.DeepEqual(got, want) {
    t.Fatalf("got %#v want %#v", got, want)
  }
}
