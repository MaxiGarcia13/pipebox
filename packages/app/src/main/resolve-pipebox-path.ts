import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { app } from 'electron';

export function resolvePipeboxPath(): string {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, 'pipebox');
  }

  // packages/app/.vite/build → packages/runner/pipebox
  const candidatePath = path.resolve(__dirname, '../../../runner/pipebox');
  if (!fs.existsSync(candidatePath)) {
    throw new Error(
      `pipebox binary not found at ${candidatePath}. Run: pnpm runner:build`,
    );
  }
  return candidatePath;
}
