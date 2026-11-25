---
name: pipebox-provider
description: >-
  Add or change PipeBox CI providers (Discover, List, Resolve, Graph, Run),
  per-workflow routing, and NDJSON event emission. Use when editing
  packages/runner/internal/provider, adding a provider, fixing workflow
  resolution, or mirroring runner events into the app.
---

# PipeBox providers

## Invariants

- Route `graph` / `run` **per selected workflow** via `provider` or `--provider` — never a single repo-wide provider.
- A repo may expose **zero or more** providers (`Discover`).
- App and MCP never parse CI files; they consume tagged NDJSON only.
- GitHub-only code stays under `packages/runner/internal/provider/github/`.
- Jenkins-only code stays under `packages/runner/internal/provider/jenkins/`.

## File map

| Path                                            | Role                                                                       |
| ----------------------------------------------- | -------------------------------------------------------------------------- |
| `packages/runner/internal/provider/provider.go` | `Provider` (`ID` / `List` / `Graph` / `Run`) + `RunOptions`                |
| `discover.go`                                   | `Discover(repo)` → present adapters                                        |
| `list.go`                                       | Multi-provider workflow list                                               |
| `resolve.go`                                    | Pick provider from hint, path (`.github/…`, Jenkinsfile), or sole provider |
| `adapters.go`                                   | Thin wrappers around github/jenkins packages                               |
| `packages/runner/internal/event/`               | NDJSON event structs + emit                                                |
| `packages/app/src/domain/runner-events.ts`      | Types-only mirror of Go events                                             |

## Protocol (v1)

Stdout is one JSON object per line: `workflow_list`, `run_started`, `run_finished`, `job_graph`, `step_started`, `step_finished`, `log`, `error`, `artifacts` (and `doctor`).

`workflow_list` includes `providers`; each workflow includes `provider` (`github` | `jenkins`).

## Checklist (change a provider)

1. Implement or extend `List` / `Graph` / `Run` in the provider package.
2. Keep discovery in that package’s `Present` (wired from `Discover`).
3. Emit events via `internal/event` — do not invent parallel shapes.
4. If the event schema changes, update `packages/app/src/domain/runner-events.ts`.
5. Add/adjust Go tests; use repos under `packages/runner/testdata/repos/` when behavior needs a fixture (`github-only`, `jenkins-only`, `mixed`, …).
6. Do not teach the Vue app to detect workflows or parse YAML/Jenkinsfile.

## Do not

- Put act/Jenkins log parsing in `packages/app` or `packages/mcp`.
- Resolve a whole-repo provider and ignore `workflow.provider`.
- Depend on a local `act` or Jenkins binary — Docker images only (v1).
