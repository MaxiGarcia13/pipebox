import type { JobAliasMap } from './step-identity';
import type { JobRunStatus, JobSummary, StepGroup, StepRunStatus } from './types';
import type { JobNode } from '@/domain/runner-events';
import {
  buildJobAliasMap,
  canonicalJobId,
  EMPTY_JOB_ALIASES,

  sameJobLabel,
} from './step-identity';

export function summarizeJobs(
  graphJobs: JobNode[],
  steps: StepGroup[],
  runFinished: boolean,
  aliases: JobAliasMap = buildJobAliasMap(graphJobs),
): JobSummary[] {
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
    byMatch.set(canonicalJobId(summary.id, aliases), summary);
  }

  for (const step of steps) {
    let summary = byMatch.get(step.job) ?? byMatch.get(canonicalJobId(step.job, aliases));
    if (!summary) {
      // Graph present: skip unmatched labels (avoids duplicate root-column pills).
      if (graphJobs.length > 0) {
        continue;
      }
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
  }

  // Without job_finished, keep the job "running" only while a step is running
  // (pending peers stay pending; finished jobs stay passed).
  if (!runFinished) {
    for (const summary of summaries) {
      if (
        summary.status === 'pending'
        || summary.status === 'failed'
        || summary.status === 'cancelled'
      ) {
        continue;
      }
      const jobSteps = steps.filter((step) => matchesJob(summary, step.job, aliases));
      if (jobSteps.some((step) => step.status === 'running')) {
        summary.status = 'running';
      }
    }
  }

  for (const summary of summaries) {
    applyJobTiming(summary, steps, aliases);
  }

  return summaries;
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

function applyJobTiming(
  summary: JobSummary,
  steps: StepGroup[],
  aliases: JobAliasMap,
): void {
  let startedAt: number | undefined;
  let finishedAt: number | undefined;
  let live = false;
  let actMs = 0;

  for (const step of steps) {
    if (!matchesJob(summary, step.job, aliases)) {
      continue;
    }
    if (step.startedAt != null) {
      startedAt = startedAt == null
        ? step.startedAt
        : Math.min(startedAt, step.startedAt);
    }
    if (step.status === 'running') {
      live = true;
    }
    if (step.finishedAt != null) {
      finishedAt = finishedAt == null
        ? step.finishedAt
        : Math.max(finishedAt, step.finishedAt);
    }
    if (step.durationMs != null && step.durationMs > 0) {
      actMs += step.durationMs;
    }
  }

  if (startedAt != null) {
    summary.startedAt = startedAt;
    if (!live && finishedAt != null) {
      summary.finishedAt = finishedAt;
    }
    return;
  }

  if (actMs > 0) {
    summary.durationMs = actMs;
  }
}

function matchesJob(
  job: JobSummary,
  label: string,
  aliases: JobAliasMap = EMPTY_JOB_ALIASES,
): boolean {
  return sameJobLabel(job.id, label, aliases) || sameJobLabel(job.name, label, aliases);
}

function mergeJobStatus(current: JobRunStatus, stepStatus: StepRunStatus): JobRunStatus {
  if (stepStatus === 'pending') {
    return current;
  }
  if (current === 'failed' || stepStatus === 'failed') {
    return 'failed';
  }
  if (current === 'cancelled' || stepStatus === 'cancelled') {
    return 'cancelled';
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
