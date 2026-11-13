import type { BrowserWindow } from 'electron';
import type { SaveArtifactOptions } from '../services/ipc-channels';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { dialog } from 'electron';

const ARTIFACTS_ROOT = path.join(os.tmpdir(), 'pipebox-artifacts');

export async function saveArtifact(
  window: BrowserWindow | null,
  options: SaveArtifactOptions,
): Promise<string | null> {
  const sourcePath = path.resolve(options.sourcePath);
  if (!isUnderArtifactsRoot(sourcePath)) {
    throw new Error('artifact path is outside the pipebox artifacts directory');
  }

  const sourceStat = await fs.stat(sourcePath);
  if (sourceStat.isDirectory()) {
    return saveArtifactDirectory(window, sourcePath, options.defaultName);
  }
  return saveArtifactFile(window, sourcePath, options.defaultName);
}

async function saveArtifactFile(
  window: BrowserWindow | null,
  sourcePath: string,
  defaultName: string,
): Promise<string | null> {
  const result = window
    ? await dialog.showSaveDialog(window, {
        defaultPath: defaultName,
      })
    : await dialog.showSaveDialog({
        defaultPath: defaultName,
      });
  if (result.canceled || !result.filePath) {
    return null;
  }
  await fs.copyFile(sourcePath, result.filePath);
  return result.filePath;
}

async function saveArtifactDirectory(
  window: BrowserWindow | null,
  sourcePath: string,
  defaultName: string,
): Promise<string | null> {
  const result = window
    ? await dialog.showOpenDialog(window, {
        title: `Save ${defaultName}`,
        properties: ['openDirectory', 'createDirectory'],
      })
    : await dialog.showOpenDialog({
        title: `Save ${defaultName}`,
        properties: ['openDirectory', 'createDirectory'],
      });
  if (result.canceled || result.filePaths.length === 0) {
    return null;
  }
  const destination = path.join(result.filePaths[0], defaultName);
  await fs.cp(sourcePath, destination, { recursive: true });
  return destination;
}

function isUnderArtifactsRoot(sourcePath: string): boolean {
  const relativePath = path.relative(ARTIFACTS_ROOT, sourcePath);
  return relativePath !== ''
    && !relativePath.startsWith('..')
    && !path.isAbsolute(relativePath);
}
