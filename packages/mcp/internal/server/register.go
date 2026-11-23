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
    mcp.WithDescription("Return the job/step graph for one workflow without running it."),
    mcp.WithString("repo",
      mcp.Description("Repo path. Defaults to the MCP process working directory."),
    ),
    mcp.WithString("workflow",
      mcp.Required(),
      mcp.Description("Workflow id or path from list_workflows."),
    ),
    mcp.WithString("provider",
      mcp.Description("Provider hint: github or jenkins."),
      mcp.Enum("github", "jenkins"),
    ),
  ), s.graphWorkflow)

  mcpServer.AddTool(mcp.NewTool("run_workflow",
    mcp.WithDescription("Run a workflow locally. Returns a skinny summary (status, failed steps). Use get_step_logs for log tails."),
    mcp.WithString("repo",
      mcp.Description("Repo path. Defaults to the MCP process working directory."),
    ),
    mcp.WithString("workflow",
      mcp.Required(),
      mcp.Description("Workflow id or path from list_workflows."),
    ),
    mcp.WithString("provider",
      mcp.Description("Provider hint: github or jenkins."),
      mcp.Enum("github", "jenkins"),
    ),
    mcp.WithString("job",
      mcp.Description("Optional single job id to run."),
    ),
    mcp.WithArray("inputs",
      mcp.Description("Workflow inputs as KEY=VALUE strings."),
      mcp.Items(map[string]any{"type": "string"}),
    ),
  ), s.runWorkflow)

  mcpServer.AddTool(mcp.NewTool("get_run_summary",
    mcp.WithDescription("Return the cached skinny summary for a prior run_workflow call."),
    mcp.WithString("runId",
      mcp.Description("Run id from run_workflow. Defaults to the latest run."),
    ),
  ), s.getRunSummary)

  mcpServer.AddTool(mcp.NewTool("get_step_logs",
    mcp.WithDescription("Return a tailed log slice for one job/step from a cached run. Prefer failed steps from the summary."),
    mcp.WithString("runId",
      mcp.Description("Run id from run_workflow. Defaults to the latest run."),
    ),
    mcp.WithString("job",
      mcp.Required(),
      mcp.Description("Job id."),
    ),
    mcp.WithString("step",
      mcp.Required(),
      mcp.Description("Step name."),
    ),
    mcp.WithNumber("tail",
      mcp.Description("Max lines from the end of the step log. Defaults to 80."),
    ),
  ), s.getStepLogs)
}
