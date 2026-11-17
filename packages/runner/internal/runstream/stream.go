package runstream

import (
  "bufio"
  "context"
  "errors"
  "fmt"
  "io"
  "os/exec"
  "sync"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type LineHandler interface {
  HandleLine(line string) []any
}

func Run(ctx context.Context, cmd *exec.Cmd, handler LineHandler, startError string) (exitCode int, status string) {
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
      Message: fmt.Sprintf("%s: %v", startError, err),
    })
    return 1, "failed"
  }

  var parseMu sync.Mutex
  var emitMu sync.Mutex
  var waitGroup sync.WaitGroup
  waitGroup.Add(2)
  go streamLines(&waitGroup, stdout, handler, &parseMu, &emitMu)
  go streamLines(&waitGroup, stderr, handler, &parseMu, &emitMu)

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
  handler LineHandler,
  parseMu *sync.Mutex,
  emitMu *sync.Mutex,
) {
  defer waitGroup.Done()
  scanner := bufio.NewScanner(reader)
  scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for scanner.Scan() {
    line := scanner.Text()
    parseMu.Lock()
    items := handler.HandleLine(line)
    parseMu.Unlock()
    emitMu.Lock()
    for _, item := range items {
      _ = event.EmitStdout(item)
    }
    emitMu.Unlock()
  }
}
