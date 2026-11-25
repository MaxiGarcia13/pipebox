---
name: pipebox-act
description: >-
  Work on PipeBox GitHub Actions local runs via Dockerized act (pipebox-act
  image, actparse, actimage). Use when editing provider/github, packages/runner/docker/act,
  act log parsing, GHA list/graph/run, or act image build/ensure.
---

# PipeBox act (GitHub Actions)

Deeper flags and parse notes: [reference.md](reference.md).

## Invariants

- v1: **Docker only** — no host `act` binary.
- Image tag: `pipebox-act:latest` from `packages/runner/docker/act`.
- **Do not** pull or depend on public `ghcr.io/nektos/act` (not publicly pullable).
- Secrets come from the project’s `.env` / `.env.*` (runner loads them); do not invent a parallel secrets path in the app.
- All GHA knowledge stays under `packages/runner/internal/provider/github/`.

## File map

| Path                                | Role                                                  |
| ----------------------------------- | ----------------------------------------------------- |
| `provider/github/list.go`           | List workflows from `.github/workflows`               |
| `provider/github/graph.go`          | Job/step graph (`needs`)                              |
| `provider/github/run.go` / `act.go` | Docker run of act                                     |
| `provider/github/actparse/`         | act stdout → NDJSON step/log events                   |
| `provider/github/actimage/`         | Ensure/build `pipebox-act:latest`                     |
| `packages/runner/docker/act/`       | Dockerfile + `actrc`                                  |
| `testdata/repos/github-*`           | Fixtures (`github-only`, `github-cache-artifacts`, …) |

## Commands

```bash
pnpm runner:image    # build local pipebox-act (+ jenkins)
pnpm runner:doctor   # or: ./pipebox doctor
./pipebox list --repo /path/to/repo
./pipebox run --workflow .github/workflows/ci.yml --repo /path/to/repo
```

## Checklist

1. Prefer changing parse logic in `actparse`, not in the Electron app.
2. After Dockerfile/`actrc` changes, rebuild with `pnpm runner:image`.
3. Cover parser and run behavior with package tests + testdata repos.
4. If new event fields appear, mirror them in `packages/app/src/domain/runner-events.ts`.

## Do not

- Install or shell out to a system `act`.
- Parse workflow YAML in Vue/preload/main for discovery (runner owns that).
- Paste full act console dumps into agent context — use step events / MCP log tails.
