import type { VueWrapper } from '@vue/test-utils';
import {
  FIXTURE_REPO_GITHUB,
  githubWorkflow,
  parallelGraph,
  parallelRunEvents,
  passedRunEvents,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

function jobStatus(wrapper: VueWrapper, jobId: string): string | undefined {
  return wrapper
    .get(`[data-testid="job-pill"][data-job-id="${jobId}"]`)
    .attributes('data-status');
}

describe('app session workflow run', () => {
  it('blocks run when required inputs are missing', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();

    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(mock.startRunCalls).not.toHaveBeenCalled();
    expect(wrapper.get('[data-testid="status-message"]').text())
      .toContain('Required inputs');
  });

  it('runs a workflow after filling required inputs and shows finished state', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      runEvents: passedRunEvents(githubWorkflow.path),
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();

    const noteInput = wrapper.get('#workflow-input-note');
    await noteInput.setValue('ship it');
    await flushPromises();

    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(mock.startRunCalls).toHaveBeenCalledOnce();
    const startArgs = mock.startRunCalls.mock.calls[0]?.[0] as {
      inputs?: string[];
    };
    expect(startArgs.inputs).toEqual(expect.arrayContaining([
      'environment=staging',
      'dry_run=false',
      'note=ship it',
    ]));

    expect(wrapper.get('[data-testid="run-finished"]').text()).toContain('passed');
    expect(wrapper.find('[data-testid="artifact-list"]').exists()).toBe(true);

    await wrapper.get('[data-testid="artifact-download"]').trigger('click');
    await flushPromises();
    expect(mock.saveArtifactCalls).toHaveBeenCalledOnce();
  });

  it('stops a running workflow', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      runEvents: [],
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('x');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-testid="workflow-stop"]').exists()).toBe(true);
    await wrapper.get('[data-testid="workflow-stop"]').trigger('click');
    await flushPromises();

    expect(mock.stopRunCalls).toHaveBeenCalledOnce();
    expect(wrapper.get('[data-testid="run-finished"]').text()).toContain('cancelled');
  });

  it('surfaces startRun errors in the status message', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      startRunError: new Error('boom'),
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('x');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="status-message"]').text()).toBe('boom');
  });

  it('shows job statuses while a failing run streams', async () => {
    const events = parallelRunEvents(githubWorkflow.path, 'failed');
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: parallelGraph },
      runEvents: [],
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    // Through e2e step_started — lint/test finished, e2e live.
    for (const event of events.slice(0, 9)) {
      mock.emit(event);
    }
    await flushPromises();

    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('running');
    expect(wrapper.find('[data-testid="run-finished"]').exists()).toBe(false);

    for (const event of events.slice(9)) {
      mock.emit(event);
    }
    await flushPromises();

    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('failed');
    expect(wrapper.get('[data-testid="run-finished"]').text()).toBe(
      'run failed (exit 1)',
    );
  });
});
