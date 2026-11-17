import { expect, test } from '@playwright/test';
import { launchAppWithRepo } from './launch-app';

test('empty fixture shows no workflows', async () => {
  const { electronApp, window } = await launchAppWithRepo('empty');
  try {
    await expect(window.getByTestId('workflow-list-none')).toBeVisible();
    await expect(window.getByTestId('project-status')).toContainText('empty');
    await expect(window.getByTestId('doctor-status')).toBeVisible();
  } finally {
    await electronApp.close();
  }
});

test('jenkins-only fixture lists Jenkinsfile workflow', async () => {
  const { electronApp, window } = await launchAppWithRepo('jenkins-only');
  try {
    await expect(window.getByTestId('project-status')).toContainText('jenkins-only');
    await expect(window.getByTestId('workflow-list')).toBeVisible();
    const row = window.getByTestId('workflow-row');
    await expect(row).toHaveAttribute('data-provider', 'jenkins');
    await expect(row).toHaveAttribute('data-workflow-path', 'Jenkinsfile');
    await expect(row.getByTestId('workflow-name')).toHaveText('Jenkinsfile');
  } finally {
    await electronApp.close();
  }
});
