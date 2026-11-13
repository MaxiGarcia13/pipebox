<script setup lang="ts">
import type { JobRunStatus, StepRunStatus } from '@/utils';
import { CheckCircle2, Circle, CircleX, LoaderCircle, MinusCircle } from 'lucide-vue-next';

defineProps<{
  status: JobRunStatus | StepRunStatus;
}>();
</script>

<template>
  <CheckCircle2
    v-if="status === 'passed'"
    class="size-4 shrink-0 text-chart-3"
    aria-hidden="true"
  />
  <CircleX
    v-else-if="status === 'failed'"
    class="size-4 shrink-0 text-destructive"
    aria-hidden="true"
  />
  <MinusCircle
    v-else-if="status === 'skipped'"
    class="size-4 shrink-0 text-muted-foreground"
    aria-hidden="true"
  />
  <LoaderCircle
    v-else-if="status === 'running'"
    class="size-4 shrink-0 animate-spin text-chart-2"
    aria-hidden="true"
  />
  <Circle
    v-else
    class="size-4 shrink-0 text-muted-foreground"
    aria-hidden="true"
  />
</template>
