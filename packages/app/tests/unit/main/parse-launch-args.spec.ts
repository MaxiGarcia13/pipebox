import { describe, expect, it } from 'vitest';
import { parseLaunchArgs } from '@/main/parse-launch-args';

describe('parseLaunchArgs', () => {
  it('returns empty defaults', () => {
    expect(parseLaunchArgs([])).toEqual({ repo: null, newWindow: false });
  });

  it('parses --repo and --repo=', () => {
    expect(parseLaunchArgs(['--repo', '/tmp/a'])).toEqual({
      repo: '/tmp/a',
      newWindow: false,
    });
    expect(parseLaunchArgs(['--repo=/tmp/b'])).toEqual({
      repo: '/tmp/b',
      newWindow: false,
    });
  });

  it('parses -n and --new', () => {
    expect(parseLaunchArgs(['-n'])).toEqual({
      repo: null,
      newWindow: true,
    });
    expect(parseLaunchArgs(['--new', '--repo', '/tmp/c'])).toEqual({
      repo: '/tmp/c',
      newWindow: true,
    });
    expect(parseLaunchArgs(['--repo=.', '-n'])).toEqual({
      repo: '.',
      newWindow: true,
    });
  });

  it('ignores incomplete --repo', () => {
    expect(parseLaunchArgs(['--repo'])).toEqual({
      repo: null,
      newWindow: false,
    });
    expect(parseLaunchArgs(['--repo', '-n'])).toEqual({
      repo: null,
      newWindow: true,
    });
  });
});
