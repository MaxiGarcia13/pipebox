import { describe, expect, it } from 'vitest';
import {
  buildSearchLineParts,
  splitSearchSegments,
} from '@/domain/run/log-search';

describe('splitSearchSegments', () => {
  it('returns the whole line when the query is empty', () => {
    expect(splitSearchSegments('hello world', '')).toEqual([
      { text: 'hello world', match: false },
    ]);
    expect(splitSearchSegments('hello world', '   ')).toEqual([
      { text: 'hello world', match: false },
    ]);
  });

  it('splits case-insensitive matches and keeps surrounding text', () => {
    expect(splitSearchSegments('Failure - Main Lint', 'lint')).toEqual([
      { text: 'Failure - Main ', match: false },
      { text: 'Lint', match: true },
    ]);
  });

  it('finds multiple matches on one line', () => {
    expect(splitSearchSegments('run run run', 'run')).toEqual([
      { text: 'run', match: true },
      { text: ' ', match: false },
      { text: 'run', match: true },
      { text: ' ', match: false },
      { text: 'run', match: true },
    ]);
  });
});

describe('buildSearchLineParts', () => {
  it('assigns global match indexes across lines', () => {
    const { parts, matchCount } = buildSearchLineParts(
      ['alpha fail', 'beta FAIL ok'],
      'fail',
    );
    expect(matchCount).toBe(2);
    expect(parts[0]).toEqual([
      { text: 'alpha ', matchIndex: null },
      { text: 'fail', matchIndex: 0 },
    ]);
    expect(parts[1]).toEqual([
      { text: 'beta ', matchIndex: null },
      { text: 'FAIL', matchIndex: 1 },
      { text: ' ok', matchIndex: null },
    ]);
  });
});
