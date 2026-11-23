package server

import (
  "fmt"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runcache"
  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
)

type Server struct {
  Runner *runner.Client
  Cache  *runcache.Cache
}

func New(client *runner.Client) *Server {
  if client == nil {
    client = runner.Default()
  }
  return &Server{
    Runner: client,
    Cache:  runcache.New(),
  }
}

func (s *Server) Serve() error {
  if s == nil || s.Runner == nil {
    return fmt.Errorf("server is not configured")
  }
  // Stdio MCP tool registration lands in the next plan task.
  return fmt.Errorf("MCP stdio server not implemented yet")
}
