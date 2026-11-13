import type { StartRunOptions } from './ipc-channels';
import type { DoctorEvent, RunnerEvent, WorkflowListEvent } from './runner-events';

export interface PipeboxApi {
  doctor: () => Promise<DoctorEvent>;
  listWorkflows: (repoPath: string) => Promise<WorkflowListEvent>;
  selectRepo: () => Promise<string | null>;
  startRun: (options: StartRunOptions) => Promise<void>;
  stopRun: () => Promise<void>;
  onEvent: (callback: (event: RunnerEvent) => void) => () => void;
}
