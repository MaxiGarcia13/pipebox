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

  artifactDir := t.TempDir()
  args, err := buildActArgs(repoPath, Options{
    Repo:         repoPath,
    Workflow:     ".github/workflows/deploy.yml",
    Inputs:       []string{"tag=v1"},
    ArtifactDir:  artifactDir,
    ArtifactPort: 34567,
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
  if !containsPair(args, "--artifact-server-path", artifactMountPath) {
    t.Fatalf("missing artifact-server-path in %#v", args)
  }
  if !containsPair(args, "--artifact-server-addr", artifactBindAddr) {
    t.Fatalf("missing artifact-server-addr in %#v", args)
  }
  if !containsPair(args, "--artifact-server-port", "34567") {
    t.Fatalf("missing artifact-server-port in %#v", args)
  }
  if !containsPair(args, "-v", artifactDir+":"+artifactMountPath) {
    t.Fatalf("missing artifact mount in %#v", args)
  }
  if !containsPair(args, "-p", "34567:34567") {
    t.Fatalf("missing artifact port publish in %#v", args)
  }
  if !containsPair(args, "-e", "ACTIONS_RUNTIME_URL=http://host.docker.internal:34567/") {
    t.Fatalf("missing ACTIONS_RUNTIME_URL in %#v", args)
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
    Repo:         repoPath,
    Workflow:     ".github/workflows/ci.yml",
    ArtifactDir:  t.TempDir(),
    ArtifactPort: 34567,
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
