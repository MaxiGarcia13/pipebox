<script setup lang="ts">
import type { RunnerEvent } from '@/shared/runner-events';

defineProps<{
  events: RunnerEvent[];
}>();
</script>

<template>
  <section class="flex flex-col gap-2">
    <h2 class="text-lg font-medium">
      Events
    </h2>
    <pre class="overflow-auto rounded-xl bg-secondary p-4 text-xs text-secondary-foreground">{{ events.map((event) => JSON.stringify(event)).join('\n') }}</pre>
  </section>
</template>
