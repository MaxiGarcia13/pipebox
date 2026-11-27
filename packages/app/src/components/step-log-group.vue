<script setup lang="ts">
import type { SearchLinePart, StepGroup } from '@/domain/run';
import { ChevronRight } from 'lucide-vue-next';
import { computed } from 'vue';
import LogMarkedLines from '@/components/log-marked-lines.vue';
import { LOG_SURFACE_CLASS } from '@/components/log-surface';
import RunStatusIcon from '@/components/run-status-icon.vue';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Skeleton } from '@/components/ui/skeleton';
import { visibleDurationMs } from '@/domain/run';
import { formatDuration } from '@/utils';

const props = defineProps<{
  group: StepGroup;
  now?: number;
  lineParts?: SearchLinePart[][];
  activeMatch?: number;
  forceOpen?: boolean;
}>();

// Remount when search forces open so defaultOpen applies (controlled v-model
// through the collapsible wrapper does not reliably expand closed panels).
const collapsibleKey = computed(() =>
  `${props.group.key}:${props.forceOpen ? 'search' : 'idle'}`,
);

const startOpen = computed(() =>
  Boolean(props.forceOpen || props.group.openDefault),
);

function stepDuration(group: StepGroup): number | undefined {
  return visibleDurationMs(group, props.now ?? 0, group.status === 'running');
}

const markedLines = computed(() => {
  if (props.lineParts && props.lineParts.length > 0) {
    return props.lineParts.map((parts) => ({ parts }));
  }
  return props.group.lines.map((text) => ({
    parts: [
      { text, matchIndex: null as number | null },
    ],
  }));
});
</script>

<template>
  <div
    v-if="group.status === 'pending'"
    data-testid="step-log-group"
    class="flex items-center gap-2 border-b border-border px-4 py-2 text-sm text-muted-foreground opacity-50"
    aria-disabled="true"
  >
    <ChevronRight
      class="size-4 shrink-0"
      aria-hidden="true"
    />
    <RunStatusIcon :status="group.status" />
    <span class="min-w-0 flex-1 truncate font-medium">
      {{ group.step }}
    </span>
  </div>

  <Collapsible
    v-else
    v-slot="{ open }"
    :key="collapsibleKey"
    :default-open="startOpen"
    data-testid="step-log-group"
    class="border-b border-border"
  >
    <CollapsibleTrigger
      class="flex w-full cursor-pointer items-center gap-2 px-4 py-2 text-left text-sm hover:bg-muted/40"
    >
      <ChevronRight
        class="size-4 shrink-0 text-muted-foreground transition-transform"
        :class="open ? 'rotate-90' : ''"
        aria-hidden="true"
      />
      <RunStatusIcon :status="group.status" />
      <span class="min-w-0 flex-1 truncate font-medium">
        {{ group.step }}
      </span>
      <span
        v-if="formatDuration(stepDuration(group))"
        class="shrink-0 font-mono text-xs text-muted-foreground"
      >
        {{ formatDuration(stepDuration(group)) }}
      </span>
    </CollapsibleTrigger>

    <CollapsibleContent>
      <div
        v-if="group.lines.length > 0"
        :class="LOG_SURFACE_CLASS"
      >
        <LogMarkedLines
          :lines="markedLines"
          :active-match="activeMatch ?? -1"
        />
      </div>
      <p
        v-else-if="group.status !== 'running'"
        class="px-4 py-2 text-xs text-muted-foreground"
      >
        No log lines
      </p>
      <div
        v-if="group.status === 'running'"
        class="flex flex-row gap-2 px-4 py-2"
        aria-label="Waiting for step output"
      >
        <Skeleton class="h-3 w-30" />
        <Skeleton class="h-3 w-full" />
      </div>
    </CollapsibleContent>
  </Collapsible>
</template>
