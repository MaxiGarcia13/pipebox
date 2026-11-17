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
