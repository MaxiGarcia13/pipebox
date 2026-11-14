export function parseLaunchRepo(argv: string[]): string | null {
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === '--repo') {
      const value = argv[index + 1];
      if (!value || value.startsWith('-')) {
        return null;
      }
      return value;
    }
    if (arg.startsWith('--repo=')) {
      const value = arg.slice('--repo='.length);
      return value === ''
        ? null
        : value;
    }
  }
  return null;
}
