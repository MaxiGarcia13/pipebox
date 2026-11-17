import type { ChildProcess } from 'node:child_process';
import type { RunnerEvent } from '../domain/runner-events';
import { spawn } from 'node:child_process';
import readline from 'node:readline';
import { resolvePipeboxPath } from './resolve-pipebox-path';

type EventHandler = (event: RunnerEvent) => void;

export async function runToCompletion(args: string[]): Promise<RunnerEvent[]> {
  const events: RunnerEvent[] = [];
  await spawnRunner(args, (event) => {
    events.push(event);
  });
  return events;
}

export class RunnerClient {
  private activeProcess: ChildProcess | null = null;

  start(args: string[], onEvent: EventHandler): Promise<number> {
    if (this.activeProcess) {
      throw new Error('A run is already in progress');
    }
    return spawnRunner(args, onEvent, {
      onSpawn: (child) => {
        this.activeProcess = child;
      },
      onEnd: (child) => {
        if (this.activeProcess === child) {
          this.activeProcess = null;
        }
      },
    });
  }

  stop(): void {
    if (!this.activeProcess) {
      return;
    }
    const child = this.activeProcess;
    this.activeProcess = null;
    child.kill('SIGTERM');
  }

  get isRunning(): boolean {
    return this.activeProcess !== null;
  }
}

interface SpawnHooks {
  onSpawn?: (child: ChildProcess) => void;
  onEnd?: (child: ChildProcess) => void;
}

function spawnRunner(
  args: string[],
  onEvent: EventHandler,
  hooks: SpawnHooks = {},
): Promise<number> {
  const binaryPath = resolvePipeboxPath();
  const child = spawn(binaryPath, args, {
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  hooks.onSpawn?.(child);

  if (child.stdout) {
    const stdout = readline.createInterface({ input: child.stdout });
    stdout.on('line', (line) => {
      const event = parseRunnerEvent(line);
      if (event) {
        onEvent(event);
      }
    });
  }

  if (child.stderr) {
    const stderr = readline.createInterface({ input: child.stderr });
    stderr.on('line', (line) => {
      if (!line.trim()) {
        return;
      }
      // Progress / diagnostics often land on stderr (e.g. docker build). Keep them
      // as unscoped logs so the UI does not paint them as failures.
      onEvent({ type: 'log', line });
    });
  }

  return new Promise((resolve, reject) => {
    child.on('error', (error) => {
      hooks.onEnd?.(child);
      reject(error);
    });
    child.on('close', (code) => {
      hooks.onEnd?.(child);
      resolve(code ?? 0);
    });
  });
}

function parseRunnerEvent(line: string): RunnerEvent | null {
  const trimmed = line.trim();
  if (!trimmed) {
    return null;
  }
  try {
    return JSON.parse(trimmed) as RunnerEvent;
  } catch {
    return { type: 'error', message: `invalid NDJSON: ${trimmed}` };
  }
}
