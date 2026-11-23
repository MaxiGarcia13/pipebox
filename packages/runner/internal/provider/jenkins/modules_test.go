package jenkins

import (
  "os"
  "path/filepath"
  "testing"
)

func TestWorkspaceModuleDirs(t *testing.T) {
  root := t.TempDir()
  for _, dir := range []string{
    filepath.Join(root, "node_modules"),
    filepath.Join(root, "packages", "app", "node_modules"),
    filepath.Join(root, "packages", "runner"),
  } {
    if err := os.MkdirAll(dir, 0o755); err != nil {
      t.Fatal(err)
    }
  }

  got := workspaceModuleDirs(root)
  if len(got) != 2 || got[0] != filepath.Join(root, "node_modules") || got[1] != filepath.Join(root, "packages", "app", "node_modules") {
    t.Fatalf("got %#v", got)
  }
}
