package provider

import (
  "os"
  "path/filepath"
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestResolveJenkinsfile(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "jenkins-only")
  resolved, err := Resolve(repo, "Jenkinsfile", "")
  if err != nil {
    t.Fatalf("Resolve: %v", err)
  }
  if resolved.ID() != event.ProviderJenkins {
    t.Fatalf("id: %q", resolved.ID())
  }
}

func TestResolveMixedByPath(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "mixed")

  githubProvider, err := Resolve(repo, ".github/workflows/ci.yml", "")
  if err != nil {
    t.Fatalf("github Resolve: %v", err)
  }
  if githubProvider.ID() != event.ProviderGitHub {
    t.Fatalf("github id: %q", githubProvider.ID())
  }

  jenkinsProvider, err := Resolve(repo, "Jenkinsfile", "")
  if err != nil {
    t.Fatalf("jenkins Resolve: %v", err)
  }
  if jenkinsProvider.ID() != event.ProviderJenkins {
    t.Fatalf("jenkins id: %q", jenkinsProvider.ID())
  }
}

func TestResolveProviderHint(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "mixed")
  resolved, err := Resolve(repo, "Jenkinsfile", event.ProviderJenkins)
  if err != nil {
    t.Fatalf("Resolve: %v", err)
  }
  if resolved.ID() != event.ProviderJenkins {
    t.Fatalf("id: %q", resolved.ID())
  }
}

func TestResolveWorkflowGitHubID(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "github-only")
  path, err := ResolveWorkflow(repo, "ci", "")
  if err != nil {
    t.Fatalf("ResolveWorkflow: %v", err)
  }
  if path != ".github/workflows/ci.yml" {
    t.Fatalf("path: %q", path)
  }
}

func TestResolveWorkflowPathPassthrough(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "github-only")
  path, err := ResolveWorkflow(repo, ".github/workflows/ci.yml", "")
  if err != nil {
    t.Fatalf("ResolveWorkflow: %v", err)
  }
  if path != ".github/workflows/ci.yml" {
    t.Fatalf("path: %q", path)
  }
}

func TestResolveWorkflowJenkinsfile(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "jenkins-only")
  path, err := ResolveWorkflow(repo, "Jenkinsfile", "")
  if err != nil {
    t.Fatalf("ResolveWorkflow: %v", err)
  }
  if path != "Jenkinsfile" {
    t.Fatalf("path: %q", path)
  }
}

func TestResolveWorkflowAmbiguousID(t *testing.T) {
  repo := t.TempDir()
  workflowsDir := filepath.Join(repo, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  for _, name := range []string{"ci.yml", "ci.yaml"} {
    if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte("on: push\n"), 0o600); err != nil {
      t.Fatal(err)
    }
  }

  _, err := ResolveWorkflow(repo, "ci", "")
  if err == nil {
    t.Fatal("expected ambiguous id error")
  }
  if !strings.Contains(err.Error(), "ambiguous") {
    t.Fatalf("error: %v", err)
  }

  path, err := ResolveWorkflow(repo, ".github/workflows/ci.yml", "")
  if err != nil {
    t.Fatalf("path resolve: %v", err)
  }
  if path != ".github/workflows/ci.yml" {
    t.Fatalf("path: %q", path)
  }
}

func TestResolveWorkflowMissing(t *testing.T) {
  repo := filepath.Join("..", "..", "testdata", "repos", "github-only")
  _, err := ResolveWorkflow(repo, "missing", "")
  if err == nil {
    t.Fatal("expected missing workflow error")
  }
}
