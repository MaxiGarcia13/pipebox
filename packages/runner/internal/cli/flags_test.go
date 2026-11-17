package cli

import "testing"

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
