<script setup lang="ts">
import type { StepGroup } from '@/domain/run';
import type { Artifact, ErrorEvent, RunFinishedEvent } from '@/domain/runner-events';
import { LOG_PRE_CLASS, LOG_SURFACE_CLASS } from '@/components/log-surface';
import RunOutcome from '@/components/run-outcome.vue';
import StepLogGroup from '@/components/step-log-group.vue';
import { cn } from '@/utils';

defineProps<{
  steps: StepGroup[];
  systemLogs: string[];
  isRunning: boolean;
  now?: number;
  artifacts: Artifact[];
  runFinished?: RunFinishedEvent | null;
  errors?: ErrorEvent[];
}>();
</script>

<template>
  <div class="min-h-0 flex-1 overflow-auto">
    <StepLogGroup
      v-for="group in steps"
      :key="group.key"
      :group="group"
      :now="now"
    />

    <p
      v-if="steps.length === 0 && !isRunning && systemLogs.length === 0"
      class="p-4 text-sm text-muted-foreground"
    >
      No steps for this job yet
    </p>

    <div
      v-if="systemLogs.length > 0"
      data-testid="event-log-system"
      :class="cn('m-4 overflow-hidden', LOG_SURFACE_CLASS)"
    >
      <pre :class="cn(LOG_PRE_CLASS, 'text-muted-foreground')">{{ systemLogs.join('\n') }}</pre>
    </div>

    <RunOutcome
      :artifacts="artifacts"
      :run-finished="runFinished"
      :errors="errors"
    />
  </div>
</template>
