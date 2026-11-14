package workflows

import (
  "os"
  "path/filepath"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestLoadDispatchInputs(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "deploy.yml")
  content := `
name: Deploy
on:
  workflow_dispatch:
    inputs:
      environment:
        description: Target env
        required: true
        type: choice
        options:
          - staging
          - production
        default: staging
      dry_run:
        description: Dry run
        type: boolean
        default: false
      count:
        type: number
        default: 2
      note:
        type: string
`
  if err := os.WriteFile(workflowPath, []byte(content), 0o600); err != nil {
    t.Fatal(err)
  }

  meta, err := LoadDispatch(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if meta.Name != "Deploy" {
    t.Fatalf("name: got %q", meta.Name)
  }
  if !meta.Callable {
    t.Fatal("expected callable")
  }
  if len(meta.Inputs) != 4 {
    t.Fatalf("inputs: got %#v", meta.Inputs)
  }

  env := meta.Inputs[0]
  if env.ID != "environment" || env.Type != "choice" || !env.Required || env.Default != "staging" {
    t.Fatalf("environment: %#v", env)
  }
  if len(env.Options) != 2 || env.Options[0] != "staging" || env.Options[1] != "production" {
    t.Fatalf("environment options: %#v", env.Options)
  }
  if meta.Inputs[1].Type != "boolean" || meta.Inputs[1].Default != "false" {
    t.Fatalf("dry_run: %#v", meta.Inputs[1])
  }
  if meta.Inputs[2].Type != "number" || meta.Inputs[2].Default != "2" {
    t.Fatalf("count: %#v", meta.Inputs[2])
  }
  if meta.Inputs[3].Type != "string" {
    t.Fatalf("note: %#v", meta.Inputs[3])
  }
}

func TestLoadDispatchScalarEvent(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "manual.yml")
  if err := os.WriteFile(workflowPath, []byte("on: workflow_dispatch\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  meta, err := LoadDispatch(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if !meta.Callable {
    t.Fatal("expected callable")
  }
  if len(meta.Inputs) != 0 {
    t.Fatalf("inputs: got %#v", meta.Inputs)
  }
}

func TestLoadDispatchEnvironmentTypeAsString(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "env.yml")
  content := `
on:
  workflow_dispatch:
    inputs:
      target:
        type: environment
`
  if err := os.WriteFile(workflowPath, []byte(content), 0o600); err != nil {
    t.Fatal(err)
  }

  meta, err := LoadDispatch(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(meta.Inputs) != 1 || meta.Inputs[0].Type != "string" {
    t.Fatalf("inputs: %#v", meta.Inputs)
  }
}

func TestLoadDispatchNotCallable(t *testing.T) {
  t.Parallel()

  workflowPath := filepath.Join(t.TempDir(), "push.yml")
  if err := os.WriteFile(workflowPath, []byte("on: push\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  meta, err := LoadDispatch(workflowPath)
  if err != nil {
    t.Fatal(err)
  }
  if meta.Callable {
    t.Fatal("expected not callable")
  }
  if len(meta.Events) != 1 || meta.Events[0] != "push" {
    t.Fatalf("events: %#v", meta.Events)
  }
}

func TestResolveActEvent(t *testing.T) {
  t.Parallel()

  if got := ResolveActEvent(nil); got != "push" {
    t.Fatalf("empty: %q", got)
  }
  if got := ResolveActEvent([]string{"pull_request", "workflow_call"}); got != "pull_request" {
    t.Fatalf("pr+call: %q", got)
  }
  if got := ResolveActEvent([]string{"workflow_call"}); got != "workflow_call" {
    t.Fatalf("call: %q", got)
  }
  if got := ResolveActEvent([]string{"push", "pull_request", "workflow_dispatch"}); got != "workflow_dispatch" {
    t.Fatalf("dispatch preferred: %q", got)
  }
  if got := ResolveActEvent([]string{"schedule"}); got != "schedule" {
    t.Fatalf("fallback: %q", got)
  }
}

func TestListEnrichesDispatch(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  workflowsDir := filepath.Join(repoPath, ".github", "workflows")
  if err := os.MkdirAll(workflowsDir, 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "deploy.yml"), []byte(`
name: Deploy App
on:
  workflow_dispatch:
    inputs:
      tag:
        required: true
`), 0o600); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("on: push\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  list, err := List(repoPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(list.Workflows) != 2 {
    t.Fatalf("workflows: got %#v", list.Workflows)
  }

  byID := map[string]event.Workflow{}
  for _, workflow := range list.Workflows {
    byID[workflow.ID] = workflow
  }

  deploy := byID["deploy"]
  if deploy.Name != "Deploy App" || !deploy.Callable || len(deploy.Inputs) != 1 {
    t.Fatalf("deploy: %#v", deploy)
  }
  ci := byID["ci"]
  if ci.Name != "ci" || ci.Callable || len(ci.Inputs) != 0 {
    t.Fatalf("ci: %#v", ci)
  }
  if len(ci.Events) != 1 || ci.Events[0] != "push" {
    t.Fatalf("ci events: %#v", ci.Events)
  }
}
