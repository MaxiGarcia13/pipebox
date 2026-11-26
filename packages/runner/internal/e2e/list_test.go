package e2e_test

import (
  "path/filepath"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestListGitHubOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "github-only"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Providers) != 1 || list.Providers[0] != event.ProviderGitHub {
    t.Fatalf("providers: %#v", list.Providers)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.ID != "ci" || workflow.Name != "CI" {
    t.Fatalf("workflow identity: %#v", workflow)
  }
  wantPath := filepath.ToSlash(filepath.Join(".github", "workflows", "ci.yml"))
  if filepath.ToSlash(workflow.Path) != wantPath {
    t.Fatalf("path: got %q want %q", workflow.Path, wantPath)
  }
  if workflow.Provider != event.ProviderGitHub {
    t.Fatalf("provider: got %q", workflow.Provider)
  }
  if !workflow.Callable {
    t.Fatal("expected callable")
  }
  if len(workflow.Inputs) < 2 {
    t.Fatalf("inputs: %#v", workflow.Inputs)
  }
}

func TestListEmpty(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "empty"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Workflows) != 0 {
    t.Fatalf("expected no workflows, got %#v", list.Workflows)
  }
}

func TestListJenkinsOnly(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "jenkins-only"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if workflow.Path != "Jenkinsfile" || workflow.Provider != event.ProviderJenkins {
    t.Fatalf("workflow: %#v", workflow)
  }
  if len(list.Providers) != 1 || list.Providers[0] != event.ProviderJenkins {
    t.Fatalf("providers: %#v", list.Providers)
  }
}

func TestListMixedGitHubAndJenkins(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "list", "--repo", repoPath(t, "mixed"))
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if len(list.Workflows) != 2 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }

  byProvider := map[string]event.Workflow{}
  for _, workflow := range list.Workflows {
    byProvider[workflow.Provider] = workflow
  }
  githubWorkflow, ok := byProvider[event.ProviderGitHub]
  if !ok || githubWorkflow.Name != "Mixed CI" {
    t.Fatalf("github workflow: %#v", list.Workflows)
  }
  jenkinsWorkflow, ok := byProvider[event.ProviderJenkins]
  if !ok || jenkinsWorkflow.Path != "Jenkinsfile" {
    t.Fatalf("jenkins workflow: %#v", list.Workflows)
  }
  if len(list.Providers) != 2 {
    t.Fatalf("providers: %#v", list.Providers)
  }
}

func TestListUsesWorkingDirectory(t *testing.T) {
  stdout, stderr, code := runPipeboxDir(t, repoPath(t, "github-only"), "list")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  list := decodeLine[event.WorkflowList](t, stdout)
  if list.Type != event.TypeWorkflowList {
    t.Fatalf("type: got %q", list.Type)
  }
  if len(list.Workflows) != 1 || list.Workflows[0].Name != "CI" {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
}
