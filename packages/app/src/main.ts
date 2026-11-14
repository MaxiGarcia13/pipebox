import { existsSync } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { app, BrowserWindow } from 'electron';
import started from 'electron-squirrel-startup';
import { parseLaunchRepo } from './main/parse-launch-repo';
import { registerIpc } from './main/register-ipc';

if (started) {
  app.quit();
}

const launchRepo = parseLaunchRepo(process.argv);

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

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 960,
    height: 720,
    icon: resolveWindowIcon(),
    webPreferences: {
      preload: path.join(__dirname, 'preload.cjs'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  if (MAIN_WINDOW_VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(MAIN_WINDOW_VITE_DEV_SERVER_URL);
  } else {
    mainWindow.loadFile(
      path.join(__dirname, `../renderer/${MAIN_WINDOW_VITE_NAME}/index.html`),
    );
  }
}

app.whenReady().then(() => {
  registerIpc(launchRepo);
  createWindow();

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
