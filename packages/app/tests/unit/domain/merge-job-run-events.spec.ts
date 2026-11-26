import type { RunnerEvent } from '@/domain/runner-events';
import { describe, expect, it } from 'vitest';
import {
  prepareJobRerun,
  shouldAcceptJobGraph,
} from '@/domain/run/merge-job-run-events';

const fullGraph: RunnerEvent = {
  type: 'job_graph',
  jobs: [
    {
      id: 'lint',
      name: 'Lint',
      needs: [],
      steps: ['echo lint'],
    },
    {
      id: 'test',
      name: 'Test',
      needs: [],
      steps: ['echo test'],
    },
    {
      id: 'e2e',
      name: 'E2E',
      needs: ['lint', 'test'],
      steps: ['echo e2e'],
    },
  ],
};

function parallelPassedRun(): RunnerEvent[] {
  return [
    {
      type: 'run_started',
      repo: '/repo',
      workflow: '.github/workflows/ci.yml',
    },
    fullGraph,
    { type: 'step_started', job: 'lint', step: 'echo lint' },
    {
      type: 'log',
      job: 'lint',
      step: 'echo lint',
      line: 'lint ok',
    },
    {
      type: 'step_finished',
      job: 'lint',
      step: 'echo lint',
      status: 'passed',
      durationMs: 5,
    },
    { type: 'step_started', job: 'test', step: 'echo test' },
    {
      type: 'log',
      job: 'test',
      step: 'echo test',
      line: 'test ok',
    },
    {
      type: 'step_finished',
      job: 'test',
      step: 'echo test',
      status: 'passed',
      durationMs: 7,
    },
    { type: 'step_started', job: 'e2e', step: 'echo e2e' },
    {
      type: 'log',
      job: 'e2e',
      step: 'echo e2e',
      line: 'e2e fail',
    },
    {
      type: 'step_finished',
      job: 'e2e',
      step: 'echo e2e',
      status: 'failed',
      durationMs: 9,
    },
    {
      type: 'artifacts',
      artifacts: [
        { name: 'report', path: '/tmp/report', size: 10 },
      ],
    },
    { type: 'log', line: 'secrets: .env' },
    {
      type: 'error',
      message: 'setup note',
    },
    {
      type: 'run_finished',
      status: 'failed',
      exitCode: 1,
    },
  ];
}

describe('prepareJobRerun', () => {
  it('keeps sibling job events and strips the target job', () => {
    const next = prepareJobRerun(parallelPassedRun(), 'e2e');

    expect(next.some((event) => event.type === 'run_started')).toBe(false);
    expect(next.some((event) => event.type === 'run_finished')).toBe(false);

    const graph = next.find((event) => event.type === 'job_graph');
    expect(graph?.type === 'job_graph' && graph.jobs).toHaveLength(3);

    expect(
      next.filter(
        (event) =>
          (event.type === 'step_started'
            || event.type === 'step_finished'
            || event.type === 'log')
          && 'job' in event
          && event.job === 'e2e',
      ),
    ).toHaveLength(0);

    expect(
      next.some(
        (event) => event.type === 'step_finished' && event.job === 'lint',
      ),
    ).toBe(true);
    expect(
      next.some(
        (event) => event.type === 'step_finished' && event.job === 'test',
      ),
    ).toBe(true);
    expect(next.some((event) => event.type === 'artifacts')).toBe(true);
    expect(
      next.some((event) => event.type === 'log' && event.line === 'secrets: .env'),
    ).toBe(true);
    expect(next.some((event) => event.type === 'error')).toBe(true);
  });

  it('keeps the richest job_graph when a filtered graph is present', () => {
    const events: RunnerEvent[] = [
      fullGraph,
      {
        type: 'job_graph',
        jobs: [
          {
            id: 'e2e',
            name: 'E2E',
            needs: [],
            steps: ['echo e2e'],
          },
        ],
      },
      {
        type: 'step_finished',
        job: 'lint',
        step: 'echo lint',
        status: 'passed',
      },
      {
        type: 'step_finished',
        job: 'e2e',
        step: 'echo e2e',
        status: 'failed',
      },
    ];

    const next = prepareJobRerun(events, 'e2e');
    const graphs = next.filter((event) => event.type === 'job_graph');
    expect(graphs).toHaveLength(1);
    expect(graphs[0]?.type === 'job_graph' && graphs[0].jobs).toHaveLength(3);
  });
});

describe('shouldAcceptJobGraph', () => {
  it('accepts the first graph', () => {
    expect(
      shouldAcceptJobGraph([], {
        type: 'job_graph',
        jobs: [
          { id: 'lint', name: 'Lint', needs: [] },
        ],
      }),
    ).toBe(true);
  });

  it('rejects a filtered graph that shrinks the timeline', () => {
    expect(
      shouldAcceptJobGraph([fullGraph], {
        type: 'job_graph',
        jobs: [
          { id: 'e2e', name: 'E2E', needs: [] },
        ],
      }),
    ).toBe(false);
  });

  it('accepts a graph with at least as many jobs', () => {
    expect(
      shouldAcceptJobGraph(
        [
          {
            type: 'job_graph',
            jobs: [
              { id: 'lint', name: 'Lint', needs: [] },
            ],
          },
        ],
        fullGraph,
      ),
    ).toBe(true);
  });
});
