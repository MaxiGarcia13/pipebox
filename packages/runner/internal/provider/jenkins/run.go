package jenkins

import (
  "context"
  "fmt"
  "os"
  "os/signal"
  "path/filepath"
  "strings"
  "syscall"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/envfile"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins/jfrimage"
)

type Options struct {
  Repo       string
  Workflow   string
  Job        string
  Secrets    []string
  SecretFile string
  Inputs     []string
  RunHome    string
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

  runHome, err := os.MkdirTemp("", "pipebox-jenkins-home-*")
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("create run home: %v", err),
    })
    return 1, nil
  }
  defer func() { _ = os.RemoveAll(runHome) }()
  options.RunHome = runHome

  if !jfrimage.Exists() {
    _ = event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "building " + jfrimage.Image + " (missing or outdated ephemeral Jenkins image)",
    })
  }

  exitCode, status := runController(ctx, options)
  if err := emitRunHomeArtifacts(runHome); err != nil {
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
  if options.Job == "" {
    return graph
  }
  filtered := make([]event.JobNode, 0, 1)
  for _, job := range graph.Jobs {
    if job.ID == options.Job || job.Name == options.Job {
      filtered = append(filtered, event.JobNode{
        ID:    job.ID,
        Name:  job.Name,
        Needs: []string{},
        Steps: job.Steps,
      })
    }
  }
  if len(filtered) == 0 {
    return graph
  }
  return event.JobGraph{
    Type: event.TypeJobGraph,
    Jobs: filtered,
  }
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
  if len(secretFiles) == 0 && len(options.Secrets) == 0 {
    return event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "secrets: none (.env / .env.* not found)",
    })
  }
  names := envfile.SourceNames(repoPath, secretFiles)
  if len(options.Secrets) > 0 {
    names = append(names, "cli")
  }
  return event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: "secrets: " + strings.Join(names, ", "),
  })
}
