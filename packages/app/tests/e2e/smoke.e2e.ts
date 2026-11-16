import type { ElectronApplication, Page } from '@playwright/test';
import { expect, test } from '@playwright/test';
import { launchAppWithRepo } from './launch-app';

let electronApp: ElectronApplication;
let window: Page;

test.beforeAll(async () => {
  const launched = await launchAppWithRepo('github-only');
  electronApp = launched.electronApp;
  window = launched.window;
});

test.afterAll(async () => {
  await electronApp?.close();
});

test('opens github fixture workflows with real pipebox', async () => {
  await expect(window.getByTestId('workflow-list')).toBeVisible();
  await expect(window.getByTestId('workflow-name')).toHaveText('CI');
  await expect(
    window.getByTestId('workflow-row').getByTestId('provider-icon'),
  ).toHaveAttribute('data-provider', 'github');

  await expect(window.getByTestId('doctor-status')).toBeVisible();
  await expect(window.getByTestId('doctor-label')).toBeVisible();
  await expect(window.getByTestId('project-status')).toContainText('github-only');
});

test('selects workflow and shows graph preview with inputs', async () => {
  await window.getByTestId('workflow-row').click();

  await expect(window.getByTestId('event-log-empty')).toHaveCount(0);
  await expect(window.getByTestId('event-log')).toContainText('Lint');
  await expect(window.getByTestId('workflow-inputs')).toBeVisible();
  await expect(
    window.locator('[data-testid="workflow-input"][data-input-id="environment"]'),
  ).toBeVisible();
  await expect(
    window.locator('[data-testid="workflow-input"][data-input-id="dry_run"]'),
  ).toBeVisible();
});

test('refreshes doctor from the status panel', async () => {
  const label = window.getByTestId('doctor-label');
  await expect(label).toBeVisible();
  await expect(label).not.toHaveText('…');

  await window.getByTestId('doctor-refresh').click();
  await expect(label).toBeVisible();
  await expect(label).not.toHaveText('…');
});
