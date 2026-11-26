<script setup lang="ts">
import DoctorStatus from '@/components/doctor-status.vue';
import EventLog from '@/components/event-log.vue';
import ProjectStatus from '@/components/project-status.vue';
import RunToolbar from '@/components/run-toolbar.vue';
import SetupBanner from '@/components/setup-banner.vue';
import SwitchProjectDialog from '@/components/switch-project-dialog.vue';
import WorkbenchShell from '@/components/workbench-shell.vue';
import WorkflowInputs from '@/components/workflow-inputs.vue';
import WorkflowList from '@/components/workflow-list.vue';
import { usePipeboxSession } from '@/composables/use-pipebox-session';

const {
  repoPath,
  branch,
  workflows,
  providers,
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
  refreshWorkflows,
  startRun,
  startJob,
  stopRun,
  switchProjectOpen,
  confirmSwitchProject,
  cancelSwitchProject,
} = usePipeboxSession();

function onRunJob(jobId: string) {
  if (!selectedWorkflow.value) {
    return;
  }
  void startJob(selectedWorkflow.value, jobId);
}
</script>

<template>
  <SwitchProjectDialog
    v-model:open="switchProjectOpen"
    @confirm="confirmSwitchProject"
    @cancel="cancelSwitchProject"
  />
  <WorkbenchShell>
    <template #toolbar>
      <RunToolbar
        :status-message="statusMessage"
        @open-project="selectRepo"
        @refresh-doctor="runDoctor"
      />
    </template>

    <template #banner>
      <SetupBanner
        :doctor="doctor"
        :providers="providers"
      />
    </template>

    <template #sidebar>
      <WorkflowList
        :workflows="workflows"
        :has-repo="Boolean(repoPath)"
        :selected-path="selectedWorkflow"
        :running-path="runningWorkflow"
        :is-running="isRunning"
        @open-project="selectRepo"
        @refresh="refreshWorkflows"
        @select="selectWorkflow($event)"
        @run="startRun"
        @stop="stopRun"
      />
    </template>

    <template #main>
      <div class="flex min-h-0 flex-1">
        <EventLog
          :key="selectedWorkflow ?? 'none'"
          class="min-w-0 flex-1"
          :events="events"
          :is-running="logIsRunning"
          @run-job="onRunJob"
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
        :providers="providers"
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
