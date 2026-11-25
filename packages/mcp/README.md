# PipeBox MCP

Stdio [MCP](https://modelcontextprotocol.io) server for AI agents and IDEs. Works with any MCP-capable client (Cursor, Claude Desktop, OpenCode, and others).

## How it works

`pipebox-mcp` wraps the `pipebox` CLI and returns **skinny** results (summaries + on-demand log tails) so local CI runs do not dump full console output into the model context.

It is a thin client of the runner — same boundary as the Electron app. It does not parse GitHub Actions or Jenkins files; it spawns `pipebox` and filters NDJSON.

## Install

```bash
pipebox install-mcp
```

## Tools

| Tool              | Purpose                                                                                           |
| ----------------- | ------------------------------------------------------------------------------------------------- |
| `doctor`          | Check Docker / images / CLI                                                                       |
| `list_workflows`  | List workflows for a repo (`repo` defaults to cwd)                                                |
| `graph_workflow`  | Job/step graph (`needs`) without running — use before multi-job runs                              |
| `run_workflow`    | Start a run (async); returns `status=running` + `runId` immediately; pass `job` for one job/stage |
| `get_run_summary` | Poll until `status` is `passed` or `failed` (`errors`, `failedSteps`, optional `setupLogTail`)    |
| `get_step_logs`   | Tailed log lines for one job/step (default tail: 80); omit job/step for setup logs                |

`workflow` accepts a list **id** (e.g. `ci`) or **path** (e.g. `.github/workflows/ci.yml`).

`job` selects one GitHub Actions job or one Jenkins Declarative stage (name/slug). Prefer this for multi-job workflows (less OOM, clearer failures).

Secrets stay out of MCP args — `pipebox run` loads the repo’s `.env` / `.env.*` as usual.

## Preferred agent loop

1. `list_workflows` → pick a workflow (`id` or `path`)
2. `graph_workflow` → note job/stage ids and `needs`
3. `run_workflow` **with `job`** → get `runId` (`status=running`)
4. Poll `get_run_summary` until `passed` or `failed` (do not shell `pipebox` on timeout)
5. On failure: `get_step_logs` for failed steps (or empty job+step for setup logs)
6. Fix → re-run the failed `job` only

Prefer these tools over shelling `pipebox run` and pasting full NDJSON/logs into the chat.

## Configure your client

Register a stdio MCP server that runs `pipebox-mcp` (or `pipebox mcp`). Restart or reload MCP after saving.

### Cursor

User or project `mcp.json`:

```json
{
  "mcpServers": {
    "pipebox": {
      "command": "pipebox-mcp"
    }
  }
}
```

### Claude Desktop

Edit `claude_desktop_config.json`:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "pipebox": {
      "command": "pipebox-mcp"
    }
  }
}
```

Fully quit and reopen Claude Desktop after saving.

### Claude Code

```bash
claude mcp add --transport stdio pipebox -- pipebox-mcp
```

Or from JSON (project scope with `--scope project`, all projects with `--scope user`):

```bash
claude mcp add-json pipebox '{"command":"pipebox-mcp"}'
```

Verify with `claude mcp list`.

### OpenCode

CLI (project config; add `--global` for every project):

```bash
opencode mcp add pipebox -- pipebox-mcp
```

Or in `opencode.json` / `.opencode/opencode.json` (global: `~/.config/opencode/opencode.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "pipebox": {
        "type": "local",
        "command": ["pipebox-mcp"]
      }
    }
  }
}
```

Check with `opencode mcp list` or `/mcps` in the TUI.
