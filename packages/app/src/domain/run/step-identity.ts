import type { JobNode } from '@/domain/runner-events';

export type JobAliasMap = Map<string, string>;

export const EMPTY_JOB_ALIASES: JobAliasMap = new Map();

export function buildJobAliasMap(graphJobs: JobNode[]): JobAliasMap {
  const aliases: JobAliasMap = new Map();
  for (const job of graphJobs) {
    aliases.set(job.id, job.id);
    if (job.name !== '' && job.name !== job.id) {
      aliases.set(job.name, job.id);
    }
  }
  return aliases;
}

export function canonicalJobId(label: string, aliases: JobAliasMap): string {
  return aliases.get(label) ?? label;
}

export function sameJobLabel(
  left: string,
  right: string,
  aliases: JobAliasMap = EMPTY_JOB_ALIASES,
): boolean {
  if (left === right) {
    return true;
  }
  if (aliases.size === 0) {
    return false;
  }
  return canonicalJobId(left, aliases) === canonicalJobId(right, aliases);
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

export function seededStepKeys(
  graphJobs: JobNode[],
  aliases: JobAliasMap = EMPTY_JOB_ALIASES,
): Set<string> {
  const keys = new Set<string>();
  for (const job of graphJobs) {
    const canonical = canonicalJobId(job.id, aliases);
    for (const step of job.steps ?? []) {
      keys.add(stepIdentity(canonical, step));
      keys.add(stepIdentity(job.id, step));
      keys.add(stepIdentity(job.name, step));
    }
  }
  return keys;
}
