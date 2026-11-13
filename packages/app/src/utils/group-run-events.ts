import type {
  JobNode,
  RunnerEvent,
  RunStatus,
  StepStatus,
} from '@/shared/runner-events';

export type JobRunStatus = 'pending' | 'running' | StepStatus;
export type StepRunStatus = 'pending' | 'running' | StepStatus;

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
  const seeded = seededStepKeys(graphJobs);
  const runStatus = runFinishedStatus(events);
  const runFinished = runStatus != null;
  const steps = foldSteps(events, graphJobs, seeded);
  cascadeAfterFailure(steps);
  demoteRunningBeforeLaterProgress(steps, seeded);
  closeExpandedParents(steps, seeded);
  advancePendingSteps(steps, runFinished);
  closeExpandedParents(steps, seeded);
  settleAfterRunFinished(steps, runStatus);
  const jobs = summarizeJobs(graphJobs, steps, runFinished);
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

function foldSteps(
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

function cascadeAfterFailure(steps: StepGroup[]): void {
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

function demoteRunningBeforeLaterProgress(
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
function closeExpandedParents(
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

function isExpandedStep(job: string, step: string, seeded: Set<string>): boolean {
  if (isJobBookend(step)) {
    return false;
  }
  return !seeded.has(stepIdentity(job, step));
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

function seededStepKeys(graphJobs: JobNode[]): Set<string> {
  const keys = new Set<string>();
  for (const job of graphJobs) {
    for (const step of job.steps ?? []) {
      keys.add(stepIdentity(job.name, step));
      keys.add(stepIdentity(job.id, step));
    }
  }
  return keys;
}

function stepIdentity(job: string, step: string): string {
  return `${job}\0${normalizeStepName(step)}`;
}

function isJobBookend(step: string): boolean {
  const lower = step.toLowerCase();
  return lower.includes('set up job')
    || lower.includes('setup job')
    || lower.includes('complete job');
}

function advancePendingSteps(steps: StepGroup[], runFinished: boolean): void {
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

function settleAfterRunFinished(
  steps: StepGroup[],
  runStatus: RunStatus | null,
): void {
  if (runStatus == null) {
    return;
  }

  for (const group of steps) {
    if (group.status === 'running') {
      if (runStatus === 'passed') {
        group.status = 'passed';
        group.openDefault = false;
      } else if (runStatus === 'cancelled') {
        group.status = 'cancelled';
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

function runFinishedStatus(events: RunnerEvent[]): RunStatus | null {
  for (let index = events.length - 1; index >= 0; index -= 1) {
    const event = events[index];
    if (event?.type === 'run_finished') {
      return event.status;
    }
  }
  return null;
}

function summarizeJobs(
  graphJobs: JobNode[],
  steps: StepGroup[],
  runFinished: boolean,
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
  }

  for (const step of steps) {
    let summary = byMatch.get(step.job);
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
    if (step.durationMs != null && step.durationMs > 0) {
      summary.durationMs = (summary.durationMs ?? 0) + step.durationMs;
    }
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
      const jobSteps = steps.filter((step) => matchesJob(summary, step.job));
      if (jobSteps.some((step) => step.status === 'running')) {
        summary.status = 'running';
      }
    }
  }

  return summaries;
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

function matchesJob(job: JobSummary, label: string): boolean {
  return sameJobLabel(job.id, label) || sameJobLabel(job.name, label);
}

function sameJobLabel(left: string, right: string): boolean {
  return left === right;
}

function normalizeStepName(name: string): string {
  return name.replace(/^Main\s+/i, '').trim().toLowerCase();
}

function stepKey(job: string, step: string): string {
  return `${job}\0${step}`;
}
