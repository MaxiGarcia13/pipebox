import type { StepGroup, StepRunStatus } from './types';
import type { RunStatus } from '@/domain/runner-events';
import {
  isJobBookend,
  sameJobLabel,
  stepIdentity,
} from './step-identity';

export function cascadeAfterFailure(steps: StepGroup[]): void {
  const jobLabels = [
    ...new Set(steps.map((step) => step.job)),
  ];
  for (const job of jobLabels) {
    const jobSteps = steps.filter((step) => sameJobLabel(step.job, job));
    let failed = false;
    for (const group of jobSteps) {
      if (failed) {
        // act still emits Success for "Complete job"; don't show green after a failure.
        group.status = 'skipped';
        group.openDefault = false;
        delete group.durationMs;
        continue;
      }
      if (group.status === 'failed') {
        failed = true;
      }
    }
  }
}

export function demoteRunningBeforeLaterProgress(
  steps: StepGroup[],
  seeded: Set<string>,
): void {
  const jobLabels = [
    ...new Set(steps.map((step) => step.job)),
  ];
  for (const job of jobLabels) {
    const jobSteps = steps.filter((step) => sameJobLabel(step.job, job));
    for (let index = 0; index < jobSteps.length; index += 1) {
      const group = jobSteps[index]!;
      if (group.status !== 'running') {
        continue;
      }
      const later = jobSteps.slice(index + 1);
      const laterProgress = later.some((step) => hasProgressed(step.status));
      if (!laterProgress) {
        continue;
      }
      // Seeded parent still owns the composite; expanded children after it are not "next".
      if (
        seeded.has(stepIdentity(group.job, group.step))
        && !later.some((step) =>
          hasProgressed(step.status)
          && (
            seeded.has(stepIdentity(step.job, step.step))
            || isJobBookend(step.step)
          ),
        )
      ) {
        continue;
      }
      group.status = 'passed';
      group.openDefault = false;
    }
  }
}

// act expands composites into child steps inserted before the YAML parent;
// the parent never gets step_finished and must not stay pending/running.
export function closeExpandedParents(
  steps: StepGroup[],
  seeded: Set<string>,
): void {
  if (seeded.size === 0) {
    return;
  }

  const jobLabels = [
    ...new Set(steps.map((step) => step.job)),
  ];
  for (const job of jobLabels) {
    const jobSteps = steps.filter((step) => sameJobLabel(step.job, job));
    // While a step is live, don't mark later empty parents/siblings passed.
    if (jobSteps.some((step) => step.status === 'running')) {
      continue;
    }
    for (let index = 0; index < jobSteps.length; index += 1) {
      const group = jobSteps[index]!;
      if (group.status !== 'pending' && group.status !== 'running') {
        continue;
      }
      if (group.lines.length > 0 || group.durationMs != null) {
        continue;
      }
      if (!seeded.has(stepIdentity(group.job, group.step))) {
        continue;
      }
      if (!hasExpandedStepsBefore(jobSteps, index, seeded)) {
        continue;
      }
      group.status = 'passed';
      group.openDefault = false;
    }
  }
}

export function advancePendingSteps(steps: StepGroup[], runFinished: boolean): void {
  if (runFinished) {
    return;
  }

  const jobLabels = [
    ...new Set(steps.map((step) => step.job)),
  ];
  for (const job of jobLabels) {
    const jobSteps = steps.filter((step) => sameJobLabel(step.job, job));
    const runningSteps = jobSteps.filter((step) => step.status === 'running');
    if (runningSteps.length > 1) {
      // Keep the last running step; demote empty optimistic duplicates.
      for (const group of runningSteps.slice(0, -1)) {
        if (group.lines.length === 0 && group.durationMs == null) {
          group.status = 'pending';
          group.openDefault = false;
        }
      }
    }
    if (jobSteps.some((step) => step.status === 'running' || step.status === 'failed')) {
      continue;
    }

    let lastFinished = -1;
    for (let index = 0; index < jobSteps.length; index += 1) {
      const status = jobSteps[index]?.status;
      if (status === 'passed' || status === 'skipped') {
        lastFinished = index;
      }
    }
    if (lastFinished < 0) {
      continue;
    }

    const next = jobSteps.slice(lastFinished + 1).find((step) => step.status === 'pending');
    if (!next) {
      continue;
    }
    next.status = 'running';
    next.openDefault = true;
  }
}

export function settleAfterRunFinished(
  steps: StepGroup[],
  runStatus: RunStatus | null,
): void {
  if (runStatus == null) {
    return;
  }

  const jobsCompletedOk = new Set(
    steps
      .filter((step) =>
        isJobBookend(step.step)
        && /complete job/i.test(step.step)
        && step.status === 'passed',
      )
      .map((step) => step.job),
  );

  for (const group of steps) {
    if (group.status === 'running') {
      if (runStatus === 'passed') {
        group.status = 'passed';
        group.openDefault = false;
      } else if (runStatus === 'cancelled') {
        group.status = 'cancelled';
        group.openDefault = false;
      } else if (jobsCompletedOk.has(group.job)) {
        // Another job failed the overall run; keep this job green.
        group.status = 'passed';
        group.openDefault = false;
      } else {
        group.status = 'failed';
        group.openDefault = true;
      }
      continue;
    }
    if (group.status === 'pending') {
      group.status = 'skipped';
      group.openDefault = false;
    }
  }
}

function hasProgressed(status: StepRunStatus): boolean {
  return status === 'passed'
    || status === 'failed'
    || status === 'running'
    || status === 'skipped'
    || status === 'cancelled';
}

function hasExpandedStepsBefore(
  jobSteps: StepGroup[],
  index: number,
  seeded: Set<string>,
): boolean {
  let expanded = 0;
  for (let cursor = index - 1; cursor >= 0; cursor -= 1) {
    const previous = jobSteps[cursor]!;
    if (previous.status !== 'passed' && previous.status !== 'skipped') {
      break;
    }
    if (isJobBookend(previous.step) || seeded.has(stepIdentity(previous.job, previous.step))) {
      break;
    }
    expanded += 1;
  }
  return expanded > 0;
}
