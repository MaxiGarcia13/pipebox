import { describe, expect, it } from 'vitest';
import { chronologicalLogLines } from '@/domain/run';

describe('chronologicalLogLines', () => {
  it('keeps log and error lines in event order', () => {
    const lines = chronologicalLogLines([
      { type: 'run_started', repo: '/r', workflow: 'ci.yml' },
      {
        type: 'log',
        job: 'lint',
        step: 'Install',
        line: 'npm ci',
      },
      { type: 'error', message: 'boom' },
      { type: 'log', line: 'unscoped' },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'Install',
        status: 'passed',
      },
    ]);

    expect(lines).toEqual([
      { text: 'npm ci', kind: 'log' },
      { text: 'boom', kind: 'error' },
      { text: 'unscoped', kind: 'log' },
    ]);
  });
});
