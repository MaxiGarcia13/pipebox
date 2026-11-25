# Agent skills

Shared, tool-agnostic playbooks for coding agents. Overview: [AGENTS.md](../AGENTS.md). Always-on conventions: [`rules/`](../rules/). Local tool setup: [CONTRIBUTING.md](../CONTRIBUTING.md).

| Skill                                                 | Topic                                                              |
| ----------------------------------------------------- | ------------------------------------------------------------------ |
| [pipebox-provider](pipebox-provider/SKILL.md)         | Discover / List / Resolve, per-workflow routing, NDJSON events     |
| [pipebox-act](pipebox-act/SKILL.md)                   | Dockerized act, `actparse` / `actimage`, GitHub Actions local runs |
| [pipebox-jenkins](pipebox-jenkins/SKILL.md)           | Ephemeral controller, Casc, DooD, Declarative scan, `jfrparse`     |
| [pipebox-electron-ipc](pipebox-electron-ipc/SKILL.md) | Main / preload / renderer boundaries, spawn runner, IPC            |
| [pipebox-mcp](pipebox-mcp/SKILL.md)                   | `pipebox-mcp` tools and preferred agent CI loop                    |
| [pipebox-debug-run](pipebox-debug-run/SKILL.md)       | Doctor, missing images, failed local runs                          |

Wire these into your agent locally (see [CONTRIBUTING.md](../CONTRIBUTING.md)).
