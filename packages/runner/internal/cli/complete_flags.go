package cli

import "strings"

func awaitingFlagValue(flags []Flag, words []string) (awaiting bool, flag Flag, ok bool) {
  for index := 0; index < len(words); index++ {
    word := words[index]
    if name, value, hasValue := splitInlineFlag(word); hasValue {
      if matched, found := findFlag(flags, name); found && matched.ValueKind != ValueNone {
        flag = matched
        ok = true
        awaiting = value == ""
      }
      continue
    }
    if matched, found := findFlag(flags, word); found {
      flag = matched
      ok = true
      if matched.ValueKind == ValueNone {
        awaiting = false
        continue
      }
      if index+1 >= len(words) {
        awaiting = true
        return awaiting, flag, ok
      }
      index++
      awaiting = false
      continue
    }
  }
  return awaiting, flag, ok
}

func unusedFlagNames(flags []Flag, words []string) []string {
  used := usedFlagNames(flags, words)
  names := make([]string, 0)
  for _, flag := range flags {
    primary := flag.Names[0]
    if !flag.Repeatable {
      if _, seen := used[primary]; seen {
        continue
      }
    }
    names = append(names, flag.Names...)
  }
  return uniqueSorted(names)
}

func usedFlagNames(flags []Flag, words []string) map[string]struct{} {
  used := make(map[string]struct{})
  for index := 0; index < len(words); index++ {
    word := words[index]
    name := word
    if inlineName, _, ok := splitInlineFlag(word); ok {
      name = inlineName
    }
    flag, found := findFlag(flags, name)
    if !found {
      continue
    }
    primary := flag.Names[0]
    used[primary] = struct{}{}
    if flag.ValueKind != ValueNone && !strings.Contains(word, "=") && index+1 < len(words) {
      index++
    }
  }
  return used
}
