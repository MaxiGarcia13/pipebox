import { describe, expect, it } from 'vitest';
import { unscopedLogLines } from '@/domain/run/unscoped-logs';

describe('unscopedLogLines', () => {
  it('keeps logs without job or step', () => {
    const lines = unscopedLogLines([
      { type: 'log', line: 'building pipebox-act:latest (first run)' },
      { type: 'log', line: '#0 building with desktop-linux' },
      {
        type: 'log',
        job: 'lint',
        step: 'echo',
        line: 'hello',
      },
      { type: 'error', message: 'boom' },
    ]);
    expect(lines).toEqual([
      'building pipebox-act:latest (first run)',
      '#0 building with desktop-linux',
    ]);
  });

  it('returns empty when every log is scoped', () => {
    expect(unscopedLogLines([
      {
        type: 'log',
        job: 'lint',
        step: 'echo',
        line: 'hello',
      },
    ])).toEqual([]);
  });
});
