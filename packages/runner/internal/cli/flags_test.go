package cli

import (
  "os"
  "testing"
)

func TestFlagValueAndValues(t *testing.T) {
  t.Parallel()

  args := []string{"--repo", ".", "--input", "a=1", "--input=b=2", "--check"}
  if got := FlagValue(args, "--repo"); got != "." {
    t.Fatalf("FlagValue --repo: got %q", got)
  }
  if got := FlagValues(args, "--input"); len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
    t.Fatalf("FlagValues --input: %#v", got)
  }
  if !HasFlag(args, "--check") {
    t.Fatal("expected --check")
  }
  if HasFlag(args, "--missing") {
    t.Fatal("unexpected --missing")
  }
}

func TestRepoFromArgsUsesFlag(t *testing.T) {
  t.Parallel()

  got, err := RepoFromArgs([]string{"--repo", "/tmp/project"})
  if err != nil {
    t.Fatal(err)
  }
  if got != "/tmp/project" {
    t.Fatalf("got %q", got)
  }

  got, err = RepoFromArgs([]string{"--repo=/tmp/other"})
  if err != nil {
    t.Fatal(err)
  }
  if got != "/tmp/other" {
    t.Fatalf("got %q", got)
  }
}

func TestRepoFromArgsDefaultsToCwd(t *testing.T) {
  previous, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  dir := t.TempDir()
  if err := os.Chdir(dir); err != nil {
    t.Fatal(err)
  }
  t.Cleanup(func() {
    _ = os.Chdir(previous)
  })

  want, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  got, err := RepoFromArgs(nil)
  if err != nil {
    t.Fatal(err)
  }
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
  if got == previous {
    t.Fatal("expected the temporary directory")
  }
}
