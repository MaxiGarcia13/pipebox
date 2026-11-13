<script setup lang="ts">
import type { DoctorEvent } from '@/domain/runner-events';

defineProps<{
  doctor: DoctorEvent | null;
}>();
</script>

<template>
  <div class="flex items-center gap-4">
    <div class="flex shrink-0 items-center gap-2 text-sm font-medium">
      <span
        class="size-2 rounded-full"
        :class="doctor?.ok ? 'bg-chart-3' : 'bg-destructive'"
      />
      Doctor
      <span class="text-muted-foreground">
        {{ doctor ? (doctor.ok ? 'OK' : 'issues') : '…' }}
      </span>
    </div>

    <ul
      v-if="doctor"
      class="flex items-center gap-2"
    >
      <li
        v-for="check in doctor.checks"
        :key="check.name"
        class="rounded-xl border border-border px-4 py-2 text-xs whitespace-nowrap"
        :title="check.detail || check.path || check.name"
      >
        <span :class="check.ok ? 'text-chart-3' : 'text-destructive'">
          {{ check.ok ? '✓' : '✗' }}
        </span>
        {{ check.name }}
        <span
          v-if="check.detail"
          class="text-muted-foreground"
        >
          — {{ check.detail }}
        </span>
      </li>
    </ul>

    <p
      v-else
      class="text-xs text-muted-foreground"
    >
      Checking Docker…
    </p>
  </div>
</template>
