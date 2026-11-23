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
    expect(grouped.jobs[0]?.durationMs).toBe(10);
    expect(grouped.steps).toHaveLength(1);
    expect(grouped.steps[0]?.status).toBe('passed');
    expect(grouped.steps[0]?.durationMs).toBe(10);
  });

  it('times a job from the first step start through now, not the sum of steps', () => {
    const grouped = groupRunEvents([
      {
        type: 'job_graph',
        jobs: [
          {
            id: 'e2e',
            name: 'e2e',
            needs: [],
            steps: ['install', 'test'],
          },
        ],
      },
      {
        type: 'step_started',
        job: 'e2e',
        step: 'install',
        receivedAt: 1_000,
      },
      {
        type: 'step_finished',
        job: 'e2e',
        step: 'install',
        status: 'passed',
        durationMs: 10,
        receivedAt: 5_000,
      },
      {
        type: 'step_started',
        job: 'e2e',
        step: 'test',
        receivedAt: 5_000,
      },
    ]);

    expect(grouped.steps[0]).toMatchObject({
      startedAt: 1_000,
      finishedAt: 5_000,
    });
    expect(grouped.steps[1]).toMatchObject({
      status: 'running',
      startedAt: 5_000,
    });
    expect(grouped.steps[1]?.finishedAt).toBeUndefined();
    expect(grouped.jobs[0]).toMatchObject({
      status: 'running',
      startedAt: 1_000,
    });
    expect(grouped.jobs[0]?.finishedAt).toBeUndefined();
    expect(grouped.jobs[0]?.durationMs).toBeUndefined();
  });

  it('freezes a finished job at the last step receipt time', () => {
    const grouped = groupRunEvents([
      {
        type: 'step_started',
        job: 'lint',
        step: 'echo ok',
        receivedAt: 2_000,
      },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'echo ok',
        status: 'passed',
        durationMs: 12,
        receivedAt: 9_000,
      },
      {
        type: 'run_finished',
        status: 'passed',
        exitCode: 0,
        receivedAt: 9_100,
      },
    ]);

    expect(grouped.jobs[0]).toMatchObject({
      startedAt: 2_000,
      finishedAt: 9_000,
    });
    expect(grouped.jobs[0]?.durationMs).toBeUndefined();
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

  it('keeps a finished step passed when a later log arrives', () => {
    const grouped = groupRunEvents([
      {
        type: 'job_graph',
        jobs: [
          {
            id: 'lint',
            name: 'Lint',
            needs: [],
            steps: ['sh'],
          },
        ],
      },
      {
        type: 'step_started',
        job: 'lint',
        step: 'sh',
      },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'sh',
        status: 'passed',
        durationMs: 40,
      },
      {
        type: 'log',
        job: 'lint',
        step: 'sh',
        line: 'late line',
      },
    ]);

    expect(grouped.steps[0]?.status).toBe('passed');
    expect(grouped.steps[0]?.durationMs).toBe(40);
    expect(grouped.steps[0]?.lines).toEqual(['late line']);
  });
});
