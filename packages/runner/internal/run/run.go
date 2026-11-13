package run

import (
  "pipebox/internal/event"
)

type Options struct {
  Repo     string
  Workflow string
  Job      string
}

// Start emits the run lifecycle stubs until act integration lands.
func Start(options Options) error {
  if err := event.EmitStdout(event.RunStarted{
    Type:     event.TypeRunStarted,
    Repo:     options.Repo,
    Workflow: options.Workflow,
    Job:      options.Job,
  }); err != nil {
    return err
  }

  if err := event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: "act integration not wired yet",
  }); err != nil {
    return err
  }

  return event.EmitStdout(event.RunFinished{
    Type:     event.TypeRunFinished,
    Status:   "failed",
    ExitCode: 1,
  })
}
