import type { DoctorEvent, RunnerEvent, Workflow, WorkflowInput } from '@/domain/runner-events';
import { computed, onMounted, onUnmounted, ref } from 'vue';

function hasRunFinished(events: RunnerEvent[]): boolean {
  return events.some((event) => event.type === 'run_finished');
}

function defaultInputValue(input: WorkflowInput): string {
  if (input.type === 'boolean') {
    return input.default === 'true'
      ? 'true'
      : 'false';
  }
  return input.default ?? '';
}

function seedInputValues(inputs: WorkflowInput[]): Record<string, string> {
  const values: Record<string, string> = {};
  for (const input of inputs) {
    values[input.id] = defaultInputValue(input);
  }
  return values;
}

function missingRequiredInputs(
  inputs: WorkflowInput[],
  values: Record<string, string>,
): string[] {
  return inputs
    .filter((input) => {
      if (!input.required) {
        return false;
      }
      if (input.type === 'boolean') {
        return false;
      }
      return !(values[input.id] ?? '').trim();
    })
    .map((input) => input.id);
}

function toInputArgs(
  inputs: WorkflowInput[],
  values: Record<string, string>,
): string[] {
  return inputs
    .map((input) => {
      const value = values[input.id] ?? '';
      if (value === '' && input.type !== 'boolean') {
        return null;
      }
      return `${input.id}=${value}`;
    })
    .filter((entry): entry is string => entry != null);
}

export function usePipeboxSession() {
  const repoPath = ref<string | null>(null);
  const branch = ref<string | null>(null);
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
  const inputValues = ref<Record<string, string>>({});
  const inputValuesWorkflow = ref<string | null>(null);

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

  const selectedWorkflowInputs = computed(() => {
    const workflow = workflows.value.find(
      (item) => item.path === selectedWorkflow.value,
    );
    return workflow?.inputs ?? [];
  });

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
    void openLaunchRepo();
  });

  onUnmounted(() => {
    unsubscribe?.();
  });

  function ensureInputValues(workflowPath: string) {
    if (inputValuesWorkflow.value === workflowPath) {
      return;
    }
    const workflow = workflows.value.find((item) => item.path === workflowPath);
    inputValues.value = seedInputValues(workflow?.inputs ?? []);
    inputValuesWorkflow.value = workflowPath;
  }

  function setInputValue(id: string, value: string) {
    inputValues.value = {
      ...inputValues.value,
      [id]: value,
    };
  }

  async function openRepo(nextPath: string) {
    statusMessage.value = '';
    repoPath.value = nextPath;
    branch.value = null;
    selectedWorkflow.value = null;
    runningWorkflow.value = null;
    runEventsWorkflow.value = null;
    previewWorkflow.value = null;
    runEvents.value = [];
    previewEvents.value = [];
    isRunning.value = false;
    inputValues.value = {};
    inputValuesWorkflow.value = null;
    const [list, nextBranch] = await Promise.all([
      window.pipebox.listWorkflows(nextPath),
      window.pipebox.gitBranch(nextPath),
    ]);
    workflows.value = list.workflows;
    branch.value = nextBranch;
  }

  async function selectRepo() {
    const nextPath = await window.pipebox.selectRepo();
    if (!nextPath) {
      return;
    }
    await openRepo(nextPath);
  }

  async function openLaunchRepo() {
    const nextPath = await window.pipebox.getLaunchRepo();
    if (!nextPath) {
      return;
    }
    await openRepo(nextPath);
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
    ensureInputValues(workflowPath);
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
    ensureInputValues(workflowPath);

    const workflow = workflows.value.find((item) => item.path === workflowPath);
    const inputs = workflow?.inputs ?? [];
    const missing = missingRequiredInputs(inputs, inputValues.value);
    if (missing.length > 0) {
      statusMessage.value = `Required inputs: ${missing.join(', ')}`;
      return;
    }

    runningWorkflow.value = workflowPath;
    runEventsWorkflow.value = workflowPath;
    runEvents.value = [];
    isRunning.value = true;
    try {
      await window.pipebox.startRun({
        repoPath: repoPath.value,
        workflow: workflowPath,
        inputs: toInputArgs(inputs, inputValues.value),
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
  };
}
