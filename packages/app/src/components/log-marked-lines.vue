<script setup lang="ts">
import type { ChronologicalLogKind, SearchLinePart } from '@/domain/run';
import { LOG_PRE_CLASS } from '@/components/log-surface';

export interface MarkedLogLine {
  kind?: ChronologicalLogKind;
  parts: SearchLinePart[];
}

defineProps<{
  lines: MarkedLogLine[];
  activeMatch: number;
}>();
</script>

<template>
  <pre :class="LOG_PRE_CLASS"><span
    v-for="(line, lineIndex) in lines"
    :key="lineIndex"
    class="block"
    :class="line.kind === 'error' ? 'text-destructive' : ''"
  ><template
    v-for="(part, partIndex) in line.parts"
    :key="partIndex"
  ><mark
    v-if="part.matchIndex != null"
    :data-match-index="part.matchIndex"
    class="rounded-sm px-0.5 text-search-mark-foreground"
    :class="part.matchIndex === activeMatch
      ? 'bg-search-mark-active text-search-mark-active-foreground'
      : 'bg-search-mark'"
  >{{ part.text }}</mark><template v-else>{{ part.text }}</template></template>
  </span></pre>
</template>
