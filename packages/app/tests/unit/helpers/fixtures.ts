import type {
  DoctorEvent,
  JobGraphEvent,
  RunnerEvent,
  Workflow,
  WorkflowListEvent,
} from '@/domain/runner-events';

export const FIXTURE_REPO_GITHUB = '/fixtures/github-only';
export const FIXTURE_REPO_JENKINS = '/fixtures/jenkins-only';
export const FIXTURE_REPO_MIXED = '/fixtures/mixed';
export const FIXTURE_REPO_EMPTY = '/fixtures/empty';

export const okDoctor: DoctorEvent = {
  type: 'doctor',
  ok: true,
  version: '0.7.0',
  checks: [
    { name: 'docker', ok: true, path: '/usr/bin/docker' },
    { name: 'docker_daemon', ok: true, detail: 'reachable' },
    { name: 'act_image', ok: true, detail: 'pipebox-act:latest' },
    { name: 'jenkins_image', ok: true, detail: 'pipebox-jenkins:latest' },
  ],
};

export function doctorWithFailedCheck(
  name: 'docker' | 'docker_daemon' | 'act_image' | 'jenkins_image',
): DoctorEvent {
  const checks = okDoctor.checks.map((check) =>
    check.name === name
      ? { ...check, ok: false, detail: `${name} failed` }
      : check,
  );
  return {
    type: 'doctor',
    ok: false,
    version: '0.7.0',
    checks,
  };
}

export const githubWorkflow: Workflow = {
  id: 'ci',
  name: 'CI',
  path: '.github/workflows/ci.yml',
  provider: 'github',
  callable: true,
  events: ['workflow_dispatch'],
  inputs: [
    {
      id: 'environment',
      description: 'Target env',
      required: true,
      default: 'staging',
      type: 'choice',
      options: ['staging', 'production'],
    },
    {
      id: 'dry_run',
      description: 'Dry run',
      default: 'false',
      type: 'boolean',
    },
    {
      id: 'note',
      description: 'Note',
      required: true,
      type: 'string',
    },
  ],
};

export const jenkinsWorkflow: Workflow = {
  id: 'pipeline',
  name: 'Jenkins pipeline',
  path: 'Jenkinsfile',
  provider: 'jenkins',
};

export const mixedGithubWorkflow: Workflow = {
  id: 'ci',
  name: 'Mixed CI',
  path: '.github/workflows/ci.yml',
  provider: 'github',
  events: ['push'],
};

export function workflowList(
  workflows: Workflow[],
  providers?: Workflow['provider'][],
): WorkflowListEvent {
  return {
    type: 'workflow_list',
    providers: providers ?? [
      ...new Set(workflows.map((item) => item.provider)),
    ],
    workflows,
  };
}

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
