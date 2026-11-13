import type { JobNode } from '@/domain/runner-events';

export function sameJobLabel(left: string, right: string): boolean {
  return left === right;
}

export function normalizeStepName(name: string): string {
  return name.replace(/^Main\s+/i, '').trim().toLowerCase();
}

export function stepKey(job: string, step: string): string {
  return `${job}\0${step}`;
}

export function stepIdentity(job: string, step: string): string {
  return `${job}\0${normalizeStepName(step)}`;
}

export function isJobBookend(step: string): boolean {
  const lower = step.toLowerCase();
  return lower.includes('set up job')
    || lower.includes('setup job')
    || lower.includes('complete job');
}

export function seededStepKeys(graphJobs: JobNode[]): Set<string> {
  const keys = new Set<string>();
  for (const job of graphJobs) {
    for (const step of job.steps ?? []) {
      keys.add(stepIdentity(job.name, step));
      keys.add(stepIdentity(job.id, step));
    }
  }
  return keys;
}
