import type { JobNode } from '@/shared/runner-events';

export interface LaidOutJob extends JobNode {
  column: number;
  row: number;
}

export function layoutJobGraph(jobs: JobNode[]): LaidOutJob[] {
  if (jobs.length === 0) {
    return [];
  }

  const byId = new Map(jobs.map((job) => [job.id, job]));
  const columns = new Map<string, number>();

  function columnFor(jobId: string, visiting: Set<string>): number {
    const cached = columns.get(jobId);
    if (cached != null) {
      return cached;
    }
    if (visiting.has(jobId)) {
      return 0;
    }
    visiting.add(jobId);

    const job = byId.get(jobId);
    if (!job || job.needs.length === 0) {
      columns.set(jobId, 0);
      visiting.delete(jobId);
      return 0;
    }

    let maxNeed = -1;
    for (const needId of job.needs) {
      if (!byId.has(needId)) {
        continue;
      }
      maxNeed = Math.max(maxNeed, columnFor(needId, visiting));
    }
    const column = maxNeed + 1;
    columns.set(jobId, column);
    visiting.delete(jobId);
    return column;
  }

  for (const job of jobs) {
    columnFor(job.id, new Set());
  }

  const rowByColumn = new Map<number, number>();
  return jobs.map((job) => {
    const column = columns.get(job.id) ?? 0;
    const row = rowByColumn.get(column) ?? 0;
    rowByColumn.set(column, row + 1);
    return {
      ...job,
      column,
      row,
    };
  });
}
