package workflows

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"

  "pipebox/internal/event"
)

func List(repoPath string) (event.WorkflowList, error) {
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  entries, err := os.ReadDir(workflowsDir)
  if err != nil {
    if os.IsNotExist(err) {
      return event.WorkflowList{
        Type:      event.TypeWorkflowList,
        Workflows: []event.Workflow{},
      }, nil
    }
    return event.WorkflowList{}, fmt.Errorf("read workflows: %w", err)
  }

  workflows := make([]event.Workflow, 0)
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    name := entry.Name()
    ext := strings.ToLower(filepath.Ext(name))
    if ext != ".yml" && ext != ".yaml" {
      continue
    }
    id := strings.TrimSuffix(name, filepath.Ext(name))
    workflows = append(workflows, event.Workflow{
      ID:   id,
      Name: id,
      Path: filepath.Join(".github", "workflows", name),
    })
  }

  return event.WorkflowList{
    Type:      event.TypeWorkflowList,
    Workflows: workflows,
  }, nil
}
