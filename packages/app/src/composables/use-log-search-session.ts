import type { Ref } from 'vue';
import { nextTick, onMounted, onUnmounted, ref } from 'vue';

export interface LogSearchSession {
  open: Ref<boolean>;
  query: Ref<string>;
  activeMatch: Ref<number>;
  openSearch: () => Promise<void>;
  closeSearch: () => void;
}

export function useLogSearchSession(
  host: Ref<HTMLElement | null>,
): LogSearchSession {
  const open = ref(false);
  const query = ref('');
  const activeMatch = ref(0);

  function focusInput() {
    const input = host.value
      ?.querySelector<HTMLInputElement>('[data-testid="log-search-input"]');
    input?.focus();
    input?.select();
  }

  async function openSearch() {
    open.value = true;
    await nextTick();
    focusInput();
  }

  function closeSearch() {
    open.value = false;
    query.value = '';
    activeMatch.value = 0;
  }

  function onGlobalKeydown(event: KeyboardEvent) {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'f') {
      event.preventDefault();
      void openSearch();
      return;
    }
    if (event.key === 'Escape' && open.value) {
      event.preventDefault();
      closeSearch();
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', onGlobalKeydown);
  });

  onUnmounted(() => {
    window.removeEventListener('keydown', onGlobalKeydown);
  });

  return {
    open,
    query,
    activeMatch,
    openSearch,
    closeSearch,
  };
}
