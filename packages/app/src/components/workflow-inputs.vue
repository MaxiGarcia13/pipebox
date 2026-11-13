<script setup lang="ts">
import type { WorkflowInput } from '@/domain/runner-events';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

defineProps<{
  inputs: WorkflowInput[];
  values: Record<string, string>;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  update: [id: string, value: string];
}>();

function labelFor(input: WorkflowInput): string {
  return input.description?.trim() || input.id;
}

function isChecked(value: string | undefined): boolean {
  return value === 'true';
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="border-b border-border px-4 py-2">
      <h2 class="text-sm font-medium">
        Inputs
      </h2>
    </div>

    <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-auto p-4">
      <div
        v-for="input in inputs"
        :key="input.id"
        class="flex flex-col gap-2"
      >
        <div class="flex items-center gap-2">
          <Label :for="`workflow-input-${input.id}`">
            {{ labelFor(input) }}
          </Label>
          <span
            v-if="input.required"
            class="text-xs text-muted-foreground"
          >
            required
          </span>
        </div>

        <div
          v-if="input.type === 'boolean'"
          class="flex items-center gap-2"
        >
          <Checkbox
            :id="`workflow-input-${input.id}`"
            :checked="isChecked(values[input.id])"
            :disabled="disabled"
            @update:checked="emit('update', input.id, $event === true ? 'true' : 'false')"
          />
          <span class="text-xs text-muted-foreground">
            {{ input.id }}
          </span>
        </div>

        <Select
          v-else-if="input.type === 'choice'"
          :model-value="values[input.id] || undefined"
          :disabled="disabled"
          @update:model-value="emit('update', input.id, String($event ?? ''))"
        >
          <SelectTrigger :id="`workflow-input-${input.id}`">
            <SelectValue :placeholder="input.id" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem
              v-for="option in input.options ?? []"
              :key="option"
              :value="option"
            >
              {{ option }}
            </SelectItem>
          </SelectContent>
        </Select>

        <Input
          v-else
          :id="`workflow-input-${input.id}`"
          :model-value="values[input.id] ?? ''"
          :type="input.type === 'number' ? 'number' : 'text'"
          :disabled="disabled"
          :placeholder="input.id"
          @update:model-value="emit('update', input.id, String($event))"
        />
      </div>
    </div>
  </div>
</template>
