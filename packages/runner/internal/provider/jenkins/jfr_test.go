package jenkins

import (
  "os"
  "strings"
  "testing"
)

func TestBuildJFRArgsBasic(t *testing.T) {
  repo := t.TempDir()
  if err := writeFile(repo, "Jenkinsfile", "pipeline { agent any stages { stage('Build') { steps { echo 'x' } } } }"); err != nil {
    t.Fatal(err)
  }
  if err := writeFile(repo, ".env", "TOKEN=secret\n"); err != nil {
    t.Fatal(err)
  }

  runHome := t.TempDir()
  args, cleanup, err := buildJFRArgs(repo, Options{
    Repo:     repo,
    Workflow: "Jenkinsfile",
    Inputs:   []string{"BRANCH=main"},
    Secrets:  []string{"EXTRA=1"},
    RunHome:  runHome,
  })
  if cleanup != nil {
    defer cleanup()
  }
  if err != nil {
    t.Fatalf("buildJFRArgs: %v", err)
  }

  joined := strings.Join(args, " ")
  if !strings.Contains(joined, "pipebox-jenkins:latest") {
    t.Fatalf("missing image: %s", joined)
  }
  if !strings.Contains(joined, "--runHome") || !strings.Contains(joined, "--runWorkspace") {
    t.Fatalf("missing run mounts: %s", joined)
  }
  if !strings.Contains(joined, "-f") || !strings.Contains(joined, "/workspace/Jenkinsfile") {
    t.Fatalf("missing jenkinsfile: %s", joined)
  }
  if !strings.Contains(joined, "-a BRANCH=main") {
    t.Fatalf("missing input: %s", joined)
  }
  if !strings.Contains(joined, "-e TOKEN=secret") || !strings.Contains(joined, "-e EXTRA=1") {
    t.Fatalf("missing secrets: %s", joined)
  }
}

func TestBuildJFRArgsMaterializesJob(t *testing.T) {
  repo := t.TempDir()
  source := `
pipeline {
  agent any
  stages {
    stage('Build') { steps { echo 'build' } }
    stage('Test') { steps { echo 'test' } }
  }
}
`
  if err := writeFile(repo, "Jenkinsfile", source); err != nil {
    t.Fatal(err)
  }

  args, cleanup, err := buildJFRArgs(repo, Options{
    Repo:     repo,
    Workflow: "Jenkinsfile",
    Job:      "test",
    RunHome:  t.TempDir(),
  })
  if cleanup != nil {
    defer cleanup()
  }
  if err != nil {
    t.Fatalf("buildJFRArgs: %v", err)
  }

  joined := strings.Join(args, " ")
  if !strings.Contains(joined, materializedFileMount) {
    t.Fatalf("expected materialized mount: %s", joined)
  }

  // Find host path from -v host:mount
  hostPath := ""
  for index := 0; index < len(args)-1; index++ {
    if args[index] != "-v" {
      continue
    }
    parts := strings.SplitN(args[index+1], ":", 3)
    if len(parts) >= 2 && parts[1] == materializedFileMount {
      hostPath = parts[0]
      break
    }
  }
  if hostPath == "" {
    t.Fatalf("materialized host path not found in %v", args)
  }
  content, err := os.ReadFile(hostPath)
  if err != nil {
    t.Fatal(err)
  }
  text := string(content)
  if !strings.Contains(text, "stage('Test')") || strings.Contains(text, "stage('Build')") {
    t.Fatalf("filtered content: %s", text)
  }
}
