# pipebox-jenkins reference

## Image

- Base: `jenkins/jenkins:lts-jdk21`.
- Local tag: `pipebox-jenkins:latest` (`jfrimage.Image`).
- Contents: lean plugins (`plugins.txt`), JCasC (`casc.yaml`), curated toolchains (Go/Node/pnpm/Docker CLI, xvfb stack for desktop e2e), entrypoint.
- Label: `pipebox.jenkins.revision` from build-arg `PIPEBOX_JENKINS_REVISION` — rebuild when Casc/plugins/Dockerfile change so ensure logic refreshes the image.
- Override Dockerfile dir via `dockerimage` env pattern (same family as act).

## Controller lifecycle (`controller.go`)

Typical `runController` flow:

1. Require Docker on `PATH`; `jfrimage.Ensure()`.
2. Emit doctor checks for Jenkins.
3. `startController` — temp Jenkins home (`pipebox-jenkins-home-*`), mount repo, docker.sock for DooD.
4. `waitForJenkins` (ready timeout ~3m).
5. Load pipeline script → `createOrUpdateJob` → `triggerBuild`.
6. Stream/poll console; map lines with `jfrparse`; watch stages as needed (`stagewatch.go`).
7. Cleanup session (stop container, remove temp home).

HTTP API helpers: `job.go`, `wfapi.go`, `console.go`.

## List / graph (no controller)

- Presence: Jenkinsfile (and related paths via `IsJenkinsfilePath`).
- Graph: Declarative scan (`scan_block.go`, `scan_calls.go`, …) produces job/stage nodes + steps for `job_graph`.
- `--job` / MCP `job`: one Declarative **stage** name/slug — prefer for multi-stage pipelines (less OOM, clearer failures).

## Secrets

- Project `.env` / `.env.*` loaded by the runner; Jenkins package maps them into the ephemeral controller/build (`secrets.go`).
- Do not pass secret values through MCP tool args.

## jfrparse

- Package: `packages/runner/internal/provider/jenkins/jfrparse`.
- Maps pipeline console prefixes, stage open/close, step lifecycle, failures into NDJSON events.
- Extend matchers with `parser_test.go` coverage; stage names are slugified for ids.

## Testdata

| Repo                   | Use                             |
| ---------------------- | ------------------------------- |
| `jenkins-only`         | Basic Jenkinsfile               |
| `jenkins-agent-any`    | Built-in/`agent any` toolchains |
| `jenkins-docker-agent` | DooD `agent { docker { … } }`   |
| `mixed`                | GitHub + Jenkins together       |

## Rebuild

```bash
pnpm runner:image
```
