package main

import (
  "fmt"
  "os"
  "os/exec"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
)

func runMCP(args []string) {
  path, err := install.ResolveMCP()
  if err != nil {
    fail(err)
  }

  cmd := exec.Command(path, args...)
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
      os.Exit(exitErr.ExitCode())
    }
    fail(fmt.Errorf("run mcp: %w", err))
  }
}
