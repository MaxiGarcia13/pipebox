import type { StepGroup } from './types';
import type { JobNode, RunnerEvent } from '@/domain/runner-events';
import {
  isJobBookend,
  normalizeStepName,
  sameJobLabel,
  stepIdentity,
  stepKey,
} from './step-identity';

export function foldSteps(
  events: RunnerEvent[],
  graphJobs: JobNode[],
  seeded: Set<string>,
): StepGroup[] {
  const groups = seedStepsFromGraph(graphJobs);
  const byKey = new Map(groups.map((group) => [group.key, group]));

  for (const event of events) {
    if (event.type === 'step_started') {
      demoteOtherRunningSteps(groups, event.job, event.step, seeded);
      const group = findOrInsertStep(groups, byKey, event.job, event.step);
      group.status = 'running';
      group.openDefault = true;
      delete group.durationMs;
      continue;
    }

    if (event.type === 'step_finished') {
      const group = findOrInsertStep(groups, byKey, event.job, event.step);
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
      const group = findOrInsertStep(groups, byKey, job, step || '(log)');
      group.lines.push(event.line);
      // Late/out-of-order act output after a premature pass must reopen the step.
      if (group.status === 'pending' || group.status === 'passed') {
        group.status = 'running';
        group.openDefault = true;
        delete group.durationMs;
      }
    }
  }

  return groups;
}

function demoteOtherRunningSteps(
  steps: StepGroup[],
  job: string,
  activeStep: string,
  seeded: Set<string>,
): void {
  const normalizedActive = normalizeStepName(activeStep);
  const activeIsExpanded = isExpandedStep(job, activeStep, seeded);
  for (const group of steps) {
    if (!sameJobLabel(group.job, job) || group.status !== 'running') {
      continue;
    }
    if (normalizeStepName(group.step) === normalizedActive) {
      continue;
    }
    // Nested act steps under a YAML parent must not close that parent.
    if (activeIsExpanded && seeded.has(stepIdentity(group.job, group.step))) {
      continue;
    }

    group.status = 'passed';
    group.openDefault = false;
  }
}

function isExpandedStep(job: string, step: string, seeded: Set<string>): boolean {
  if (isJobBookend(step)) {
    return false;
  }
  return !seeded.has(stepIdentity(job, step));
}

function seedStepsFromGraph(graphJobs: JobNode[]): StepGroup[] {
  const groups: StepGroup[] = [];
  for (const job of graphJobs) {
    for (const step of job.steps ?? []) {
      groups.push({
        key: stepKey(job.name, step),
        job: job.name,
        step,
        status: 'pending',
        lines: [],
        openDefault: false,
      });
    }
  }
  return groups;
}

function findOrInsertStep(
  groups: StepGroup[],
  byKey: Map<string, StepGroup>,
  job: string,
  step: string,
): StepGroup {
  const existing = findStepGroup(groups, job, step);
  if (existing) {
    return existing;
  }

  const key = stepKey(job, step);
  const group: StepGroup = {
    key,
    job,
    step,
    status: 'pending',
    lines: [],
    openDefault: false,
  };
  byKey.set(key, group);
  insertStepGroup(groups, group);
  return group;
}

function findStepGroup(
  groups: StepGroup[],
  job: string,
  step: string,
): StepGroup | undefined {
  const normalizedStep = normalizeStepName(step);
  return groups.find((group) =>
    sameJobLabel(group.job, job)
    && normalizeStepName(group.step) === normalizedStep,
  );
}

function insertStepGroup(groups: StepGroup[], group: StepGroup): void {
  const jobIndexes = indexesForJob(groups, group.job);
  if (jobIndexes.length === 0) {
    groups.push(group);
    return;
  }

  const lower = group.step.toLowerCase();
  if (lower.includes('set up job') || lower.includes('setup job')) {
    groups.splice(jobIndexes[0]!, 0, group);
    return;
  }
  if (lower.includes('complete job')) {
    groups.splice(jobIndexes[jobIndexes.length - 1]! + 1, 0, group);
    return;
  }

  const firstPending = jobIndexes.find((index) => groups[index]?.status === 'pending');
  if (firstPending != null) {
    groups.splice(firstPending, 0, group);
    return;
  }
  groups.splice(jobIndexes[jobIndexes.length - 1]! + 1, 0, group);
}

function indexesForJob(groups: StepGroup[], job: string): number[] {
  const indexes: number[] = [];
  for (let index = 0; index < groups.length; index += 1) {
    if (sameJobLabel(groups[index]!.job, job)) {
      indexes.push(index);
    }
  }
  return indexes;
}
