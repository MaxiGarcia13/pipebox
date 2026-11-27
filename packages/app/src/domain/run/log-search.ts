export interface SearchSegment {
  text: string;
  match: boolean;
}

export interface SearchLinePart {
  text: string;
  matchIndex: number | null;
}

export function splitSearchSegments(text: string, query: string): SearchSegment[] {
  const needle = query.trim();
  if (needle === '') {
    return [
      { text, match: false },
    ];
  }

  const lowerText = text.toLowerCase();
  const lowerNeedle = needle.toLowerCase();
  const segments: SearchSegment[] = [];
  let cursor = 0;

  while (cursor < text.length) {
    const foundAt = lowerText.indexOf(lowerNeedle, cursor);
    if (foundAt === -1) {
      segments.push({
        text: text.slice(cursor),
        match: false,
      });
      break;
    }
    if (foundAt > cursor) {
      segments.push({
        text: text.slice(cursor, foundAt),
        match: false,
      });
    }
    segments.push({
      text: text.slice(foundAt, foundAt + needle.length),
      match: true,
    });
    cursor = foundAt + needle.length;
  }

  return segments.length > 0
    ? segments
    : [
        { text: '', match: false },
      ];
}

export function buildSearchLineParts(
  texts: string[],
  query: string,
): { parts: SearchLinePart[][]; matchCount: number } {
  let matchIndex = 0;
  const parts = texts.map((text) =>
    splitSearchSegments(text, query).map((segment) => {
      if (!segment.match) {
        return { text: segment.text, matchIndex: null };
      }
      const index = matchIndex;
      matchIndex += 1;
      return { text: segment.text, matchIndex: index };
    }),
  );
  return { parts, matchCount: matchIndex };
}
