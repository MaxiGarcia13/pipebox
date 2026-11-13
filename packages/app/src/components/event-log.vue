<script setup lang="ts">
import type { Artifact, ErrorEvent, RunnerEvent } from '@/domain/runner-events';
import { computed, ref } from 'vue';
import ArtifactList from '@/components/artifact-list.vue';
import JobTimeline from '@/components/job-timeline.vue';
import StepLogGroup from '@/components/step-log-group.vue';
import {
  defaultSelectedJobId,
  groupRunEvents,
  stepsForJob,
} from '@/domain/run';

const props = defineProps<{
  events: RunnerEvent[];
  isRunning: boolean;
}>();

const selectedId = ref<string | null>(null);

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

const isEmpty = computed(() =>
  !props.isRunning && props.events.length === 0,
);
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
    </div>

    <p
      v-if="isEmpty"
      class="p-4 text-sm text-muted-foreground"
    >
      Select a workflow to preview jobs and steps
    </p>

    <template v-else>
      <JobTimeline
        :jobs="grouped.jobs"
        :selected-id="activeJobId"
        @select="selectedId = $event"
      />

      <div class="min-h-0 flex-1 overflow-auto">
        <StepLogGroup
          v-for="group in visibleSteps"
          :key="group.key"
          :group="group"
        />

        <p
          v-if="visibleSteps.length === 0 && !isRunning"
          class="p-4 text-sm text-muted-foreground"
        >
          No steps for this job yet
        </p>

        <ul
          v-if="runErrors.length > 0"
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
          class="border-t border-border px-4 py-2 text-sm text-muted-foreground"
        >
          run {{ runFinished.status }} (exit {{ runFinished.exitCode }})
        </p>
      </div>
    </template>
  </div>
</template>
