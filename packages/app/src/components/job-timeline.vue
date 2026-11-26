<script setup lang="ts">
import type { JobSummary } from '@/domain/run';
import { computed } from 'vue';
import JobGraphConnector from '@/components/job-graph-connector.vue';
import RunStatusIcon from '@/components/run-status-icon.vue';
import { groupJobsByColumn, visibleDurationMs } from '@/domain/run';
import { formatDuration } from '@/utils';

const props = defineProps<{
  jobs: JobSummary[];
  selectedId: string | null;
  now?: number;
}>();

defineEmits<{
  select: [id: string];
}>();

const columns = computed(() => groupJobsByColumn(props.jobs));

function jobDuration(job: JobSummary): number | undefined {
  return visibleDurationMs(job, props.now ?? 0, job.status === 'running');
}
</script>

<template>
  <div
    v-if="columns.length > 0"
    class="overflow-x-auto border-b border-border p-4"
  >
    <div class="flex min-w-full items-center">
      <template
        v-for="(column, columnIndex) in columns"
        :key="columnIndex"
      >
        <div class="flex min-w-48 flex-col gap-4">
          <button
            v-for="job in column"
            :key="job.id"
            type="button"
            data-testid="job-pill"
            :data-job-id="job.id"
            class="flex h-10 w-full max-w-64 min-w-48 shrink-0 cursor-pointer items-center gap-2 rounded-xl border bg-card p-2 text-left text-sm shadow-sm transition-colors hover:bg-muted/40"
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
              v-if="formatDuration(jobDuration(job))"
              class="shrink-0 font-mono text-xs text-muted-foreground"
            >
              {{ formatDuration(jobDuration(job)) }}
            </span>
          </button>
        </div>

        <JobGraphConnector
          v-if="columnIndex < columns.length - 1"
          :from="column"
          :to="columns[columnIndex + 1] ?? []"
        />
      </template>
    </div>
  </div>
</template>
