package main

import (
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestFormatWorkflowList(t *testing.T) {
  t.Parallel()

  got := formatWorkflowList(event.WorkflowList{
    Workflows: []event.Workflow{
      {Provider: "github", Name: "CI", Path: ".github/workflows/ci.yml"},
      {Provider: "jenkins", Name: "Build", Path: "Jenkinsfile"},
    },
  })
  want := strings.Join([]string{
    "PROVIDER  NAME   PATH",
    "github    CI     .github/workflows/ci.yml",
    "jenkins   Build  Jenkinsfile",
  }, "\n")
  if got != want {
    t.Fatalf("got:\n%s\nwant:\n%s", got, want)
  }
}

func TestFormatWorkflowListEmpty(t *testing.T) {
  t.Parallel()

  if got := formatWorkflowList(event.WorkflowList{}); got != "no workflows" {
    t.Fatalf("got %q", got)
  }
}

func TestFormatJobGraph(t *testing.T) {
  t.Parallel()

  got := formatJobGraph(event.JobGraph{
    Jobs: []event.JobNode{
      {ID: "lint", Name: "Lint", Steps: []string{"echo ok"}},
      {ID: "build", Name: "Build", Needs: []string{"lint", "test"}, Steps: []string{"npm test"}},
    },
  })
  want := strings.Join([]string{
    "lint   Lint",
    "  echo ok",
    "build  Build",
    "  needs: lint, test",
    "  npm test",
  }, "\n")
  if got != want {
    t.Fatalf("got:\n%s\nwant:\n%s", got, want)
  }
}

func TestFormatJobGraphEmpty(t *testing.T) {
  t.Parallel()

  if got := formatJobGraph(event.JobGraph{}); got != "no jobs" {
    t.Fatalf("got %q", got)
  }
}
