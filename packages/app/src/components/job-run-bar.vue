<script setup lang="ts">
import { Play, RotateCcw } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';

defineProps<{
  label: 'Run' | 'Re-run';
  disabled: boolean;
}>();

defineEmits<{
  run: [];
}>();
</script>

<template>
  <div
    data-testid="job-run-bar"
    class="flex items-center justify-end border-b border-border px-4 py-2"
  >
    <Button
      data-testid="job-run-button"
      variant="outline"
      size="sm"
      class="rounded-xl"
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
