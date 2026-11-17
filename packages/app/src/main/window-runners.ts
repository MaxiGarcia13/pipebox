import type { WebContents } from 'electron';
import { RunnerClient } from './runner-client';

const clients = new Map<number, RunnerClient>();

export function runnerFor(webContents: WebContents): RunnerClient {
  const id = webContents.id;
  const existing = clients.get(id);
  if (existing) {
    return existing;
  }

  const client = new RunnerClient();
  clients.set(id, client);
  webContents.once('destroyed', () => {
    client.stop();
    clients.delete(id);
  });
  return client;
}
