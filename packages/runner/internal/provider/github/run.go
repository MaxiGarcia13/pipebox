package github

import (
  "bufio"
  "context"
  "errors"
  "fmt"
  "io"
  "os"
  "os/exec"
  "os/signal"
  "path/filepath"
  "strings"
  "sync"
  "syscall"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actparse"
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

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("stdout pipe: %v", err),
    })
    return 1, "failed"
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("stderr pipe: %v", err),
    })
    return 1, "failed"
  }

  if err := cmd.Start(); err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("start dockerized act: %v", err),
    })
    return 1, "failed"
  }

  parser := actparse.NewParser()
  var parseMu sync.Mutex
  var emitMu sync.Mutex
  var waitGroup sync.WaitGroup
  waitGroup.Add(2)
  go streamLines(&waitGroup, stdout, parser, &parseMu, &emitMu)
  go streamLines(&waitGroup, stderr, parser, &parseMu, &emitMu)

  waitErr := cmd.Wait()
  waitGroup.Wait()

  if waitErr == nil {
    return 0, "passed"
  }

  var exitError *exec.ExitError
  if errors.Is(ctx.Err(), context.Canceled) {
    return 130, "cancelled"
  }
  if errors.As(waitErr, &exitError) {
    return exitError.ExitCode(), "failed"
  }
  _ = event.EmitStdout(event.Error{
    Type:    event.TypeError,
    Message: waitErr.Error(),
  })
  return 1, "failed"
}

func streamLines(
  waitGroup *sync.WaitGroup,
  reader io.Reader,
  parser *actparse.Parser,
  parseMu *sync.Mutex,
  emitMu *sync.Mutex,
) {
  defer waitGroup.Done()
  scanner := bufio.NewScanner(reader)
  scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for scanner.Scan() {
    line := scanner.Text()
    parseMu.Lock()
    items := parser.HandleLine(line)
    parseMu.Unlock()
    emitMu.Lock()
    for _, item := range items {
      _ = event.EmitStdout(item)
    }
    emitMu.Unlock()
  }
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
  secretFiles, err := resolveSecretFiles(repoPath, options.SecretFile)
  if err != nil {
    return err
  }
  if len(secretFiles) == 0 {
    return event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "secrets: none (.env / .env.* not found)",
    })
  }
  names := secretSourceNames(repoPath, secretFiles)
  return event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: "secrets: " + strings.Join(names, ", "),
  })
}
