package provider

import (
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins"
)

func Discover(repo string) []Provider {
  providers := make([]Provider, 0, 2)
  if github.Present(repo) {
    providers = append(providers, githubAdapter{inner: github.New()})
  }
  if jenkins.Present(repo) {
    providers = append(providers, jenkinsAdapter{inner: jenkins.New()})
  }
  return providers
}
