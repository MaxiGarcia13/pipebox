package main

import (
  "fmt"
  "os"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

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

func printCommandUsage(command string) {
  fmt.Fprintln(os.Stderr, strings.TrimRight(cli.CommandUsage(command), "\n"))
}
