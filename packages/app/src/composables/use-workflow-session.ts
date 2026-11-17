import type { Ref } from 'vue';
import type { RunnerEvent, Workflow } from '@/domain/runner-events';
import { computed, ref } from 'vue';
import { seedInputValues } from '@/domain/workflow-inputs';

export function useWorkflowSession(workflows: Ref<Workflow[]>) {
  const selectedWorkflow = ref<string | null>(null);
  const previewWorkflow = ref<string | null>(null);
  const previewEvents = ref<RunnerEvent[]>([]);
  const inputValues = ref<Record<string, string>>({});
  const inputValuesWorkflow = ref<string | null>(null);

  const selectedWorkflowInputs = computed(() => {
    const workflow = workflows.value.find(
      (item) => item.path === selectedWorkflow.value,
    );
    return workflow?.inputs ?? [];
  });

  function inputsFor(workflowPath: string) {
    return workflows.value.find((item) => item.path === workflowPath)?.inputs ?? [];
  }

  function reset() {
    selectedWorkflow.value = null;
    previewWorkflow.value = null;
    previewEvents.value = [];
    inputValues.value = {};
    inputValuesWorkflow.value = null;
  }

  function ensureInputValues(workflowPath: string) {
    if (inputValuesWorkflow.value === workflowPath) {
      return;
    }
    inputValues.value = seedInputValues(inputsFor(workflowPath));
    inputValuesWorkflow.value = workflowPath;
  }

  function select(workflowPath: string) {
    selectedWorkflow.value = workflowPath;
    ensureInputValues(workflowPath);
  }

  function setInputValue(id: string, value: string) {
    inputValues.value = {
      ...inputValues.value,
      [id]: value,
    };
  }

  async function loadPreview(
    repoPath: string,
    workflowPath: string,
  ): Promise<
    | { status: 'cached' }
    | { status: 'loaded' }
    | { status: 'error'; message: string }
  > {
    if (previewWorkflow.value === workflowPath && previewEvents.value.length > 0) {
      return { status: 'cached' };
    }
    try {
      const graph = await window.pipebox.graphWorkflow({
        repoPath,
        workflow: workflowPath,
      });
      previewEvents.value = [graph];
      previewWorkflow.value = workflowPath;
      return { status: 'loaded' };
    } catch (error) {
      previewEvents.value = [];
      previewWorkflow.value = null;
      return {
        status: 'error',
        message: error instanceof Error
          ? error.message
          : String(error),
      };
    }
  }

  return {
    selectedWorkflow,
    previewEvents,
    inputValues,
    selectedWorkflowInputs,
    inputsFor,
    reset,
    select,
    setInputValue,
    loadPreview,
  };
}
