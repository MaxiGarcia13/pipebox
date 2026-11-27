<script setup lang="ts">
import { ChevronDown, ChevronUp, X } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

defineProps<{
  matchCount: number;
  matchLabel: string;
}>();

const emit = defineEmits<{
  next: [];
  prev: [];
  close: [];
  keydown: [event: KeyboardEvent];
}>();

const query = defineModel<string>('query', { required: true });
</script>

<template>
  <div
    data-testid="log-search-bar"
    class="flex items-center gap-2 border-b border-border px-3 py-2"
  >
    <Input
      v-model="query"
      data-testid="log-search-input"
      type="search"
      placeholder="Find in logs…"
      class="h-8"
      @keydown="emit('keydown', $event)"
    />
    <span
      v-if="matchLabel"
      data-testid="log-search-count"
      class="shrink-0 text-xs text-muted-foreground tabular-nums"
    >
      {{ matchLabel }}
    </span>
    <Button
      type="button"
      variant="ghost"
      size="sm"
      class="size-8 shrink-0 p-0"
      :disabled="matchCount === 0"
      aria-label="Previous match"
      data-testid="log-search-prev"
      @click="emit('prev')"
    >
      <ChevronUp class="size-4" />
    </Button>
    <Button
      type="button"
      variant="ghost"
      size="sm"
      class="size-8 shrink-0 p-0"
      :disabled="matchCount === 0"
      aria-label="Next match"
      data-testid="log-search-next"
      @click="emit('next')"
    >
      <ChevronDown class="size-4" />
    </Button>
    <Button
      type="button"
      variant="ghost"
      size="sm"
      class="size-8 shrink-0 p-0"
      aria-label="Close find"
      data-testid="log-search-close"
      @click="emit('close')"
    >
      <X class="size-4" />
    </Button>
  </div>
</template>
