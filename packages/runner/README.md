# PipeBox runner

Go binary `pipebox` — multi-provider CI runner and public CLI. Shared by the Electron app and `pipebox-mcp`.

## How it works

Discovers providers per repo (`Discover` / `List` / `Resolve`), then routes `graph` / `run` **per selected workflow** via `provider` (or `--provider`).

```mermaid
flowchart LR
  Desktop[Desktop] -->|spawn| Runner[pipebox]
  Terminal[Terminal] -->|run| Runner
  MCP[pipebox_mcp] -->|spawn| Runner
  Runner --> Discover[Discover]
  Discover --> GH[provider_github]
  Discover --> JK[provider_jenkins]
  Runner --> Resolve[Resolve_per_workflow]
  Resolve --> GH
  Resolve --> JK
  GH -->|docker_run| Act[pipebox_act]
  JK -->|ephemeral_controller| Jfr[pipebox_jenkins]
  Act --> Docker[Docker]
  Jfr -->|docker.sock_DooD| Docker
  Runner -->|NDJSON| Desktop
  Runner -->|tables_or_NDJSON| Terminal
  Runner -->|NDJSON| MCP
```

| Provider       | Execution                                                       |
| -------------- | --------------------------------------------------------------- |
| GitHub Actions | Dockerized act (`pipebox-act` from `docker/act`)                |
| Jenkins        | Ephemeral `pipebox-jenkins` controller (`docker/jenkins`, DooD) |

v1 needs Docker only. Secrets load from the repo’s `.env` / `.env.*`. Machines get NDJSON on stdout; interactive terminals get tables for `list` / `graph`.

## Protocol (v1)

Event types: `workflow_list`, `run_started`, `run_finished`, `job_graph`, `step_started`, `step_finished`, `log`, `error`, `artifacts`, `doctor`.

App mirror: `packages/app/src/domain/runner-events.ts` (keep in sync when the Go shape changes).

## Related

- [packages/app/README.md](../app/README.md)
- [packages/mcp/README.md](../mcp/README.md)
- [README.md](../../README.md) (CLI)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
