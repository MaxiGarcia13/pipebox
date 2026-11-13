<script setup lang="ts">
import AppHeader from '@/components/app-header.vue';
import DoctorStatus from '@/components/doctor-status.vue';
import EventLog from '@/components/event-log.vue';
import RunToolbar from '@/components/run-toolbar.vue';
import WorkflowList from '@/components/workflow-list.vue';
import { usePipeboxSession } from '@/composables/use-pipebox-session';

const {
  repoPath,
  workflows,
  selectedWorkflow,
  doctor,
  events,
  isRunning,
  statusMessage,
  selectRepo,
  runDoctor,
  startRun,
  stopRun,
} = usePipeboxSession();
</script>

<template>
  <div class="flex flex-col gap-6 p-6">
    <AppHeader />

    <RunToolbar
      :can-run="Boolean(repoPath && selectedWorkflow)"
      :is-running="isRunning"
      @select-repo="selectRepo"
      @doctor="runDoctor"
      @start-run="startRun"
      @stop-run="stopRun"
    />

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

    <DoctorStatus
      v-if="doctor"
      :doctor="doctor"
    />

    <WorkflowList
      v-if="workflows.length > 0"
      :workflows="workflows"
      :selected-path="selectedWorkflow"
      @select="selectedWorkflow = $event"
    />

    <EventLog
      v-if="events.length > 0"
      :events="events"
    />
  </div>
</template>
