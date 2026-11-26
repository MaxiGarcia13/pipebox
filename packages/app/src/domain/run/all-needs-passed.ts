import type { JobSummary } from './types';

export function allNeedsPassed(
  job: Pick<JobSummary, 'needs'>,
  jobs: JobSummary[],
): boolean {
  if (job.needs.length === 0) {
    return true;
  }

  return job.needs.every((needId) => {
    const dependency = jobs.find(
      (candidate) => candidate.id === needId || candidate.name === needId,
    );
    return dependency?.status === 'passed';
  });
}
