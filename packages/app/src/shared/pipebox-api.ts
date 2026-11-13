import type { GraphWorkflowOptions, StartRunOptions } from './ipc-channels';
import type {
  DoctorEvent,
  JobGraphEvent,
  RunnerEvent,
  WorkflowListEvent,
} from './runner-events';

export interface PipeboxApi {
  doctor: () => Promise<DoctorEvent>;
  listWorkflows: (repoPath: string) => Promise<WorkflowListEvent>;
  graphWorkflow: (options: GraphWorkflowOptions) => Promise<JobGraphEvent>;
  selectRepo: () => Promise<string | null>;
  startRun: (options: StartRunOptions) => Promise<void>;
  stopRun: () => Promise<void>;
  onEvent: (callback: (event: RunnerEvent) => void) => () => void;
}
