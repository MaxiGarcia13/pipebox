<script setup lang="ts">
import type { LogSearchSession } from '@/composables/use-log-search-session';
import type { ChronologicalLogLine } from '@/domain/run';
import { nextTick, ref, toRef, watch } from 'vue';
import LogMarkedLines from '@/components/log-marked-lines.vue';
import LogSearchBar from '@/components/log-search-bar.vue';
import { useLogSearch } from '@/composables/use-log-search';

const props = defineProps<{
  lines: ChronologicalLogLine[];
  isLive: boolean;
  search: LogSearchSession;
}>();

const scroller = ref<HTMLElement | null>(null);
const stickToBottom = ref(true);

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
} = useLogSearch(
  toRef(props, 'lines'),
  scroller,
  props.search,
  () => {
    stickToBottom.value = false;
  },
);

function onScroll() {
  const element = scroller.value;
  if (!element) {
    return;
  }
  const remaining = element.scrollHeight - element.scrollTop - element.clientHeight;
  stickToBottom.value = remaining < 48;
}

watch(
  () => [props.lines.length, props.isLive, isSearching.value] as const,
  async () => {
    if (isSearching.value || !stickToBottom.value || !props.isLive) {
      return;
    }
    await nextTick();
    const element = scroller.value;
    if (element) {
      element.scrollTop = element.scrollHeight;
    }
  },
);
</script>

<template>
  <div
    data-testid="run-terminal"
    class="flex min-h-0 flex-1 flex-col bg-terminal text-foreground"
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
      @scroll="onScroll"
    >
      <p
        v-if="lines.length === 0"
        class="p-4 text-xs text-muted-foreground"
      >
        {{ isLive ? 'Waiting for output…' : 'No log lines' }}
      </p>
      <LogMarkedLines
        v-else
        :lines="renderedLines"
        :active-match="activeMatch"
      />
      <p
        v-if="isLive && lines.length > 0"
        class="border-t border-secondary px-4 py-2 text-xs text-muted-foreground"
      >
        streaming…
      </p>
    </div>
  </div>
</template>
