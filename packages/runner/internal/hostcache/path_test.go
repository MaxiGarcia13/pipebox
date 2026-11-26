package hostcache

import (
  "path/filepath"
  "testing"
)

func TestPathJoinsRoot(t *testing.T) {
  root, err := Root()
  if err != nil {
    t.Fatalf("Root: %v", err)
  }
  got, err := Path("update-check.json")
  if err != nil {
    t.Fatalf("Path: %v", err)
  }
  want := filepath.Join(root, "update-check.json")
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
}
