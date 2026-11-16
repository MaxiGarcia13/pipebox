package jenkins

import (
  "context"
  "fmt"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

const ID = event.ProviderJenkins

type Provider struct{}

func New() *Provider {
  return &Provider{}
}

func Present(_ string) bool {
  return false
}

func (Provider) ID() string {
  return ID
}

func (Provider) List(_ string) (event.WorkflowList, error) {
  return event.WorkflowList{
    Type:      event.TypeWorkflowList,
    Providers: nil,
    Workflows: []event.Workflow{},
  }, nil
}

func (Provider) Graph(_, _ string) (event.JobGraph, error) {
  return event.JobGraph{}, fmt.Errorf("jenkins provider not implemented")
}

func (Provider) Run(_ context.Context, _, _, _ string, _ []string, _ string, _ []string) (int, error) {
  return 1, fmt.Errorf("jenkins provider not implemented")
}
