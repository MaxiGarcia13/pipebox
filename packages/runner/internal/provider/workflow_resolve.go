package provider

import (
  "fmt"
  "path/filepath"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func ResolveWorkflow(repo, ref, providerHint string) (string, error) {
  if strings.TrimSpace(ref) == "" {
    return "", fmt.Errorf("workflow is required")
  }

  list, err := List(repo)
  if err != nil {
    return "", err
  }

  slashRef := normalizeWorkflowRef(repo, ref)

  var pathMatches []event.Workflow
  var idMatches []event.Workflow
  for _, workflow := range list.Workflows {
    if providerHint != "" && workflow.Provider != providerHint {
      continue
    }
    path := filepath.ToSlash(workflow.Path)
    if path == slashRef {
      pathMatches = append(pathMatches, workflow)
      continue
    }
    if workflow.ID == ref || workflow.ID == slashRef {
      idMatches = append(idMatches, workflow)
    }
  }

  if len(pathMatches) == 1 {
    return filepath.ToSlash(pathMatches[0].Path), nil
  }
  if len(pathMatches) > 1 {
    return "", fmt.Errorf("ambiguous workflow path %q", slashRef)
  }
  if len(idMatches) == 1 {
    return filepath.ToSlash(idMatches[0].Path), nil
  }
  if len(idMatches) > 1 {
    return "", fmt.Errorf("ambiguous workflow id %q; pass a path or --provider", ref)
  }
  return "", fmt.Errorf("workflow %q not found in %s", ref, repo)
}

func normalizeWorkflowRef(repo, ref string) string {
  if !filepath.IsAbs(ref) {
    return filepath.ToSlash(ref)
  }
  rel, err := filepath.Rel(repo, ref)
  if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
    return filepath.ToSlash(ref)
  }
  return filepath.ToSlash(rel)
}
