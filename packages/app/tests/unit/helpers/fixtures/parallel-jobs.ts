import type { JobGraphEvent, RunnerEvent } from '@/domain/runner-events';
import { FIXTURE_REPO_GITHUB } from './repos';

export const parallelGraph: JobGraphEvent = {
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

export function parallelRunEvents(
  workflowPath: string,
  e2eStatus: 'passed' | 'failed' = 'failed',
): RunnerEvent[] {
  return [
    {
      type: 'run_started',
      repo: FIXTURE_REPO_GITHUB,
      workflow: workflowPath,
    },
    parallelGraph,
    {
      type: 'step_started',
      job: 'lint',
      step: 'echo lint',
    },
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
    {
      type: 'step_started',
      job: 'test',
      step: 'echo test',
    },
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
    {
      type: 'step_started',
      job: 'e2e',
      step: 'echo e2e',
    },
    {
      type: 'log',
      job: 'e2e',
      step: 'echo e2e',
      line: e2eStatus === 'passed'
        ? 'e2e ok'
        : 'e2e fail',
    },
    {
      type: 'step_finished',
      job: 'e2e',
      step: 'echo e2e',
      status: e2eStatus,
      durationMs: 9,
    },
    {
      type: 'run_finished',
      status: e2eStatus,
      exitCode: e2eStatus === 'passed'
        ? 0
        : 1,
    },
  ];
}

export function e2eJobRerunEvents(workflowPath: string): RunnerEvent[] {
  return [
    {
      type: 'run_started',
      repo: FIXTURE_REPO_GITHUB,
      workflow: workflowPath,
      job: 'e2e',
    },
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
      type: 'step_started',
      job: 'e2e',
      step: 'echo e2e',
    },
    {
      type: 'log',
      job: 'e2e',
      step: 'echo e2e',
      line: 'e2e ok',
    },
    {
      type: 'step_finished',
      job: 'e2e',
      step: 'echo e2e',
      status: 'passed',
      durationMs: 4,
    },
    {
      type: 'run_finished',
      status: 'passed',
      exitCode: 0,
    },
  ];
}
