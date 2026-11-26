import {
  FIXTURE_REPO_EMPTY,
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
  FIXTURE_REPO_MIXED,
  githubGraph,
  githubWorkflow,
  workflowList,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

describe('app session workflow list', () => {
  it('lists jenkins workflows with jenkins icon', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_JENKINS,
    });

    const row = wrapper.get('[data-testid="workflow-row"]');
    expect(row.attributes('data-provider')).toBe('jenkins');
    expect(row.get('[data-testid="provider-icon"]').attributes('data-provider'))
      .toBe('jenkins');
    expect(row.get('[data-testid="workflow-name"]').text()).toBe('Jenkins pipeline');
  });

  it('lists mixed github and jenkins providers', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_MIXED,
    });

    const rows = wrapper.findAll('[data-testid="workflow-row"]');
    expect(rows).toHaveLength(2);
    expect(rows[0]?.attributes('data-provider')).toBe('github');
    expect(rows[1]?.attributes('data-provider')).toBe('jenkins');
  });

  it('shows empty workflows message', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_EMPTY,
    });

    expect(wrapper.get('[data-testid="workflow-list-none"]').text())
      .toContain('No workflows found');
  });

  it('refreshes the workflow list for the open project', async () => {
    const list = workflowList([githubWorkflow], ['github']);
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      listsByRepo: { [FIXTURE_REPO_GITHUB]: list },
    });

    expect(wrapper.findAll('[data-testid="workflow-row"]')).toHaveLength(1);
    expect(wrapper.find('[data-testid="workflow-refresh"]').exists()).toBe(true);

    Object.assign(list, workflowList([
      githubWorkflow,
      {
        ...githubWorkflow,
        id: 'publish',
        name: 'Publish CLI',
        path: '.github/workflows/publish.yml',
      },
    ], ['github']));

    await wrapper.get('[data-testid="workflow-refresh"]').trigger('click');
    await flushPromises();

    const rows = wrapper.findAll('[data-testid="workflow-row"]');
    expect(rows).toHaveLength(2);
    expect(rows[1]?.get('[data-testid="workflow-name"]').text()).toBe('Publish CLI');
  });

  it('refreshes the selected workflow graph preview', async () => {
    const graph = {
      type: 'job_graph' as const,
      jobs: [...githubGraph.jobs],
    };
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: graph },
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('Lint');
    expect(wrapper.get('[data-testid="event-log"]').text()).not.toContain('Test');

    graph.jobs = [
      {
        id: 'test',
        name: 'Test',
        needs: [],
        steps: ['npm test'],
      },
    ];

    await wrapper.get('[data-testid="workflow-refresh"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('Test');
    expect(wrapper.get('[data-testid="event-log"]').text()).not.toContain('Lint');
  });

  it('loads graph preview when selecting a workflow', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-testid="event-log-empty"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('Lint');
    expect(wrapper.find('[data-testid="workflow-inputs"]').exists()).toBe(true);
  });
});
