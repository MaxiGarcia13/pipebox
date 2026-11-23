package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/server"
)

func main() {
  if err := server.New(runner.Default()).Serve(); err != nil {
    fmt.Fprintf(os.Stderr, "pipebox-mcp: %v\n", err)
    os.Exit(1)
  }
}
