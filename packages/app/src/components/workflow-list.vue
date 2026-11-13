<script setup lang="ts">
import type { Workflow } from '@/shared/runner-events';
import { Button } from '@/components/ui/button';

defineProps<{
  workflows: Workflow[];
  selectedPath: string | null;
}>();

defineEmits<{
  select: [path: string];
}>();
</script>

<template>
  <section class="flex flex-col gap-2">
    <h2 class="text-lg font-medium">
      Workflows
    </h2>
    <ul class="flex flex-col gap-2">
      <li
        v-for="workflow in workflows"
        :key="workflow.path"
      >
        <Button
          :variant="selectedPath === workflow.path ? 'default' : 'outline'"
          @click="$emit('select', workflow.path)"
        >
          {{ workflow.name }}
        </Button>
      </li>
    </ul>
  </section>
</template>
