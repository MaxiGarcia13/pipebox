package run

import (
  "os"
  "path/filepath"
  "slices"
  "testing"
)

func TestBuildActArgsDispatchInputs(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "deploy.yml"), []byte(`
on:
  workflow_dispatch:
    inputs:
      tag:
        required: true
`), 0o600); err != nil {
    t.Fatal(err)
  }

  args, err := buildActArgs(repoPath, Options{
    Repo:     repoPath,
    Workflow: ".github/workflows/deploy.yml",
    Inputs:   []string{"tag=v1"},
  })
  if err != nil {
    t.Fatal(err)
  }

  imageIndex := slices.Index(args, "pipebox-act:latest")
  if imageIndex < 0 {
    t.Fatalf("missing image in %#v", args)
  }
  if imageIndex+1 >= len(args) || args[imageIndex+1] != "workflow_dispatch" {
    t.Fatalf("expected workflow_dispatch after image, got %#v", args[imageIndex:])
  }
  if !containsPair(args, "--input", "tag=v1") {
    t.Fatalf("missing --input tag=v1 in %#v", args)
  }
  if !containsPair(args, "-W", ".github/workflows/deploy.yml") {
    t.Fatalf("missing -W in %#v", args)
  }
}

func TestBuildActArgsPushWorkflow(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("on: push\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  args, err := buildActArgs(repoPath, Options{
    Repo:     repoPath,
    Workflow: ".github/workflows/ci.yml",
  })
  if err != nil {
    t.Fatal(err)
  }
  if slices.Contains(args, "workflow_dispatch") {
    t.Fatalf("unexpected workflow_dispatch in %#v", args)
  }
}

func containsPair(args []string, flag, value string) bool {
  for index := 0; index+1 < len(args); index++ {
    if args[index] == flag && args[index+1] == value {
      return true
    }
  }
  return false
}
