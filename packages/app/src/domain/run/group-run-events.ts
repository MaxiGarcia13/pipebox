import type { GroupedRun } from './types';
import type {
  JobNode,
  RunnerEvent,
  RunStatus,
} from '@/domain/runner-events';
import { foldSteps } from './fold-steps';
import {
  advancePendingSteps,
  cascadeAfterFailure,
  closeExpandedParents,
  demoteRunningBeforeLaterProgress,
  settleAfterRunFinished,
} from './settle-steps';
import { buildJobAliasMap, seededStepKeys } from './step-identity';
import { summarizeJobs } from './summarize-jobs';

export function groupRunEvents(events: RunnerEvent[]): GroupedRun {
  const graphJobs = jobsFromGraph(events);
  const aliases = buildJobAliasMap(graphJobs);
  const seeded = seededStepKeys(graphJobs, aliases);
  const runStatus = runFinishedStatus(events);
  const runFinished = runStatus != null;
  const steps = foldSteps(events, graphJobs, seeded, aliases);
  cascadeAfterFailure(steps, aliases);
  demoteRunningBeforeLaterProgress(steps, seeded, aliases);
  closeExpandedParents(steps, seeded, aliases);
  advancePendingSteps(steps, runFinished, aliases);
  closeExpandedParents(steps, seeded, aliases);
  settleAfterRunFinished(steps, runStatus);
  const jobs = summarizeJobs(graphJobs, steps, runFinished, aliases);
  return { jobs, steps };
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

function runFinishedStatus(events: RunnerEvent[]): RunStatus | null {
  for (let index = events.length - 1; index >= 0; index -= 1) {
    const event = events[index];
    if (event?.type === 'run_finished') {
      return event.status;
    }
  }
  return null;
}
