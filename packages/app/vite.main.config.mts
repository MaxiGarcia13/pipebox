import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';

const rootDir = path.dirname(fileURLToPath(import.meta.url));
const monorepoPackage = JSON.parse(
  readFileSync(path.resolve(rootDir, '../../package.json'), 'utf8'),
) as { repository?: { url?: string } };

const repositoryUrl = monorepoPackage.repository?.url;
if (!repositoryUrl) {
  throw new Error('repository.url missing from monorepo package.json');
}

const reportIssueUrl = `${repositoryUrl
  .replace(/^git\+/, '')
  .replace(/\.git$/, '')}/issues/new`;

export default defineConfig({
  define: {
    __PIPEBOX_REPORT_ISSUE_URL__: JSON.stringify(reportIssueUrl),
  },
});
