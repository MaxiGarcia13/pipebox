import type { RunnerEvent } from '@/domain/runner-events';

export type ChronologicalLogKind = 'log' | 'error';

export interface ChronologicalLogLine {
  text: string;
  kind: ChronologicalLogKind;
}

export function chronologicalLogLines(events: RunnerEvent[]): ChronologicalLogLine[] {
  const lines: ChronologicalLogLine[] = [];
  for (const event of events) {
    if (event.type === 'log') {
      lines.push({ text: event.line, kind: 'log' });
      continue;
    }
    if (event.type === 'error') {
      lines.push({ text: event.message, kind: 'error' });
    }
  }
  return lines;
}
