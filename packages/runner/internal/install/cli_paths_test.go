package install

import (
  "os"
  "path/filepath"
  "runtime"
  "testing"
)

func TestCliGoBinPathPrefersGOBIN(t *testing.T) {
  gobin := t.TempDir()
  t.Setenv("GOBIN", gobin)
  t.Setenv("GOPATH", filepath.Join(t.TempDir(), "unused"))

  got, err := cliGoBinPath()
  if err != nil {
    t.Fatal(err)
  }
  want := filepath.Join(gobin, cliBinaryName())
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
}

func TestCliGoBinPathUsesGOPATH(t *testing.T) {
  gopath := t.TempDir()
  t.Setenv("GOBIN", "")
  t.Setenv("GOPATH", gopath)

  got, err := cliGoBinPath()
  if err != nil {
    t.Fatal(err)
  }
  want := filepath.Join(gopath, "bin", cliBinaryName())
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
}

func TestCliUpdateTargetsUsesExistingManagedPaths(t *testing.T) {
  home := t.TempDir()
  t.Setenv("HOME", home)
  t.Setenv("USERPROFILE", home)
  gobin := filepath.Join(home, "gobin")
  t.Setenv("GOBIN", gobin)
  t.Setenv("GOPATH", "")

  goCLI := filepath.Join(gobin, cliBinaryName())
  if err := os.MkdirAll(gobin, 0o755); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(goCLI, []byte("fake"), 0o755); err != nil {
    t.Fatal(err)
  }

  targets, err := cliUpdateTargets()
  if err != nil {
    t.Fatal(err)
  }
  if len(targets) != 1 || targets[0] != goCLI {
    t.Fatalf("got %#v want [%q]", targets, goCLI)
  }
}

func TestCliUpdateTargetsFallsBackToLocalBin(t *testing.T) {
  home := t.TempDir()
  t.Setenv("HOME", home)
  t.Setenv("USERPROFILE", home)
  t.Setenv("GOBIN", filepath.Join(home, "gobin"))
  t.Setenv("GOPATH", "")

  targets, err := cliUpdateTargets()
  if err != nil {
    t.Fatal(err)
  }
  want := filepath.Join(home, ".local", "bin", cliBinaryName())
  if len(targets) != 1 || targets[0] != want {
    t.Fatalf("got %#v want [%q]", targets, want)
  }
}

func TestUninstallRemovesGoBinAndLocalBin(t *testing.T) {
  if runtime.GOOS == "windows" {
    t.Skip("home path env differs on windows CI shape")
  }

  home := t.TempDir()
  t.Setenv("HOME", home)
  t.Setenv("GOBIN", "")
  t.Setenv("GOPATH", "")

  localCLI := filepath.Join(home, ".local", "bin", cliBinaryName())
  goCLI := filepath.Join(home, "go", "bin", cliBinaryName())
  for _, path := range []string{localCLI, goCLI} {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
      t.Fatal(err)
    }
    if err := os.WriteFile(path, []byte("fake"), 0o755); err != nil {
      t.Fatal(err)
    }
  }

  removed, desktop, err := Uninstall()
  if err != nil {
    t.Fatal(err)
  }
  if desktop != "" {
    t.Fatalf("unexpected desktop path %q", desktop)
  }
  if len(removed) != 2 {
    t.Fatalf("removed %#v", removed)
  }
  for _, path := range []string{localCLI, goCLI} {
    if _, err := os.Stat(path); !os.IsNotExist(err) {
      t.Fatalf("expected %s removed, stat err=%v", path, err)
    }
  }
}

func TestUniquePathsDedupes(t *testing.T) {
  a := filepath.Join("tmp", "pipebox")
  got := uniquePaths(a, a, "")
  if len(got) != 1 || got[0] != filepath.Clean(a) {
    t.Fatalf("got %#v", got)
  }
}
