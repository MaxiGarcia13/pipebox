package jenkins

import (
  "context"
  "fmt"
  "os"
  "path/filepath"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

const ID = event.ProviderJenkins

type Provider struct{}

func New() *Provider {
  return &Provider{}
}

func Present(repo string) bool {
  entries, err := os.ReadDir(repo)
  if err != nil {
    return false
  }
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    if isJenkinsfileName(entry.Name()) {
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

func (Provider) Graph(_, _ string) (event.JobGraph, error) {
  return event.JobGraph{}, fmt.Errorf("jenkins provider not implemented")
}

func (Provider) Run(_ context.Context, _, _, _ string, _ []string, _ string, _ []string) (int, error) {
  return 1, fmt.Errorf("jenkins provider not implemented")
}

func IsJenkinsfilePath(workflow string) bool {
  base := filepath.Base(filepath.ToSlash(workflow))
  return isJenkinsfileName(base)
}
