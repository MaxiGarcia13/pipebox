export const IPC_CHANNELS = {
  doctor: 'pipebox:doctor',
  listWorkflows: 'pipebox:list-workflows',
  selectRepo: 'pipebox:select-repo',
  startRun: 'pipebox:start-run',
  stopRun: 'pipebox:stop-run',
  event: 'pipebox:event',
} as const;

export interface StartRunOptions {
  repoPath: string;
  workflow: string;
  job?: string;
}
