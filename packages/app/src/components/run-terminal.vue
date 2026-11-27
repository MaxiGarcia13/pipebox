<script setup lang="ts">
import type { ChronologicalLogLine } from '@/domain/run';
import { ChevronDown, ChevronUp, X } from 'lucide-vue-next';
import { nextTick, ref, toRef, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useLogSearch } from '@/composables/use-log-search';

const props = defineProps<{
  lines: ChronologicalLogLine[];
  isLive: boolean;
}>();

const root = ref<HTMLElement | null>(null);
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
  root,
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
    ref="root"
    data-testid="run-terminal"
    class="flex min-h-0 flex-1 flex-col bg-terminal text-foreground"
  >
    <div
      v-if="open"
      data-testid="run-terminal-search-bar"
      class="flex items-center gap-2 border-b border-border px-3 py-2"
    >
      <Input
        v-model="query"
        data-testid="run-terminal-search"
        type="search"
        placeholder="Find in logs…"
        class="h-8"
        @keydown="onSearchKeydown"
      />
      <span
        v-if="matchLabel"
        data-testid="run-terminal-search-count"
        class="shrink-0 text-xs text-muted-foreground tabular-nums"
      >
        {{ matchLabel }}
      </span>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="size-8 shrink-0 p-0"
        :disabled="matchCount === 0"
        aria-label="Previous match"
        data-testid="run-terminal-search-prev"
        @click="goPrev"
      >
        <ChevronUp class="size-4" />
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="size-8 shrink-0 p-0"
        :disabled="matchCount === 0"
        aria-label="Next match"
        data-testid="run-terminal-search-next"
        @click="goNext"
      >
        <ChevronDown class="size-4" />
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="size-8 shrink-0 p-0"
        aria-label="Close find"
        data-testid="run-terminal-search-close"
        @click="closeSearch"
      >
        <X class="size-4" />
      </Button>
    </div>

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
      <pre
        v-else
        class="overflow-x-auto p-4 font-mono text-xs leading-relaxed break-all whitespace-pre-wrap"
      ><span
        v-for="(line, lineIndex) in renderedLines"
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
      <p
        v-if="isLive && lines.length > 0"
        class="border-t border-secondary px-4 py-2 text-xs text-muted-foreground"
      >
        streaming…
      </p>
    </div>
  </div>
</template>
