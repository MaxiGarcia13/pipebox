import type { DoctorEvent, RunnerEvent, Workflow } from '@/shared/runner-events';
import { onMounted, onUnmounted, ref } from 'vue';

export function usePipeboxSession() {
  const repoPath = ref<string | null>(null);
  const workflows = ref<Workflow[]>([]);
  const selectedWorkflow = ref<string | null>(null);
  const runningWorkflow = ref<string | null>(null);
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
    events.value = [];
    isRunning.value = false;
    const list = await window.pipebox.listWorkflows(nextPath);
    workflows.value = list.workflows;
  }

  async function runDoctor() {
    statusMessage.value = '';
    doctor.value = await window.pipebox.doctor();
  }

  async function startRun(workflowPath: string) {
    if (!repoPath.value || !workflowPath || isRunning.value) {
      return;
    }
    statusMessage.value = '';
    selectedWorkflow.value = workflowPath;
    runningWorkflow.value = workflowPath;
    events.value = [];
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
  }

  return {
    repoPath,
    workflows,
    selectedWorkflow,
    runningWorkflow,
    doctor,
    events,
    isRunning,
    statusMessage,
    selectRepo,
    runDoctor,
    startRun,
    stopRun,
  };
}
