import {
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

describe('app session switch project', () => {
  it('asks before switching project while a run is active', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      runEvents: [],
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('x');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    mock.emitOpenRepo(FIXTURE_REPO_JENKINS);
    await flushPromises();

    expect(document.body.querySelector('[data-testid="switch-project-dialog"]')).not.toBeNull();
    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('CI');
    expect(mock.stopRunCalls).not.toHaveBeenCalled();

    document.body.querySelector<HTMLElement>('[data-testid="switch-project-continue"]')
      ?.click();
    await flushPromises();

    expect(document.body.querySelector('[data-testid="switch-project-dialog"]')).toBeNull();
    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('CI');
    expect(mock.stopRunCalls).not.toHaveBeenCalled();
  });

  it('stops the run and opens the pending project on confirm', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      runEvents: [],
    });

    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('x');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    mock.emitOpenRepo(FIXTURE_REPO_JENKINS);
    await flushPromises();

    document.body.querySelector<HTMLElement>('[data-testid="switch-project-stop"]')
      ?.click();
    await flushPromises();

    expect(mock.stopRunCalls).toHaveBeenCalledOnce();
    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('Jenkins pipeline');
    expect(document.body.querySelector('[data-testid="switch-project-dialog"]')).toBeNull();
  });
});
