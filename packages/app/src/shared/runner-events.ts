export type RunnerEventType
  = | 'workflow_list'
    | 'run_started'
    | 'run_finished'
    | 'job_graph'
    | 'step_started'
    | 'step_finished'
    | 'log'
    | 'error'
    | 'doctor';

export type StepStatus = 'passed' | 'failed' | 'skipped';
export type RunStatus = 'passed' | 'failed';

export interface Workflow {
  id: string;
  name: string;
  path: string;
}

export interface WorkflowListEvent {
  type: 'workflow_list';
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
}

export interface StepStartedEvent {
  type: 'step_started';
  job: string;
  step: string;
}

export interface JobNode {
  id: string;
  name: string;
  needs: string[];
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
  checks: DoctorCheck[];
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
    | DoctorEvent;
