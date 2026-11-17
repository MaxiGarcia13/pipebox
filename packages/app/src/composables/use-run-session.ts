import type { RunnerEvent } from '@/domain/runner-events';
import type { StartRunOptions } from '@/services/ipc-channels';
import { ref } from 'vue';

function hasRunFinished(events: RunnerEvent[]): boolean {
  return events.some((event) => event.type === 'run_finished');
}

export function useRunSession() {
  const runningWorkflow = ref<string | null>(null);
  const runEventsWorkflow = ref<string | null>(null);
  const runEvents = ref<RunnerEvent[]>([]);
  const isRunning = ref(false);

  function reset() {
    runningWorkflow.value = null;
    runEventsWorkflow.value = null;
    runEvents.value = [];
    isRunning.value = false;
  }

  function onRunnerEvent(event: RunnerEvent) {
    if (event.type === 'run_finished' && hasRunFinished(runEvents.value)) {
      return;
    }
    runEvents.value = [...runEvents.value, event];
    if (event.type === 'run_finished') {
      isRunning.value = false;
      runningWorkflow.value = null;
    }
  }

  function subscribe(): () => void {
    return window.pipebox.onEvent(onRunnerEvent);
  }

  async function start(options: StartRunOptions): Promise<string | null> {
    runningWorkflow.value = options.workflow;
    runEventsWorkflow.value = options.workflow;
    runEvents.value = [];
    isRunning.value = true;
    try {
      await window.pipebox.startRun(options);
      return null;
    } catch (error) {
      isRunning.value = false;
      runningWorkflow.value = null;
      return error instanceof Error
        ? error.message
        : String(error);
    }
  }

  async function stop() {
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
    runningWorkflow,
    runEventsWorkflow,
    runEvents,
    isRunning,
    reset,
    subscribe,
    start,
    stop,
  };
}
