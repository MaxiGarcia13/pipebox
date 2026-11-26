import type { JobGraphEvent, RunnerEvent } from '@/domain/runner-events';
import { FIXTURE_REPO_GITHUB } from './repos';

export const githubGraph: JobGraphEvent = {
  type: 'job_graph',
  jobs: [
    {
      id: 'lint',
      name: 'Lint',
      needs: [],
      steps: ['echo ok'],
    },
  ],
};

export function passedRunEvents(workflowPath: string): RunnerEvent[] {
  return [
    {
      type: 'run_started',
      repo: FIXTURE_REPO_GITHUB,
      workflow: workflowPath,
    },
    githubGraph,
    {
      type: 'step_started',
      job: 'lint',
      step: 'echo ok',
    },
    {
      type: 'log',
      job: 'lint',
      step: 'echo ok',
      line: 'ok',
    },
    {
      type: 'step_finished',
      job: 'lint',
      step: 'echo ok',
      status: 'passed',
      durationMs: 12,
    },
    {
      type: 'artifacts',
      artifacts: [
        {
          name: 'report',
          path: '/tmp/report.txt',
          size: 128,
        },
      ],
    },
    {
      type: 'run_finished',
      status: 'passed',
      exitCode: 0,
    },
  ];
}
