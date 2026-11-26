package hostcache

import (
  "encoding/json"
  "fmt"
  "os"
  "path/filepath"
  "time"
)

type envelope struct {
  CheckedAt string          `json:"checkedAt"`
  Value     json.RawMessage `json:"value"`
}

func LoadFresh[T any](path string, now time.Time, ttl time.Duration) (T, bool) {
  var zero T
  data, err := os.ReadFile(path)
  if err != nil {
    return zero, false
  }
  var env envelope
  if err := json.Unmarshal(data, &env); err != nil {
    return zero, false
  }
  checkedAt, err := time.Parse(time.RFC3339, env.CheckedAt)
  if err != nil {
    return zero, false
  }
  if ttl > 0 && now.Sub(checkedAt) > ttl {
    return zero, false
  }
  var value T
  if err := json.Unmarshal(env.Value, &value); err != nil {
    return zero, false
  }
  return value, true
}

func Store[T any](path string, now time.Time, value T) error {
  if now.IsZero() {
    now = time.Now()
  }
  raw, err := json.Marshal(value)
  if err != nil {
    return fmt.Errorf("marshal cache value: %w", err)
  }
  data, err := json.Marshal(envelope{
    CheckedAt: now.UTC().Format(time.RFC3339),
    Value:     raw,
  })
  if err != nil {
    return fmt.Errorf("marshal cache: %w", err)
  }
  if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
    return fmt.Errorf("create cache dir: %w", err)
  }
  if err := os.WriteFile(path, data, 0o644); err != nil {
    return fmt.Errorf("write cache: %w", err)
  }
  return nil
}
