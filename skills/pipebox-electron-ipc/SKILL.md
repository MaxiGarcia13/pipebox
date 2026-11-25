---
name: pipebox-electron-ipc
description: >-
  Change PipeBox Electron process boundaries (main, preload, renderer), IPC
  channels, and spawning the pipebox runner. Use when editing packages/app main
  or preload, runner-client, ipc-channels, pipebox-api, or forwarding NDJSON
  events to Vue.
---

# PipeBox Electron IPC

## Process boundaries

```text
Vue renderer  →  preload (typed bridge)  →  Electron main  →  spawn pipebox
     ↑                                              |
     └──────── IPC events (RunnerEvent) ────────────┘
```

| Layer    | Owns                                                | Does not own                           |
| -------- | --------------------------------------------------- | -------------------------------------- |
| Renderer | Typed `RunnerEvent` UI, shadcn-vue                  | act/Jenkins parsing, spawning binaries |
| Preload  | Narrow `window.pipebox` API                         | Generic shell / arbitrary invoke       |
| Main     | Spawn/kill `pipebox`, forward NDJSON → IPC, dialogs | CI file detection                      |
| Runner   | Providers, Docker, NDJSON emit                      | Vue / preload                          |

## File map

| Path                                        | Role                                            |
| ------------------------------------------- | ----------------------------------------------- |
| `packages/app/src/preload.ts`               | `contextBridge.exposeInMainWorld('pipebox', …)` |
| `packages/app/src/services/ipc-channels.ts` | Channel names + option types                    |
| `packages/app/src/services/pipebox-api.ts`  | Renderer-facing API type                        |
| `packages/app/src/main/runner-client.ts`    | Spawn `pipebox`, readline NDJSON, kill          |
| `packages/app/src/main/register-ipc.ts`     | Main handlers for invoke channels               |
| `packages/app/src/domain/runner-events.ts`  | Protocol types (mirror of Go)                   |
| `packages/app/src/components/`              | Vue UI only                                     |

## Adding a capability

1. **Runner first** — emit or accept the behavior in Go (CLI/NDJSON).
2. **Types** — extend `domain/runner-events.ts` if the wire shape changes.
3. **IPC** — add channel + options in `ipc-channels.ts`; handler in main; method on preload + `pipebox-api`.
4. **UI** — call `window.pipebox…` from services/composables; keep domain transforms in `domain/`.

## Do not

- Expose a generic shell or `exec` API through preload.
- Parse act/Jenkins logs or CI files in main/preload/renderer.
- Put IPC contracts under `domain/` or business rules under `utils/`.
- Mix main/preload/renderer concerns in one module.
