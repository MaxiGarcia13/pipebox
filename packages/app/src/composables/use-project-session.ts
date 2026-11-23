import type { Workflow, WorkflowProvider } from '@/domain/runner-events';
import { ref } from 'vue';

export function useProjectSession() {
  const repoPath = ref<string | null>(null);
  const branch = ref<string | null>(null);
  const workflows = ref<Workflow[]>([]);
  const providers = ref<WorkflowProvider[]>([]);
  const pendingRepo = ref<string | null>(null);
  const switchProjectOpen = ref(false);

  async function load(nextPath: string) {
    repoPath.value = nextPath;
    branch.value = null;
    workflows.value = [];
    providers.value = [];
    await refresh();
  }

  async function refresh() {
    const path = repoPath.value;
    if (!path) {
      return;
    }
    const [list, nextBranch] = await Promise.all([
      window.pipebox.listWorkflows(path),
      window.pipebox.gitBranch(path),
    ]);
    workflows.value = list.workflows;
    providers.value = list.providers ?? [
      ...new Set(list.workflows.map((workflow) => workflow.provider)),
    ];
    branch.value = nextBranch;
  }

  function requestSwitch(nextPath: string) {
    pendingRepo.value = nextPath;
    switchProjectOpen.value = true;
  }

  function takePendingRepo(): string | null {
    const nextPath = pendingRepo.value;
    pendingRepo.value = null;
    switchProjectOpen.value = false;
    return nextPath;
  }

  function cancelSwitch() {
    pendingRepo.value = null;
    switchProjectOpen.value = false;
  }

  return {
    repoPath,
    branch,
    workflows,
    providers,
    switchProjectOpen,
    load,
    refresh,
    requestSwitch,
    takePendingRepo,
    cancelSwitch,
  };
}
