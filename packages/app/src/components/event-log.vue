<script setup lang="ts">
import type { RunnerEvent } from '@/shared/runner-events';

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
    <div class="flex items-center justify-between border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Logs
      </h2>
      <span
        v-if="isRunning"
        class="text-xs text-primary"
      >
        running…
      </span>
    </div>

    <div class="min-h-0 flex-1 overflow-auto bg-background p-4 font-mono text-xs leading-relaxed">
      <p
        v-if="events.length === 0"
        class="text-muted-foreground"
      >
        Run a workflow to stream logs here
      </p>
      <pre
        v-else
        class="break-all whitespace-pre-wrap text-foreground"
      >{{ events.map(formatEvent).join('\n') }}</pre>
    </div>
  </div>
</template>
