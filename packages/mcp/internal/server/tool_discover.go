package server

import (
  "context"

  "github.com/mark3labs/mcp-go/mcp"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
)

func (s *Server) doctor(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  out, err := s.Runner.Doctor(ctx)
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  return mcp.NewToolResultJSON(out)
}

func (s *Server) listWorkflows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  repo, err := resolveRepo(req.GetString("repo", ""))
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  out, err := s.Runner.List(ctx, runner.ListOptions{Repo: repo})
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  return mcp.NewToolResultJSON(out)
}

func (s *Server) graphWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  workflow, err := req.RequireString("workflow")
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  repo, err := resolveRepo(req.GetString("repo", ""))
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  out, err := s.Runner.Graph(ctx, runner.GraphOptions{
    Repo:     repo,
    Workflow: workflow,
    Provider: req.GetString("provider", ""),
  })
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  return mcp.NewToolResultJSON(out)
}
