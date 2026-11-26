import type { VueWrapper } from '@vue/test-utils';
import {
  e2eJobRerunEvents,
  FIXTURE_REPO_GITHUB,
  githubWorkflow,
  parallelGraph,
  parallelRunEvents,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

function jobStatus(wrapper: VueWrapper, jobId: string): string | undefined {
  return wrapper
    .get(`[data-testid="job-pill"][data-job-id="${jobId}"]`)
    .attributes('data-status');
}

describe('app session job re-run', () => {
  it('re-runs e2e with job flag, skips needs modal when lint and test passed, keeps siblings', async () => {
    let runCount = 0;
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: parallelGraph },
      runEvents: (options) => {
        runCount += 1;
        if (options.job === 'e2e') {
          return e2eJobRerunEvents(githubWorkflow.path);
        }
        return parallelRunEvents(githubWorkflow.path, 'failed');
      },
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(runCount).toBe(1);
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('Lint');
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('Test');
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('E2E');

    await wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="job-run-button"]').text()).toContain('Re-run');
    await wrapper.get('[data-testid="job-run-button"]').trigger('click');
    await flushPromises();

    expect(document.body.querySelector('[data-testid="job-needs-dialog"]')).toBeNull();
    expect(mock.startRunCalls).toHaveBeenCalledTimes(2);
    const jobArgs = mock.startRunCalls.mock.calls[1]?.[0] as {
      job?: string;
      workflow: string;
    };
    expect(jobArgs.job).toBe('e2e');
    expect(jobArgs.workflow).toBe(githubWorkflow.path);
    expect(runCount).toBe(2);

    expect(wrapper.find('[data-testid="job-pill"][data-job-id="lint"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="job-pill"][data-job-id="test"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="job-pill"][data-job-id="e2e"]').exists()).toBe(true);

    await wrapper.get('[data-testid="job-pill"][data-job-id="lint"]').trigger('click');
    await flushPromises();
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('echo lint');

    await wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').trigger('click');
    await flushPromises();
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('echo e2e');
    expect(wrapper.get('[data-testid="run-finished"]').text()).toContain('passed');
  });

  it('opens needs dialog when dependencies are unmet and runs after confirm', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: parallelGraph },
      runEvents: e2eJobRerunEvents(githubWorkflow.path),
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');

    await wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="job-run-button"]').text()).toContain('Run');
    await wrapper.get('[data-testid="job-run-button"]').trigger('click');
    await flushPromises();

    expect(document.body.querySelector('[data-testid="job-needs-dialog"]')).not.toBeNull();
    expect(mock.startRunCalls).not.toHaveBeenCalled();

    document.body.querySelector<HTMLElement>('[data-testid="job-needs-confirm"]')
      ?.click();
    await flushPromises();

    expect(mock.startRunCalls).toHaveBeenCalledOnce();
    const startArgs = mock.startRunCalls.mock.calls[0]?.[0] as { job?: string };
    expect(startArgs.job).toBe('e2e');
    expect(document.body.querySelector('[data-testid="job-needs-dialog"]')).toBeNull();
  });

  it('keeps e2e selected when running that job after siblings already passed', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: parallelGraph },
      runEvents: (options) => {
        if (options.job === 'e2e') {
          return e2eJobRerunEvents(githubWorkflow.path);
        }
        return parallelRunEvents(githubWorkflow.path, 'failed');
      },
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    await wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').trigger('click');
    await flushPromises();
    expect(
      wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').classes(),
    ).toContain('border-primary');

    await wrapper.get('[data-testid="job-run-button"]').trigger('click');
    await flushPromises();

    expect(
      wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').classes(),
    ).toContain('border-primary');
    expect(wrapper.get('[data-testid="event-log"]').text()).toContain('echo e2e');
  });

  it('updates job statuses when re-running a failed job', async () => {
    const rerunEvents = e2eJobRerunEvents(githubWorkflow.path);
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      graphsByWorkflow: { [githubWorkflow.path]: parallelGraph },
      runEvents: (options) => {
        if (options.job === 'e2e') {
          return [];
        }
        return parallelRunEvents(githubWorkflow.path, 'failed');
      },
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('failed');
    expect(wrapper.get('[data-testid="run-finished"]').text()).toBe(
      'run failed (exit 1)',
    );

    await wrapper.get('[data-testid="job-pill"][data-job-id="e2e"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="job-run-button"]').trigger('click');
    await flushPromises();

    expect(mock.startRunCalls).toHaveBeenCalledTimes(2);
    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('pending');
    expect(wrapper.find('[data-testid="run-finished"]').exists()).toBe(false);

    // Through e2e step_started — filtered job_graph is ignored.
    for (const event of rerunEvents.slice(0, 3)) {
      mock.emit(event);
    }
    await flushPromises();

    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('running');
    expect(wrapper.find('[data-testid="run-finished"]').exists()).toBe(false);

    for (const event of rerunEvents.slice(3)) {
      mock.emit(event);
    }
    await flushPromises();

    expect(jobStatus(wrapper, 'lint')).toBe('passed');
    expect(jobStatus(wrapper, 'test')).toBe('passed');
    expect(jobStatus(wrapper, 'e2e')).toBe('passed');
    expect(wrapper.get('[data-testid="run-finished"]').text()).toBe(
      'run passed (exit 0)',
    );
  });
});
