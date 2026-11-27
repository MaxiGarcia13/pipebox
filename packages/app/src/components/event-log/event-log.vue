<script setup lang="ts">
import type { RunnerEvent, WorkflowProvider } from '@/domain/runner-events';
import { computed, ref, toRef } from 'vue';
import JobNeedsDialog from '@/components/job-needs-dialog.vue';
import JobRunBar from '@/components/job-run-bar.vue';
import JobTimeline from '@/components/job-timeline.vue';
import RunOutcome from '@/components/run-outcome.vue';
import RunTerminal from '@/components/run-terminal.vue';
import StepLogPanel from '@/components/step-log-panel.vue';
import { useLogSearchSession } from '@/composables/use-log-search-session';
import { useNow } from '@/composables/use-now';
import ParallelJobsToggle from './parallel-jobs-toggle.vue';
import { useEventLogView } from './use-event-log-view';

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

const root = ref<HTMLElement | null>(null);
const now = useNow(toRef(props, 'isRunning'));
const logSearch = useLogSearchSession(root);

const {
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
} = useEventLogView(
  () => props.events,
  {
    isRunning: () => props.isRunning,
    parallelJobs: () => props.parallelJobs,
    provider: () => props.provider,
    onRunJob: (jobId) => emit('runJob', jobId),
    onRunWorkflow: () => emit('runWorkflow'),
  },
);

const parallelChecked = computed(() => props.parallelJobs === true);

function onParallelChecked(value: boolean) {
  if (props.isRunning) {
    return;
  }
  emit('update:parallelJobs', value);
}
</script>

<template>
  <div
    ref="root"
    data-testid="event-log"
    class="flex min-h-0 flex-1 flex-col"
  >
    <div class="flex items-center justify-between gap-4 border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
      <ParallelJobsToggle
        v-if="showParallel"
        :checked="parallelChecked"
        :disabled="isRunning"
        @update:checked="onParallelChecked"
      />
    </div>
    <p
      v-if="showParallel && parallelChecked"
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
        @select="selectJob"
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
            :search="logSearch"
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
          :search="logSearch"
        />
      </div>
    </template>
  </div>
</template>
