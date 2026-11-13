import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execFileAsync = promisify(execFile);

export async function resolveGitBranch(repoPath: string): Promise<string | null> {
  try {
    const { stdout } = await execFileAsync(
      'git',
      [
        '-C',
        repoPath,
        'rev-parse',
        '--abbrev-ref',
        'HEAD',
      ],
      { timeout: 5_000 },
    );
    const branch = stdout.trim();
    if (!branch) {
      return null;
    }
    if (branch !== 'HEAD') {
      return branch;
    }

    const { stdout: sha } = await execFileAsync(
      'git',
      [
        '-C',
        repoPath,
        'rev-parse',
        '--short',
        'HEAD',
      ],
      { timeout: 5_000 },
    );
    const shortSha = sha.trim();
    return shortSha || null;
  } catch {
    return null;
  }
}
