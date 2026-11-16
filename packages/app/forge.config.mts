import type { ForgeConfig } from '@electron-forge/shared-types';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { MakerDeb } from '@electron-forge/maker-deb';
import { MakerSquirrel } from '@electron-forge/maker-squirrel';
import { MakerZIP } from '@electron-forge/maker-zip';
import { FusesPlugin } from '@electron-forge/plugin-fuses';
import { VitePlugin } from '@electron-forge/plugin-vite';
import { FuseV1Options, FuseVersion } from '@electron/fuses';

const rootDir = path.dirname(fileURLToPath(import.meta.url));
const appIcon = path.join(rootDir, 'assets', 'icon');
const runnerBinary = path.join(
  rootDir,
  'resources',
  process.platform === 'win32'
    ? 'pipebox.exe'
    : 'pipebox',
);

const config: ForgeConfig = {
  packagerConfig: {
    name: 'PipeBox',
    executableName: 'PipeBox',
    asar: true,
    icon: appIcon,
    extraResource: [`${appIcon}.png`, runnerBinary],
  },
  rebuildConfig: {},
  makers: [
    new MakerSquirrel({
      name: 'PipeBox',
      authors: 'Maximiliano Garcia Mortigliengo',
      description:
        'Local CI/CD control panel for running GitHub Actions with Dockerized act.',
      setupIcon: `${appIcon}.ico`,
    }),
    new MakerZIP({}, ['darwin', 'linux', 'win32']),
    new MakerDeb({
      options: {
        // Avoid claiming /usr/bin/pipebox — that name is the CLI.
        name: 'pipebox-app',
        productName: 'PipeBox',
        bin: 'PipeBox',
        description:
          'Local CI/CD control panel for running GitHub Actions with Dockerized act.',
        icon: `${appIcon}.png`,
      },
    }),
  ],
  plugins: [
    new VitePlugin({
      build: [
        {
          entry: 'src/main.ts',
          config: 'vite.main.config.mts',
          target: 'main',
        },
        {
          entry: 'src/preload.ts',
          config: 'vite.preload.config.mts',
          target: 'preload',
        },
      ],
      renderer: [
        {
          name: 'main_window',
          config: 'vite.renderer.config.mts',
        },
      ],
    }),
    new FusesPlugin({
      version: FuseVersion.V1,
      [FuseV1Options.RunAsNode]: false,
      [FuseV1Options.EnableCookieEncryption]: true,
      [FuseV1Options.EnableNodeOptionsEnvironmentVariable]: false,
      // Required for Playwright Electron to attach in e2e.
      [FuseV1Options.EnableNodeCliInspectArguments]: true,
      [FuseV1Options.EnableEmbeddedAsarIntegrityValidation]: true,
      [FuseV1Options.OnlyLoadAppFromAsar]: true,
    }),
  ],
};

export default config;
