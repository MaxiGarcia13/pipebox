import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, nextTick, ref } from 'vue';
import { useLogSearchSession } from '@/composables/use-log-search-session';

describe('useLogSearchSession', () => {
  it('owns open/query state and clears on close', async () => {
    let session: ReturnType<typeof useLogSearchSession> | undefined;
    const host = ref<HTMLElement | null>(null);
    const Host = defineComponent({
      setup() {
        session = useLogSearchSession(host);
        return () => null;
      },
    });
    mount(Host);

    await session!.openSearch();
    expect(session!.open.value).toBe(true);
    session!.query.value = 'lint';
    session!.activeMatch.value = 1;

    session!.closeSearch();
    expect(session!.open.value).toBe(false);
    expect(session!.query.value).toBe('');
    expect(session!.activeMatch.value).toBe(0);
    await nextTick();
  });
});
