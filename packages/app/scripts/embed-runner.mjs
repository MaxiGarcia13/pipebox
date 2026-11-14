import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const appDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const runnerDir = path.resolve(appDir, '..', 'runner');
const resourcesDir = path.join(appDir, 'resources');
const isWindows = process.platform === 'win32';
const outName = isWindows ? 'pipebox.exe' : 'pipebox';
const outPath = path.join(resourcesDir, outName);

fs.mkdirSync(resourcesDir, { recursive: true });

const stale = path.join(resourcesDir, isWindows ? 'pipebox' : 'pipebox.exe');
if (fs.existsSync(stale)) {
  fs.unlinkSync(stale);
}

const result = spawnSync(
  'go',
  [
    'build',
    '-trimpath',
    '-ldflags',
    '-s -w',
    '-o',
    outPath,
    './cmd/pipebox',
  ],
  {
    cwd: runnerDir,
    stdio: 'inherit',
    env: { ...process.env, CGO_ENABLED: '0' },
  },
);

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

console.log(`embedded ${outPath}`);
