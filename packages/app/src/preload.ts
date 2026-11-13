import type { RunnerEvent } from './domain/runner-events';
import type {
  GraphWorkflowOptions,
  SaveArtifactOptions,
  StartRunOptions,
} from './services/ipc-channels';
import type { PipeboxApi } from './services/pipebox-api';
import { contextBridge, ipcRenderer } from 'electron';
import { IPC_CHANNELS } from './services/ipc-channels';

const pipeboxApi: PipeboxApi = {
  doctor: () => ipcRenderer.invoke(IPC_CHANNELS.doctor),
  listWorkflows: (repoPath) =>
    ipcRenderer.invoke(IPC_CHANNELS.listWorkflows, repoPath),
  graphWorkflow: (options: GraphWorkflowOptions) =>
    ipcRenderer.invoke(IPC_CHANNELS.graphWorkflow, options),
  selectRepo: () => ipcRenderer.invoke(IPC_CHANNELS.selectRepo),
  gitBranch: (repoPath) => ipcRenderer.invoke(IPC_CHANNELS.gitBranch, repoPath),
  startRun: (options: StartRunOptions) =>
    ipcRenderer.invoke(IPC_CHANNELS.startRun, options),
  stopRun: () => ipcRenderer.invoke(IPC_CHANNELS.stopRun),
  saveArtifact: (options: SaveArtifactOptions) =>
    ipcRenderer.invoke(IPC_CHANNELS.saveArtifact, options),
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
