# PipeBox

Local CI/CD control panel for GitHub Actions: pick a repo, run workflows with Dockerized **act**, inject secrets from `.env` / `.env.*`, and watch step logs—without pushing test commits. Jenkins support comes later.

## Install

### Prerequisites

- **Go 1.22+** on `PATH` (for `go install`)
- **Docker** running (for `doctor` / `run`)
- After install, `$GOBIN` or `$GOPATH/bin` must be on your `PATH`

### 1. CLI

```bash
go install github.com/MaxiGarcia13/pipebox/packages/runner/cmd/pipebox@latest
```

Puts `pipebox` on your `PATH` (`$GOBIN` or `$GOPATH/bin`). `@latest` resolves the newest `packages/runner/v*` release. Pin with `@v0.1.0` (module version—not the git tag path prefix).

Or download the CLI binary for your OS/arch from [Releases](https://github.com/MaxiGarcia13/pipebox/releases) and put it on your `PATH`.

### 2. Desktop app

```bash
pipebox install-desktop
```

Downloads the matching desktop app from [Releases](https://github.com/MaxiGarcia13/pipebox/releases) into a user-local folder (no admin/sudo).

Or run `pipebox install` to copy the CLI into `~/.local/bin/pipebox` and fetch the desktop app in one go.

### Verify

```bash
pipebox doctor
pipebox                 # open desktop UI
pipebox --repo .        # open UI with this repo
```

Builds are unsigned for now. On macOS, if Gatekeeper blocks the app: right-click `PipeBox.app` → Open → Open. On Windows, accept the SmartScreen prompt if shown.

Desktop lookup order: `PIPEBOX_APP`, user install (`~/Applications/PipeBox.app`, `%LOCALAPPDATA%\PipeBox\PipeBox.exe`, `~/.local/share/pipebox-app/PipeBox`), system installs (`/Applications`, `/opt/pipebox-app`), then a packaged monorepo build under `packages/app/out/`.

### Alternative — GitHub Release binary

1. Download the CLI asset for your OS/arch from [Releases](https://github.com/MaxiGarcia13/pipebox/releases)
2. Put it on your `PATH`, or run `./pipebox install` (CLI + desktop)
3. Or download a desktop zip/setup/`.deb` manually from the same release

### Alternative — build from source

```bash
pnpm runner:build
# or: make -C packages/runner build
./packages/runner/pipebox install   # CLI + desktop from the published release for this version
```

Local desktop: `pnpm start` (launches Electron in dev). To package locally: `pnpm --filter @pipebox/app package` or `make` (embeds the runner binary).

## CLI

```bash
pipebox                      # open desktop UI
pipebox --repo <path>        # open UI with that repo (e.g. --repo .)
pipebox --help | -h
pipebox --version | -v
pipebox version
pipebox doctor
pipebox list --repo <path>
pipebox graph --repo <path> --workflow <file>
pipebox run --repo <path> --workflow <file> [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]
pipebox install              # CLI → ~/.local/bin + fetch desktop
pipebox install-desktop      # fetch desktop only
pipebox uninstall
```

`run` loads `.env` and any `.env.*` from the repo root (`.env` first, then alphabetical) when `--secret-file` is omitted. `install` copies the current binary into `~/.local/bin/pipebox` and downloads the desktop app for this CLI version; `uninstall` removes both.

## Stack

| Package           | Role                     | Tech                                                                     |
| ----------------- | ------------------------ | ------------------------------------------------------------------------ |
| `packages/app`    | Desktop shell + UI       | Electron, Vue 3, Vite, Tailwind v4, [shadcn-vue](https://shadcn-vue.com) |
| `packages/runner` | CI runner + terminal CLI | Go binary `pipebox`                                                      |

## Architecture

```mermaid
flowchart LR
  Vue[Vue_renderer] -->|IPC| Preload[preload_bridge]
  Preload --> Main[Electron_main]
  Main -->|spawn_stdin_stdout| Runner[Go_runner]
  Runner -->|docker_run| ActImage[nektos_act_image]
  ActImage --> Docker[Docker]
  Runner -->|NDJSON_events| Main
  Main -->|IPC_events| Vue
```

The Go binary is both the app’s runner (spawned by Electron) and the CLI above.

## Develop

```bash
pnpm install
pnpm start
pnpm test
pnpm lint
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

See [AGENTS.md](AGENTS.md) for package boundaries and agent conventions.
