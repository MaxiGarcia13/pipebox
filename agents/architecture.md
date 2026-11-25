# Architecture

## Packages

- `packages/app` — Electron shell + Vue UI (Forge, Vite, Tailwind, shadcn-vue)
- `packages/runner` — Go binary `pipebox` (CI runner + public CLI)

## Do not

- Put act/log parsing in `packages/app` (not in Vue, preload, or main as a permanent home — main only spawns and forwards)
- Put CI-provider parsing or detection in the app — runner owns providers; UI only consumes tagged NDJSON
- Reinvent act; wrap the CLI and normalize output to NDJSON events
- Expose a generic shell API through preload
- Build ad-hoc UI primitives; use shadcn-vue under `packages/app/src/components/ui/`
- After `shadcn-vue add`, rename generated Vue files to **kebab-case** (`Button.vue` → `button.vue`) and fix local imports — CLI defaults to PascalCase
- Add shadcn components only when a screen needs them

## Process boundaries

- **Renderer:** typed runner events + shadcn-vue only (provider icons from `workflow.provider`)
- **Preload:** narrow typed bridge
- **Main:** spawn/kill `pipebox`, forward NDJSON → IPC, file dialogs
- **Runner (`pipebox`):** multi-provider discovery under `internal/provider/` — `Discover` / `List` / `Resolve`; GitHub in `provider/github` (Dockerized act, local `pipebox-act` from `packages/runner/docker/act`); Jenkins in `provider/jenkins` (ephemeral official Jenkins controller, local `pipebox-jenkins` from `packages/runner/docker/jenkins`); load project `.env` / `.env.*` as secrets; emit events; CLI commands `doctor`, `list`, `run`, `install`
- **v1 dependency:** Docker only — no local `act` or Jenkins binary; do not depend on a public `ghcr.io/nektos/act` image (it is not publicly pullable); Jenkins image is built FROM `jenkins/jenkins:lts-jdk21` and tagged locally as `pipebox-jenkins` (ephemeral controller per run, Casc, lean plugins, curated built-in toolchains, DooD via docker.sock for `agent { docker { … } }`)

## Providers

- A repo may expose **zero or more** providers (e.g. migration with both GitHub and Jenkins)
- Route `graph` / `run` **per selected workflow** via `provider` field or `--provider`, never a single repo-wide provider
- GitHub-only knowledge (`.github` YAML, act, actparse, actimage) stays under `internal/provider/github/`
- Jenkins-only knowledge (`Jenkinsfile`, console parse, controller lifecycle, Declarative scan) stays under `internal/provider/jenkins/`

## Runner protocol (v1)

Stdout is NDJSON (one object per line): `workflow_list`, `run_started`, `run_finished`, `job_graph`, `step_started`, `step_finished`, `log`, `error`, `artifacts`.

`workflow_list` includes `providers` and each workflow includes `provider` (`github` | `jenkins`).

Mirror types in the app at `packages/app/src/domain/runner-events.ts` (types-only copy of Go structs).
