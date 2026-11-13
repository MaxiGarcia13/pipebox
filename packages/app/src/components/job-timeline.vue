<script setup lang="ts">
import type { JobSummary } from '@/utils';
import { ChevronRight } from 'lucide-vue-next';
import { computed } from 'vue';
import RunStatusIcon from '@/components/run-status-icon.vue';
import { formatDuration, groupJobsByColumn } from '@/utils';

const props = defineProps<{
  jobs: JobSummary[];
  selectedId: string | null;
}>();

defineEmits<{
  select: [id: string];
}>();

const columns = computed(() => groupJobsByColumn(props.jobs));
</script>

<template>
  <div
    v-if="columns.length > 0"
    class="overflow-x-auto border-b border-border p-4"
  >
    <div class="flex min-w-full items-stretch gap-4">
      <template
        v-for="(column, columnIndex) in columns"
        :key="columnIndex"
      >
        <div class="flex min-w-48 flex-col gap-4">
          <button
            v-for="job in column"
            :key="job.id"
            type="button"
            class="flex w-full max-w-64 min-w-48 cursor-pointer items-center gap-2 rounded-xl border bg-card p-2 text-left text-sm shadow-sm transition-colors hover:bg-muted/40"
            :class="selectedId === job.id
              ? 'border-primary ring-2 ring-ring/40'
              : 'border-border'"
            @click="$emit('select', job.id)"
          >
            <RunStatusIcon :status="job.status" />
            <span class="min-w-0 flex-1 truncate font-medium">
              {{ job.name }}
            </span>
            <span
              v-if="formatDuration(job.durationMs)"
              class="shrink-0 font-mono text-xs text-muted-foreground"
            >
              {{ formatDuration(job.durationMs) }}
            </span>
          </button>
        </div>

        <div
          v-if="columnIndex < columns.length - 1"
          class="flex shrink-0 items-center"
          aria-hidden="true"
        >
          <ChevronRight class="size-4 text-muted-foreground" />
        </div>
      </template>
    </div>
  </div>
</template>
