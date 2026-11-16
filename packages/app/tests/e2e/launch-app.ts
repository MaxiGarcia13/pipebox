import type { ElectronApplication, Page } from '@playwright/test';
import type { FixtureRepoName } from './resolve-packaged-app';
import { _electron as electron } from '@playwright/test';
import {
  resolveFixtureRepo,
  resolvePackagedExecutable,
} from './resolve-packaged-app';

export interface LaunchedApp {
  electronApp: ElectronApplication;
  window: Page;
}

export async function launchAppWithRepo(
  fixture: FixtureRepoName,
): Promise<LaunchedApp> {
  const electronApp = await electron.launch({
    executablePath: resolvePackagedExecutable(),
    args: [`--repo=${resolveFixtureRepo(fixture)}`],
  });
  const window = await electronApp.firstWindow();
  await window.waitForLoadState('domcontentloaded');
  return { electronApp, window };
}
