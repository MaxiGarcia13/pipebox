package server

import (
  "context"
  "encoding/json"

  "github.com/mark3labs/mcp-go/mcp"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
)

func (s *Server) runWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  workflow, err := req.RequireString("workflow")
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  repo, err := resolveRepo(req.GetString("repo", ""))
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }

  job := req.GetString("job", "")
  runID := s.Cache.Begin(repo, workflow, job)
  exitCode, runErr := s.Runner.Run(ctx, runner.RunOptions{
    Repo:     repo,
    Workflow: workflow,
    Provider: req.GetString("provider", ""),
    Job:      job,
    Inputs:   req.GetStringSlice("inputs", nil),
  }, func(raw json.RawMessage) error {
    return s.Cache.Ingest(runID, raw)
  })

  summary := s.Cache.Finish(runID, exitCode, runErr)
  return mcp.NewToolResultJSON(summary)
}

func (s *Server) getRunSummary(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  summary, err := s.Cache.Summary(req.GetString("runId", ""))
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }
  return mcp.NewToolResultJSON(summary)
}

func (s *Server) getStepLogs(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
  // Empty job+step returns unscoped setup/preflight logs from the run.
  job := req.GetString("job", "")
  step := req.GetString("step", "")

  tail := req.GetInt("tail", defaultTail)
  if tail <= 0 {
    tail = defaultTail
  }

  lines, err := s.Cache.StepLogs(req.GetString("runId", ""), job, step, tail)
  if err != nil {
    return mcp.NewToolResultError(err.Error()), nil
  }

  payload := map[string]any{
    "runId": req.GetString("runId", s.Cache.LatestID()),
    "job":   job,
    "step":  step,
    "tail":  tail,
    "lines": lines,
  }
  return mcp.NewToolResultJSON(payload)
}
