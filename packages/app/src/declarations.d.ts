/// <reference types="@electron-forge/plugin-vite/forge-vite-env" />

declare global {
  interface Window {
    pipebox: import('./services/pipebox-api').PipeboxApi;
  }
}

export {};
