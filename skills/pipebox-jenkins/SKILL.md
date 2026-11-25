---
name: pipebox-jenkins
description: >-
  Work on PipeBox ephemeral Jenkins (pipebox-jenkins image, Casc, plugins, DooD,
  Declarative scan, jfrparse, controller lifecycle). Use when editing
  provider/jenkins, packages/runner/docker/jenkins, Jenkinsfile list/graph/run,
  or console/stage parsing.
---

# PipeBox Jenkins

Controller/Casc details: [reference.md](reference.md).

## Invariants

- v1: **ephemeral official controller** per run — no host Jenkins binary.
- Image: `pipebox-jenkins:latest` built FROM `jenkins/jenkins:lts-jdk21` (`packages/runner/docker/jenkins`).
- List/graph: Declarative scan of the Jenkinsfile (no controller required).
- Run: start controller → create/update job → trigger build → stream console via `jfrparse`.
- Agents: **DooD** via docker.sock for `agent { docker { … } }` — not DinD in v1.
- All Jenkins knowledge stays under `packages/runner/internal/provider/jenkins/`.

## File map

| Path                                        | Role                                                                   |
| ------------------------------------------- | ---------------------------------------------------------------------- |
| `provider/jenkins/list.go`                  | Discover Jenkinsfile workflows                                         |
| `provider/jenkins/graph.go` + `scan_*.go`   | Declarative stage/step graph                                           |
| `provider/jenkins/run.go` / `controller.go` | Ephemeral controller lifecycle                                         |
| `provider/jenkins/jfrparse/`                | Console lines → NDJSON events                                          |
| `provider/jenkins/jfrimage/`                | Ensure/build `pipebox-jenkins:latest`                                  |
| `provider/jenkins/secrets.go`               | Inject env/secrets into the run                                        |
| `packages/runner/docker/jenkins/`           | Dockerfile, `casc.yaml`, `plugins.txt`, `entrypoint.sh`                |
| `testdata/repos/jenkins-*`                  | Fixtures (`jenkins-only`, `jenkins-agent-any`, `jenkins-docker-agent`) |

## Commands

```bash
pnpm runner:image    # rebuild when Casc/plugins/Dockerfile/revision label change
./pipebox doctor
./pipebox list --repo /path/to/repo
./pipebox run --workflow Jenkinsfile --repo /path/to/repo
# Prefer a single stage for multi-stage pipelines:
./pipebox run --workflow Jenkinsfile --job build --repo /path/to/repo
```

## Checklist

1. Scan/list/graph changes: keep pure Go scan code; add tests around `scan_*.go` / `graph_test.go`.
2. Console mapping: change `jfrparse` + tests, not the app.
3. Image/Casc/plugin changes: bump/rebuild so `PIPEBOX_JENKINS_REVISION` / label logic picks up the new image (`pnpm runner:image`).
4. Keep plugin set lean; prefer curated `agent any` toolchains already in the image.

## Do not

- Require a preinstalled Jenkins on the host.
- Switch v1 to DinD unless product direction changes.
- Reimplement Jenkinsfile parsing in the Electron app or MCP.
