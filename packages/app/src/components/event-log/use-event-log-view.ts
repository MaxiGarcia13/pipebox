import type { MaybeRefOrGetter } from 'vue';
import type { JobNeedDependency } from '@/components/job-needs-dialog.vue';
import type { RunnerEvent, WorkflowProvider } from '@/domain/runner-events';
import { computed, ref, toValue } from 'vue';
import {
  allNeedsPassed,
  chronologicalLogLines,
  defaultSelectedJobId,
  findLastArtifacts,
  findLastRunFinished,
  findRunErrors,
  groupRunEvents,
  stepsForJob,
  unscopedLogLines,
} from '@/domain/run';

export function useEventLogView(
  events: MaybeRefOrGetter<RunnerEvent[]>,
  options: {
    isRunning: MaybeRefOrGetter<boolean>;
    parallelJobs: MaybeRefOrGetter<boolean | undefined>;
    provider: MaybeRefOrGetter<WorkflowProvider | null | undefined>;
    onRunJob: (jobId: string) => void;
    onRunWorkflow: () => void;
  },
) {
  const selectedId = ref<string | null>(null);
  const needsDialogOpen = ref(false);

  const grouped = computed(() => groupRunEvents(toValue(events)));

  const activeJobId = computed(() => {
    const jobs = grouped.value.jobs;
    if (selectedId.value && jobs.some((job) => job.id === selectedId.value)) {
      return selectedId.value;
    }
    return defaultSelectedJobId(jobs);
  });

  const activeJob = computed(() =>
    grouped.value.jobs.find((job) => job.id === activeJobId.value) ?? null,
  );

  const visibleSteps = computed(() =>
    stepsForJob(grouped.value.steps, activeJob.value),
  );

  const runLabel = computed((): 'Run' | 'Re-run' => {
    if (toValue(options.parallelJobs)) {
      return 'Run';
    }
    return activeJob.value && activeJob.value.status !== 'pending'
      ? 'Re-run'
      : 'Run';
  });

  const showParallel = computed(() => toValue(options.provider) === 'github');

  const showTerminal = computed(() => toValue(options.parallelJobs) === true);

  const terminalLines = computed(() => chronologicalLogLines(toValue(events)));

  const needDependencies = computed((): JobNeedDependency[] => {
    const job = activeJob.value;
    if (!job) {
      return [];
    }
    return job.needs.map((needId) => {
      const dependency = grouped.value.jobs.find(
        (candidate) => candidate.id === needId || candidate.name === needId,
      );
      return {
        id: needId,
        name: dependency?.name ?? needId,
        status: dependency?.status ?? 'pending',
      };
    });
  });

  const runFinished = computed(() => findLastRunFinished(toValue(events)));

  const artifacts = computed(() => findLastArtifacts(toValue(events)));

  const runErrors = computed(() => findRunErrors(toValue(events)));

  const systemLogs = computed(() => unscopedLogLines(toValue(events)));

  const isEmpty = computed(() =>
    !toValue(options.isRunning) && toValue(events).length === 0,
  );

  function selectJob(jobId: string) {
    selectedId.value = jobId;
  }

  function requestRun() {
    if (toValue(options.isRunning)) {
      return;
    }
    if (toValue(options.parallelJobs)) {
      options.onRunWorkflow();
      return;
    }
    const job = activeJob.value;
    if (!job) {
      return;
    }
    if (allNeedsPassed(job, grouped.value.jobs)) {
      options.onRunJob(job.id);
      return;
    }
    needsDialogOpen.value = true;
  }

  function confirmNeedsRun() {
    const job = activeJob.value;
    if (!job) {
      return;
    }
    options.onRunJob(job.id);
  }

  return {
    selectedId,
    needsDialogOpen,
    grouped,
    activeJobId,
    activeJob,
    visibleSteps,
    runLabel,
    showParallel,
    showTerminal,
    terminalLines,
    needDependencies,
    runFinished,
    artifacts,
    runErrors,
    systemLogs,
    isEmpty,
    selectJob,
    requestRun,
    confirmNeedsRun,
  };
}
