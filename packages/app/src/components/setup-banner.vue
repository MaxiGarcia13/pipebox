<script setup lang="ts">
import type { DoctorEvent } from '@/domain/runner-events';
import { computed } from 'vue';

const props = defineProps<{
  doctor: DoctorEvent | null;
}>();

const message = computed(() => {
  if (!props.doctor) {
    return null;
  }
  const checks = Object.fromEntries(
    props.doctor.checks.map((check) => [check.name, check]),
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
  if (!props.doctor.ok) {
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
