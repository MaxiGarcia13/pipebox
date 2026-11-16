import type { ElectronApplication, Page } from '@playwright/test';
import { _electron as electron, expect, test } from '@playwright/test';
import {
  resolveGithubFixtureRepo,
  resolvePackagedExecutable,
} from './resolve-packaged-app';

let electronApp: ElectronApplication;
let window: Page;

test.beforeAll(async () => {
  const executablePath = resolvePackagedExecutable();
  const repoPath = resolveGithubFixtureRepo();

  electronApp = await electron.launch({
    executablePath,
    args: [`--repo=${repoPath}`],
  });
  window = await electronApp.firstWindow();
  await window.waitForLoadState('domcontentloaded');
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
