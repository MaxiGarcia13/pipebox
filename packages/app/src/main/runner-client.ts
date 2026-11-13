import type { ChildProcess } from 'node:child_process';
import type { RunnerEvent } from '../shared/runner-events';
import { spawn } from 'node:child_process';
import readline from 'node:readline';
import { resolvePipeboxPath } from './resolve-pipebox-path';

type EventHandler = (event: RunnerEvent) => void;

export class RunnerClient {
  private activeProcess: ChildProcess | null = null;

  async runToCompletion(args: string[]): Promise<RunnerEvent[]> {
    const events: RunnerEvent[] = [];
    await this.spawn(args, (event) => {
      events.push(event);
    });
    return events;
  }

  start(args: string[], onEvent: EventHandler): Promise<number> {
    if (this.activeProcess) {
      throw new Error('A run is already in progress');
    }
    return this.spawn(args, onEvent).finally(() => {
      this.activeProcess = null;
    });
  }

  stop(): void {
    if (!this.activeProcess) {
      return;
    }
    this.activeProcess.kill('SIGTERM');
    this.activeProcess = null;
  }

  get isRunning(): boolean {
    return this.activeProcess !== null;
  }

  private spawn(args: string[], onEvent: EventHandler): Promise<number> {
    const binaryPath = resolvePipeboxPath();
    const child = spawn(binaryPath, args, {
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    this.activeProcess = child;

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
        onEvent({ type: 'error', message: line });
      });
    }

    return new Promise((resolve, reject) => {
      child.on('error', (error) => {
        this.activeProcess = null;
        reject(error);
      });
      child.on('close', (code) => {
        this.activeProcess = null;
        resolve(code ?? 0);
      });
    });
  }
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
