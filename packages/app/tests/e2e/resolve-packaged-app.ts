import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const appDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');

export type FixtureRepoName
  = | 'empty'
    | 'github-only'
    | 'jenkins-only'
    | 'mixed';

export function resolveAppDir(): string {
  return appDir;
}

export function resolveFixtureRepo(name: FixtureRepoName): string {
  const repoPath = path.resolve(
    appDir,
    '../runner/testdata/repos',
    name,
  );
  if (!fs.existsSync(repoPath)) {
    throw new Error(`Fixture repo not found: ${repoPath}`);
  }
  return repoPath;
}

export function resolveGithubFixtureRepo(): string {
  return resolveFixtureRepo('github-only');
}

export function resolvePackagedExecutable(): string {
  const outDir = path.join(appDir, 'out');
  if (!fs.existsSync(outDir)) {
    throw new Error(
      `Packaged app not found at ${outDir}. Run: pnpm -F @pipebox/app package`,
    );
  }

  const packageDirs = fs.readdirSync(outDir)
    .filter((entry) => entry.startsWith('PipeBox-'))
    .map((entry) => path.join(outDir, entry))
    .filter((entry) => fs.statSync(entry).isDirectory());

  if (packageDirs.length === 0) {
    throw new Error(`No PipeBox-* package directory under ${outDir}`);
  }

  const packageDir = packageDirs[0]!;

  if (process.platform === 'darwin') {
    return path.join(
      packageDir,
      'PipeBox.app',
      'Contents',
      'MacOS',
      'PipeBox',
    );
  }

  if (process.platform === 'win32') {
    return path.join(packageDir, 'PipeBox.exe');
  }

  return path.join(packageDir, 'PipeBox');
}
