<script setup lang="ts">
import type { JobRunStatus } from '@/domain/run';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';

export interface JobNeedDependency {
  id: string;
  name: string;
  status: JobRunStatus;
}

defineProps<{
  jobName: string;
  dependencies: JobNeedDependency[];
}>();

const emit = defineEmits<{
  confirm: [];
  cancel: [];
}>();

const open = defineModel<boolean>('open', { required: true });

let closedByButton = false;

function onOpenChange(next: boolean) {
  if (!next && open.value && !closedByButton) {
    emit('cancel');
  }
  closedByButton = false;
  open.value = next;
}

function onCancel() {
  closedByButton = true;
  open.value = false;
  emit('cancel');
}

function onConfirm() {
  closedByButton = true;
  open.value = false;
  emit('confirm');
}
</script>

<template>
  <AlertDialog
    :open="open"
    @update:open="onOpenChange"
  >
    <AlertDialogContent data-testid="job-needs-dialog">
      <AlertDialogHeader>
        <AlertDialogTitle>Run {{ jobName }}?</AlertDialogTitle>
        <AlertDialogDescription>
          This job depends on other jobs that are not all passed yet. You can still force a local run.
        </AlertDialogDescription>
      </AlertDialogHeader>

      <ul
        data-testid="job-needs-list"
        class="space-y-2 text-sm"
      >
        <li
          v-for="dependency in dependencies"
          :key="dependency.id"
          class="flex items-center justify-between gap-4"
        >
          <span class="truncate font-medium">
            {{ dependency.name }}
          </span>
          <span
            class="shrink-0 text-muted-foreground tabular-nums"
            :data-need-status="dependency.status"
          >
            {{ dependency.status }}
          </span>
        </li>
      </ul>

      <AlertDialogFooter>
        <Button
          variant="outline"
          data-testid="job-needs-cancel"
          class="rounded-xl"
          @click="onCancel"
        >
          Cancel
        </Button>
        <Button
          data-testid="job-needs-confirm"
          class="rounded-xl"
          @click="onConfirm"
        >
          Run anyway
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
