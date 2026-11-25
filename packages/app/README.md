# PipeBox app

Electron shell + Vue UI. Spawns the `pipebox` runner and renders typed NDJSON events.

## How it works

```text
Vue renderer → preload (typed bridge) → Electron main → spawn pipebox
     ↑                                              |
     └──────── IPC events (RunnerEvent) ────────────┘
```

Main spawns/kills `pipebox` and forwards NDJSON over IPC. The UI consumes tagged events only — no CI file parsing. Provider icons come from `workflow.provider`.

## Support

| OS / arch     | Desktop |
| ------------- | ------- |
| macOS arm64   | ✓       |
| macOS amd64   | ✓       |
| Windows amd64 | ✓       |
| Linux amd64   | ✓       |
| Linux arm64   | —       |

## Related

- [packages/runner/README.md](../runner/README.md)
- [packages/mcp/README.md](../mcp/README.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
