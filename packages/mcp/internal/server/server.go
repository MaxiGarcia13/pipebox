package server

import (
  "fmt"
  "sync"

  mcpserver "github.com/mark3labs/mcp-go/server"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runcache"
  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
)

const (
  serverName    = "pipebox"
  serverVersion = "0.23.0"
  defaultTail   = 80
)

type Server struct {
  Runner *runner.Client
  Cache  *runcache.Cache
  runs   sync.WaitGroup
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

func (s *Server) Wait() {
  if s == nil {
    return
  }
  s.runs.Wait()
}

func (s *Server) Serve() error {
  if s == nil || s.Runner == nil {
    return fmt.Errorf("server is not configured")
  }
  if s.Cache == nil {
    s.Cache = runcache.New()
  }

  mcpServer := mcpserver.NewMCPServer(
    serverName,
    serverVersion,
    mcpserver.WithToolCapabilities(true),
    mcpserver.WithRecovery(),
  )
  s.registerTools(mcpServer)
  return mcpserver.ServeStdio(mcpServer)
}
