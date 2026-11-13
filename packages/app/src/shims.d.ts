declare module '*.css' {
  const styles: string;
  export default styles;
}

declare module '*.png' {
  const src: string;
  export default src;
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<object, object, unknown>;
  export default component;
}
