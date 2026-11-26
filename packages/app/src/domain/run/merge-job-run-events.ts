import type { JobAliasMap } from './step-identity';
import type { JobGraphEvent, RunnerEvent } from '@/domain/runner-events';
import {
  buildJobAliasMap,
  EMPTY_JOB_ALIASES,

  sameJobLabel,
} from './step-identity';

export function prepareJobRerun(events: RunnerEvent[], jobId: string): RunnerEvent[] {
  const graph = richestJobGraph(events);
  const aliases = graph
    ? buildJobAliasMap(graph.jobs)
    : EMPTY_JOB_ALIASES;

  const kept: RunnerEvent[] = [];
  let keptGraph = false;

  for (const event of events) {
    if (event.type === 'run_started' || event.type === 'run_finished') {
      continue;
    }
    if (event.type === 'job_graph') {
      if (keptGraph || !graph) {
        continue;
      }
      kept.push(graph);
      keptGraph = true;
      continue;
    }
    if (isJobScopedEvent(event) && eventBelongsToJob(event.job, jobId, aliases)) {
      continue;
    }
    kept.push(event);
  }

  if (graph && !keptGraph) {
    kept.unshift(graph);
  }

  return kept;
}

export function shouldAcceptJobGraph(
  events: RunnerEvent[],
  incoming: JobGraphEvent,
): boolean {
  const current = richestJobGraph(events);
  if (!current || current.jobs.length === 0) {
    return true;
  }
  return incoming.jobs.length >= current.jobs.length;
}

function richestJobGraph(events: RunnerEvent[]): JobGraphEvent | null {
  let best: JobGraphEvent | null = null;
  for (const event of events) {
    if (event.type !== 'job_graph') {
      continue;
    }
    if (!best || event.jobs.length > best.jobs.length) {
      best = event;
    }
  }
  return best;
}

function isJobScopedEvent(
  event: RunnerEvent,
): event is RunnerEvent & { job: string } {
  return (
    event.type === 'step_started'
    || event.type === 'step_finished'
    || (event.type === 'log' && event.job != null && event.job !== '')
  );
}

function eventBelongsToJob(
  eventJob: string,
  targetJobId: string,
  aliases: JobAliasMap,
): boolean {
  return sameJobLabel(eventJob, targetJobId, aliases);
}
