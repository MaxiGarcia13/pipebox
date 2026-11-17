import type { Workflow } from '@/domain/runner-events';
import { ref } from 'vue';

export function useProjectSession() {
  const repoPath = ref<string | null>(null);
  const branch = ref<string | null>(null);
  const workflows = ref<Workflow[]>([]);
  const pendingRepo = ref<string | null>(null);
  const switchProjectOpen = ref(false);

  async function load(nextPath: string) {
    repoPath.value = nextPath;
    branch.value = null;
    workflows.value = [];
    const [list, nextBranch] = await Promise.all([
      window.pipebox.listWorkflows(nextPath),
      window.pipebox.gitBranch(nextPath),
    ]);
    workflows.value = list.workflows;
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
    switchProjectOpen,
    load,
    requestSwitch,
    takePendingRepo,
    cancelSwitch,
  };
}
