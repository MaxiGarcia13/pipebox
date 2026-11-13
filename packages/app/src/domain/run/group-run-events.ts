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
import { seededStepKeys } from './step-identity';
import { summarizeJobs } from './summarize-jobs';

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
