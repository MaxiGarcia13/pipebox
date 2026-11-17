package jenkins

import (
  "strings"
  "unicode"
)

func extractFirstBlock(source, keyword string) (body string, ok bool) {
  body, _, ok = extractBlockAfter(source, keyword, 0)
  return body, ok
}

func extractBlockAfter(source, keyword string, from int) (body string, end int, ok bool) {
  index := from
  for {
    found := indexOfIdent(source, keyword, index)
    if found < 0 {
      return "", from, false
    }
    cursor := skipSpace(source, found+len(keyword))
    if cursor < len(source) && source[cursor] == '{' {
      return balancedBody(source, cursor)
    }
    index = found + len(keyword)
  }
}

func balancedBody(source string, openBrace int) (body string, end int, ok bool) {
  if openBrace >= len(source) || source[openBrace] != '{' {
    return "", openBrace, false
  }

  var tracker quoteTracker
  depth := 0
  for index := openBrace; index < len(source); {
    if !tracker.active() {
      switch source[index] {
      case '{':
        depth++
        index++
        continue
      case '}':
        depth--
        if depth == 0 {
          return source[openBrace+1 : index], index + 1, true
        }
        index++
        continue
      }
    }
    index += tracker.advance(source, index)
  }
  return "", openBrace, false
}

func indexOfIdent(source, ident string, from int) int {
  identLen := len(ident)
  for index := from; index+identLen <= len(source); index++ {
    if !strings.HasPrefix(source[index:], ident) {
      continue
    }
    if !identBoundaryOK(source, index, identLen) {
      continue
    }
    if inQuotes(source, index) {
      continue
    }
    return index
  }
  return -1
}

func hasIdentAt(source string, index int, ident string) bool {
  return indexOfIdent(source, ident, index) == index
}

func identBoundaryOK(source string, index, identLen int) bool {
  if index > 0 && isIdentChar(rune(source[index-1])) {
    return false
  }
  after := index + identLen
  if after < len(source) && isIdentChar(rune(source[after])) {
    return false
  }
  return true
}

func isIdentChar(char rune) bool {
  return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_'
}

func skipSpace(source string, index int) int {
  for index < len(source) && unicode.IsSpace(rune(source[index])) {
    index++
  }
  return index
}
