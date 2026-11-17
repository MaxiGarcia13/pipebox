package jenkins

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func isJenkinsfileName(name string) bool {
  return name == "Jenkinsfile" || strings.HasPrefix(name, "Jenkinsfile.")
}

func List(repoPath string) (event.WorkflowList, error) {
  entries, err := os.ReadDir(repoPath)
  if err != nil {
    if os.IsNotExist(err) {
      return event.WorkflowList{
        Type:      event.TypeWorkflowList,
        Workflows: []event.Workflow{},
      }, nil
    }
    return event.WorkflowList{}, fmt.Errorf("read repo: %w", err)
  }

  workflows := make([]event.Workflow, 0)
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    name := entry.Name()
    if !isJenkinsfileName(name) {
      continue
    }
    workflow := event.Workflow{
      ID:       name,
      Name:     name,
      Path:     filepath.ToSlash(name),
      Provider: event.ProviderJenkins,
      Events:   []string{"pipeline"},
      Inputs:   []event.WorkflowInput{},
    }

    data, err := os.ReadFile(filepath.Join(repoPath, name))
    if err != nil {
      return event.WorkflowList{}, fmt.Errorf("read %s: %w", name, err)
    }
    inputs := LoadParams(string(data))
    if inputs == nil {
      inputs = []event.WorkflowInput{}
    }
    workflow.Inputs = inputs
    workflow.Callable = len(inputs) > 0
    workflows = append(workflows, workflow)
  }

  providers := []string{}
  if len(workflows) > 0 {
    providers = []string{event.ProviderJenkins}
  }

  return event.WorkflowList{
    Type:      event.TypeWorkflowList,
    Providers: providers,
    Workflows: workflows,
  }, nil
}
