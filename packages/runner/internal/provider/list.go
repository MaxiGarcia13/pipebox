package provider

import (
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func List(repo string) (event.WorkflowList, error) {
  providers := Discover(repo)
  providerIDs := make([]string, 0, len(providers))
  workflows := make([]event.Workflow, 0)

  for _, item := range providers {
    providerIDs = append(providerIDs, item.ID())
    result, err := item.List(repo)
    if err != nil {
      return event.WorkflowList{}, err
    }
    workflows = append(workflows, result.Workflows...)
  }

  return event.WorkflowList{
    Type:      event.TypeWorkflowList,
    Providers: providerIDs,
    Workflows: workflows,
  }, nil
}
