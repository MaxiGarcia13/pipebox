<script setup lang="ts">
import type { Artifact, ErrorEvent, RunFinishedEvent } from '@/domain/runner-events';
import ArtifactList from '@/components/artifact-list.vue';

defineProps<{
  artifacts: Artifact[];
  runFinished?: RunFinishedEvent | null;
  errors?: ErrorEvent[];
}>();
</script>

<template>
  <ul
    v-if="errors && errors.length > 0"
    data-testid="event-log-errors"
    class="space-y-2 border-t border-border p-4 text-sm text-destructive"
  >
    <li
      v-for="(error, index) in errors"
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
</template>
