package main

import (
  "fmt"
  "os"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func emitInstallLog(line string) {
  if err := event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: line,
  }); err != nil {
    fail(err)
  }
}

func emitError(err error) {
  _ = event.EmitStdout(event.Error{
    Type:    event.TypeError,
    Message: err.Error(),
  })
}

func fail(err error) {
  emitError(err)
  printUsage()
  os.Exit(2)
}

func printUsage() {
  fmt.Fprintln(os.Stderr, strings.TrimRight(cli.Usage(), "\n"))
}
