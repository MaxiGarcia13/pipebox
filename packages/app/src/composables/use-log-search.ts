import type { MaybeRefOrGetter, Ref } from 'vue';
import type { LogSearchSession } from '@/composables/use-log-search-session';
import type { ChronologicalLogKind, ChronologicalLogLine } from '@/domain/run';
import { computed, nextTick, toValue, watch } from 'vue';
import { buildSearchLineParts } from '@/domain/run';

export interface SearchedLogLine {
  kind: ChronologicalLogKind;
  parts: Array<{ text: string; matchIndex: number | null }>;
}

export function useLogSearch(
  lines: MaybeRefOrGetter<ChronologicalLogLine[]>,
  scroller: Ref<HTMLElement | null>,
  session: LogSearchSession,
  onNavigateAwayFromBottom?: () => void,
) {
  const { open, query, activeMatch, closeSearch } = session;

  const search = computed(() => {
    if (!open.value) {
      return buildSearchLineParts(
        toValue(lines).map((line) => line.text),
        '',
      );
    }
    return buildSearchLineParts(
      toValue(lines).map((line) => line.text),
      query.value,
    );
  });

  const matchCount = computed(() => search.value.matchCount);

  const matchLabel = computed(() => {
    if (!open.value || query.value.trim() === '') {
      return '';
    }
    if (matchCount.value === 0) {
      return '0/0';
    }
    return `${activeMatch.value + 1}/${matchCount.value}`;
  });

  const renderedLines = computed((): SearchedLogLine[] =>
    toValue(lines).map((line, index) => ({
      kind: line.kind,
      parts: search.value.parts[index] ?? [
        { text: line.text, matchIndex: null },
      ],
    })),
  );

  const isSearching = computed(() => open.value && query.value.trim() !== '');

  function clampActiveMatch() {
    if (matchCount.value === 0) {
      activeMatch.value = 0;
      return;
    }
    if (activeMatch.value >= matchCount.value) {
      activeMatch.value = matchCount.value - 1;
    }
  }

  async function scrollActiveMatchIntoView() {
    if (matchCount.value === 0) {
      return;
    }
    // Collapsible content unmounts while closed; wait until the mark exists
    // after force-open remounts matching step groups.
    for (let attempt = 0; attempt < 12; attempt += 1) {
      await nextTick();
      const mark = scroller.value?.querySelector(
        `[data-match-index="${activeMatch.value}"]`,
      );
      if (mark instanceof HTMLElement && mark.getClientRects().length > 0) {
        mark.scrollIntoView({ block: 'center' });
        return;
      }
      await new Promise<void>((resolve) => {
        requestAnimationFrame(() => resolve());
      });
    }
  }

  function goToMatch(index: number) {
    if (matchCount.value === 0) {
      return;
    }
    activeMatch.value = ((index % matchCount.value) + matchCount.value)
      % matchCount.value;
    onNavigateAwayFromBottom?.();
    void scrollActiveMatchIntoView();
  }

  function goNext() {
    goToMatch(activeMatch.value + 1);
  }

  function goPrev() {
    goToMatch(activeMatch.value - 1);
  }

  function onSearchKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      closeSearch();
      return;
    }
    if (event.key !== 'Enter') {
      return;
    }
    event.preventDefault();
    if (event.shiftKey) {
      goPrev();
      return;
    }
    goNext();
  }

  watch(query, () => {
    activeMatch.value = 0;
    if (isSearching.value) {
      onNavigateAwayFromBottom?.();
      void scrollActiveMatchIntoView();
    }
  });

  watch(matchCount, () => {
    clampActiveMatch();
  });

  return {
    open,
    query,
    activeMatch,
    matchCount,
    matchLabel,
    renderedLines,
    isSearching,
    closeSearch,
    goNext,
    goPrev,
    onSearchKeydown,
  };
}
