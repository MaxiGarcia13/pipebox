/// <reference types="@electron-forge/plugin-vite/forge-vite-env" />

declare module '*.css' {
  const styles: string;
  export default styles;
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<object, object, unknown>;
  export default component;
}

declare global {
  interface Window {
    pipebox: import('./shared/pipebox-api').PipeboxApi;
  }
}

export {};
