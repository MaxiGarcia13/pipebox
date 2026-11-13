package workflows

import (
  "os"
  "path/filepath"
  "testing"
)

func TestLoadJobGraphNeedsShapes(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "ci.yml")
  content := `
jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
  bump:
    needs: lint
    runs-on: ubuntu-latest
  release-notes:
    name: Release notes
    needs: [bump]
    runs-on: ubuntu-latest
  deploy:
    needs:
      - bump
    runs-on: ubuntu-latest
`
  if err := os.WriteFile(workflowPath, []byte(content), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if graph.Type != "job_graph" {
    t.Fatalf("type: got %q", graph.Type)
  }
  if len(graph.Jobs) != 4 {
    t.Fatalf("jobs: got %d %#v", len(graph.Jobs), graph.Jobs)
  }

  byID := map[string]struct {
    name  string
    needs []string
  }{}
  for _, job := range graph.Jobs {
    byID[job.ID] = struct {
      name  string
      needs []string
    }{name: job.Name, needs: job.Needs}
  }

  assertJob(t, byID, "lint", "Lint", nil)
  assertJob(t, byID, "bump", "bump", []string{"lint"})
  assertJob(t, byID, "release-notes", "Release notes", []string{"bump"})
  assertJob(t, byID, "deploy", "deploy", []string{"bump"})

  if graph.Jobs[0].ID != "lint" || graph.Jobs[1].ID != "bump" {
    t.Fatalf("expected YAML order preserved, got %#v", graph.Jobs)
  }
}

func TestLoadJobGraphEmptyJobs(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "empty.yml")
  if err := os.WriteFile(workflowPath, []byte("name: empty\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 0 {
    t.Fatalf("expected no jobs, got %#v", graph.Jobs)
  }
}

func assertJob(
  t *testing.T,
  byID map[string]struct {
    name  string
    needs []string
  },
  id string,
  name string,
  needs []string,
) {
  t.Helper()
  job, ok := byID[id]
  if !ok {
    t.Fatalf("missing job %q", id)
  }
  if job.name != name {
    t.Fatalf("job %q name: got %q want %q", id, job.name, name)
  }
  if needs == nil {
    needs = []string{}
  }
  if len(job.needs) != len(needs) {
    t.Fatalf("job %q needs: got %#v want %#v", id, job.needs, needs)
  }
  for index := range needs {
    if job.needs[index] != needs[index] {
      t.Fatalf("job %q needs: got %#v want %#v", id, job.needs, needs)
    }
  }
}
