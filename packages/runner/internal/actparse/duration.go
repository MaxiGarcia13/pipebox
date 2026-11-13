package actparse

import (
  "strings"
  "time"
)

func parseActDurationMs(raw string) int64 {
  text := strings.TrimSpace(raw)
  if text == "" {
    return 0
  }
  duration, err := time.ParseDuration(text)
  if err != nil {
    return 0
  }
  if duration <= 0 {
    return 0
  }
  return duration.Milliseconds()
}
