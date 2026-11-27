<script setup lang="ts">
import type { StepGroup } from '@/domain/run';
import { ChevronRight } from 'lucide-vue-next';
import { LOG_PRE_CLASS, LOG_SURFACE_CLASS } from '@/components/log-surface';
import RunStatusIcon from '@/components/run-status-icon.vue';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Skeleton } from '@/components/ui/skeleton';
import { visibleDurationMs } from '@/domain/run';
import { cn, formatDuration } from '@/utils';

const props = defineProps<{
  group: StepGroup;
  now?: number;
}>();

function stepDuration(group: StepGroup): number | undefined {
  return visibleDurationMs(group, props.now ?? 0, group.status === 'running');
}
</script>

<template>
  <div
    v-if="group.status === 'pending'"
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
    :default-open="group.openDefault"
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
        :class="cn('mx-4 mb-4 overflow-hidden', LOG_SURFACE_CLASS)"
      >
        <pre :class="cn(LOG_PRE_CLASS, 'text-foreground')">{{ group.lines.join('\n') }}</pre>
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
