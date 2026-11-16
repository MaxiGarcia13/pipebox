import { describe, expect, it } from 'vitest';
import { groupRunEvents } from '@/domain/run/group-run-events';

describe('groupRunEvents', () => {
  it('returns empty groups for no events', () => {
    expect(groupRunEvents([])).toEqual({ jobs: [], steps: [] });
  });

  it('seeds steps from job_graph and marks finished steps', () => {
    const grouped = groupRunEvents([
      {
        type: 'job_graph',
        jobs: [
          {
            id: 'lint',
            name: 'Lint',
            needs: [],
            steps: ['echo ok'],
          },
        ],
      },
      {
        type: 'step_started',
        job: 'lint',
        step: 'echo ok',
      },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'echo ok',
        status: 'passed',
        durationMs: 10,
      },
      {
        type: 'run_finished',
        status: 'passed',
        exitCode: 0,
      },
    ]);

    expect(grouped.jobs).toHaveLength(1);
    expect(grouped.jobs[0]?.status).toBe('passed');
    expect(grouped.steps).toHaveLength(1);
    expect(grouped.steps[0]?.status).toBe('passed');
    expect(grouped.steps[0]?.durationMs).toBe(10);
  });

  it('marks remaining steps failed after run failure', () => {
    const grouped = groupRunEvents([
      {
        type: 'job_graph',
        jobs: [
          {
            id: 'lint',
            name: 'Lint',
            needs: [],
            steps: ['one', 'two'],
          },
        ],
      },
      {
        type: 'step_started',
        job: 'lint',
        step: 'one',
      },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'one',
        status: 'failed',
      },
      {
        type: 'run_finished',
        status: 'failed',
        exitCode: 1,
      },
    ]);

    expect(grouped.steps.map((step) => step.status)).toEqual(
      expect.arrayContaining(['failed']),
    );
    expect(grouped.jobs[0]?.status).toBe('failed');
  });
});
