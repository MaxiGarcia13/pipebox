<script setup lang="ts">
import DoctorStatus from '@/components/doctor-status.vue';
import EventLog from '@/components/event-log.vue';
import ProjectStatus from '@/components/project-status.vue';
import RunToolbar from '@/components/run-toolbar.vue';
import SetupBanner from '@/components/setup-banner.vue';
import WorkbenchShell from '@/components/workbench-shell.vue';
import WorkflowInputs from '@/components/workflow-inputs.vue';
import WorkflowList from '@/components/workflow-list.vue';
import { usePipeboxSession } from '@/composables/use-pipebox-session';

const {
  repoPath,
  branch,
  workflows,
  selectedWorkflow,
  runningWorkflow,
  doctor,
  events,
  isRunning,
  logIsRunning,
  statusMessage,
  selectedWorkflowInputs,
  inputValues,
  setInputValue,
  selectRepo,
  selectWorkflow,
  runDoctor,
  startRun,
  stopRun,
} = usePipeboxSession();
</script>

<template>
  <WorkbenchShell>
    <template #toolbar>
      <RunToolbar
        :status-message="statusMessage"
        @open-project="selectRepo"
        @refresh-doctor="runDoctor"
      />
    </template>

    <template #banner>
      <SetupBanner :doctor="doctor" />
    </template>

    <template #sidebar>
      <WorkflowList
        :workflows="workflows"
        :has-repo="Boolean(repoPath)"
        :selected-path="selectedWorkflow"
        :running-path="runningWorkflow"
        :is-running="isRunning"
        @open-project="selectRepo"
        @select="selectWorkflow($event)"
        @run="startRun"
        @stop="stopRun"
      />
    </template>

    <template #main>
      <div class="flex min-h-0 flex-1">
        <EventLog
          :key="`${selectedWorkflow ?? 'none'}:${logIsRunning ? 'live' : 'preview'}`"
          class="min-w-0 flex-1"
          :events="events"
          :is-running="logIsRunning"
        />
        <aside
          v-if="selectedWorkflowInputs.length > 0"
          class="flex w-72 shrink-0 flex-col border-l border-border"
        >
          <WorkflowInputs
            :inputs="selectedWorkflowInputs"
            :values="inputValues"
            :disabled="isRunning"
            @update="setInputValue"
          />
        </aside>
      </div>
    </template>

    <template #panel>
      <DoctorStatus
        :doctor="doctor"
        @refresh="runDoctor"
      />
    </template>

    <template #status>
      <div class="flex items-center gap-4">
        <span
          v-if="doctor?.version"
          class="shrink-0 text-xs text-muted-foreground tabular-nums"
        >
          v{{ doctor.version }}
        </span>
        <ProjectStatus
          :repo-path="repoPath"
          :branch="branch"
        />
      </div>
    </template>
  </WorkbenchShell>
</template>
