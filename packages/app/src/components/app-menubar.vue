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
}>();
</script>

<template>
  <Menubar class="h-8 border-0 bg-transparent p-0 shadow-none">
    <MenubarMenu>
      <MenubarTrigger>
        File
      </MenubarTrigger>
      <MenubarContent>
        <MenubarItem @click="$emit('openProject')">
          Open Project…
        </MenubarItem>
      </MenubarContent>
    </MenubarMenu>

    <MenubarMenu>
      <MenubarTrigger>
        View
      </MenubarTrigger>
      <MenubarContent>
        <MenubarItem @click="$emit('refreshDoctor')">
          Refresh Doctor
        </MenubarItem>
      </MenubarContent>
    </MenubarMenu>
  </Menubar>
</template>
