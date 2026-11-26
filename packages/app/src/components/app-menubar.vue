<script setup lang="ts">
import {
  Menubar,
  MenubarContent,
  MenubarItem,
  MenubarMenu,
  MenubarTrigger,
} from '@/components/ui/menubar';

defineEmits<{
  openProject: [];
  refreshDoctor: [];
  refreshWorkflows: [];
}>();

function reportIssue() {
  void window.pipebox.reportIssue();
}
</script>

<template>
  <Menubar class="h-8 border-0 bg-transparent p-0 shadow-none">
    <MenubarMenu>
      <MenubarTrigger>
        File
      </MenubarTrigger>
      <MenubarContent>
        <MenubarItem
          data-testid="menubar-open-project"
          @click="$emit('openProject')"
        >
          Open Project…
        </MenubarItem>
      </MenubarContent>
    </MenubarMenu>

    <MenubarMenu>
      <MenubarTrigger>
        View
      </MenubarTrigger>
      <MenubarContent>
        <MenubarItem
          data-testid="menubar-refresh-workflows"
          @click="$emit('refreshWorkflows')"
        >
          Refresh Workflows
        </MenubarItem>
        <MenubarItem
          data-testid="menubar-refresh-doctor"
          @click="$emit('refreshDoctor')"
        >
          Refresh Doctor
        </MenubarItem>
      </MenubarContent>
    </MenubarMenu>

    <MenubarMenu>
      <MenubarTrigger data-testid="menubar-help">
        Help
      </MenubarTrigger>
      <MenubarContent>
        <MenubarItem
          data-testid="menubar-report-issue"
          @click="reportIssue"
        >
          Report Issue or Idea…
        </MenubarItem>
      </MenubarContent>
    </MenubarMenu>
  </Menubar>
</template>
