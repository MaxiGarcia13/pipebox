package jenkins

import "strings"

func stripJenkinsNoise(source string) string {
  var builder strings.Builder
  builder.Grow(len(source))
  var tracker quoteTracker
  length := len(source)

  for index := 0; index < length; {
    if !tracker.active() {
      if hasPrefixAt(source, index, "//") {
        index = skipUntil(source, index+2, '\n')
        continue
      }
      if hasPrefixAt(source, index, "/*") {
        index = skipBlockComment(source, index+2)
        continue
      }
    }
    next := tracker.advance(source, index)
    builder.WriteString(source[index : index+next])
    index += next
  }
  return builder.String()
}

func hasPrefixAt(source string, index int, prefix string) bool {
  return index+len(prefix) <= len(source) && source[index:index+len(prefix)] == prefix
}

func skipUntil(source string, index int, stop byte) int {
  for index < len(source) && source[index] != stop {
    index++
  }
  return index
}

func skipBlockComment(source string, index int) int {
  for index+1 < len(source) {
    if source[index] == '*' && source[index+1] == '/' {
      return index + 2
    }
    index++
  }
  return len(source)
}
