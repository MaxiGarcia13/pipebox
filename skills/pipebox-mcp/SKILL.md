---
name: pipebox-mcp
description: >-
  Use or change PipeBox MCP (pipebox-mcp): skinny tools over runner NDJSON,
  preferred agent CI loop, install/config. Use when editing packages/mcp,
  wiring MCP clients, or running local CI via MCP instead of raw pipebox logs.
---

# PipeBox MCP

Full setup and client snippets: [packages/mcp/README.md](../../packages/mcp/README.md).

## Invariants

- `pipebox-mcp` is a **thin client** of the runner — same boundary as Electron main.
- It must **not** reimplement providers or parse GitHub Actions / Jenkins files.
- Prefer MCP tools over shelling `pipebox run` and pasting full NDJSON/console into chat.
- Secrets stay in the repo’s `.env` / `.env.*` — never pass secret values as MCP args.

## Tools

| Tool              | Purpose                                                          |
| ----------------- | ---------------------------------------------------------------- |
| `doctor`          | Docker / images / CLI                                            |
| `list_workflows`  | Workflows for a repo (`repo` defaults to cwd)                    |
| `graph_workflow`  | Job/step graph without running                                   |
| `run_workflow`    | Start run (async); returns `runId`; pass `job` for one job/stage |
| `get_run_summary` | Poll until `passed` / `failed`                                   |
| `get_step_logs`   | Tailed lines for one job/step (or setup logs)                    |

`workflow` accepts list **id** or **path**. `job` = one GHA job or one Jenkins Declarative stage.

## Preferred agent loop

1. `list_workflows` → pick workflow (`id` or `path`)
2. `graph_workflow` → note job/stage ids and `needs`
3. `run_workflow` **with `job`** → get `runId`
4. Poll `get_run_summary` until terminal (do not shell `pipebox` on timeout)
5. On failure: `get_step_logs` for failed steps (empty job+step for setup)
6. Fix → re-run the failed `job` only

## Commands

```bash
pnpm mcp:build
pipebox install-mcp
# or: pipebox mcp / ./pipebox-mcp
```

## Do not

- Duplicate `provider/github` or `provider/jenkins` logic inside MCP.
- Dump full run logs into the model context when `get_step_logs` / summary suffice.
- Put MCP transport or tool handlers in `packages/app`.
