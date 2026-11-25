# pipebox-act reference

## Image

- Dockerfile: `packages/runner/docker/act/Dockerfile` (Alpine, downloads nektos/act release into the image).
- Config: `packages/runner/docker/act/actrc` copied to `/root/.config/act/actrc`.
- Go ensure/build: `packages/runner/internal/provider/github/actimage` (`Image = "pipebox-act:latest"`).
- Override Dockerfile dir: `PIPEBOX_ACT_DOCKERFILE_DIR` (see `dockerimage.Config`).

## Runtime notes

- Act runs inside Docker; host needs Docker on `PATH`.
- Cache/artifacts helpers live beside the provider (`cache.go`, `artifacts.go`) — keep host↔container path assumptions in github package tests.
- Job selection: pass a single job when multi-job workflows are heavy (CLI `--job` / MCP `job`).
- Always pass `--concurrent-jobs 1` so independent jobs (e.g. lint + test) do not interleave stdout; actparse cannot safely attribute unprefixed `| …` lines across parallel jobs.

## actparse

- Package: `packages/runner/internal/provider/github/actparse`.
- `Parser.HandleLine` turns raw act lines into `step_started` / `step_finished` / `log` / `error` events.
- Job prefix splitting and step start/finish matchers are the fragile surface — extend with tests in `parser_test.go` before broadening regexes.
- Unprefixed `|` continuations stick to the active job/step; with `--concurrent-jobs 1` that sticky rule stays correct.
- Duration parsing: `duration.go` (`parseActDurationMs`).

## Testdata

Under `packages/runner/testdata/repos/`:

| Repo                     | Use                        |
| ------------------------ | -------------------------- |
| `github-only`            | Basic GHA discovery/run    |
| `github-cache-artifacts` | Cache + artifacts paths    |
| `mixed`                  | Both providers in one repo |
| `empty`                  | No workflows               |

## Rebuild

```bash
pnpm runner:image
# or from packages/runner: make image / actimage ensure via doctor
```
