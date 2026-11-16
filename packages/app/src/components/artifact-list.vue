<script setup lang="ts">
import type { Artifact } from '@/domain/runner-events';
import { ref } from 'vue';
import { Button } from '@/components/ui/button';
import { formatBytes } from '@/utils/format-bytes';

const props = defineProps<{
  artifacts: Artifact[];
}>();

const savingPath = ref<string | null>(null);
const statusMessage = ref('');

async function downloadArtifact(artifact: Artifact) {
  if (savingPath.value) {
    return;
  }
  statusMessage.value = '';
  savingPath.value = artifact.path;
  try {
    const defaultName = defaultDownloadName(artifact);
    const savedPath = await window.pipebox.saveArtifact({
      sourcePath: artifact.path,
      defaultName,
    });
    if (savedPath) {
      statusMessage.value = `Saved ${artifact.name}`;
    }
  } catch (error) {
    statusMessage.value = error instanceof Error
      ? error.message
      : String(error);
  } finally {
    savingPath.value = null;
  }
}

function defaultDownloadName(artifact: Artifact): string {
  const baseName = artifact.path.split(/[/\\]/).pop() ?? artifact.name;
  if (baseName.includes('.')) {
    return baseName;
  }
  return artifact.name;
}
</script>

<template>
  <section
    v-if="props.artifacts.length > 0"
    data-testid="artifact-list"
    class="border-t border-border p-4"
  >
    <h3 class="text-sm font-medium">
      Artifacts
    </h3>
    <ul class="mt-2 space-y-2">
      <li
        v-for="artifact in props.artifacts"
        :key="artifact.path"
        data-testid="artifact-row"
        class="flex items-center justify-between gap-4"
      >
        <div class="min-w-0">
          <p class="truncate text-sm">
            {{ artifact.name }}
          </p>
          <p class="text-xs text-muted-foreground">
            {{ formatBytes(artifact.size) }}
          </p>
        </div>
        <Button
          data-testid="artifact-download"
          variant="outline"
          size="sm"
          class="rounded-xl"
          :disabled="savingPath === artifact.path"
          @click="downloadArtifact(artifact)"
        >
          Download
        </Button>
      </li>
    </ul>
    <p
      v-if="statusMessage"
      class="mt-2 text-xs text-muted-foreground"
    >
      {{ statusMessage }}
    </p>
  </section>
</template>
