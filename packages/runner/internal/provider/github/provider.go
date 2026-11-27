package github

import (
  "context"
  "os"
  "path/filepath"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

const ID = event.ProviderGitHub

type Provider struct{}

func New() *Provider {
  return &Provider{}
}

func Present(repo string) bool {
  workflowsDir := filepath.Join(repo, ".github", "workflows")
  entries, err := os.ReadDir(workflowsDir)
  if err != nil {
    return false
  }
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    ext := strings.ToLower(filepath.Ext(entry.Name()))
    if ext == ".yml" || ext == ".yaml" {
      return true
    }
  }
  return false
}

func (Provider) ID() string {
  return ID
}

func (Provider) List(repo string) (event.WorkflowList, error) {
  result, err := List(repo)
  if err != nil {
    return event.WorkflowList{}, err
  }
  result.Providers = []string{ID}
  for index := range result.Workflows {
    result.Workflows[index].Provider = ID
  }
  return result, nil
}

func (Provider) Graph(repo, workflow string) (event.JobGraph, error) {
  return Graph(repo, workflow)
}

func (Provider) Run(
  _ context.Context,
  repo, workflow, job string,
  secrets []string,
  secretFile string,
  inputs []string,
  parallel bool,
) (int, error) {
  return Start(Options{
    Repo:       repo,
    Workflow:   workflow,
    Job:        job,
    Secrets:    secrets,
    SecretFile: secretFile,
    Inputs:     inputs,
    Parallel:   parallel,
  })
}
