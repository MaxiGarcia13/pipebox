# Contributing

## Prerequisites

- **Node.js 26+** and **pnpm** (`engines` / `packageManager` in root `package.json`)
- **Go 1.22+** on `PATH` (see `packages/runner/go.mod`)
- **golangci-lint** on `PATH` (for `pnpm lint` / `pnpm runner:lint`; CI uses v2.15.0)
- **Docker** running (GitHub via `pipebox-act`; Jenkins via ephemeral `pipebox-jenkins` from `jenkins/jenkins:lts-jdk21`)

## Setup

```bash
pnpm install
pnpm runner:image   # once: build local pipebox-act + pipebox-jenkins images (Jenkins is rebuilt when its revision label changes)
pnpm start          # build runner + launch Electron app
```

## Commands

```bash
pnpm start
pnpm test              # Go runner tests + app unit (Vitest)
pnpm test:e2e          # package Electron app + Playwright smoke
pnpm lint
pnpm lint:fix
pnpm runner:build
pnpm runner:test
pnpm runner:image
pnpm runner:lint
pnpm runner:lint:fix
pnpm runner:doctor
pnpm mcp:build
pnpm mcp:test
pnpm mcp:lint
pnpm mcp:lint:fix
```

From `packages/runner`:

```bash
make build
./pipebox doctor
./pipebox                    # open desktop UI (reuses a running app)
./pipebox --repo /path/to/repo
./pipebox -n --repo /path/to/repo   # new window
./pipebox list
./pipebox list --repo /path/to/repo
./pipebox run --workflow .github/workflows/ci.yml
```

From `packages/mcp` (or via `pipebox mcp`):

```bash
make build                   # or: pnpm mcp:build
make test
./pipebox-mcp                # stdio server; point a client at this binary
```

`pipebox mcp` resolves managed install → `PATH` → monorepo build → `PIPEBOX_MCP`.

## Package boundaries

| Package           | Owns                                                 | Does not own                           |
| ----------------- | ---------------------------------------------------- | -------------------------------------- |
| `packages/app`    | Electron UI, typed runner events, IPC bridge         | act/log parsing, CI-provider detection |
| `packages/runner` | Providers, Dockerized act, NDJSON events, public CLI | Vue / preload UI                       |
| `packages/mcp`    | Stdio MCP server, skinny tools over runner NDJSON    | Providers, CI file parsing             |

- Route `graph` / `run` **per workflow** via `provider` (or `--provider`), never a single repo-wide provider.
- GitHub-only code stays under `packages/runner/internal/provider/github/`.
- App UI uses shadcn-vue under `packages/app/src/components/ui/`. After `shadcn-vue add`, rename generated files to **kebab-case** and fix imports.
- Do not expose a generic shell API through preload.
- Package notes: [packages/app/README.md](packages/app/README.md), [packages/runner/README.md](packages/runner/README.md), [packages/mcp/README.md](packages/mcp/README.md).

## AI agents

Shared guidance lives in [AGENTS.md](AGENTS.md), [`rules/`](rules/), and [`skills/`](skills/).

Wire **rules** and **skills** locally for whatever agent you use (e.g. Cursor `.cursor/`, OpenCode `.opencode/`, Claude `CLAUDE.md`). They are not installed automatically on clone. Do not commit tool-specific config (`.cursor/`, `.opencode/`, `CLAUDE.md`, `.claude/`).

MCP: [packages/mcp/README.md](packages/mcp/README.md).

## Pull requests

- Keep changes focused; match existing naming (kebab-case files, camelCase functions).
- Run `pnpm lint` and `pnpm test` before opening a PR.
- Prefer small modules with a single responsibility (app main / preload / renderer stay separate).
