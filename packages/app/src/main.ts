import { existsSync } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { app, BrowserWindow } from 'electron';
import started from 'electron-squirrel-startup';
import { parseLaunchArgs } from './main/parse-launch-args';
import { registerIpc } from './main/register-ipc';
import {
  clearWindowLaunchRepo,
  setWindowLaunchRepo,
} from './main/window-launch-repo';
import { IPC_CHANNELS } from './services/ipc-channels';

const launchArgs = parseLaunchArgs(process.argv);
const gotSingleInstanceLock = !started && app.requestSingleInstanceLock();

if (started || !gotSingleInstanceLock) {
  app.quit();
}

function resolveWindowIcon(): string | undefined {
  if (process.platform === 'darwin') {
    return undefined;
  }

  const candidates = [
    path.join(__dirname, '../../assets/icon.png'),
    path.join(process.resourcesPath, 'icon.png'),
  ];

  return candidates.find((candidate) => existsSync(candidate));
}

const WINDOW_BACKGROUND = '#282a36';

interface CreateWindowOptions {
  repo?: string | null;
}

function createWindow(options: CreateWindowOptions = {}) {
  const mainWindow = new BrowserWindow({
    width: 960,
    height: 720,
    show: false,
    backgroundColor: WINDOW_BACKGROUND,
    icon: resolveWindowIcon(),
    webPreferences: {
      preload: path.join(__dirname, 'preload.cjs'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  const webContentsId = mainWindow.webContents.id;
  setWindowLaunchRepo(webContentsId, options.repo ?? null);
  mainWindow.on('closed', () => {
    clearWindowLaunchRepo(webContentsId);
  });

  mainWindow.once('ready-to-show', () => {
    mainWindow.show();
  });

  if (MAIN_WINDOW_VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(MAIN_WINDOW_VITE_DEV_SERVER_URL);
  } else {
    mainWindow.loadFile(
      path.join(__dirname, `../renderer/${MAIN_WINDOW_VITE_NAME}/index.html`),
    );
  }

  return mainWindow;
}

function focusWindow(window: BrowserWindow) {
  if (window.isMinimized()) {
    window.restore();
  }
  window.show();
  window.focus();
}

function resolveTargetWindow(): BrowserWindow | null {
  return BrowserWindow.getFocusedWindow()
    ?? BrowserWindow.getAllWindows().at(-1)
    ?? null;
}

function deliverOpenRepo(window: BrowserWindow, repo: string) {
  window.webContents.send(IPC_CHANNELS.openRepo, repo);
}

if (gotSingleInstanceLock) {
  app.on('second-instance', (_event, commandLine) => {
    const next = parseLaunchArgs(commandLine);
    if (next.newWindow) {
      createWindow({ repo: next.repo });
      return;
    }

    const window = resolveTargetWindow();
    if (!window) {
      createWindow({ repo: next.repo });
      return;
    }

    focusWindow(window);
    if (next.repo) {
      deliverOpenRepo(window, next.repo);
    }
  });

  app.whenReady().then(() => {
    registerIpc();
    createWindow({ repo: launchArgs.repo });

    app.on('activate', () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
      }
    });
  });

  app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
      app.quit();
    }
  });
}
