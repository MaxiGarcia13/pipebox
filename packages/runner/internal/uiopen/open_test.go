package uiopen

import (
  "os"
  "path/filepath"
  "runtime"
  "testing"
)

func TestLaunchArgsUsesEqualsRepoForm(t *testing.T) {
  t.Parallel()

  got := launchArgs("/tmp/project", false)
  if len(got) != 1 || got[0] != "--repo=/tmp/project" {
    t.Fatalf("launchArgs repo: got %#v", got)
  }

  got = launchArgs("/tmp/project", true)
  if len(got) != 2 || got[0] != "--repo=/tmp/project" || got[1] != "--new" {
    t.Fatalf("launchArgs repo+new: got %#v", got)
  }

  got = launchArgs("", true)
  if len(got) != 1 || got[0] != "--new" {
    t.Fatalf("launchArgs new only: got %#v", got)
  }

  got = launchArgs("", false)
  if len(got) != 0 {
    t.Fatalf("launchArgs empty: got %#v", got)
  }
}

func TestImplicitRepoKeepsExplicitPath(t *testing.T) {
  withCwd(t, repoFixture(t, "github-only"))
  if got := implicitRepo("/tmp/other"); got != "/tmp/other" {
    t.Fatalf("got %q", got)
  }
}

func TestImplicitRepoUsesCwdWhenProvidersPresent(t *testing.T) {
  for _, name := range []string{"github-only", "jenkins-only", "mixed"} {
    t.Run(name, func(t *testing.T) {
      withCwd(t, repoFixture(t, name))
      want, err := os.Getwd()
      if err != nil {
        t.Fatal(err)
      }
      if got := implicitRepo(""); got != want {
        t.Fatalf("got %q want %q", got, want)
      }
    })
  }
}

func TestImplicitRepoSkipsCwdWithoutProviders(t *testing.T) {
  withCwd(t, t.TempDir())
  if got := implicitRepo(""); got != "" {
    t.Fatalf("got %q", got)
  }
}

func withCwd(t *testing.T, dir string) {
  t.Helper()
  previous, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  if err := os.Chdir(dir); err != nil {
    t.Fatal(err)
  }
  t.Cleanup(func() {
    _ = os.Chdir(previous)
  })
}

func repoFixture(t *testing.T, name string) string {
  t.Helper()
  _, file, _, ok := runtime.Caller(0)
  if !ok {
    t.Fatal("runtime.Caller failed")
  }
  path := filepath.Join(filepath.Dir(file), "..", "..", "testdata", "repos", name)
  if _, err := os.Stat(path); err != nil {
    t.Fatalf("fixture %s: %v", name, err)
  }
  return path
}
