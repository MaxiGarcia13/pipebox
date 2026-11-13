<script setup lang="ts">
import logoUrl from '@/assets/logo-32.png';
import AppMenubar from '@/components/app-menubar.vue';

defineProps<{
  statusMessage: string;
}>();

defineEmits<{
  openProject: [];
  refreshDoctor: [];
}>();
</script>

<template>
  <div class="flex min-w-0 flex-1 items-center gap-2">
    <div class="flex shrink-0 items-center gap-2">
      <img
        :src="logoUrl"
        alt=""
        class="size-6"
      >
      <span class="text-sm font-semibold tracking-wide">
        PipeBox
      </span>
    </div>

    <AppMenubar
      @open-project="$emit('openProject')"
      @refresh-doctor="$emit('refreshDoctor')"
    />

    <p
      v-if="statusMessage"
      class="min-w-0 flex-1 truncate text-sm text-destructive"
    >
      {{ statusMessage }}
    </p>
  </div>
</template>
