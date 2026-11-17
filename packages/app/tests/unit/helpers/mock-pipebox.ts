import type {
  DoctorEvent,
  JobGraphEvent,
  RunnerEvent,
  WorkflowListEvent,
} from '@/domain/runner-events';
import type { PipeboxApi } from '@/services/pipebox-api';
import { vi } from 'vitest';
import {
  FIXTURE_REPO_EMPTY,
  FIXTURE_REPO_GITHUB,
  FIXTURE_REPO_JENKINS,
  FIXTURE_REPO_MIXED,
  githubGraph,
  githubWorkflow,
  jenkinsWorkflow,
  mixedGithubWorkflow,
  okDoctor,
  workflowList,
} from './fixtures';

export interface MockPipeboxOptions {
  doctor?: DoctorEvent;
  launchRepo?: string | null;
  selectRepoPath?: string | null | (() => string | null | Promise<string | null>);
  listsByRepo?: Record<string, WorkflowListEvent>;
  graphsByWorkflow?: Record<string, JobGraphEvent>;
  branchesByRepo?: Record<string, string | null>;
  runEvents?: RunnerEvent[] | ((options: {
    repoPath: string;
    workflow: string;
    inputs?: string[];
  }) => RunnerEvent[]);
  startRunError?: Error;
}

export interface MockPipeboxHandle {
  api: PipeboxApi;
  emit: (event: RunnerEvent) => void;
  emitOpenRepo: (repoPath: string) => void;
  setDoctor: (next: DoctorEvent) => void;
  doctorCalls: ReturnType<typeof vi.fn>;
  selectRepoCalls: ReturnType<typeof vi.fn>;
  startRunCalls: ReturnType<typeof vi.fn>;
  stopRunCalls: ReturnType<typeof vi.fn>;
  saveArtifactCalls: ReturnType<typeof vi.fn>;
  reportIssueCalls: ReturnType<typeof vi.fn>;
}

const defaultLists: Record<string, WorkflowListEvent> = {
  [FIXTURE_REPO_GITHUB]: workflowList([githubWorkflow], ['github']),
  [FIXTURE_REPO_JENKINS]: workflowList([jenkinsWorkflow], ['jenkins']),
  [FIXTURE_REPO_MIXED]: workflowList(
    [mixedGithubWorkflow, jenkinsWorkflow],
    ['github', 'jenkins'],
  ),
  [FIXTURE_REPO_EMPTY]: workflowList([], []),
};

export function createMockPipebox(
  options: MockPipeboxOptions = {},
): MockPipeboxHandle {
  const listeners = new Set<(event: RunnerEvent) => void>();
  const openRepoListeners = new Set<(repoPath: string) => void>();
  const doctorCalls = vi.fn();
  const selectRepoCalls = vi.fn();
  const startRunCalls = vi.fn();
  const stopRunCalls = vi.fn();
  const saveArtifactCalls = vi.fn();
  const reportIssueCalls = vi.fn();

  let doctorResult = options.doctor ?? okDoctor;
  const listsByRepo = {
    ...defaultLists,
    ...options.listsByRepo,
  };
  const graphsByWorkflow = {
    [githubWorkflow.path]: githubGraph,
    [mixedGithubWorkflow.path]: githubGraph,
    [jenkinsWorkflow.path]: githubGraph,
    ...options.graphsByWorkflow,
  };
  const branchesByRepo = {
    [FIXTURE_REPO_GITHUB]: 'main',
    [FIXTURE_REPO_JENKINS]: 'main',
    [FIXTURE_REPO_MIXED]: 'develop',
    [FIXTURE_REPO_EMPTY]: 'main',
    ...options.branchesByRepo,
  };

  function emit(event: RunnerEvent) {
    for (const listener of listeners) {
      listener(event);
    }
  }

  function emitOpenRepo(repoPath: string) {
    for (const listener of openRepoListeners) {
      listener(repoPath);
    }
  }

  const api: PipeboxApi = {
    doctor: async () => {
      doctorCalls();
      return doctorResult;
    },
    listWorkflows: async (repoPath) => {
      const list = listsByRepo[repoPath];
      if (!list) {
        return workflowList([]);
      }
      return list;
    },
    graphWorkflow: async ({ workflow }) => {
      const graph = graphsByWorkflow[workflow];
      if (!graph) {
        throw new Error(`no graph for ${workflow}`);
      }
      return graph;
    },
    selectRepo: async () => {
      selectRepoCalls();
      const next = options.selectRepoPath;
      if (typeof next === 'function') {
        return next();
      }
      return next === undefined
        ? FIXTURE_REPO_GITHUB
        : next;
    },
    getLaunchRepo: async () => options.launchRepo ?? null,
    gitBranch: async (repoPath) => branchesByRepo[repoPath] ?? null,
    startRun: async (startOptions) => {
      startRunCalls(startOptions);
      if (options.startRunError) {
        throw options.startRunError;
      }
      const events = typeof options.runEvents === 'function'
        ? options.runEvents(startOptions)
        : (options.runEvents ?? []);
      queueMicrotask(() => {
        for (const event of events) {
          emit(event);
        }
      });
    },
    stopRun: async () => {
      stopRunCalls();
    },
    saveArtifact: async (saveOptions) => {
      saveArtifactCalls(saveOptions);
      return `/saved/${saveOptions.defaultName}`;
    },
    reportIssue: async () => {
      reportIssueCalls();
    },
    onEvent: (callback) => {
      listeners.add(callback);
      return () => {
        listeners.delete(callback);
      };
    },
    onOpenRepo: (callback) => {
      openRepoListeners.add(callback);
      return () => {
        openRepoListeners.delete(callback);
      };
    },
  };

  return {
    api,
    emit,
    emitOpenRepo,
    setDoctor: (next) => {
      doctorResult = next;
    },
    doctorCalls,
    selectRepoCalls,
    startRunCalls,
    stopRunCalls,
    saveArtifactCalls,
    reportIssueCalls,
  };
}
