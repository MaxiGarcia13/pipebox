import type { JobSummary } from '@/domain/run/types';
import { describe, expect, it } from 'vitest';
import { allNeedsPassed } from '@/domain/run/all-needs-passed';

const jobs: JobSummary[] = [
  {
    id: 'lint',
    name: 'Lint',
    needs: [],
    status: 'passed',
  },
  {
    id: 'test',
    name: 'Test',
    needs: [],
    status: 'passed',
  },
  {
    id: 'e2e',
    name: 'E2E',
    needs: ['lint', 'test'],
    status: 'failed',
  },
];

describe('allNeedsPassed', () => {
  it('returns true when the job has no needs', () => {
    expect(allNeedsPassed({ needs: [] }, jobs)).toBe(true);
  });

  it('returns true when lint and test passed even if e2e failed', () => {
    expect(allNeedsPassed({
      needs: ['lint', 'test'],
    }, jobs)).toBe(true);
  });

  it('returns false when a need is pending', () => {
    const withPending = jobs.map((job) =>
      job.id === 'test'
        ? { ...job, status: 'pending' as const }
        : job,
    );
    expect(allNeedsPassed({
      needs: ['lint', 'test'],
    }, withPending)).toBe(false);
  });

  it('returns false when a need failed', () => {
    const withFailed = jobs.map((job) =>
      job.id === 'lint'
        ? { ...job, status: 'failed' as const }
        : job,
    );
    expect(allNeedsPassed({
      needs: ['lint', 'test'],
    }, withFailed)).toBe(false);
  });

  it('returns false when a need is missing from the session', () => {
    expect(allNeedsPassed({
      needs: ['lint', 'missing'],
    }, jobs)).toBe(false);
  });
});
