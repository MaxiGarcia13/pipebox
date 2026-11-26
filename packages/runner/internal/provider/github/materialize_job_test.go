package github

import (
  "os"
  "path/filepath"
  "strings"
  "testing"

  "go.yaml.in/yaml/v4"
)

func TestIsolateJobStripsNeedsAndKeepsOnlyTarget(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  workflowPath := filepath.Join(workflowsDir, "ci.yml")
  if err := os.WriteFile(workflowPath, []byte(`
on: push
jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - run: echo lint
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - run: echo test
  e2e:
    name: E2E
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - run: echo e2e
`), 0o600); err != nil {
    t.Fatal(err)
  }

  isolated, err := IsolateJob(workflowPath, "e2e")
  if err != nil {
    t.Fatal(err)
  }
  defer isolated.Cleanup()
  if isolated.JobID != "e2e" {
    t.Fatalf("JobID: %q", isolated.JobID)
  }
  if isolated.Path == workflowPath {
    t.Fatal("expected a temp isolated workflow")
  }

  data, err := os.ReadFile(isolated.Path)
  if err != nil {
    t.Fatal(err)
  }
  body := string(data)
  if strings.Contains(body, "needs:") {
    t.Fatalf("needs should be stripped: %s", body)
  }
  if strings.Contains(body, "echo lint") || strings.Contains(body, "echo test") {
    t.Fatalf("sibling jobs should be removed: %s", body)
  }
  if !strings.Contains(body, "echo e2e") {
    t.Fatalf("missing e2e steps: %s", body)
  }

  var document yaml.Node
  if err := yaml.Unmarshal(data, &document); err != nil {
    t.Fatal(err)
  }
  jobs := mapValue(document.Content[0], "jobs")
  if jobs == nil || jobs.Kind != yaml.MappingNode {
    t.Fatalf("jobs: %#v", jobs)
  }
  if len(jobs.Content) != 2 {
    t.Fatalf("expected one job, got %#v", jobs.Content)
  }
  if jobs.Content[0].Value != "e2e" {
    t.Fatalf("job id: %q", jobs.Content[0].Value)
  }
}

func TestIsolateJobMatchesDisplayName(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  workflowPath := filepath.Join(workflowsDir, "ci.yml")
  if err := os.WriteFile(workflowPath, []byte(`
on: push
jobs:
  e2e:
    name: End to end
    needs: lint
    runs-on: ubuntu-latest
    steps:
      - run: echo e2e
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: echo lint
`), 0o600); err != nil {
    t.Fatal(err)
  }

  isolated, err := IsolateJob(workflowPath, "End to end")
  if err != nil {
    t.Fatal(err)
  }
  defer isolated.Cleanup()
  if isolated.JobID != "e2e" {
    t.Fatalf("JobID: %q", isolated.JobID)
  }
}

func TestIsolateJobMissing(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowPath := filepath.Join(repoPath, "ci.yml")
  if err := os.WriteFile(workflowPath, []byte(`
on: push
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: echo lint
`), 0o600); err != nil {
    t.Fatal(err)
  }

  _, err := IsolateJob(workflowPath, "missing")
  if err == nil {
    t.Fatal("expected error")
  }
}
