<script setup lang="ts">
import { FolderGit2 } from 'lucide-vue-next';

defineProps<{
  repoPath: string | null;
  branch: string | null;
}>();
</script>

<template>
  <div
    class="flex max-w-md items-center gap-2 text-xs text-muted-foreground"
    :title="[repoPath, branch].filter(Boolean).join(' · ') || undefined"
  >
    <FolderGit2
      class="size-4 shrink-0"
      aria-hidden="true"
    />
    <span class="min-w-0 truncate">
      {{ repoPath ?? 'No project open' }}
    </span>
    <span
      v-if="branch"
      class="max-w-32 shrink-0 truncate"
    >
      · {{ branch }}
    </span>
  </div>
</template>
