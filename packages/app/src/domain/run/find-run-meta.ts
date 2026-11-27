import type {
  Artifact,
  ErrorEvent,
  RunFinishedEvent,
  RunnerEvent,
} from '@/domain/runner-events';

export function findLastRunFinished(
  events: RunnerEvent[],
): RunFinishedEvent | null {
  for (let index = events.length - 1; index >= 0; index -= 1) {
    const event = events[index];
    if (event?.type === 'run_finished') {
      return event;
    }
  }
  return null;
}

export function findLastArtifacts(events: RunnerEvent[]): Artifact[] {
  for (let index = events.length - 1; index >= 0; index -= 1) {
    const event = events[index];
    if (event?.type === 'artifacts') {
      return event.artifacts;
    }
  }
  return [];
}

export function findRunErrors(events: RunnerEvent[]): ErrorEvent[] {
  return events.filter((event): event is ErrorEvent => event.type === 'error');
}
