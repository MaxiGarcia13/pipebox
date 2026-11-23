export type RunnerEventType
  = | 'workflow_list'
    | 'run_started'
    | 'run_finished'
    | 'job_graph'
    | 'step_started'
    | 'step_finished'
    | 'log'
    | 'error'
    | 'doctor'
    | 'artifacts';

export type StepStatus = 'passed' | 'failed' | 'skipped' | 'cancelled';
export type RunStatus = 'passed' | 'failed' | 'cancelled';
export type WorkflowProvider = 'github' | 'jenkins';

export interface WorkflowInput {
  id: string;
  description?: string;
  required?: boolean;
  default?: string;
  type?: 'string' | 'boolean' | 'choice' | 'number';
  options?: string[];
}

export interface Workflow {
  id: string;
  name: string;
  path: string;
  provider: WorkflowProvider;
  callable?: boolean;
  events?: string[];
  inputs?: WorkflowInput[];
}

export interface WorkflowListEvent {
  type: 'workflow_list';
  providers?: WorkflowProvider[];
  workflows: Workflow[];
}

export interface RunStartedEvent {
  type: 'run_started';
  repo: string;
  workflow: string;
  job?: string;
}

export interface RunFinishedEvent {
  type: 'run_finished';
  status: RunStatus;
  exitCode: number;
  receivedAt?: number;
}

export interface StepStartedEvent {
  type: 'step_started';
  job: string;
  step: string;
  receivedAt?: number;
}

export interface JobNode {
  id: string;
  name: string;
  needs: string[];
  steps?: string[];
}

export interface JobGraphEvent {
  type: 'job_graph';
  jobs: JobNode[];
}

export interface StepFinishedEvent {
  type: 'step_finished';
  job: string;
  step: string;
  status: StepStatus;
  durationMs?: number;
  receivedAt?: number;
}

export interface LogEvent {
  type: 'log';
  job?: string;
  step?: string;
  line: string;
}

export interface ErrorEvent {
  type: 'error';
  message: string;
}

export interface DoctorCheck {
  name: string;
  ok: boolean;
  detail?: string;
  path?: string;
}

export interface DoctorEvent {
  type: 'doctor';
  ok: boolean;
  version: string;
  checks: DoctorCheck[];
}

export interface Artifact {
  name: string;
  path: string;
  size: number;
}

export interface ArtifactsEvent {
  type: 'artifacts';
  artifacts: Artifact[];
}

export type RunnerEvent
  = | WorkflowListEvent
    | RunStartedEvent
    | RunFinishedEvent
    | JobGraphEvent
    | StepStartedEvent
    | StepFinishedEvent
    | LogEvent
    | ErrorEvent
    | DoctorEvent
    | ArtifactsEvent;
