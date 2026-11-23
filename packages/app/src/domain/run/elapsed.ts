export function visibleDurationMs(
  timing: {
    startedAt?: number;
    finishedAt?: number;
    durationMs?: number;
  },
  now: number,
  live: boolean,
): number | undefined {
  if (timing.startedAt != null && (live || timing.finishedAt != null)) {
    const end = live ? now : timing.finishedAt;
    if (end != null) {
      return Math.max(0, end - timing.startedAt);
    }
  }
  return timing.durationMs;
}
