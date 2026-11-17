import { describe, expect, it } from 'vitest';
import {
  parseLaunchArgs,
  resolveLaunchArgs,
} from '@/main/parse-launch-args';

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

  it('keeps --repo= across Chromium-inserted switches', () => {
    expect(parseLaunchArgs([
      '/Apps/PipeBox',
      '--repo=/tmp/project',
      '--allow-file-access-from-files',
      '--original-process-start-time=1',
    ])).toEqual({
      repo: '/tmp/project',
      newWindow: false,
    });
  });
});

describe('resolveLaunchArgs', () => {
  it('prefers additionalData from the second instance', () => {
    expect(resolveLaunchArgs(
      { repo: '/tmp/from-lock', newWindow: true },
      ['--repo=/tmp/from-argv'],
    )).toEqual({
      repo: '/tmp/from-lock',
      newWindow: true,
    });
  });

  it('falls back to commandLine when additionalData is missing', () => {
    expect(resolveLaunchArgs(undefined, ['--repo=/tmp/from-argv', '--new'])).toEqual({
      repo: '/tmp/from-argv',
      newWindow: true,
    });
  });

  it('ignores malformed additionalData', () => {
    expect(resolveLaunchArgs({ repo: 1 }, ['--repo=/tmp/ok'])).toEqual({
      repo: '/tmp/ok',
      newWindow: false,
    });
  });
});
