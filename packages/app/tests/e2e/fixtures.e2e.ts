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

test('jenkins-only fixture lists no workflows until Present is implemented', async () => {
  const { electronApp, window } = await launchAppWithRepo('jenkins-only');
  try {
    await expect(window.getByTestId('workflow-list-none')).toBeVisible();
    await expect(window.getByTestId('project-status')).toContainText('jenkins-only');
    await expect(window.getByTestId('workflow-list')).toHaveCount(0);
  } finally {
    await electronApp.close();
  }
});
