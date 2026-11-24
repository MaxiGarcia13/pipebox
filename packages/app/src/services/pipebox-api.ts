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
  doctor: (repoPath?: string | null) => Promise<DoctorEvent>;
  listWorkflows: (repoPath: string) => Promise<WorkflowListEvent>;
  graphWorkflow: (options: GraphWorkflowOptions) => Promise<JobGraphEvent>;
  selectRepo: () => Promise<string | null>;
  getLaunchRepo: () => Promise<string | null>;
  gitBranch: (repoPath: string) => Promise<string | null>;
  startRun: (options: StartRunOptions) => Promise<void>;
  stopRun: () => Promise<void>;
  saveArtifact: (options: SaveArtifactOptions) => Promise<string | null>;
  reportIssue: () => Promise<void>;
  onEvent: (callback: (event: RunnerEvent) => void) => () => void;
  onOpenRepo: (callback: (repoPath: string) => void) => () => void;
}
