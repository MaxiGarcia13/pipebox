import type { WorkflowInput } from '@/domain/runner-events';
import { describe, expect, it } from 'vitest';
import {
  missingRequiredInputs,
  seedInputValues,
  toInputArgs,
} from '@/domain/workflow-inputs';

const inputs: WorkflowInput[] = [
  {
    id: 'environment',
    required: true,
    type: 'choice',
    default: 'staging',
    options: ['staging', 'prod'],
  },
  {
    id: 'dry_run',
    type: 'boolean',
    default: 'false',
  },
  {
    id: 'note',
    required: true,
    type: 'string',
  },
];

describe('workflow-inputs', () => {
  it('seeds defaults including boolean false', () => {
    expect(seedInputValues(inputs)).toEqual({
      environment: 'staging',
      dry_run: 'false',
      note: '',
    });
  });

  it('reports missing required non-boolean inputs', () => {
    expect(missingRequiredInputs(inputs, {
      environment: 'staging',
      dry_run: 'false',
      note: '',
    })).toEqual(['note']);
  });

  it('builds input args and skips empty optional strings', () => {
    expect(toInputArgs(inputs, {
      environment: 'prod',
      dry_run: 'true',
      note: 'ship',
    })).toEqual([
      'environment=prod',
      'dry_run=true',
      'note=ship',
    ]);
  });
});
