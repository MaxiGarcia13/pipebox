export interface LaunchArgs {
  repo: string | null;
  newWindow: boolean;
}

export function parseLaunchArgs(argv: string[]): LaunchArgs {
  let repo: string | null = null;
  let newWindow = false;

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === '-n' || arg === '--new') {
      newWindow = true;
      continue;
    }
    if (arg === '--repo') {
      const value = argv[index + 1];
      if (!value || value.startsWith('-')) {
        repo = null;
      } else {
        repo = value;
        index += 1;
      }
      continue;
    }
    if (arg.startsWith('--repo=')) {
      const value = arg.slice('--repo='.length);
      repo = value === ''
        ? null
        : value;
    }
  }

  return { repo, newWindow };
}

export function resolveLaunchArgs(
  additionalData: unknown,
  commandLine: string[],
): LaunchArgs {
  if (isLaunchArgs(additionalData)) {
    return additionalData;
  }
  return parseLaunchArgs(commandLine);
}

function isLaunchArgs(value: unknown): value is LaunchArgs {
  if (value === null || typeof value !== 'object') {
    return false;
  }
  const record = value as Record<string, unknown>;
  const repo = record.repo;
  return (repo === null || typeof repo === 'string')
    && typeof record.newWindow === 'boolean';
}
