<script setup lang="ts">
import type { DoctorEvent, WorkflowProvider } from '@/domain/runner-events';
import { computed } from 'vue';
import { relevantDoctorChecks } from '@/domain/relevant-doctor-checks';

const props = defineProps<{
  doctor: DoctorEvent | null;
  providers: WorkflowProvider[];
}>();

const message = computed(() => {
  if (!props.doctor) {
    return null;
  }
  const relevant = relevantDoctorChecks(props.doctor.checks, props.providers);
  const checks = Object.fromEntries(
    relevant.map((check) => [check.name, check]),
  );
  if (checks.docker && !checks.docker.ok) {
    return 'Install Docker and ensure it is on your PATH to run workflows locally.';
  }
  if (checks.docker_daemon && !checks.docker_daemon.ok) {
    return 'Docker is installed but the daemon is not reachable. Start Docker Desktop (or your Docker service), then refresh doctor.';
  }
  if (checks.act_image && !checks.act_image.ok) {
    return 'Act image is not built yet. The first workflow run will build pipebox-act locally.';
  }
  if (checks.jenkins_image && !checks.jenkins_image.ok) {
    return 'Jenkins image is not built yet. The first workflow run will build pipebox-jenkins locally.';
  }
  if (relevant.some((check) => !check.ok)) {
    return 'Doctor reported issues. Fix Docker setup before running workflows.';
  }
  return null;
});
</script>

<template>
  <div
    v-if="message"
    data-testid="setup-banner"
    class="border-b border-border bg-secondary/40 px-4 py-2 text-sm text-foreground"
  >
    {{ message }}
  </div>
</template>
