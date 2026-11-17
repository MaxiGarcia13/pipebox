import { computed, onMounted, onUnmounted, ref } from 'vue';
import { missingRequiredInputs, toInputArgs } from '@/domain/workflow-inputs';
import { useDoctorSession } from './use-doctor-session';
import { useProjectSession } from './use-project-session';
import { useRunSession } from './use-run-session';
import { useWorkflowSession } from './use-workflow-session';

export function usePipeboxSession() {
  const statusMessage = ref('');
  const project = useProjectSession();
  const doctor = useDoctorSession();
  const workflow = useWorkflowSession(project.workflows);
  const run = useRunSession();

  let unsubscribeEvents: (() => void) | undefined;
  let unsubscribeOpenRepo: (() => void) | undefined;

  const viewingRun = computed(() =>
    workflow.selectedWorkflow.value != null
    && workflow.selectedWorkflow.value === run.runEventsWorkflow.value,
  );

  const events = computed(() =>
    viewingRun.value ? run.runEvents.value : workflow.previewEvents.value,
  );

  const logIsRunning = computed(() =>
    viewingRun.value && run.isRunning.value,
  );

  onMounted(() => {
    unsubscribeOpenRepo = window.pipebox.onOpenRepo((nextPath) => {
      void requestOpenRepo(nextPath);
    });
    unsubscribeEvents = run.subscribe();
    void refreshDoctor();
    void openLaunchRepo();
  });

  onUnmounted(() => {
    unsubscribeOpenRepo?.();
    unsubscribeEvents?.();
  });

  async function openRepo(nextPath: string) {
    statusMessage.value = '';
    workflow.reset();
    run.reset();
    await project.load(nextPath);
  }

  async function requestOpenRepo(nextPath: string) {
    if (!nextPath || nextPath === project.repoPath.value) {
      return;
    }
    if (run.isRunning.value) {
      project.requestSwitch(nextPath);
      return;
    }
    await openRepo(nextPath);
  }

  async function selectRepo() {
    const nextPath = await window.pipebox.selectRepo();
    if (!nextPath) {
      return;
    }
    await requestOpenRepo(nextPath);
  }

  async function openLaunchRepo() {
    const nextPath = await window.pipebox.getLaunchRepo();
    if (!nextPath) {
      return;
    }
    await requestOpenRepo(nextPath);
  }

  async function confirmSwitchProject() {
    const nextPath = project.takePendingRepo();
    if (!nextPath) {
      return;
    }
    await run.stop();
    await openRepo(nextPath);
  }

  function cancelSwitchProject() {
    project.cancelSwitch();
  }

  async function refreshDoctor() {
    statusMessage.value = '';
    await doctor.runDoctor();
  }

  async function selectWorkflow(workflowPath: string) {
    workflow.select(workflowPath);
    if (workflowPath === run.runEventsWorkflow.value) {
      return;
    }
    const repoPath = project.repoPath.value;
    if (!repoPath) {
      return;
    }
    const preview = await workflow.loadPreview(repoPath, workflowPath);
    if (preview.status === 'cached') {
      return;
    }
    statusMessage.value = preview.status === 'error'
      ? preview.message
      : '';
  }

  async function startRun(workflowPath: string) {
    const repoPath = project.repoPath.value;
    if (!repoPath || !workflowPath || run.isRunning.value) {
      return;
    }
    statusMessage.value = '';
    workflow.select(workflowPath);

    const inputs = workflow.inputsFor(workflowPath);
    const missing = missingRequiredInputs(inputs, workflow.inputValues.value);
    if (missing.length > 0) {
      statusMessage.value = `Required inputs: ${missing.join(', ')}`;
      return;
    }

    const error = await run.start({
      repoPath,
      workflow: workflowPath,
      inputs: toInputArgs(inputs, workflow.inputValues.value),
    });
    if (error) {
      statusMessage.value = error;
    }
  }

  return {
    repoPath: project.repoPath,
    branch: project.branch,
    workflows: project.workflows,
    selectedWorkflow: workflow.selectedWorkflow,
    runningWorkflow: run.runningWorkflow,
    doctor: doctor.doctor,
    events,
    isRunning: run.isRunning,
    logIsRunning,
    statusMessage,
    selectedWorkflowInputs: workflow.selectedWorkflowInputs,
    inputValues: workflow.inputValues,
    setInputValue: workflow.setInputValue,
    selectRepo,
    selectWorkflow,
    runDoctor: refreshDoctor,
    startRun,
    stopRun: run.stop,
    switchProjectOpen: project.switchProjectOpen,
    confirmSwitchProject,
    cancelSwitchProject,
  };
}
