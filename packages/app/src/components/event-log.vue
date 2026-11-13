<script setup lang="ts">
import type { RunnerEvent } from '@/shared/runner-events';
import { Skeleton } from '@/components/ui/skeleton';

defineProps<{
  events: RunnerEvent[];
  isRunning: boolean;
}>();

function formatEvent(event: RunnerEvent): string {
  switch (event.type) {
    case 'log':
      return event.line;
    case 'step_started':
      return `▶ ${event.job} / ${event.step}`;
    case 'step_finished':
      return `${statusIcon(event.status)} ${event.job} / ${event.step}`;
    case 'run_started':
      return `run started: ${event.workflow}`;
    case 'run_finished':
      return `run ${event.status} (exit ${event.exitCode})`;
    case 'error':
      return `error: ${event.message}`;
    default:
      return JSON.stringify(event);
  }
}

function statusIcon(status: string): string {
  if (status === 'passed') {
    return '✓';
  }
  if (status === 'failed') {
    return '✗';
  }
  return '○';
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
    </div>

    <div class="min-h-0 flex-1 overflow-auto bg-background p-4 font-mono text-xs leading-relaxed">
      <p
        v-if="!isRunning && events.length === 0"
        class="text-muted-foreground"
      >
        Run a workflow to stream logs here
      </p>

      <template v-else>
        <pre
          v-if="events.length > 0"
          class="break-all whitespace-pre-wrap text-foreground"
        >{{ events.map(formatEvent).join('\n') }}</pre>

        <div
          v-if="isRunning"
          class="mt-2 flex flex-row gap-2"
          aria-label="Waiting for log stream"
        >
          <Skeleton class="h-3 w-30" />
          <Skeleton class="h-3 w-full" />
        </div>
      </template>
    </div>
  </div>
</template>
