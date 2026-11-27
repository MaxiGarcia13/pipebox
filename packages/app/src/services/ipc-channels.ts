export const IPC_CHANNELS = {
  doctor: 'pipebox:doctor',
  listWorkflows: 'pipebox:list-workflows',
  graphWorkflow: 'pipebox:graph-workflow',
  selectRepo: 'pipebox:select-repo',
  getLaunchRepo: 'pipebox:get-launch-repo',
  openRepo: 'pipebox:open-repo',
  gitBranch: 'pipebox:git-branch',
  startRun: 'pipebox:start-run',
  stopRun: 'pipebox:stop-run',
  saveArtifact: 'pipebox:save-artifact',
  reportIssue: 'pipebox:report-issue',
  event: 'pipebox:event',
} as const;

export interface SaveArtifactOptions {
  sourcePath: string;
  defaultName: string;
}

export interface StartRunOptions {
  repoPath: string;
  workflow: string;
  job?: string;
  inputs?: string[];
  parallel?: boolean;
}

export interface GraphWorkflowOptions {
  repoPath: string;
  workflow: string;
}
