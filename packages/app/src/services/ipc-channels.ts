export const IPC_CHANNELS = {
  doctor: 'pipebox:doctor',
  listWorkflows: 'pipebox:list-workflows',
  graphWorkflow: 'pipebox:graph-workflow',
  selectRepo: 'pipebox:select-repo',
  gitBranch: 'pipebox:git-branch',
  startRun: 'pipebox:start-run',
  stopRun: 'pipebox:stop-run',
  event: 'pipebox:event',
} as const;

export interface StartRunOptions {
  repoPath: string;
  workflow: string;
  job?: string;
  inputs?: string[];
}

export interface GraphWorkflowOptions {
  repoPath: string;
  workflow: string;
}
