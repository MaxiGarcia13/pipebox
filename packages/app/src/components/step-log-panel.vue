<script setup lang="ts">
import type { ChronologicalLogLine, SearchLinePart, StepGroup } from '@/domain/run';
import type { Artifact, ErrorEvent, RunFinishedEvent } from '@/domain/runner-events';
import { computed, ref } from 'vue';
import LogMarkedLines from '@/components/log-marked-lines.vue';
import LogSearchBar from '@/components/log-search-bar.vue';
import { LOG_SURFACE_CLASS } from '@/components/log-surface';
import RunOutcome from '@/components/run-outcome.vue';
import StepLogGroup from '@/components/step-log-group.vue';
import { useLogSearch } from '@/composables/use-log-search';

const props = defineProps<{
  steps: StepGroup[];
  systemLogs: string[];
  isRunning: boolean;
  now?: number;
  artifacts: Artifact[];
  runFinished?: RunFinishedEvent | null;
  errors?: ErrorEvent[];
}>();

const root = ref<HTMLElement | null>(null);
const scroller = ref<HTMLElement | null>(null);

const searchLines = computed((): ChronologicalLogLine[] => {
  const lines: ChronologicalLogLine[] = [];
  for (const step of props.steps) {
    for (const text of step.lines) {
      lines.push({ text, kind: 'log' });
    }
  }
  for (const text of props.systemLogs) {
    lines.push({ text, kind: 'log' });
  }
  return lines;
});

const {
  open,
  query,
  activeMatch,
  matchCount,
  matchLabel,
  renderedLines,
  isSearching,
  closeSearch,
  goNext,
  goPrev,
  onSearchKeydown,
} = useLogSearch(searchLines, scroller, root);

const stepSearchParts = computed((): SearchLinePart[][][] => {
  let offset = 0;
  return props.steps.map((step) => {
    const slice = renderedLines.value
      .slice(offset, offset + step.lines.length)
      .map((line) => line.parts);
    offset += step.lines.length;
    return slice;
  });
});

const systemSearchParts = computed(() => {
  let offset = 0;
  for (const step of props.steps) {
    offset += step.lines.length;
  }
  return renderedLines.value.slice(offset).map((line) => line.parts);
});

function stepHasMatch(parts: SearchLinePart[][]): boolean {
  return parts.some((line) => line.some((part) => part.matchIndex != null));
}
</script>

<template>
  <div
    ref="root"
    data-testid="step-log-panel"
    class="flex min-h-0 flex-1 flex-col"
  >
    <LogSearchBar
      v-if="open"
      v-model:query="query"
      :match-count="matchCount"
      :match-label="matchLabel"
      @next="goNext"
      @prev="goPrev"
      @close="closeSearch"
      @keydown="onSearchKeydown"
    />

    <div
      ref="scroller"
      class="min-h-0 flex-1 overflow-auto"
    >
      <StepLogGroup
        v-for="(group, index) in steps"
        :key="group.key"
        :group="group"
        :now="now"
        :line-parts="stepSearchParts[index]"
        :active-match="activeMatch"
        :force-open="isSearching && stepHasMatch(stepSearchParts[index] ?? [])"
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
        :class="LOG_SURFACE_CLASS"
      >
        <LogMarkedLines
          :lines="systemSearchParts.map((parts) => ({ parts }))"
          :active-match="activeMatch"
        />
      </div>

      <RunOutcome
        :artifacts="artifacts"
        :run-finished="runFinished"
        :errors="errors"
      />
    </div>
  </div>
</template>
