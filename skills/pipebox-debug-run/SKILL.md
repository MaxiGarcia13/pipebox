---
name: pipebox-debug-run
description: >-
  Diagnose failed PipeBox local CI runs: Docker down, missing pipebox-act or
  pipebox-jenkins images, empty workflow lists, wrong provider, secrets, or
  multi-job OOM. Use when doctor fails, list is empty, run errors, or debugging
  act/Jenkins local execution.
---

# Debug a local PipeBox run

Prefer skinny MCP tools ([pipebox-mcp](../pipebox-mcp/SKILL.md)) or CLI events over pasting full console dumps into chat.

## Ladder

1. **Docker** — daemon running; `docker` on `PATH`.
2. **Images** — `pnpm runner:image` (builds `pipebox-act:latest` + `pipebox-jenkins:latest`).
3. **Doctor** — `pnpm runner:doctor` or `./pipebox doctor` (optionally with `--repo`).
4. **List** — `./pipebox list --repo /path/to/repo` (expect `provider` on each workflow).
5. **Graph** — `./pipebox graph --workflow <id|path>` / MCP `graph_workflow` before a heavy run.
6. **Run** — prefer a single `--job` / MCP `job` on multi-job or multi-stage workflows.
7. **Logs** — MCP `get_run_summary` + `get_step_logs`, or NDJSON `error` / failed `step_finished` events.

## Common failures

| Symptom                                  | Likely cause                                                  | Fix                                                        |
| ---------------------------------------- | ------------------------------------------------------------- | ---------------------------------------------------------- |
| Doctor / run: docker not found           | Docker missing or not on `PATH`                               | Install/start Docker                                       |
| Image / jenkins image / act image errors | Local images not built or stale after Casc/Dockerfile change  | `pnpm runner:image`                                        |
| Empty workflow list                      | No `.github/workflows` and no Jenkinsfile (or wrong `--repo`) | Check path; use `testdata/repos/*` fixtures                |
| `provider "…" not available`             | Hint doesn’t match Discover                                   | Omit hint or pass matching `--provider`                    |
| `cannot resolve provider`                | Ambiguous path in a mixed repo                                | Pass `--provider github\|jenkins` or a clear workflow path |
| Secrets missing in steps                 | No `.env` / `.env.*` in project                               | Add env files; runner loads them (not MCP args)            |
| OOM / huge multi-job run                 | Full workflow without job filter                              | Re-run with `--job` / stage name                           |
| App shows nothing useful                 | Parsing in UI or dropped events                               | Fix runner emit + `runner-events.ts`; main only forwards   |

## Testdata

Reproduce with `packages/runner/testdata/repos/` (`github-only`, `jenkins-only`, `mixed`, `empty`, …) before blaming the user’s project.

## Do not

- Install host `act` or Jenkins to “fix” v1 — images only.
- Re-parse CI files in the app to debug discovery.
- Flood the chat with entire act/Jenkins console output.
