<script setup lang="ts">
import { Button } from '@/components/ui/button';

defineProps<{
  canRun: boolean;
  isRunning: boolean;
}>();

defineEmits<{
  selectRepo: [];
  doctor: [];
  startRun: [];
  stopRun: [];
}>();
</script>

<template>
  <section class="flex flex-wrap gap-2">
    <Button @click="$emit('selectRepo')">
      Select repo
    </Button>
    <Button
      variant="secondary"
      @click="$emit('doctor')"
    >
      Doctor
    </Button>
    <Button
      :disabled="!canRun || isRunning"
      @click="$emit('startRun')"
    >
      Run
    </Button>
    <Button
      variant="destructive"
      :disabled="!isRunning"
      @click="$emit('stopRun')"
    >
      Stop
    </Button>
  </section>
</template>
