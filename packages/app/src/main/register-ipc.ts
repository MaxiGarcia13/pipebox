import type { DoctorEvent, JobGraphEvent, WorkflowListEvent } from '../domain/runner-events';
import type {
  GraphWorkflowOptions,
  SaveArtifactOptions,
  StartRunOptions,
} from '../services/ipc-channels';
import { BrowserWindow, dialog, ipcMain } from 'electron';
import { IPC_CHANNELS } from '../services/ipc-channels';
import { openReportIssue } from './open-report-issue';
import { resolveGitBranch } from './resolve-git-branch';
import { RunnerClient } from './runner-client';
import { saveArtifact } from './save-artifact';

const runnerClient = new RunnerClient();

export function registerIpc(): void {
  ipcMain.handle(IPC_CHANNELS.doctor, async () => {
    const events = await runnerClient.runToCompletion(['doctor']);
    const doctorEvent = events.find((event) => event.type === 'doctor');
    if (!doctorEvent || doctorEvent.type !== 'doctor') {
      throw new Error('doctor did not return a doctor event');
    }
    return doctorEvent satisfies DoctorEvent;
  });

  ipcMain.handle(IPC_CHANNELS.listWorkflows, async (_event, repoPath: string) => {
    const events = await runnerClient.runToCompletion([
      'list',
      '--repo',
      repoPath,
    ]);
    const listEvent = events.find((event) => event.type === 'workflow_list');
    if (!listEvent || listEvent.type !== 'workflow_list') {
      throw new Error('list did not return a workflow_list event');
    }
    return listEvent satisfies WorkflowListEvent;
  });

  ipcMain.handle(IPC_CHANNELS.graphWorkflow, async (_event, options: GraphWorkflowOptions) => {
    const events = await runnerClient.runToCompletion([
      'graph',
      '--repo',
      options.repoPath,
      '--workflow',
      options.workflow,
    ]);
    const graphEvent = events.find((event) => event.type === 'job_graph');
    if (!graphEvent || graphEvent.type !== 'job_graph') {
      throw new Error('graph did not return a job_graph event');
    }
    return graphEvent satisfies JobGraphEvent;
  });

  ipcMain.handle(IPC_CHANNELS.selectRepo, async (event) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    const result = window
      ? await dialog.showOpenDialog(window, {
          properties: ['openDirectory'],
        })
      : await dialog.showOpenDialog({
          properties: ['openDirectory'],
        });
    if (result.canceled || result.filePaths.length === 0) {
      return null;
    }
    return result.filePaths[0];
  });

  ipcMain.handle(IPC_CHANNELS.gitBranch, async (_event, repoPath: string) => {
    return resolveGitBranch(repoPath);
  });

  ipcMain.handle(IPC_CHANNELS.startRun, async (event, options: StartRunOptions) => {
    const args = [
      'run',
      '--repo',
      options.repoPath,
      '--workflow',
      options.workflow,
    ];
    if (options.job) {
      args.push('--job', options.job);
    }
    for (const input of options.inputs ?? []) {
      args.push('--input', input);
    }

    const window = BrowserWindow.fromWebContents(event.sender);
    void runnerClient.start(args, (runnerEvent) => {
      window?.webContents.send(IPC_CHANNELS.event, runnerEvent);
    });
  });

  ipcMain.handle(IPC_CHANNELS.stopRun, async () => {
    runnerClient.stop();
  });

  ipcMain.handle(IPC_CHANNELS.saveArtifact, async (event, options: SaveArtifactOptions) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    return saveArtifact(window, options);
  });

  ipcMain.handle(IPC_CHANNELS.reportIssue, async () => {
    await openReportIssue();
  });
}
