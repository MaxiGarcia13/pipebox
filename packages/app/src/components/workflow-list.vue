<script setup lang="ts">
import type { Workflow } from '@/domain/runner-events';
import { Play, Square } from 'lucide-vue-next';
import ProviderIcon from '@/components/provider-icon.vue';
import { Button } from '@/components/ui/button';

defineProps<{
  workflows: Workflow[];
  hasRepo: boolean;
  selectedPath: string | null;
  runningPath: string | null;
  isRunning: boolean;
}>();

defineEmits<{
  openProject: [];
  select: [path: string];
  run: [path: string];
  stop: [];
}>();
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="border-b border-sidebar-border px-4 py-2">
      <h2 class="text-sm font-medium text-sidebar-foreground">
        Workflows
      </h2>
    </div>

    <div
      v-if="!hasRepo"
      class="flex flex-col items-center justify-center gap-4 p-4 text-center"
    >
      <div class="flex flex-col gap-2">
        <p class="text-sm font-medium text-sidebar-foreground">
          No project open
        </p>
        <p class="text-xs text-muted-foreground">
          Open a project to list its CI workflows.
        </p>
      </div>
      <Button
        variant="outline"
        class="w-full max-w-48 rounded-xl"
        @click="$emit('openProject')"
      >
        Open Project…
      </Button>
    </div>

    <p
      v-else-if="workflows.length === 0"
      class="p-4 text-sm text-muted-foreground"
    >
      No workflows found in this project
    </p>

    <ul
      v-else
      class="flex min-h-0 flex-1 flex-col overflow-auto py-2"
    >
      <li
        v-for="workflow in workflows"
        :key="`${workflow.provider}:${workflow.path}`"
      >
        <div
          class="group flex cursor-pointer items-center gap-2 p-2 text-sm transition-colors"
          :class="selectedPath === workflow.path
            ? 'bg-sidebar-accent text-sidebar-accent-foreground'
            : 'text-sidebar-foreground hover:bg-sidebar-accent/60'"
          @click="$emit('select', workflow.path)"
        >
          <ProviderIcon :provider="workflow.provider" />
          <div class="min-w-0 flex-1">
            <p class="truncate font-medium">
              {{ workflow.name }}
            </p>
            <p class="truncate text-xs text-muted-foreground">
              {{ workflow.path }}
            </p>
          </div>

          <div class="flex shrink-0 items-center gap-1">
            <button
              v-if="runningPath !== workflow.path"
              type="button"
              class="cursor-pointer rounded-md p-1 text-chart-3 transition-opacity enabled:hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-30"
              :disabled="isRunning"
              title="Run"
              aria-label="Run"
              @click.stop="$emit('run', workflow.path)"
            >
              <Play class="size-4 fill-current" />
            </button>
            <button
              v-else
              type="button"
              class="cursor-pointer rounded-md p-1 text-destructive transition-opacity hover:opacity-80"
              title="Stop"
              aria-label="Stop"
              @click.stop="$emit('stop')"
            >
              <Square class="size-3.5 fill-current" />
            </button>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>
