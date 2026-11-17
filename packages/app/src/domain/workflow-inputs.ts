import type { WorkflowInput } from '@/domain/runner-events';

export function seedInputValues(inputs: WorkflowInput[]): Record<string, string> {
  const values: Record<string, string> = {};
  for (const input of inputs) {
    values[input.id] = defaultInputValue(input);
  }
  return values;
}

export function missingRequiredInputs(
  inputs: WorkflowInput[],
  values: Record<string, string>,
): string[] {
  return inputs
    .filter((input) => {
      if (!input.required) {
        return false;
      }
      if (input.type === 'boolean') {
        return false;
      }
      return !(values[input.id] ?? '').trim();
    })
    .map((input) => input.id);
}

export function toInputArgs(
  inputs: WorkflowInput[],
  values: Record<string, string>,
): string[] {
  return inputs
    .map((input) => {
      const value = values[input.id] ?? '';
      if (value === '' && input.type !== 'boolean') {
        return null;
      }
      return `${input.id}=${value}`;
    })
    .filter((entry): entry is string => entry != null);
}

function defaultInputValue(input: WorkflowInput): string {
  if (input.type === 'boolean') {
    return input.default === 'true'
      ? 'true'
      : 'false';
  }
  return input.default ?? '';
}
