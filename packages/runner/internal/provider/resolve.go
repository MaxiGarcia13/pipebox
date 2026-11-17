package provider

import (
  "fmt"
  "path/filepath"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins"
)

func Resolve(repo, workflow, providerHint string) (Provider, error) {
  providers := Discover(repo)
  byID := make(map[string]Provider, len(providers))
  for _, item := range providers {
    byID[item.ID()] = item
  }

  if providerHint != "" {
    item, ok := byID[providerHint]
    if !ok {
      return nil, fmt.Errorf("provider %q not available in %s", providerHint, repo)
    }
    return item, nil
  }

  slashPath := filepath.ToSlash(workflow)
  if strings.HasPrefix(slashPath, ".github/") {
    item, ok := byID[event.ProviderGitHub]
    if !ok {
      return nil, fmt.Errorf("github provider not available in %s", repo)
    }
    return item, nil
  }

  if jenkins.IsJenkinsfilePath(workflow) {
    item, ok := byID[event.ProviderJenkins]
    if !ok {
      return nil, fmt.Errorf("jenkins provider not available in %s", repo)
    }
    return item, nil
  }

  if len(providers) == 1 {
    return providers[0], nil
  }

  return nil, fmt.Errorf("cannot resolve provider for workflow %q", workflow)
}
