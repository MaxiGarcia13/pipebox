import type {
  JobNode,
  RunnerEvent,
  StepStatus,
} from '@/shared/runner-events';

export type JobRunStatus = 'pending' | 'running' | StepStatus;
export type StepRunStatus = 'running' | StepStatus;

export interface StepGroup {
  key: string;
  job: string;
  step: string;
  status: StepRunStatus;
  durationMs?: number;
  lines: string[];
  openDefault: boolean;
}

export interface JobSummary {
  id: string;
  name: string;
  needs: string[];
  status: JobRunStatus;
  durationMs?: number;
}

export interface GroupedRun {
  jobs: JobSummary[];
  steps: StepGroup[];
}

export function groupRunEvents(events: RunnerEvent[]): GroupedRun {
  const graphJobs = jobsFromGraph(events);
  const steps = foldSteps(events);
  const jobs = summarizeJobs(graphJobs, steps);
  return { jobs, steps };
}

export function stepsForJob(steps: StepGroup[], job: JobSummary | null): StepGroup[] {
  if (!job) {
    return steps;
  }
  return steps.filter((step) => matchesJob(job, step.job));
}

export function defaultSelectedJobId(jobs: JobSummary[]): string | null {
  const withActivity = jobs.find((job) => job.status !== 'pending');
  return withActivity?.id ?? jobs[0]?.id ?? null;
}

function jobsFromGraph(events: RunnerEvent[]): JobNode[] {
  for (let index = events.length - 1; index >= 0; index -= 1) {
    const event = events[index];
    if (event?.type === 'job_graph') {
      return event.jobs;
    }
  }
  return [];
}

function foldSteps(events: RunnerEvent[]): StepGroup[] {
  const groups: StepGroup[] = [];
  const byKey = new Map<string, StepGroup>();

  for (const event of events) {
    if (event.type === 'step_started') {
      const key = stepKey(event.job, event.step);
      const existing = byKey.get(key);
      if (existing) {
        existing.status = 'running';
        existing.openDefault = true;
        delete existing.durationMs;
        continue;
      }
      const group: StepGroup = {
        key,
        job: event.job,
        step: event.step,
        status: 'running',
        lines: [],
        openDefault: true,
      };
      byKey.set(key, group);
      groups.push(group);
      continue;
    }

    if (event.type === 'step_finished') {
      const key = stepKey(event.job, event.step);
      let group = byKey.get(key);
      if (!group) {
        group = {
          key,
          job: event.job,
          step: event.step,
          status: event.status,
          lines: [],
          openDefault: event.status === 'failed',
        };
        byKey.set(key, group);
        groups.push(group);
      }
      group.status = event.status;
      group.durationMs = event.durationMs;
      group.openDefault = event.status === 'failed';
      continue;
    }

    if (event.type === 'log') {
      const job = event.job ?? '';
      const step = event.step ?? '';
      if (!job && !step) {
        continue;
      }
      const key = stepKey(job, step || '(log)');
      let group = byKey.get(key);
      if (!group) {
        group = {
          key,
          job,
          step: step || '(log)',
          status: 'running',
          lines: [],
          openDefault: true,
        };
        byKey.set(key, group);
        groups.push(group);
      }
      group.lines.push(event.line);
    }
  }

  return groups;
}

function summarizeJobs(graphJobs: JobNode[], steps: StepGroup[]): JobSummary[] {
  const summaries: JobSummary[] = graphJobs.map((job) => ({
    id: job.id,
    name: job.name,
    needs: job.needs,
    status: 'pending' as JobRunStatus,
  }));

  const byMatch = new Map<string, JobSummary>();
  for (const summary of summaries) {
    byMatch.set(summary.id, summary);
    byMatch.set(summary.name, summary);
  }

  for (const step of steps) {
    let summary = byMatch.get(step.job);
    if (!summary) {
      summary = {
        id: step.job,
        name: step.job,
        needs: [],
        status: 'pending',
      };
      summaries.push(summary);
      byMatch.set(summary.id, summary);
    }

    summary.status = mergeJobStatus(summary.status, step.status);
    if (step.durationMs != null && step.durationMs > 0) {
      summary.durationMs = (summary.durationMs ?? 0) + step.durationMs;
    }
  }

  return summaries;
}

function mergeJobStatus(current: JobRunStatus, stepStatus: StepRunStatus): JobRunStatus {
  if (current === 'failed' || stepStatus === 'failed') {
    return 'failed';
  }
  if (current === 'running' || stepStatus === 'running') {
    return 'running';
  }
  if (current === 'pending') {
    return stepStatus;
  }
  if (current === 'skipped' && stepStatus === 'passed') {
    return 'passed';
  }
  if (current === 'passed' && stepStatus === 'skipped') {
    return 'passed';
  }
  return stepStatus;
}

function matchesJob(job: JobSummary, label: string): boolean {
  return job.id === label || job.name === label;
}

function stepKey(job: string, step: string): string {
  return `${job}\0${step}`;
}
