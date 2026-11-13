import type { StartRunOptions } from './shared/ipc-channels';
import type { PipeboxApi } from './shared/pipebox-api';
import type { RunnerEvent } from './shared/runner-events';
import { contextBridge, ipcRenderer } from 'electron';
import { IPC_CHANNELS } from './shared/ipc-channels';

const pipeboxApi: PipeboxApi = {
  doctor: () => ipcRenderer.invoke(IPC_CHANNELS.doctor),
  listWorkflows: (repoPath) =>
    ipcRenderer.invoke(IPC_CHANNELS.listWorkflows, repoPath),
  selectRepo: () => ipcRenderer.invoke(IPC_CHANNELS.selectRepo),
  startRun: (options: StartRunOptions) =>
    ipcRenderer.invoke(IPC_CHANNELS.startRun, options),
  stopRun: () => ipcRenderer.invoke(IPC_CHANNELS.stopRun),
  onEvent: (callback) => {
    const listener = (_event: Electron.IpcRendererEvent, data: RunnerEvent) => {
      callback(data);
    };
    ipcRenderer.on(IPC_CHANNELS.event, listener);
    return () => {
      ipcRenderer.removeListener(IPC_CHANNELS.event, listener);
    };
  },
};

contextBridge.exposeInMainWorld('pipebox', pipeboxApi);
