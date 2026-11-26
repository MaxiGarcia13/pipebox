import {
  doctorWithFailedCheck,
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
  githubWorkflow,
  okDoctor,
  passedRunEvents,
} from '@unit-tests/helpers/fixtures';
import { mountApp } from '@unit-tests/helpers/mount-app';
import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

describe('app session doctor banner', () => {
  it('shows doctor banner for docker failure and refreshes', async () => {
    const { wrapper, mock } = await mountApp({
      doctor: doctorWithFailedCheck('docker'),
    });

    expect(wrapper.get('[data-testid="setup-banner"]').text()).toContain('Install Docker');
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('issues');

    mock.setDoctor(doctorWithFailedCheck('docker_daemon'));
    await wrapper.get('[data-testid="doctor-refresh"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="setup-banner"]').text()).toContain('daemon');
    expect(mock.doctorCalls.mock.calls.length).toBeGreaterThanOrEqual(2);
  });

  it('updates the banner from a doctor event on the run stream', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      doctor: doctorWithFailedCheck('act_image'),
    });

    expect(wrapper.get('[data-testid="setup-banner"]').text()).toContain('Act image');

    mock.emit({ ...okDoctor, version: '' });
    await flushPromises();

    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('OK');
    expect(wrapper.text()).toContain('v0.7.0');
  });

  it('refreshes doctor when a run finishes', async () => {
    const { wrapper, mock } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      doctor: doctorWithFailedCheck('act_image'),
      runEvents: passedRunEvents(githubWorkflow.path),
    });

    expect(wrapper.get('[data-testid="setup-banner"]').text()).toContain('Act image');
    const callsBefore = mock.doctorCalls.mock.calls.length;

    mock.setDoctor(okDoctor);
    await wrapper.get('[data-testid="workflow-row"]').trigger('click');
    await flushPromises();
    await wrapper.get('#workflow-input-note').setValue('ship it');
    await wrapper.get('[data-testid="workflow-run"]').trigger('click');
    await flushPromises();

    expect(mock.doctorCalls.mock.calls.length).toBeGreaterThan(callsBefore);
    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
  });

  it('hides setup banner when doctor is ok', async () => {
    const { wrapper } = await mountApp();
    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('OK');
  });

  it('hides act image hints when the project has no github workflows', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_JENKINS,
      doctor: doctorWithFailedCheck('act_image'),
    });

    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('OK');
    expect(wrapper.text()).not.toContain('act_image');
  });

  it('hides jenkins image hints when the project has no jenkins workflows', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_GITHUB,
      doctor: doctorWithFailedCheck('jenkins_image'),
    });

    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('OK');
    expect(wrapper.text()).not.toContain('jenkins_image');
  });

  it('shows jenkins image hint for jenkins projects', async () => {
    const { wrapper } = await mountApp({
      launchRepo: FIXTURE_REPO_JENKINS,
      doctor: doctorWithFailedCheck('jenkins_image'),
    });

    expect(wrapper.get('[data-testid="setup-banner"]').text()).toContain('Jenkins image');
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('issues');
  });

  it('ignores provider image hints when no project is open', async () => {
    const { wrapper } = await mountApp({
      doctor: doctorWithFailedCheck('act_image'),
    });

    expect(wrapper.find('[data-testid="setup-banner"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="doctor-label"]').text()).toBe('OK');
  });
});
