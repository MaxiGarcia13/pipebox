import path from 'node:path';
import { fileURLToPath } from 'node:url';
import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vitest/config';

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(rootDir, './src'),
      '@unit-tests': path.resolve(rootDir, './tests/unit'),
    },
  },
  define: {
    __PIPEBOX_REPORT_ISSUE_URL__: JSON.stringify('https://example.test/issues'),
  },
  test: {
    environment: 'happy-dom',
    setupFiles: ['./tests/unit/setup.ts'],
    include: ['tests/unit/**/*.spec.ts'],
  },
});
