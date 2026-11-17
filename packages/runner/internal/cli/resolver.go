package cli

import (
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
)

type ProviderResolver struct{}

func (ProviderResolver) Workflows(repo string) []string {
  if repo == "" {
    return nil
  }
  list, err := provider.List(repo)
  if err != nil {
    return nil
  }
  paths := make([]string, 0, len(list.Workflows))
  for _, workflow := range list.Workflows {
    if workflow.Path != "" {
      paths = append(paths, workflow.Path)
    }
  }
  return uniqueSorted(paths)
}

func (ProviderResolver) Jobs(repo, workflow, providerHint string) []string {
  if repo == "" || workflow == "" {
    return nil
  }
  resolved, err := provider.Resolve(repo, workflow, providerHint)
  if err != nil {
    return nil
  }
  graph, err := resolved.Graph(repo, workflow)
  if err != nil {
    return nil
  }
  ids := make([]string, 0, len(graph.Jobs))
  for _, job := range graph.Jobs {
    if job.ID != "" {
      ids = append(ids, job.ID)
    }
  }
  return uniqueSorted(ids)
}

func (ProviderResolver) Providers(repo string) []string {
  if repo == "" {
    return nil
  }
  list, err := provider.List(repo)
  if err != nil {
    return nil
  }
  return uniqueSorted(list.Providers)
}
