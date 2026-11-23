import type { StepStatus } from '@/domain/runner-events';

export type JobRunStatus = 'pending' | 'running' | StepStatus;
export type StepRunStatus = 'pending' | 'running' | StepStatus;

export interface StepGroup {
  key: string;
  job: string;
  step: string;
  status: StepRunStatus;
  startedAt?: number;
  finishedAt?: number;
  durationMs?: number;
  lines: string[];
  openDefault: boolean;
}

export interface JobSummary {
  id: string;
  name: string;
  needs: string[];
  status: JobRunStatus;
  startedAt?: number;
  finishedAt?: number;
  durationMs?: number;
}

export interface GroupedRun {
  jobs: JobSummary[];
  steps: StepGroup[];
}
