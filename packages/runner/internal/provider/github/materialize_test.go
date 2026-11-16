package github

import (
  "os"
  "path/filepath"
  "strings"
  "testing"

  "go.yaml.in/yaml/v4"
)

func TestMaterializeLocalReusablesInlinesCallee(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "lint.yml"), []byte(`
on:
  workflow_call:
jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - run: pnpm lint
`), 0o600); err != nil {
    t.Fatal(err)
  }
  caller := filepath.Join(workflowsDir, "ci.yml")
  if err := os.WriteFile(caller, []byte(`
on:
  pull_request:
jobs:
  lint:
    uses: ./.github/workflows/lint.yml
`), 0o600); err != nil {
    t.Fatal(err)
  }

  materialized, err := MaterializeLocalReusables(repoPath, caller)
  if err != nil {
    t.Fatal(err)
  }
  defer materialized.Cleanup()
  if materialized.Path == caller {
    t.Fatal("expected a temp materialized workflow")
  }

  data, err := os.ReadFile(materialized.Path)
  if err != nil {
    t.Fatal(err)
  }
  body := string(data)
  if strings.Contains(body, "uses: ./.github/workflows/lint.yml") {
    t.Fatalf("caller uses should be expanded: %s", body)
  }
  if !strings.Contains(body, "pnpm lint") {
    t.Fatalf("missing callee steps: %s", body)
  }
  if !strings.Contains(body, "pull_request") {
    t.Fatalf("expected parent on: preserved: %s", body)
  }

  var root yaml.Node
  if err := yaml.Unmarshal(data, &root); err != nil {
    t.Fatal(err)
  }
  jobs := mapValue(root.Content[0], "jobs")
  if jobs == nil || jobs.Kind != yaml.MappingNode || len(jobs.Content) < 2 {
    t.Fatalf("jobs: %#v", jobs)
  }
  if jobs.Content[0].Value != "lint" {
    t.Fatalf("job id: %q", jobs.Content[0].Value)
  }
  if scalarMapValue(jobs.Content[1], "uses") != "" {
    t.Fatalf("expanded job still has uses")
  }
}

func TestMaterializeLocalReusablesNoopWithoutUses(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowPath := filepath.Join(repoPath, "ci.yml")
  if err := os.WriteFile(workflowPath, []byte(`
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
`), 0o600); err != nil {
    t.Fatal(err)
  }

  materialized, err := MaterializeLocalReusables(repoPath, workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  defer materialized.Cleanup()
  if materialized.Path != workflowPath {
    t.Fatalf("expected original path, got %s", materialized.Path)
  }
}
