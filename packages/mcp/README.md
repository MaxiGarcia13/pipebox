# PipeBox MCP

Stdio [MCP](https://modelcontextprotocol.io) server for AI agents and IDEs. It wraps the `pipebox` CLI and returns **skinny** results (summaries + on-demand log tails) so local CI runs do not dump full console output into the model context.

Works with any MCP-capable client (Cursor, Claude Desktop, and others).

## Prerequisites

- Docker running
- `pipebox` CLI on `PATH` (`pipebox doctor` OK)

## Install

```bash
pipebox install-mcp
```

This downloads `pipebox-mcp` to `~/.local/bin` (same managed layout as the CLI) and prints a config snippet.

Override the binary with `PIPEBOX_MCP` if needed.

## Configure your client

Add a stdio MCP server that runs `pipebox-mcp` (or `pipebox mcp`):

```json
{
  "mcpServers": {
    "pipebox": {
      "command": "pipebox-mcp"
    }
  }
}
```

**Cursor** — MCP settings / `mcp.json` (user or project).  
**Claude Desktop** — `claude_desktop_config.json` under the app support directory for your OS.  
**Other agents** — any client that can launch a stdio MCP command.

After saving, restart or reload the MCP connection.

## Tools

| Tool              | Purpose                                                                                  |
| ----------------- | ---------------------------------------------------------------------------------------- |
| `doctor`          | Check Docker / images / CLI                                                              |
| `list_workflows`  | List workflows for a repo (`repo` defaults to cwd)                                       |
| `graph_workflow`  | Job/step graph without running                                                           |
| `run_workflow`    | Run locally; skinny summary (`status`, `errors`, `failedSteps`, optional `setupLogTail`) |
| `get_run_summary` | Cached summary from a prior `run_workflow`                                               |
| `get_step_logs`   | Tailed log lines for one job/step (default tail: 80); omit job/step for setup logs       |

`workflow` accepts a list **id** (e.g. `ci`) or **path** (e.g. `.github/workflows/ci.yml`).

Secrets stay out of MCP args — `pipebox run` loads the repo’s `.env` / `.env.*` as usual.

### Preferred agent loop

1. `list_workflows` → pick a workflow (`id` or `path`)
2. `run_workflow` → read `errors` / `setupLogTail` / `failedSteps` (do not shell `pipebox`)
3. `get_step_logs` on failed steps only (or empty job+step for more setup logs)
4. Fix → `run_workflow` again

Prefer these tools over shelling `pipebox run` and pasting full NDJSON/logs into the chat.

## Develop

From the monorepo:

```bash
pnpm mcp:build   # or: make -C packages/mcp build
pnpm mcp:test
./packages/mcp/pipebox-mcp   # stdio server
```

Point a client at the built binary, or:

```bash
pipebox mcp
```

(`pipebox mcp` resolves managed install → `PATH` → monorepo build → `PIPEBOX_MCP`.)

## Architecture

`pipebox-mcp` is a thin client of the runner — same boundary as the Electron app. It does not parse GitHub Actions or Jenkins files; it spawns `pipebox` and filters NDJSON.
