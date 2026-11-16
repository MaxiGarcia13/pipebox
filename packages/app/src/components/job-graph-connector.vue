<script setup lang="ts">
import type { JobGraphConnectorJob } from '@/components/job-graph-edges';
import { computed } from 'vue';
import { buildJobGraphConnector } from '@/components/job-graph-edges';

const props = defineProps<{
  from: JobGraphConnectorJob[];
  to: JobGraphConnectorJob[];
}>();

const layout = computed(() => buildJobGraphConnector(props.from, props.to));
</script>

<template>
  <svg
    v-if="layout.height > 0"
    data-testid="job-graph-connector"
    class="shrink-0 text-muted-foreground"
    :width="layout.width"
    :height="layout.height"
    :viewBox="`0 0 ${layout.width} ${layout.height}`"
    aria-hidden="true"
  >
    <path
      v-if="layout.lines.length > 0"
      :d="layout.lines.join(' ')"
      fill="none"
      stroke="currentColor"
      stroke-width="1"
    />
    <circle
      v-for="(dot, index) in layout.dots"
      :key="index"
      :cx="dot.x"
      :cy="dot.y"
      r="4"
      fill="currentColor"
    />
  </svg>
</template>
