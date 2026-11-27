<script setup lang="ts">
import type { JobNeedDependency } from '@/components/job-needs-dialog.vue';
import type { Artifact, ErrorEvent, RunnerEvent, WorkflowProvider } from '@/domain/runner-events';
import { Check } from 'lucide-vue-next';
import { computed, ref, toRef } from 'vue';
import JobNeedsDialog from '@/components/job-needs-dialog.vue';
import JobRunBar from '@/components/job-run-bar.vue';
import JobTimeline from '@/components/job-timeline.vue';
import RunOutcome from '@/components/run-outcome.vue';
import RunTerminal from '@/components/run-terminal.vue';
import StepLogPanel from '@/components/step-log-panel.vue';
import { useNow } from '@/composables/use-now';
import {
  allNeedsPassed,
  chronologicalLogLines,
  defaultSelectedJobId,
  groupRunEvents,
  stepsForJob,
  unscopedLogLines,
} from '@/domain/run';

const props = defineProps<{
  events: RunnerEvent[];
  isRunning: boolean;
  provider?: WorkflowProvider | null;
  parallelJobs?: boolean;
}>();

const emit = defineEmits<{
  'runJob': [jobId: string];
  'runWorkflow': [];
  'update:parallelJobs': [value: boolean];
}>();

const selectedId = ref<string | null>(null);
const needsDialogOpen = ref(false);
const now = useNow(toRef(props, 'isRunning'));

const grouped = computed(() => groupRunEvents(props.events));

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
  // Parallel runs the whole workflow; the control is never a single-job re-run.
  if (props.parallelJobs) {
    return 'Run';
  }
  return activeJob.value && activeJob.value.status !== 'pending'
    ? 'Re-run'
    : 'Run';
});

const showParallel = computed(() => props.provider === 'github');

const showTerminal = computed(() => props.parallelJobs === true);

const terminalLines = computed(() => chronologicalLogLines(props.events));

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

const runFinished = computed(() => {
  for (let index = props.events.length - 1; index >= 0; index -= 1) {
    const event = props.events[index];
    if (event?.type === 'run_finished') {
      return event;
    }
  }
  return null;
});

const artifacts = computed((): Artifact[] => {
  for (let index = props.events.length - 1; index >= 0; index -= 1) {
    const event = props.events[index];
    if (event?.type === 'artifacts') {
      return event.artifacts;
    }
  }
  return [];
});

const runErrors = computed(() =>
  props.events.filter((event): event is ErrorEvent => event.type === 'error'),
);

const systemLogs = computed(() => unscopedLogLines(props.events));

const isEmpty = computed(() =>
  !props.isRunning && props.events.length === 0,
);

function requestRun() {
  if (props.isRunning) {
    return;
  }
  // Parallel = full workflow + terminal stream (never --job / step split).
  if (props.parallelJobs) {
    emit('runWorkflow');
    return;
  }
  const job = activeJob.value;
  if (!job) {
    return;
  }
  if (allNeedsPassed(job, grouped.value.jobs)) {
    emit('runJob', job.id);
    return;
  }
  needsDialogOpen.value = true;
}

function confirmNeedsRun() {
  const job = activeJob.value;
  if (!job) {
    return;
  }
  emit('runJob', job.id);
}

function toggleParallelJobs() {
  if (props.isRunning) {
    return;
  }
  emit('update:parallelJobs', !props.parallelJobs);
}
</script>

<template>
  <div
    data-testid="event-log"
    class="flex min-h-0 flex-1 flex-col"
  >
    <div class="flex items-center justify-between gap-4 border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
      <div
        v-if="showParallel"
        data-testid="parallel-jobs-toggle"
        role="checkbox"
        class="flex min-w-0 items-center gap-2 text-sm text-foreground"
        :class="isRunning ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'"
        :aria-checked="parallelJobs === true"
        :aria-disabled="isRunning || undefined"
        :tabindex="isRunning ? -1 : 0"
        @click="toggleParallelJobs"
        @keydown.enter.prevent="toggleParallelJobs"
        @keydown.space.prevent="toggleParallelJobs"
      >
        <span
          data-slot="checkbox"
          class="flex size-4 shrink-0 items-center justify-center rounded-sm border shadow-xs"
          :class="parallelJobs
            ? 'border-primary bg-primary text-primary-foreground'
            : 'border-input dark:bg-input/30'"
          :data-state="parallelJobs ? 'checked' : 'unchecked'"
        >
          <Check
            v-if="parallelJobs"
            class="size-3.5"
          />
        </span>
        <span class="truncate">
          Run jobs in parallel
        </span>
      </div>
    </div>
    <p
      v-if="showParallel && parallelJobs"
      data-testid="parallel-jobs-hint"
      class="border-b border-border px-4 py-2 text-xs text-muted-foreground"
    >
      Live terminal stream for the whole workflow (step logs are hidden — parallel output is interleaved).
    </p>

    <p
      v-if="isEmpty"
      data-testid="event-log-empty"
      class="p-4 text-sm text-muted-foreground"
    >
      Select a workflow to preview jobs and steps
    </p>

    <template v-else>
      <JobTimeline
        :jobs="grouped.jobs"
        :selected-id="activeJobId"
        :now="now"
        @select="selectedId = $event"
      />

      <JobRunBar
        v-if="activeJob || parallelJobs"
        :label="runLabel"
        :disabled="isRunning"
        @run="requestRun"
      />

      <JobNeedsDialog
        v-model:open="needsDialogOpen"
        :job-name="activeJob?.name ?? ''"
        :dependencies="needDependencies"
        @confirm="confirmNeedsRun"
      />

      <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
        <template v-if="showTerminal">
          <RunTerminal
            :lines="terminalLines"
            :is-live="isRunning"
          />
          <RunOutcome
            :artifacts="artifacts"
            :run-finished="runFinished"
          />
        </template>

        <StepLogPanel
          v-else
          :steps="visibleSteps"
          :system-logs="systemLogs"
          :is-running="isRunning"
          :now="now"
          :artifacts="artifacts"
          :run-finished="runFinished"
          :errors="runErrors"
        />
      </div>
    </template>
  </div>
</template>
