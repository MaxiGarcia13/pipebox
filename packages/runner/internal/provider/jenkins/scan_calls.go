package jenkins

import "unicode"

func splitTopLevelCalls(body string) []string {
  calls := make([]string, 0)
  start := -1
  depth := 0
  var tracker quoteTracker

  flush := func(end int) {
    if start < 0 {
      return
    }
    call := trimSpaceRange(body, start, end)
    if call != "" {
      calls = append(calls, call)
    }
    start = -1
  }

  for index := 0; index < len(body); {
    if tracker.active() {
      index += tracker.advance(body, index)
      continue
    }

    char := body[index]
    switch char {
    case '\'', '"':
      if depth == 0 && start < 0 {
        start = index
      }
      index += tracker.advance(body, index)
    case '(', '{':
      if depth == 0 && start < 0 {
        start = index
      }
      depth++
      index++
    case ')', '}':
      if depth > 0 {
        depth--
      }
      index++
    case '\n', ';':
      if depth == 0 {
        flush(index)
      }
      index++
    default:
      if depth == 0 && start < 0 && !unicode.IsSpace(rune(char)) {
        start = index
      }
      index++
    }
  }
  flush(len(body))
  return calls
}

func trimSpaceRange(source string, start, end int) string {
  for start < end && unicode.IsSpace(rune(source[start])) {
    start++
  }
  for end > start && unicode.IsSpace(rune(source[end-1])) {
    end--
  }
  if start >= end {
    return ""
  }
  return source[start:end]
}
