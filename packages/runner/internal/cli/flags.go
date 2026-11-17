package cli

import "strings"

func HasFlag(args []string, name string) bool {
  for _, arg := range args {
    if arg == name {
      return true
    }
  }
  return false
}

func FlagValue(args []string, name string) string {
  values := FlagValues(args, name)
  if len(values) == 0 {
    return ""
  }
  return values[0]
}

func FlagValues(args []string, name string) []string {
  values := make([]string, 0)
  prefix := name + "="
  for index := 0; index < len(args); index++ {
    arg := args[index]
    if arg == name {
      if index+1 >= len(args) {
        continue
      }
      values = append(values, args[index+1])
      index++
      continue
    }
    if strings.HasPrefix(arg, prefix) {
      values = append(values, arg[len(prefix):])
    }
  }
  return values
}

func splitInlineFlag(word string) (name, value string, ok bool) {
  if !strings.HasPrefix(word, "--") {
    return "", "", false
  }
  name, value, found := strings.Cut(word, "=")
  if !found {
    return "", "", false
  }
  return name, value, true
}
