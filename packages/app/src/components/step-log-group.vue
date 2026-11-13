<script setup lang="ts">
import type { StepGroup } from '@/utils';
import { ChevronRight } from 'lucide-vue-next';
import RunStatusIcon from '@/components/run-status-icon.vue';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDuration } from '@/utils';

defineProps<{
  group: StepGroup;
}>();
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
        v-if="formatDuration(group.durationMs)"
        class="shrink-0 font-mono text-xs text-muted-foreground"
      >
        {{ formatDuration(group.durationMs) }}
      </span>
    </CollapsibleTrigger>

    <CollapsibleContent>
      <pre
        v-if="group.lines.length > 0"
        class="overflow-x-auto bg-background px-4 py-2 font-mono text-xs leading-relaxed break-all whitespace-pre-wrap text-foreground"
      >{{ group.lines.join('\n') }}</pre>
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
