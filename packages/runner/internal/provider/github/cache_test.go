package github

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestPrepareCacheRuntimeCreatesDir(t *testing.T) {
  t.Parallel()

  runtime, err := prepareCacheRuntime()
  if err != nil {
    t.Fatal(err)
  }
  if runtime.Port <= 0 {
    t.Fatalf("expected port, got %d", runtime.Port)
  }
  info, err := os.Stat(runtime.Dir)
  if err != nil {
    t.Fatal(err)
  }
  if !info.IsDir() {
    t.Fatalf("expected directory at %s", runtime.Dir)
  }
  if !strings.Contains(runtime.Dir, filepath.Join("pipebox-act-cache")) {
    t.Fatalf("expected pipebox-act-cache path, got %q", runtime.Dir)
  }
}

func TestPrepareCacheRuntimeUniqueDirs(t *testing.T) {
  t.Parallel()

  first, err := prepareCacheRuntime()
  if err != nil {
    t.Fatal(err)
  }
  second, err := prepareCacheRuntime()
  if err != nil {
    t.Fatal(err)
  }
  if first.Dir == second.Dir {
    t.Fatalf("expected unique dirs, both %q", first.Dir)
  }
}
