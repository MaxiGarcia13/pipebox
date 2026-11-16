package provider

import (
  "context"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type RunOptions struct {
  Repo       string
  Workflow   string
  Job        string
  Secrets    []string
  SecretFile string
  Inputs     []string
}

type Provider interface {
  ID() string
  List(repo string) (event.WorkflowList, error)
  Graph(repo, workflow string) (event.JobGraph, error)
  Run(ctx context.Context, opts RunOptions) (int, error)
}
