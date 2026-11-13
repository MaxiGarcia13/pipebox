/// <reference types="@electron-forge/plugin-vite/forge-vite-env" />

declare global {
  interface Window {
    pipebox: import('./shared/pipebox-api').PipeboxApi;
  }
}

export {};
