package e2e_test

import (
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

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

func TestGraphUsesWorkingDirectory(t *testing.T) {
  stdout, stderr, code := runPipeboxDir(
    t,
    repoPath(t, "github-only"),
    "graph",
    "--workflow",
    ".github/workflows/ci.yml",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  graph := decodeLine[event.JobGraph](t, stdout)
  if graph.Type != event.TypeJobGraph {
    t.Fatalf("type: got %q", graph.Type)
  }
  if len(graph.Jobs) != 1 || graph.Jobs[0].ID != "lint" {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
}
