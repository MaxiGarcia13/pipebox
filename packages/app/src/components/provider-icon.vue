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
  <GithubIcon
    v-if="provider === 'github'"
    :class="cn('size-4 shrink-0', props.class)"
  />
  <JenkinsIcon
    v-else
    :class="cn('size-4 shrink-0', props.class)"
  />
</template>
