<script setup lang="ts">
import type { DoctorEvent, WorkflowProvider } from '@/domain/runner-events';
import { RefreshCw } from 'lucide-vue-next';
import { computed } from 'vue';
import { Button } from '@/components/ui/button';
import { relevantDoctorChecks } from '@/domain/relevant-doctor-checks';

const props = defineProps<{
  doctor: DoctorEvent | null;
  providers: WorkflowProvider[];
}>();

defineEmits<{
  refresh: [];
}>();

const checks = computed(() =>
  props.doctor
    ? relevantDoctorChecks(props.doctor.checks, props.providers)
    : [],
);

const isOk = computed(() =>
  checks.value.length > 0 && checks.value.every((check) => check.ok),
);
</script>

<template>
  <div
    data-testid="doctor-status"
    class="flex items-center gap-4"
  >
    <div class="flex shrink-0 items-center gap-2 text-sm font-medium">
      <span
        class="size-2 rounded-full"
        :class="doctor
          ? (isOk ? 'bg-chart-3' : 'bg-destructive')
          : 'bg-muted-foreground'"
      />
      Doctor
      <span
        data-testid="doctor-label"
        class="text-muted-foreground"
      >
        {{ doctor ? (isOk ? 'OK' : 'issues') : '…' }}
      </span>

      <Button
        data-testid="doctor-refresh"
        variant="ghost"
        size="icon-xs"
        class="rounded-xl text-muted-foreground"
        title="Refresh Doctor"
        @click="$emit('refresh')"
      >
        <RefreshCw class="size-4" />
      </Button>
    </div>

    <ul
      v-if="doctor"
      class="flex items-center gap-2"
    >
      <li
        v-for="check in checks"
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
