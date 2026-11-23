import { eslintConfig } from '@maxigarcia/eslint-config';

export default eslintConfig(
  {
    typescript: true,
  },
  {
    ignores: [
      'packages/app/out/**',
      'packages/app/test-results/**',
      'packages/app/playwright-report/**',
    ],
  },
);
