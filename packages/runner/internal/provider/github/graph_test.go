package github

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestLoadJobGraphNeedsShapes(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  workflowPath := filepath.Join(dir, "ci.yml")
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

  graph, err := LoadJobGraph(dir, workflowPath)
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

  dir := t.TempDir()
  workflowPath := filepath.Join(dir, "empty.yml")
  if err := os.WriteFile(workflowPath, []byte("name: empty\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(dir, workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 0 {
    t.Fatalf("expected no jobs, got %#v", graph.Jobs)
  }
}

func TestGraphResolvesRelativeWorkflow(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(`
jobs:
  lint:
    name: Lint
    steps:
      - run: echo ok
`), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := Graph(repoPath, ".github/workflows/ci.yml")
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 1 || graph.Jobs[0].Name != "Lint" {
    t.Fatalf("unexpected graph: %#v", graph.Jobs)
  }
  if len(graph.Jobs[0].Steps) != 1 || graph.Jobs[0].Steps[0] != "echo ok" {
    t.Fatalf("unexpected steps: %#v", graph.Jobs[0].Steps)
  }
}

func TestLoadJobGraphSteps(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  workflowPath := filepath.Join(dir, "test.yml")
  content := `
jobs:
  test-unit:
    name: Unit Tests & E2E Tests
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Setup Node
        uses: actions/setup-node@v4
      - run: npm ci
      - name: Run Unit Tests
        run: npm run test:unit
`
  if err := os.WriteFile(workflowPath, []byte(content), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(dir, workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 1 {
    t.Fatalf("jobs: got %d", len(graph.Jobs))
  }
  steps := graph.Jobs[0].Steps
  want := []string{"Checkout", "Setup Node", "npm ci", "Run Unit Tests"}
  if len(steps) != len(want) {
    t.Fatalf("steps: got %#v want %#v", steps, want)
  }
  for index := range want {
    if steps[index] != want[index] {
      t.Fatalf("steps: got %#v want %#v", steps, want)
    }
  }
}

func TestLoadJobGraphExpandsLocalReusable(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "lint.yml"), []byte(`
jobs:
  lint:
    name: Lint
    steps:
      - name: Run lint
        run: npm run lint
`), 0o600); err != nil {
    t.Fatal(err)
  }
  callerPath := filepath.Join(workflowsDir, "main.yml")
  if err := os.WriteFile(callerPath, []byte(`
jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  build:
    needs: lint
    steps:
      - run: npm run build
`), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(repoPath, callerPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 2 {
    t.Fatalf("jobs: got %d %#v", len(graph.Jobs), graph.Jobs)
  }
  if graph.Jobs[0].ID != "lint" || graph.Jobs[0].Name != "Lint" {
    t.Fatalf("expanded lint: %#v", graph.Jobs[0])
  }
  if len(graph.Jobs[0].Steps) != 1 || graph.Jobs[0].Steps[0] != "Run lint" {
    t.Fatalf("expanded steps: %#v", graph.Jobs[0].Steps)
  }
  if graph.Jobs[1].ID != "build" {
    t.Fatalf("build: %#v", graph.Jobs[1])
  }
  if len(graph.Jobs[1].Needs) != 1 || graph.Jobs[1].Needs[0] != "lint" {
    t.Fatalf("build needs: %#v", graph.Jobs[1].Needs)
  }
}

func TestLoadJobGraphExpandsNestedReusable(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }

  write := func(name, content string) {
    t.Helper()
    if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte(content), 0o600); err != nil {
      t.Fatal(err)
    }
  }

  write("ci.yml", `
jobs:
  test:
    name: Test
    steps:
      - run: pnpm test
  lint:
    name: Lint
    steps:
      - run: pnpm lint
`)
  write("publish-cli.yml", `
jobs:
  publish-cli:
    name: Publish CLI
    steps:
      - run: echo cli
`)
  write("publish-desktop.yml", `
jobs:
  publish-desktop:
    name: Publish Desktop
    steps:
      - run: echo desktop
`)
  write("publish.yml", `
jobs:
  publish-cli:
    uses: ./.github/workflows/publish-cli.yml
  publish-desktop:
    needs: publish-cli
    uses: ./.github/workflows/publish-desktop.yml
`)
  write("release.yml", `
jobs:
  ci:
    uses: ./.github/workflows/ci.yml
  publish:
    needs: ci
    uses: ./.github/workflows/publish.yml
`)

  graph, err := LoadJobGraph(repoPath, filepath.Join(workflowsDir, "release.yml"))
  if err != nil {
    t.Fatal(err)
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

  assertJob(t, byID, "test", "Test", nil)
  assertJob(t, byID, "lint", "Lint", nil)
  assertJob(t, byID, "publish-cli", "Publish CLI", []string{"test", "lint"})
  assertJob(t, byID, "publish-desktop", "Publish Desktop", []string{"publish-cli"})

  for _, job := range graph.Jobs {
    if job.ID == "ci" || job.ID == "publish" {
      t.Fatalf("caller stub should be replaced, got %#v", job)
    }
  }
}

func TestLoadJobGraphReusableCycle(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "a.yml"), []byte(`
jobs:
  call-b:
    uses: ./.github/workflows/b.yml
`), 0o600); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "b.yml"), []byte(`
jobs:
  call-a:
    uses: ./.github/workflows/a.yml
`), 0o600); err != nil {
    t.Fatal(err)
  }

  _, err := LoadJobGraph(repoPath, filepath.Join(workflowsDir, "a.yml"))
  if err == nil {
    t.Fatal("expected cycle error")
  }
  if !strings.Contains(err.Error(), "cycle") {
    t.Fatalf("expected cycle error, got %v", err)
  }
}

func TestLoadJobGraphRemoteReusableKeepsSyntheticStep(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  workflowPath := filepath.Join(dir, "main.yml")
  remote := "octo-org/octo-repo/.github/workflows/workflow.yml@main"
  if err := os.WriteFile(workflowPath, []byte(`
jobs:
  remote:
    uses: `+remote+`
`), 0o600); err != nil {
    t.Fatal(err)
  }

  graph, err := LoadJobGraph(dir, workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(graph.Jobs) != 1 {
    t.Fatalf("jobs: got %d %#v", len(graph.Jobs), graph.Jobs)
  }
  if graph.Jobs[0].ID != "remote" {
    t.Fatalf("job id: %#v", graph.Jobs[0])
  }
  if len(graph.Jobs[0].Steps) != 1 || graph.Jobs[0].Steps[0] != remote {
    t.Fatalf("steps: %#v", graph.Jobs[0].Steps)
  }
}

func TestLoadJobGraphReusableIDCollision(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "shared.yml"), []byte(`
jobs:
  build:
    steps:
      - run: echo build
`), 0o600); err != nil {
    t.Fatal(err)
  }
  callerPath := filepath.Join(workflowsDir, "main.yml")
  if err := os.WriteFile(callerPath, []byte(`
jobs:
  first:
    uses: ./.github/workflows/shared.yml
  second:
    uses: ./.github/workflows/shared.yml
`), 0o600); err != nil {
    t.Fatal(err)
  }

  _, err := LoadJobGraph(repoPath, callerPath)
  if err == nil {
    t.Fatal("expected duplicate id error")
  }
  if !strings.Contains(err.Error(), "duplicate job id") {
    t.Fatalf("expected duplicate id error, got %v", err)
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
