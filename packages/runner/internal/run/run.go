package run

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

  "pipebox/internal/actimage"
  "pipebox/internal/actparse"
  "pipebox/internal/event"
)

type Options struct {
  Repo       string
  Workflow   string
  Job        string
  Secrets    []string
  SecretFile string
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

  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()

  if err := emitSecretSources(options); err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: err.Error(),
    })
  }

  if !actimage.Exists() {
    _ = event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: "building " + actimage.Image + " (first run)",
    })
  }

  cmd, err := buildActCommand(ctx, options)
  if err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: err.Error(),
    })
    _ = event.EmitStdout(event.RunFinished{
      Type:     event.TypeRunFinished,
      Status:   "failed",
      ExitCode: 1,
    })
    return 1, nil
  }

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return 1, fmt.Errorf("stdout pipe: %w", err)
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return 1, fmt.Errorf("stderr pipe: %w", err)
  }

  if err := cmd.Start(); err != nil {
    _ = event.EmitStdout(event.Error{
      Type:    event.TypeError,
      Message: fmt.Sprintf("start dockerized act: %v", err),
    })
    _ = event.EmitStdout(event.RunFinished{
      Type:     event.TypeRunFinished,
      Status:   "failed",
      ExitCode: 1,
    })
    return 1, nil
  }

  parser := actparse.NewParser()
  var waitGroup sync.WaitGroup
  waitGroup.Add(2)
  go streamLines(&waitGroup, stdout, parser)
  go streamLines(&waitGroup, stderr, parser)

  waitErr := cmd.Wait()
  waitGroup.Wait()

  exitCode := 0
  status := "passed"
  if waitErr != nil {
    var exitError *exec.ExitError
    if errors.As(waitErr, &exitError) {
      exitCode = exitError.ExitCode()
    } else if errors.Is(ctx.Err(), context.Canceled) {
      exitCode = 130
    } else {
      exitCode = 1
      _ = event.EmitStdout(event.Error{
        Type:    event.TypeError,
        Message: waitErr.Error(),
      })
    }
    status = "failed"
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

func streamLines(waitGroup *sync.WaitGroup, reader io.Reader, parser *actparse.Parser) {
  defer waitGroup.Done()
  scanner := bufio.NewScanner(reader)
  scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for scanner.Scan() {
    for _, item := range parser.HandleLine(scanner.Text()) {
      _ = event.EmitStdout(item)
    }
  }
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
