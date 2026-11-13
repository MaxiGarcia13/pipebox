<template>
  <div class="flex h-screen flex-col overflow-hidden bg-background text-foreground">
    <header class="flex shrink-0 items-center gap-2 border-b border-border p-2">
      <slot name="toolbar" />
    </header>

    <div class="flex min-h-0 flex-1">
      <aside class="flex w-72 shrink-0 flex-col border-r border-border bg-sidebar">
        <slot name="sidebar" />
      </aside>

      <main class="flex min-w-0 flex-1 flex-col">
        <slot name="main" />
      </main>
    </div>

    <footer class="flex shrink-0 items-center justify-between gap-4 border-t border-border bg-sidebar px-4 py-2">
      <div class="min-w-0 flex-1 overflow-x-auto">
        <slot name="panel" />
      </div>
      <div class="shrink-0">
        <slot name="status" />
      </div>
    </footer>
  </div>
</template>
