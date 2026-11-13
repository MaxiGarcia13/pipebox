# PipeBox — agent notes

Local CI/CD desktop control panel: wrap **act** + Docker (Jenkins later) so users can pick a repo, run workflows, manage secrets, and watch step logs without pushing test commits.

## Architecture

```mermaid
flowchart LR
  Vue[Vue_renderer] -->|IPC| Preload[preload_bridge]
  Preload --> Main[Electron_main]
  Main -->|spawn_stdin_stdout| Runner[Go_runner]
  Runner -->|exec| Act[act_CLI]
  Act --> Docker[Docker]
  Runner -->|NDJSON_events| Main
  Main -->|IPC_events| Vue
```

## MVP vs later

- **MVP:** GitHub Actions via act, Electron UI, `pipebox` CLI, secrets in Electron main.
- **Later:** Jenkins adapter, richer installers (Homebrew), package runner into Forge release.

## Commands

```bash
pnpm app:start       # Electron app
pnpm runner:build    # Go CLI → packages/runner/pipebox
pnpm runner:lint       # golangci-lint
pnpm runner:lint:fix  # golangci-lint --fix
pnpm runner:doctor
```
