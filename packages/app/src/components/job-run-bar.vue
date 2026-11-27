<script setup lang="ts">
import { Play, RotateCcw } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';

withDefaults(defineProps<{
  label: 'Run' | 'Re-run';
  disabled: boolean;
  showRun?: boolean;
  showParallel?: boolean;
  parallelJobs?: boolean;
}>(), {
  showRun: true,
  showParallel: false,
  parallelJobs: false,
});

defineEmits<{
  'run': [];
  'update:parallelJobs': [value: boolean];
}>();
</script>

<template>
  <div
    data-testid="job-run-bar"
    class="flex items-center justify-between gap-4 border-b border-border px-4 py-2"
  >
    <label
      v-if="showParallel"
      data-testid="parallel-jobs-toggle"
      class="flex min-w-0 items-center gap-2 text-sm text-muted-foreground"
    >
      <Checkbox
        :checked="parallelJobs"
        :disabled="disabled"
        @update:checked="$emit('update:parallelJobs', $event === true)"
      />
      <span class="truncate">
        Run jobs in parallel
      </span>
    </label>
    <div
      v-else
      class="min-w-0 flex-1"
    />

    <Button
      v-if="showRun"
      data-testid="job-run-button"
      variant="outline"
      size="sm"
      class="shrink-0 rounded-xl"
      :disabled="disabled"
      @click="$emit('run')"
    >
      <RotateCcw
        v-if="label === 'Re-run'"
        class="size-4"
      />
      <Play
        v-else
        class="size-4"
      />
      {{ label }}
    </Button>
  </div>
</template>
