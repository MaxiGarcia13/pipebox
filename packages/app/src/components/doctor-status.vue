<script setup lang="ts">
import type { DoctorEvent } from '@/shared/runner-events';

defineProps<{
  doctor: DoctorEvent;
}>();
</script>

<template>
  <section class="flex flex-col gap-2">
    <h2 class="text-lg font-medium">
      Doctor {{ doctor.ok ? 'OK' : 'issues' }}
    </h2>
    <ul class="flex flex-col gap-2 text-sm">
      <li
        v-for="check in doctor.checks"
        :key="check.name"
      >
        {{ check.ok ? '✓' : '✗' }}
        {{ check.name }}
        <span
          v-if="check.detail"
          class="text-muted-foreground"
        >— {{ check.detail }}</span>
      </li>
    </ul>
  </section>
</template>
