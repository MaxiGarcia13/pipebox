<script setup lang="ts">
import { FolderGit2 } from 'lucide-vue-next';

defineProps<{
  repoPath: string | null;
}>();
</script>

<template>
  <div
    class="flex max-w-md items-center gap-2 truncate text-xs text-muted-foreground"
    :title="repoPath ?? undefined"
  >
    <FolderGit2
      class="size-4 shrink-0"
      aria-hidden="true"
    />
    <span class="truncate">
      {{ repoPath ?? 'No project open' }}
    </span>
  </div>
</template>
