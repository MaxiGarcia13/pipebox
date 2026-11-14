/// <reference types="@electron-forge/plugin-vite/forge-vite-env" />

declare global {
  const __PIPEBOX_REPORT_ISSUE_URL__: string;

  interface Window {
    pipebox: import('./services/pipebox-api').PipeboxApi;
  }
}

export {};
