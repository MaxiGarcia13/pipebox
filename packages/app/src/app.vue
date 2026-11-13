<script setup lang="ts">
import type { DoctorEvent, RunnerEvent, Workflow } from '@/shared/runner-events';
import { onMounted, onUnmounted, ref } from 'vue';
import { Button } from '@/components/ui/button';

const repoPath = ref<string | null>(null);
const workflows = ref<Workflow[]>([]);
const selectedWorkflow = ref<string | null>(null);
const doctor = ref<DoctorEvent | null>(null);
const events = ref<RunnerEvent[]>([]);
const isRunning = ref(false);
const statusMessage = ref('');

let unsubscribe: (() => void) | undefined;

onMounted(() => {
  unsubscribe = window.pipebox.onEvent((event) => {
    events.value = [...events.value, event];
    if (event.type === 'run_finished') {
      isRunning.value = false;
    }
  });
});

onUnmounted(() => {
  unsubscribe?.();
});

async function selectRepo() {
  statusMessage.value = '';
  const nextPath = await window.pipebox.selectRepo();
  if (!nextPath) {
    return;
  }
  repoPath.value = nextPath;
  selectedWorkflow.value = null;
  const list = await window.pipebox.listWorkflows(nextPath);
  workflows.value = list.workflows;
}

async function runDoctor() {
  statusMessage.value = '';
  doctor.value = await window.pipebox.doctor();
}

async function startRun() {
  if (!repoPath.value || !selectedWorkflow.value) {
    return;
  }
  statusMessage.value = '';
  events.value = [];
  isRunning.value = true;
  try {
    await window.pipebox.startRun({
      repoPath: repoPath.value,
      workflow: selectedWorkflow.value,
    });
  } catch (error) {
    isRunning.value = false;
    statusMessage.value = error instanceof Error
      ? error.message
      : String(error);
  }
}

async function stopRun() {
  await window.pipebox.stopRun();
  isRunning.value = false;
}
</script>

<template>
  <div class="flex flex-col gap-6 p-6">
    <header class="flex flex-col gap-2">
      <h1 class="text-2xl font-semibold">
        PipeBox
      </h1>
      <p class="text-muted-foreground">
        Local CI control panel
      </p>
    </header>

    <section class="flex flex-wrap gap-2">
      <Button @click="selectRepo">
        Select repo
      </Button>
      <Button
        variant="secondary"
        @click="runDoctor"
      >
        Doctor
      </Button>
      <Button
        :disabled="!repoPath || !selectedWorkflow || isRunning"
        @click="startRun"
      >
        Run
      </Button>
      <Button
        variant="destructive"
        :disabled="!isRunning"
        @click="stopRun"
      >
        Stop
      </Button>
    </section>

    <p
      v-if="repoPath"
      class="text-sm break-all"
    >
      Repo: {{ repoPath }}
    </p>
    <p
      v-if="statusMessage"
      class="text-sm text-destructive"
    >
      {{ statusMessage }}
    </p>

    <section
      v-if="doctor"
      class="flex flex-col gap-2"
    >
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

    <section
      v-if="workflows.length > 0"
      class="flex flex-col gap-2"
    >
      <h2 class="text-lg font-medium">
        Workflows
      </h2>
      <ul class="flex flex-col gap-2">
        <li
          v-for="workflow in workflows"
          :key="workflow.path"
        >
          <Button
            :variant="selectedWorkflow === workflow.path ? 'default' : 'outline'"
            @click="selectedWorkflow = workflow.path"
          >
            {{ workflow.name }}
          </Button>
        </li>
      </ul>
    </section>

    <section
      v-if="events.length > 0"
      class="flex flex-col gap-2"
    >
      <h2 class="text-lg font-medium">
        Events
      </h2>
      <pre class="overflow-auto rounded-xl bg-secondary p-4 text-xs text-secondary-foreground">{{ events.map((event) => JSON.stringify(event)).join('\n') }}</pre>
    </section>
  </div>
</template>
