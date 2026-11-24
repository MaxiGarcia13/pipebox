package server

import (
  "github.com/mark3labs/mcp-go/mcp"
  mcpserver "github.com/mark3labs/mcp-go/server"
)

func (s *Server) registerTools(mcpServer *mcpserver.MCPServer) {
  mcpServer.AddTool(mcp.NewTool("doctor",
    mcp.WithDescription("Check local PipeBox prerequisites (Docker, images, CLI)."),
  ), s.doctor)

  mcpServer.AddTool(mcp.NewTool("list_workflows",
    mcp.WithDescription("List local CI workflows for a repo (GitHub Actions and/or Jenkins)."),
    mcp.WithString("repo",
      mcp.Description("Repo path. Defaults to the MCP process working directory."),
    ),
  ), s.listWorkflows)

  mcpServer.AddTool(mcp.NewTool("graph_workflow",
    mcp.WithDescription("Return the job/step graph (ids, needs) without running. Call this before run_workflow when a workflow has multiple jobs/stages, then run one job at a time."),
    mcp.WithString("repo",
      mcp.Description("Repo path. Defaults to the MCP process working directory."),
    ),
    mcp.WithString("workflow",
      mcp.Required(),
      mcp.Description("Workflow id (e.g. ci) or path (e.g. .github/workflows/ci.yml) from list_workflows."),
    ),
    mcp.WithString("provider",
      mcp.Description("Provider hint: github or jenkins."),
      mcp.Enum("github", "jenkins"),
    ),
  ), s.graphWorkflow)

  mcpServer.AddTool(mcp.NewTool("run_workflow",
    mcp.WithDescription("Start a local workflow run asynchronously. Returns immediately with status=running and runId. Prefer job= for one GitHub job or Jenkins Declarative stage. Poll get_run_summary until status is passed or failed — do not shell pipebox or dump NDJSON. On failure: read errors/setupLogTail; use get_step_logs for failed step tails."),
    mcp.WithString("repo",
      mcp.Description("Repo path. Defaults to the MCP process working directory."),
    ),
    mcp.WithString("workflow",
      mcp.Required(),
      mcp.Description("Workflow id (e.g. ci) or path (e.g. .github/workflows/ci.yml) from list_workflows."),
    ),
    mcp.WithString("provider",
      mcp.Description("Provider hint: github or jenkins."),
      mcp.Enum("github", "jenkins"),
    ),
    mcp.WithString("job",
      mcp.Description("Preferred: single GitHub job id or Jenkins Declarative stage name/slug. Omit only for single-job workflows. Respect graph needs order (e.g. run test and lint before e2e)."),
    ),
    mcp.WithArray("inputs",
      mcp.Description("Workflow inputs as KEY=VALUE strings."),
      mcp.Items(map[string]any{"type": "string"}),
    ),
  ), s.runWorkflow)

  mcpServer.AddTool(mcp.NewTool("get_run_summary",
    mcp.WithDescription("Poll the skinny summary for a runId from run_workflow. status is running until the run finishes (passed or failed). Do not shell pipebox while waiting."),
    mcp.WithString("runId",
      mcp.Description("Run id from run_workflow. Defaults to the latest run."),
    ),
  ), s.getRunSummary)

  mcpServer.AddTool(mcp.NewTool("get_step_logs",
    mcp.WithDescription("Return a tailed log slice for one job/step from a cached run. Use failedSteps from the summary. Omit job and step (or pass empty) for setup/preflight logs."),
    mcp.WithString("runId",
      mcp.Description("Run id from run_workflow. Defaults to the latest run."),
    ),
    mcp.WithString("job",
      mcp.Description("Job id. Empty with empty step returns setup/preflight logs."),
    ),
    mcp.WithString("step",
      mcp.Description("Step name. Empty with empty job returns setup/preflight logs."),
    ),
    mcp.WithNumber("tail",
      mcp.Description("Max lines from the end of the step log. Defaults to 80."),
    ),
  ), s.getStepLogs)
}
