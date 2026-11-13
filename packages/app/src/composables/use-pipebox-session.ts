import type { DoctorEvent, RunnerEvent, Workflow } from '@/shared/runner-events';
import { computed, onMounted, onUnmounted, ref } from 'vue';

function hasRunFinished(events: RunnerEvent[]): boolean {
  return events.some((event) => event.type === 'run_finished');
}

export function usePipeboxSession() {
  const repoPath = ref<string | null>(null);
  const workflows = ref<Workflow[]>([]);
  const selectedWorkflow = ref<string | null>(null);
  const runningWorkflow = ref<string | null>(null);
  const runEventsWorkflow = ref<string | null>(null);
  const previewWorkflow = ref<string | null>(null);
  const doctor = ref<DoctorEvent | null>(null);
  const runEvents = ref<RunnerEvent[]>([]);
  const previewEvents = ref<RunnerEvent[]>([]);
  const isRunning = ref(false);
  const statusMessage = ref('');

  let unsubscribe: (() => void) | undefined;

  const viewingRun = computed(() =>
    selectedWorkflow.value != null
    && selectedWorkflow.value === runEventsWorkflow.value,
  );

  const events = computed(() =>
    viewingRun.value ? runEvents.value : previewEvents.value,
  );

  const logIsRunning = computed(() =>
    viewingRun.value && isRunning.value,
  );

  onMounted(() => {
    unsubscribe = window.pipebox.onEvent((event) => {
      if (event.type === 'run_finished' && hasRunFinished(runEvents.value)) {
        return;
      }
      runEvents.value = [...runEvents.value, event];
      if (event.type === 'run_finished') {
        isRunning.value = false;
        runningWorkflow.value = null;
      }
    });
    void runDoctor();
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
    runningWorkflow.value = null;
    runEventsWorkflow.value = null;
    previewWorkflow.value = null;
    runEvents.value = [];
    previewEvents.value = [];
    isRunning.value = false;
    const list = await window.pipebox.listWorkflows(nextPath);
    workflows.value = list.workflows;
  }

  async function runDoctor() {
    statusMessage.value = '';
    doctor.value = await window.pipebox.doctor();
  }

  async function loadPreview(workflowPath: string) {
    if (!repoPath.value) {
      return;
    }
    if (previewWorkflow.value === workflowPath && previewEvents.value.length > 0) {
      return;
    }
    statusMessage.value = '';
    try {
      const graph = await window.pipebox.graphWorkflow({
        repoPath: repoPath.value,
        workflow: workflowPath,
      });
      previewEvents.value = [graph];
      previewWorkflow.value = workflowPath;
    } catch (error) {
      previewEvents.value = [];
      previewWorkflow.value = null;
      statusMessage.value = error instanceof Error
        ? error.message
        : String(error);
    }
  }

  async function selectWorkflow(workflowPath: string) {
    selectedWorkflow.value = workflowPath;
    if (workflowPath === runEventsWorkflow.value) {
      return;
    }
    await loadPreview(workflowPath);
  }

  async function startRun(workflowPath: string) {
    if (!repoPath.value || !workflowPath || isRunning.value) {
      return;
    }
    statusMessage.value = '';
    selectedWorkflow.value = workflowPath;
    runningWorkflow.value = workflowPath;
    runEventsWorkflow.value = workflowPath;
    runEvents.value = [];
    isRunning.value = true;
    try {
      await window.pipebox.startRun({
        repoPath: repoPath.value,
        workflow: workflowPath,
      });
    } catch (error) {
      isRunning.value = false;
      runningWorkflow.value = null;
      statusMessage.value = error instanceof Error
        ? error.message
        : String(error);
    }
  }

  async function stopRun() {
    await window.pipebox.stopRun();
    isRunning.value = false;
    runningWorkflow.value = null;
    if (!hasRunFinished(runEvents.value)) {
      runEvents.value = [
        ...runEvents.value,
        {
          type: 'run_finished',
          status: 'cancelled',
          exitCode: 130,
        },
      ];
    }
  }

  return {
    repoPath,
    workflows,
    selectedWorkflow,
    runningWorkflow,
    doctor,
    events,
    isRunning,
    logIsRunning,
    statusMessage,
    selectRepo,
    selectWorkflow,
    runDoctor,
    startRun,
    stopRun,
  };
}
