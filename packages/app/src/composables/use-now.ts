import type { Ref } from 'vue';
import { onUnmounted, ref, watch } from 'vue';

const TICK_MS = 1000;

export function useNow(ticking: Ref<boolean>) {
  const now = ref(Date.now());
  let timer: ReturnType<typeof setInterval> | undefined;

  function stop() {
    if (timer == null) {
      return;
    }
    clearInterval(timer);
    timer = undefined;
  }

  watch(ticking, (active) => {
    stop();
    now.value = Date.now();
    if (!active) {
      return;
    }
    timer = setInterval(() => {
      now.value = Date.now();
    }, TICK_MS);
  }, { immediate: true });

  onUnmounted(stop);
  return now;
}
