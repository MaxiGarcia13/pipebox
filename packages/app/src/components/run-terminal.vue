<script setup lang="ts">
import type { ChronologicalLogLine } from '@/domain/run';
import { nextTick, ref, watch } from 'vue';

const props = defineProps<{
  lines: ChronologicalLogLine[];
  isLive: boolean;
}>();

const root = ref<HTMLElement | null>(null);
const stickToBottom = ref(true);

function onScroll() {
  const element = root.value;
  if (!element) {
    return;
  }
  const remaining = element.scrollHeight - element.scrollTop - element.clientHeight;
  stickToBottom.value = remaining < 48;
}

watch(
  () => [props.lines.length, props.isLive] as const,
  async () => {
    if (!stickToBottom.value || !props.isLive) {
      return;
    }
    await nextTick();
    const element = root.value;
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
    class="min-h-0 flex-1 overflow-auto bg-terminal text-foreground"
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
      v-for="(line, index) in lines"
      :key="index"
      class="block"
      :class="line.kind === 'error' ? 'text-destructive' : ''"
    >{{ line.text }}</span></pre>
    <p
      v-if="isLive && lines.length > 0"
      class="border-t border-secondary px-4 py-2 text-xs text-muted-foreground"
    >
      streaming…
    </p>
  </div>
</template>
