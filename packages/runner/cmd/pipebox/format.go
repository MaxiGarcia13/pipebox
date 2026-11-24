package main

import (
  "fmt"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func formatWorkflowList(list event.WorkflowList) string {
  if len(list.Workflows) == 0 {
    return "no workflows"
  }

  providerWidth := len("PROVIDER")
  nameWidth := len("NAME")
  for _, workflow := range list.Workflows {
    if len(workflow.Provider) > providerWidth {
      providerWidth = len(workflow.Provider)
    }
    if len(workflow.Name) > nameWidth {
      nameWidth = len(workflow.Name)
    }
  }

  var builder strings.Builder
  fmt.Fprintf(&builder, "%-*s  %-*s  %s", providerWidth, "PROVIDER", nameWidth, "NAME", "PATH")
  for _, workflow := range list.Workflows {
    fmt.Fprintf(
      &builder,
      "\n%-*s  %-*s  %s",
      providerWidth,
      workflow.Provider,
      nameWidth,
      workflow.Name,
      workflow.Path,
    )
  }
  return builder.String()
}

func formatJobGraph(graph event.JobGraph) string {
  if len(graph.Jobs) == 0 {
    return "no jobs"
  }

  idWidth := 0
  for _, job := range graph.Jobs {
    if len(job.ID) > idWidth {
      idWidth = len(job.ID)
    }
  }

  var builder strings.Builder
  for index, job := range graph.Jobs {
    if index > 0 {
      builder.WriteByte('\n')
    }
    fmt.Fprintf(&builder, "%-*s  %s", idWidth, job.ID, job.Name)
    if len(job.Needs) > 0 {
      fmt.Fprintf(&builder, "\n  needs: %s", strings.Join(job.Needs, ", "))
    }
    for _, step := range job.Steps {
      fmt.Fprintf(&builder, "\n  %s", step)
    }
  }
  return builder.String()
}
