import { describe, expect, it } from 'vitest';
import { visibleDurationMs } from '@/domain/run';

describe('visibleDurationMs', () => {
  it('ticks a live step from when it was received', () => {
    expect(visibleDurationMs({ startedAt: 1_000 }, 4_500, true)).toBe(3_500);
  });

  it('freezes a finished step at its own receipt span', () => {
    expect(visibleDurationMs(
      { startedAt: 1_000, finishedAt: 2_500, durationMs: 10 },
      9_000,
      false,
    )).toBe(1_500);
  });

  it('falls back to the runner duration when no clock was stamped', () => {
    expect(visibleDurationMs({ durationMs: 10 }, 9_000, false)).toBe(10);
  });
});
