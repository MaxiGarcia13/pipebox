package uiopen

import "testing"

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
