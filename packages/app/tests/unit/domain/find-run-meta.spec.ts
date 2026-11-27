import type { RunnerEvent } from '@/domain/runner-events';
import { describe, expect, it } from 'vitest';
import {
  findLastArtifacts,
  findLastRunFinished,
  findRunErrors,
} from '@/domain/run';

describe('find-run-meta', () => {
  it('returns the last run_finished and artifacts events', () => {
    const events: RunnerEvent[] = [
      { type: 'run_finished', status: 'failed', exitCode: 1 },
      {
        type: 'artifacts',
        artifacts: [
          { name: 'a', path: '/tmp/a', size: 1 },
        ],
      },
      { type: 'run_finished', status: 'passed', exitCode: 0 },
      {
        type: 'artifacts',
        artifacts: [
          { name: 'b', path: '/tmp/b', size: 2 },
        ],
      },
      { type: 'error', message: 'boom' },
    ];

    expect(findLastRunFinished(events)).toEqual({
      type: 'run_finished',
      status: 'passed',
      exitCode: 0,
    });
    expect(findLastArtifacts(events)).toEqual([
      { name: 'b', path: '/tmp/b', size: 2 },
    ]);
    expect(findRunErrors(events)).toEqual([
      { type: 'error', message: 'boom' },
    ]);
  });
});
