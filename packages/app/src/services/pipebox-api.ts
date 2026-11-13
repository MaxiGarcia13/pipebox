import type {
  GraphWorkflowOptions,
  SaveArtifactOptions,
  StartRunOptions,
} from './ipc-channels';
import type {
  DoctorEvent,
  JobGraphEvent,
  RunnerEvent,
  WorkflowListEvent,
} from '@/domain/runner-events';

export interface PipeboxApi {
  doctor: () => Promise<DoctorEvent>;
  listWorkflows: (repoPath: string) => Promise<WorkflowListEvent>;
  graphWorkflow: (options: GraphWorkflowOptions) => Promise<JobGraphEvent>;
  selectRepo: () => Promise<string | null>;
  gitBranch: (repoPath: string) => Promise<string | null>;
  startRun: (options: StartRunOptions) => Promise<void>;
  stopRun: () => Promise<void>;
  saveArtifact: (options: SaveArtifactOptions) => Promise<string | null>;
  onEvent: (callback: (event: RunnerEvent) => void) => () => void;
}
