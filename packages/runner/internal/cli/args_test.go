package cli

import "testing"

func TestParseOpenUI(t *testing.T) {
  t.Parallel()

  result := Parse(nil)
  if result.Kind != KindOpenUI || result.Repo != "" {
    t.Fatalf("got %+v", result)
  }

  result = Parse([]string{"--repo", "."})
  if result.Kind != KindOpenUI || result.Repo != "." {
    t.Fatalf("got %+v", result)
  }

  result = Parse([]string{"--repo=./demo"})
  if result.Kind != KindOpenUI || result.Repo != "./demo" {
    t.Fatalf("got %+v", result)
  }
}

func TestParseHelpAndVersion(t *testing.T) {
  t.Parallel()

  for _, args := range [][]string{
    {"--help"},
    {"-h"},
    {"help"},
    {"--repo", ".", "--help"},
  } {
    result := Parse(args)
    if result.Kind != KindHelp {
      t.Fatalf("%v: got %+v", args, result)
    }
  }

  for _, args := range [][]string{
    {"--version"},
    {"-v"},
    {"version"},
  } {
    result := Parse(args)
    if result.Kind != KindVersion {
      t.Fatalf("%v: got %+v", args, result)
    }
  }
}

func TestParseSubcommands(t *testing.T) {
  t.Parallel()

  result := Parse([]string{"doctor"})
  if result.Kind != KindSubcommand || result.Command != "doctor" {
    t.Fatalf("got %+v", result)
  }

  result = Parse([]string{"list", "--repo", "/tmp"})
  if result.Kind != KindSubcommand || result.Command != "list" {
    t.Fatalf("got %+v", result)
  }
  if len(result.CommandArgs) != 2 || result.CommandArgs[0] != "--repo" {
    t.Fatalf("got args %+v", result.CommandArgs)
  }

  result = Parse([]string{"update", "--check"})
  if result.Kind != KindSubcommand || result.Command != "update" {
    t.Fatalf("got %+v", result)
  }
  if len(result.CommandArgs) != 1 || result.CommandArgs[0] != "--check" {
    t.Fatalf("got args %+v", result.CommandArgs)
  }
}

func TestParseUnknown(t *testing.T) {
  t.Parallel()

  result := Parse([]string{"."})
  if result.Kind != KindError {
    t.Fatalf("got %+v", result)
  }
}
