<script setup lang="ts">
import type { WorkflowProvider } from '@/domain/runner-events';
import GithubIcon from '@/components/icons/github-icon.vue';
import JenkinsIcon from '@/components/icons/jenkins-icon.vue';
import { cn } from '@/utils';

const props = defineProps<{
  provider: WorkflowProvider;
  class?: string;
}>();
</script>

<template>
  <span
    data-testid="provider-icon"
    :data-provider="provider"
    :class="cn('inline-flex size-4 shrink-0 items-center justify-center', props.class)"
  >
    <GithubIcon
      v-if="provider === 'github'"
      class="size-4"
    />
    <JenkinsIcon
      v-else
      class="size-4"
    />
  </span>
</template>
