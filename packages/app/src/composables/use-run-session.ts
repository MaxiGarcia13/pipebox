import type { DoctorEvent, RunnerEvent } from '@/domain/runner-events';
import type { StartRunOptions } from '@/services/ipc-channels';
import { ref } from 'vue';
import {
  prepareJobRerun,
  shouldAcceptJobGraph,
} from '@/domain/run/merge-job-run-events';

function hasRunFinished(events: RunnerEvent[]): boolean {
  return events.some((event) => event.type === 'run_finished');
}

function withReceiptTime(event: RunnerEvent): RunnerEvent {
  if (
    event.type !== 'step_started'
    && event.type !== 'step_finished'
    && event.type !== 'run_finished'
  ) {
    return event;
  }
  if (event.receivedAt != null) {
    return event;
  }
  // Wall clock for the log. groupRunEvents only reads this; it does not sample time.
  return {
    ...event,
    receivedAt: Date.now(),
  };
}

export function useRunSession(hooks?: {
  onDoctor?: (event: DoctorEvent) => void;
}) {
  const runningWorkflow = ref<string | null>(null);
  const runEventsWorkflow = ref<string | null>(null);
  const runEvents = ref<RunnerEvent[]>([]);
  const isRunning = ref(false);
  const activeJob = ref<string | null>(null);

  function reset() {
    runningWorkflow.value = null;
    runEventsWorkflow.value = null;
    runEvents.value = [];
    isRunning.value = false;
    activeJob.value = null;
  }

  function onRunnerEvent(event: RunnerEvent) {
    if (event.type === 'doctor') {
      hooks?.onDoctor?.(event);
      return;
    }
    const stamped = withReceiptTime(event);
    if (
      stamped.type === 'job_graph'
      && activeJob.value
      && !shouldAcceptJobGraph(runEvents.value, stamped)
    ) {
      return;
    }
    if (stamped.type === 'run_finished' && hasRunFinished(runEvents.value)) {
      return;
    }
    runEvents.value = [...runEvents.value, stamped];
    if (stamped.type === 'run_finished') {
      isRunning.value = false;
      runningWorkflow.value = null;
      activeJob.value = null;
    }
  }

  function subscribe(): () => void {
    return window.pipebox.onEvent(onRunnerEvent);
  }

  async function start(options: StartRunOptions): Promise<string | null> {
    const jobId = options.job?.trim() || undefined;
    const mergeJob = Boolean(
      jobId
      && runEventsWorkflow.value === options.workflow
      && runEvents.value.length > 0,
    );

    runningWorkflow.value = options.workflow;
    runEventsWorkflow.value = options.workflow;
    activeJob.value = jobId ?? null;
    runEvents.value = mergeJob && jobId
      ? prepareJobRerun(runEvents.value, jobId)
      : [];
    isRunning.value = true;
    try {
      await window.pipebox.startRun({
        ...options,
        job: jobId,
      });
      return null;
    } catch (error) {
      isRunning.value = false;
      runningWorkflow.value = null;
      activeJob.value = null;
      return error instanceof Error
        ? error.message
        : String(error);
    }
  }

  async function stop() {
    await window.pipebox.stopRun();
    isRunning.value = false;
    runningWorkflow.value = null;
    activeJob.value = null;
    if (!hasRunFinished(runEvents.value)) {
      runEvents.value = [
        ...runEvents.value,
        withReceiptTime({
          type: 'run_finished',
          status: 'cancelled',
          exitCode: 130,
        }),
      ];
    }
  }

  return {
    runningWorkflow,
    runEventsWorkflow,
    runEvents,
    isRunning,
    activeJob,
    reset,
    subscribe,
    start,
    stop,
  };
}
