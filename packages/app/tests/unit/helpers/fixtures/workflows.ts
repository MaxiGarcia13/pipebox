import type { Workflow, WorkflowListEvent } from '@/domain/runner-events';

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
