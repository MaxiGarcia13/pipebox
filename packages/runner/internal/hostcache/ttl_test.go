package hostcache

import (
  "path/filepath"
  "testing"
  "time"
)

func TestStoreLoadFreshWithinTTL(t *testing.T) {
  path := filepath.Join(t.TempDir(), "entry.json")
  now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
  if err := Store(path, now, "0.2.0"); err != nil {
    t.Fatalf("Store: %v", err)
  }
  got, ok := LoadFresh[string](path, now.Add(30*time.Minute), time.Hour)
  if !ok {
    t.Fatal("expected fresh cache")
  }
  if got != "0.2.0" {
    t.Fatalf("got %q", got)
  }
}

func TestLoadFreshExpired(t *testing.T) {
  path := filepath.Join(t.TempDir(), "entry.json")
  now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
  if err := Store(path, now, "0.2.0"); err != nil {
    t.Fatalf("Store: %v", err)
  }
  if _, ok := LoadFresh[string](path, now.Add(2*time.Hour), time.Hour); ok {
    t.Fatal("expected expired cache")
  }
}

func TestLoadFreshMissing(t *testing.T) {
  path := filepath.Join(t.TempDir(), "missing.json")
  if _, ok := LoadFresh[string](path, time.Now(), time.Hour); ok {
    t.Fatal("expected miss")
  }
}

func TestLoadFreshStruct(t *testing.T) {
  type payload struct {
    Latest string `json:"latest"`
  }
  path := filepath.Join(t.TempDir(), "entry.json")
  now := time.Now().UTC()
  if err := Store(path, now, payload{Latest: "1.0.0"}); err != nil {
    t.Fatalf("Store: %v", err)
  }
  got, ok := LoadFresh[payload](path, now, time.Hour)
  if !ok {
    t.Fatal("expected fresh cache")
  }
  if got.Latest != "1.0.0" {
    t.Fatalf("got %+v", got)
  }
}
