<script setup lang="ts">
import type { CollapsibleContentProps } from 'reka-ui';
import { CollapsibleContent, useForwardProps } from 'reka-ui';

const props = defineProps<CollapsibleContentProps>();
const forwardedProps = useForwardProps(props);
</script>

<template>
  <CollapsibleContent v-bind="forwardedProps">
    <slot />
  </CollapsibleContent>
</template>
