# Contributing

## Prerequisites

- **Node.js** with **pnpm** (`packageManager` in root `package.json`)
- **Go 1.22+** on `PATH`
- **Docker** running (GitHub workflows via the local `pipebox-act` image)

## Setup

```bash
pnpm install
pnpm runner:image   # once: build local pipebox-act image
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
```

From `packages/runner`:

```bash
make build
./pipebox doctor
./pipebox list --repo /path/to/repo
```

## Package boundaries

| Package           | Owns                                                 | Does not own                           |
| ----------------- | ---------------------------------------------------- | -------------------------------------- |
| `packages/app`    | Electron UI, typed runner events, IPC bridge         | act/log parsing, CI-provider detection |
| `packages/runner` | Providers, Dockerized act, NDJSON events, public CLI | Vue / preload UI                       |

- Route `graph` / `run` **per workflow** via `provider` (or `--provider`), never a single repo-wide provider.
- GitHub-only code stays under `packages/runner/internal/provider/github/`.
- App UI uses shadcn-vue under `packages/app/src/components/ui/`. After `shadcn-vue add`, rename generated files to **kebab-case** and fix imports.
- Do not expose a generic shell API through preload.

Deeper agent/architecture notes: [AGENTS.md](AGENTS.md).

## Pull requests

- Keep changes focused; match existing naming (kebab-case files, camelCase functions).
- Run `pnpm lint` and `pnpm test` before opening a PR.
- Prefer small modules with a single responsibility (app main / preload / renderer stay separate).
