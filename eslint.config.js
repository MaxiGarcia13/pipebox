import { fileURLToPath } from 'node:url';
import { eslintConfig } from '@maxigarcia/eslint-config';

const cssConfigPath = fileURLToPath(
  new URL('./packages/app/src/index.css', import.meta.url),
);

export default eslintConfig(
  {
    typescript: true,
    vue: true,
    tailwindcss: true,
  },
  {
    settings: {
      tailwindcss: {
        cssConfigPath,
      },
    },
  },
);
