const launchRepos = new Map<number, string | null>();

export function setWindowLaunchRepo(
  webContentsId: number,
  repo: string | null,
): void {
  launchRepos.set(webContentsId, repo);
}

export function takeWindowLaunchRepo(webContentsId: number): string | null {
  const repo = launchRepos.get(webContentsId) ?? null;
  launchRepos.delete(webContentsId);
  return repo;
}

export function clearWindowLaunchRepo(webContentsId: number): void {
  launchRepos.delete(webContentsId);
}
