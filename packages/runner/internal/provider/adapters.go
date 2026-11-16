package provider

import (
  "context"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins"
)

type githubAdapter struct {
  inner *github.Provider
}

func (adapter githubAdapter) ID() string {
  return adapter.inner.ID()
}

func (adapter githubAdapter) List(repo string) (event.WorkflowList, error) {
  return adapter.inner.List(repo)
}

func (adapter githubAdapter) Graph(repo, workflow string) (event.JobGraph, error) {
  return adapter.inner.Graph(repo, workflow)
}

func (adapter githubAdapter) Run(ctx context.Context, opts RunOptions) (int, error) {
  return adapter.inner.Run(
    ctx,
    opts.Repo,
    opts.Workflow,
    opts.Job,
    opts.Secrets,
    opts.SecretFile,
    opts.Inputs,
  )
}

type jenkinsAdapter struct {
  inner *jenkins.Provider
}

func (adapter jenkinsAdapter) ID() string {
  return adapter.inner.ID()
}

func (adapter jenkinsAdapter) List(repo string) (event.WorkflowList, error) {
  return adapter.inner.List(repo)
}

func (adapter jenkinsAdapter) Graph(repo, workflow string) (event.JobGraph, error) {
  return adapter.inner.Graph(repo, workflow)
}

func (adapter jenkinsAdapter) Run(ctx context.Context, opts RunOptions) (int, error) {
  return adapter.inner.Run(
    ctx,
    opts.Repo,
    opts.Workflow,
    opts.Job,
    opts.Secrets,
    opts.SecretFile,
    opts.Inputs,
  )
}
