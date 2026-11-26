package install

import (
  "bytes"
  "errors"
  "os"
  "path/filepath"
  "strings"
  "testing"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostcache"
)

func TestMaybeSuggestUpdateSkipsNonTTY(t *testing.T) {
  var out bytes.Buffer
  fetches := 0
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.1.0",
    errOut:         &out,
    errTTY:         false,
    now:            time.Now(),
    ttl:            time.Hour,
    cachePath:      filepath.Join(t.TempDir(), "update-check.json"),
    fetchLatest: func() (string, error) {
      fetches++
      return "0.2.0", nil
    },
  })
  if fetches != 0 {
    t.Fatalf("expected no fetch, got %d", fetches)
  }
  if out.Len() != 0 {
    t.Fatalf("expected no output, got %q", out.String())
  }
}

func TestMaybeSuggestUpdatePrintsWhenNewer(t *testing.T) {
  var out bytes.Buffer
  cachePath := filepath.Join(t.TempDir(), "update-check.json")
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.1.0",
    errOut:         &out,
    errTTY:         true,
    now:            time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
    ttl:            time.Hour,
    cachePath:      cachePath,
    fetchLatest:    func() (string, error) { return "0.2.0", nil },
  })
  got := out.String()
  if !strings.Contains(got, "Update available: 0.1.0 → 0.2.0") {
    t.Fatalf("missing update line: %q", got)
  }
  if !strings.Contains(got, "pipebox update") {
    t.Fatalf("missing update command: %q", got)
  }
  if _, err := os.Stat(cachePath); err != nil {
    t.Fatalf("expected cache file: %v", err)
  }
}

func TestMaybeSuggestUpdateSilentWhenCurrent(t *testing.T) {
  var out bytes.Buffer
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.2.0",
    errOut:         &out,
    errTTY:         true,
    now:            time.Now(),
    ttl:            time.Hour,
    cachePath:      filepath.Join(t.TempDir(), "update-check.json"),
    fetchLatest:    func() (string, error) { return "v0.2.0", nil },
  })
  if out.Len() != 0 {
    t.Fatalf("expected no output, got %q", out.String())
  }
}

func TestMaybeSuggestUpdateUsesCacheWithinTTL(t *testing.T) {
  cachePath := filepath.Join(t.TempDir(), "update-check.json")
  now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
  if err := hostcache.Store(cachePath, now.Add(-30*time.Minute), "0.9.0"); err != nil {
    t.Fatalf("Store: %v", err)
  }

  var out bytes.Buffer
  fetches := 0
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.1.0",
    errOut:         &out,
    errTTY:         true,
    now:            now,
    ttl:            time.Hour,
    cachePath:      cachePath,
    fetchLatest: func() (string, error) {
      fetches++
      return "0.9.0", nil
    },
  })
  if fetches != 0 {
    t.Fatalf("expected cached check, got %d fetches", fetches)
  }
  if out.Len() != 0 {
    t.Fatalf("expected no re-suggest within TTL, got %q", out.String())
  }
}

func TestMaybeSuggestUpdateRefetchesAfterTTL(t *testing.T) {
  cachePath := filepath.Join(t.TempDir(), "update-check.json")
  now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
  if err := hostcache.Store(cachePath, now.Add(-2*time.Hour), "0.1.0"); err != nil {
    t.Fatalf("Store: %v", err)
  }

  var out bytes.Buffer
  fetches := 0
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.1.0",
    errOut:         &out,
    errTTY:         true,
    now:            now,
    ttl:            time.Hour,
    cachePath:      cachePath,
    fetchLatest: func() (string, error) {
      fetches++
      return "0.3.0", nil
    },
  })
  if fetches != 1 {
    t.Fatalf("expected one fetch after TTL, got %d", fetches)
  }
  if !strings.Contains(out.String(), "0.1.0 → 0.3.0") {
    t.Fatalf("expected suggestion, got %q", out.String())
  }
}

func TestMaybeSuggestUpdateIgnoresFetchErrors(t *testing.T) {
  var out bytes.Buffer
  maybeSuggestUpdate(suggestConfig{
    currentVersion: "0.1.0",
    errOut:         &out,
    errTTY:         true,
    now:            time.Now(),
    ttl:            time.Hour,
    cachePath:      filepath.Join(t.TempDir(), "update-check.json"),
    fetchLatest:    func() (string, error) { return "", errors.New("network down") },
  })
  if out.Len() != 0 {
    t.Fatalf("expected no output on fetch error, got %q", out.String())
  }
}
