package github

import (
  "context"
  "fmt"
  "os"
  "os/signal"
  "path/filepath"
  "strings"
  "syscall"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/envfile"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actparse"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/runstream"
)

type Options struct {
  Repo         string
  Workflow     string
  Job          string
  Secrets      []string
  SecretFile   string
  Inputs       []string
  ArtifactDir  string
  ArtifactPort int
}

func Start(options Options) (int, error) {
  if err := event.EmitStdout(event.RunStarted{
    Type:     event.TypeRunStarted,
    Repo:     options.Repo,
    Workflow: options.Workflow,
    Job:      options.Job,
  }); err != nil {
    return 1, err
  }

  _ = event.EmitStdout(loadJobGraph(options))

  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()

  if err := emitSecretSources(options); err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: err.Error(),
    })
  }

  artifacts, err := prepareArtifactRuntime()
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: err.Error(),
    })
    return 1, nil
  }
  options.ArtifactDir = artifacts.Dir
  options.ArtifactPort = artifacts.Port

  if !actimage.Exists() {
    _ = event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "building " + actimage.Image + " (first run)",
    })
  }

  exitCode, status := runAct(ctx, options)
  if err := emitArtifacts(artifacts.Dir); err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("list artifacts: %v", err),
    })
  }

  if err := event.EmitStdout(event.RunFinished{
    Type:     event.TypeRunFinished,
    Status:   status,
    ExitCode: exitCode,
  }); err != nil {
    return exitCode, err
  }
  return exitCode, nil
}

func runAct(ctx context.Context, options Options) (exitCode int, status string) {
  cmd, cleanup, err := buildActCommand(ctx, options)
  if cleanup != nil {
    defer cleanup()
  }
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: err.Error(),
    })
    return 1, "failed"
  }

  // buildActCommand already ensured the act image.
  doctor.Emit(ID)

  return runstream.Run(ctx, cmd, actparse.NewParser(), "start dockerized act")
}

func loadJobGraph(options Options) event.JobGraph {
  graph, err := Graph(options.Repo, options.Workflow)
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("job graph: %v", err),
    })
    return event.JobGraph{
      Type: event.TypeJobGraph,
      Jobs: []event.JobNode{},
    }
  }
  return graph
}

func emitSecretSources(options Options) error {
  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return err
  }
  secretFiles, err := envfile.Resolve(repoPath, options.SecretFile)
  if err != nil {
    return err
  }
  if len(secretFiles) == 0 {
    return event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "secrets: none (.env / .env.* not found)",
    })
  }
  names := envfile.SourceNames(repoPath, secretFiles)
  return event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: "secrets: " + strings.Join(names, ", "),
  })
}
