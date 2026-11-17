# PipeBox

Local CI/CD control panel for GitHub Actions (Jenkins coming) — run and debug pipelines locally with visual step logs. No git push required.

![PipeBox demo](docs/demo.gif)

| Provider       | Status                                                                  |
| -------------- | ----------------------------------------------------------------------- |
| GitHub Actions | Ready — run locally via Dockerized [act](https://github.com/nektos/act) |
| Jenkins        | In progress — adapter stub, not runnable yet                            |

## Quick start

**Needs:** Docker running · Go 1.22+ on `PATH`

Install the CLI:

```bash
go install github.com/MaxiGarcia13/pipebox/packages/runner/cmd/pipebox@latest
```

Install the desktop app:

```bash
pipebox install-desktop
```

Check setup:

```bash
pipebox doctor
```

Open the UI on this repo:

```bash
pipebox --repo .
```

Or download the CLI binary for your OS/arch from [Releases](https://github.com/MaxiGarcia13/pipebox/releases) and put it on your `PATH`.

Builds are unsigned for now. On macOS, if Gatekeeper blocks the app: right-click `PipeBox.app` → Open → Open. On Windows, accept the SmartScreen prompt if shown.

## What you get

- Browse workflows for a local repo
- Run jobs with Dockerized act (no local `act` binary)
- Inject secrets from `.env` / `.env.*`
- Live step logs in the desktop UI — or pure CLI (`list` / `graph` / `run`)

## Updates

Check for a newer release:

```bash
pipebox update --check
```

Install the latest CLI + desktop:

```bash
pipebox update
```

Pin a version:

```bash
pipebox update --version 0.4.1
```

No second `go install` needed once a `pipebox` binary is on your `PATH`.

Or run `pipebox install` to copy the CLI into `~/.local/bin/pipebox` and fetch the desktop app in one go.

## CLI

Open the desktop UI (reuse if already running):

```bash
pipebox
```

Open / switch to a repo:

```bash
pipebox --repo <path>
```

Open a new window:

```bash
pipebox -n --repo <path>
```

Other commands:

Show help:

```bash
pipebox --help
```

Print the CLI version:

```bash
pipebox --version
```

Check Docker and local setup:

```bash
pipebox doctor
```

List workflows in a repo:

```bash
pipebox list --repo <path>
```

Print the job graph for a workflow:

```bash
pipebox graph --repo <path> --workflow <file>
```

Run a workflow (optional job, inputs, and secrets):

```bash
pipebox run --repo <path> --workflow <file> [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]
```

Install CLI to `~/.local/bin`, fetch the desktop app, and write shell completions:

```bash
pipebox install
```

Fetch the desktop app only:

```bash
pipebox install-desktop
```

Update CLI + desktop (`--check` to report only, `--version` to pin):

```bash
pipebox update [--check] [--version <v>]
```

Remove the local CLI install and desktop app:

```bash
pipebox uninstall
```

Print a shell completion script (`bash` and `fish` also work; `install` / `update` write these automatically):

```bash
pipebox completion zsh
```

## How it works

| Package           | Role                     | Tech                                                                     |
| ----------------- | ------------------------ | ------------------------------------------------------------------------ |
| `packages/app`    | Desktop shell + UI       | Electron, Vue 3, Vite, Tailwind v4, [shadcn-vue](https://shadcn-vue.com) |
| `packages/runner` | CI runner + terminal CLI | Go binary `pipebox`                                                      |

```mermaid
flowchart LR
  Vue[Vue_renderer] -->|IPC| Preload[preload_bridge]
  Preload --> Main[Electron_main]
  Main -->|spawn_stdin_stdout| Runner[Go_runner]
  Runner -->|docker_run| ActImage[pipebox_act_image]
  ActImage --> Docker[Docker]
  Runner -->|NDJSON_events| Main
  Main -->|IPC_events| Vue
```

The Go binary is both the app’s runner (spawned by Electron) and the CLI above. Provider discovery and boundaries: [AGENTS.md](AGENTS.md).

## Develop

See [CONTRIBUTING.md](CONTRIBUTING.md).
