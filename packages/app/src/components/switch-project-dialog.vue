<script setup lang="ts">
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';

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
    <AlertDialogContent data-testid="switch-project-dialog">
      <AlertDialogHeader>
        <AlertDialogTitle>Stop the current run?</AlertDialogTitle>
        <AlertDialogDescription>
          A run is in progress. Stop it and open the other project, or continue and keep this one.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <Button
          variant="outline"
          data-testid="switch-project-continue"
          @click="onCancel"
        >
          Continue
        </Button>
        <Button
          data-testid="switch-project-stop"
          @click="onConfirm"
        >
          Stop and open
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
