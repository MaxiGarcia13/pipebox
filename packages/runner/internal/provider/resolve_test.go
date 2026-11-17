package provider

import (
  "path/filepath"
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
