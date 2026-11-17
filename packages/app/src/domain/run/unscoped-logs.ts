import type { RunnerEvent } from '@/domain/runner-events';

export function unscopedLogLines(events: RunnerEvent[]): string[] {
  const lines: string[] = [];
  for (const event of events) {
    if (event.type !== 'log') {
      continue;
    }
    if (event.job || event.step) {
      continue;
    }
    lines.push(event.line);
  }
  return lines;
}
