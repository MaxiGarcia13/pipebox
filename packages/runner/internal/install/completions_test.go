package install

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestInstallAndUninstallCompletions(t *testing.T) {
  home := t.TempDir()
  t.Setenv("HOME", home)

  installed, err := InstallCompletions()
  if err != nil {
    t.Fatalf("install: %v", err)
  }
  if len(installed) == 0 {
    t.Fatal("expected completion files")
  }
  for _, path := range installed {
    raw, readErr := os.ReadFile(path)
    if readErr != nil {
      t.Fatalf("read %s: %v", path, readErr)
    }
    text := string(raw)
    if !strings.Contains(text, "pipebox __complete") {
      t.Fatalf("%s missing __complete hook", path)
    }
  }

  removed, err := UninstallCompletions()
  if err != nil {
    t.Fatalf("uninstall: %v", err)
  }
  if len(removed) == 0 {
    t.Fatal("expected removed completion files")
  }
  for _, path := range removed {
    if _, err := os.Stat(path); !os.IsNotExist(err) {
      t.Fatalf("expected removed %s", path)
    }
  }
}

func TestCompletionTargetsIncludeBashZshFish(t *testing.T) {
  t.Parallel()

  targets := completionTargets(filepath.Join("tmp", "home"))
  shells := map[string]bool{}
  for _, target := range targets {
    shells[target.Shell] = true
  }
  for _, shell := range []string{"zsh", "bash", "fish"} {
    if !shells[shell] {
      t.Fatalf("missing %s target", shell)
    }
  }
}
