import {
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

describe('app session open project', () => {
  it('shows empty project CTA and opens a github fixture', async () => {
    const { wrapper, mock } = await mountApp({
      selectRepoPath: FIXTURE_REPO_GITHUB,
    });

    expect(wrapper.get('[data-testid="workflow-list-empty"]').text())
      .toContain('No project open');

    await wrapper.get('[data-testid="open-project-empty"]').trigger('click');
    await flushPromises();

    expect(mock.selectRepoCalls).toHaveBeenCalledOnce();
    expect(wrapper.get('[data-testid="workflow-list"]').text()).toContain('CI');
    expect(
      wrapper.get('[data-testid="workflow-row"] [data-testid="provider-icon"]')
        .attributes('data-provider'),
    ).toBe('github');
    expect(wrapper.get('[data-testid="project-status"]').text()).toContain('main');
  });

  it('cancels open project when selectRepo returns null', async () => {
    const { wrapper } = await mountApp({
      selectRepoPath: null,
    });

    await wrapper.get('[data-testid="open-project-empty"]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-testid="workflow-list-empty"]').exists()).toBe(true);
  });

  it('opens launch repo on mount', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
    });

    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('CI');
  });

  it('opens a repo pushed from main after mount', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
    });

    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('CI');

    mock.emitOpenRepo(FIXTURE_REPO_JENKINS);
    await flushPromises();

    expect(wrapper.get('[data-testid="workflow-name"]').text()).toBe('Jenkins pipeline');
    expect(wrapper.get('[data-testid="project-status"]').text()).toContain('main');
  });
});
