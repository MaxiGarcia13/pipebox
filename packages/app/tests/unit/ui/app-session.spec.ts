import { flushPromises } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import {
  doctorWithFailedCheck,
  FIXTURE_REPO_EMPTY,
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
  FIXTURE_REPO_MIXED,
  githubWorkflow,
  okDoctor,
  passedRunEvents,
  workflowList,
} from '../helpers/fixtures';
import { mountApp } from '../helpers/mount-app';

describe('app session UI', () => {
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
});
