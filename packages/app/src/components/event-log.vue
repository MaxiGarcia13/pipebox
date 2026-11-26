<script setup lang="ts">
import type { Artifact, ErrorEvent, RunnerEvent } from '@/domain/runner-events';
import { computed, ref, toRef } from 'vue';
import ArtifactList from '@/components/artifact-list.vue';
import JobTimeline from '@/components/job-timeline.vue';
import StepLogGroup from '@/components/step-log-group.vue';
import { useNow } from '@/composables/use-now';
import {
  defaultSelectedJobId,
  groupRunEvents,
  stepsForJob,
  unscopedLogLines,
} from '@/domain/run';

const props = defineProps<{
  events: RunnerEvent[];
  isRunning: boolean;
}>();

defineEmits<{
  runJob: [jobId: string];
}>();

const selectedId = ref<string | null>(null);
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
</script>

<template>
  <div
    data-testid="event-log"
    class="flex min-h-0 flex-1 flex-col"
  >
    <div class="border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
    </div>

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

      <div class="min-h-0 flex-1 overflow-auto">
        <StepLogGroup
          v-for="group in visibleSteps"
          :key="group.key"
          :group="group"
          :now="now"
        />

        <p
          v-if="visibleSteps.length === 0 && !isRunning && systemLogs.length === 0"
          class="p-4 text-sm text-muted-foreground"
        >
          No steps for this job yet
        </p>

        <pre
          v-if="systemLogs.length > 0"
          data-testid="event-log-system"
          class="overflow-x-auto border-t border-border p-4 font-mono text-xs leading-relaxed text-muted-foreground"
        >{{ systemLogs.join('\n') }}</pre>

        <ul
          v-if="runErrors.length > 0"
          data-testid="event-log-errors"
          class="space-y-2 border-t border-border p-4 text-sm text-destructive"
        >
          <li
            v-for="(error, index) in runErrors"
            :key="index"
          >
            {{ error.message }}
          </li>
        </ul>

        <ArtifactList :artifacts="artifacts" />

        <p
          v-if="runFinished"
          data-testid="run-finished"
          class="border-t border-border px-4 py-2 text-sm text-muted-foreground"
        >
          run {{ runFinished.status }} (exit {{ runFinished.exitCode }})
        </p>
      </div>
    </template>
  </div>
</template>
