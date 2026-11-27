<script setup lang="ts">
import { Check } from 'lucide-vue-next';

const props = defineProps<{
  checked: boolean;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  'update:checked': [value: boolean];
}>();

function toggle() {
  if (props.disabled) {
    return;
  }
  emit('update:checked', !props.checked);
}
</script>

<template>
  <div
    data-testid="parallel-jobs-toggle"
    role="checkbox"
    class="flex min-w-0 items-center gap-2 text-sm text-foreground"
    :class="disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'"
    :aria-checked="checked"
    :aria-disabled="disabled || undefined"
    :tabindex="disabled ? -1 : 0"
    @click="toggle"
    @keydown.enter.prevent="toggle"
    @keydown.space.prevent="toggle"
  >
    <span
      data-slot="checkbox"
      class="flex size-4 shrink-0 items-center justify-center rounded-sm border shadow-xs"
      :class="checked
        ? 'border-primary bg-primary text-primary-foreground'
        : 'border-input dark:bg-input/30'"
      :data-state="checked ? 'checked' : 'unchecked'"
    >
      <Check
        v-if="checked"
        class="size-3.5"
      />
    </span>
    <span class="truncate">
      Run jobs in parallel
    </span>
  </div>
</template>
