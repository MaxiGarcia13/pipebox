package cli

import (
  "sort"
  "strings"
)

func filterPrefix(candidates []string, partial string) []string {
  if partial == "" {
    return candidates
  }
  filtered := make([]string, 0, len(candidates))
  for _, candidate := range candidates {
    if candidate == PathSentinel || strings.HasPrefix(candidate, partial) {
      filtered = append(filtered, candidate)
    }
  }
  return filtered
}

func uniqueSorted(values []string) []string {
  seen := make(map[string]struct{}, len(values))
  out := make([]string, 0, len(values))
  for _, value := range values {
    if value == "" {
      continue
    }
    if _, ok := seen[value]; ok {
      continue
    }
    seen[value] = struct{}{}
    out = append(out, value)
  }
  sort.Strings(out)
  return out
}
