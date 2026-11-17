package cli

import (
  "reflect"
  "strings"
  "testing"
)

type stubResolver struct {
  workflows []string
  jobs      []string
  providers []string
}

func (s stubResolver) Workflows(string) []string { return s.workflows }
func (s stubResolver) Jobs(string, string, string) []string {
  return s.jobs
}
func (s stubResolver) Providers(string) []string { return s.providers }

func TestCompleteRootSubcommands(t *testing.T) {
  t.Parallel()

  got := Complete([]string{""}, CompleteOptions{Catalog: DefaultCatalog()})
  assertContains(t, got, "doctor", "list", "run", "completion")
  assertNotContains(t, got, "__complete")
}

func TestCompleteRootPartial(t *testing.T) {
  t.Parallel()

  got := Complete([]string{"li"}, CompleteOptions{Catalog: DefaultCatalog()})
  if !reflect.DeepEqual(got, []string{"list"}) {
    t.Fatalf("got %#v", got)
  }
}

func TestCompleteRunFlags(t *testing.T) {
  t.Parallel()

  got := Complete([]string{"run", ""}, CompleteOptions{Catalog: DefaultCatalog()})
  assertContains(t, got, "--repo", "--workflow", "--job", "--provider")
}

func TestCompleteSkipsUsedFlags(t *testing.T) {
  t.Parallel()

  got := Complete(
    []string{"run", "--repo", ".", ""},
    CompleteOptions{Catalog: DefaultCatalog()},
  )
  assertNotContains(t, got, "--repo")
  assertContains(t, got, "--workflow")
}

func TestCompletePathSentinel(t *testing.T) {
  t.Parallel()

  got := Complete(
    []string{"list", "--repo", ""},
    CompleteOptions{Catalog: DefaultCatalog()},
  )
  if !reflect.DeepEqual(got, []string{PathSentinel}) {
    t.Fatalf("got %#v", got)
  }
}

func TestCompleteWorkflows(t *testing.T) {
  t.Parallel()

  got := Complete(
    []string{"run", "--repo", ".", "--workflow", ""},
    CompleteOptions{
      Catalog: DefaultCatalog(),
      Resolver: stubResolver{
        workflows: []string{".github/workflows/ci.yml", ".github/workflows/release.yml"},
      },
    },
  )
  if !reflect.DeepEqual(got, []string{
    ".github/workflows/ci.yml",
    ".github/workflows/release.yml",
  }) {
    t.Fatalf("got %#v", got)
  }
}

func TestCompleteJobs(t *testing.T) {
  t.Parallel()

  got := Complete(
    []string{"run", "--repo", ".", "--workflow", "ci.yml", "--job", "t"},
    CompleteOptions{
      Catalog:  DefaultCatalog(),
      Resolver: stubResolver{jobs: []string{"test", "build"}},
    },
  )
  if !reflect.DeepEqual(got, []string{"test"}) {
    t.Fatalf("got %#v", got)
  }
}

func TestCompleteShells(t *testing.T) {
  t.Parallel()

  got := Complete([]string{"completion", ""}, CompleteOptions{Catalog: DefaultCatalog()})
  if !reflect.DeepEqual(got, []string{"bash", "fish", "zsh"}) {
    t.Fatalf("got %#v", got)
  }
}

func TestCompletionScript(t *testing.T) {
  t.Parallel()

  for _, shell := range []string{"zsh", "bash", "fish"} {
    script, err := CompletionScript(shell)
    if err != nil {
      t.Fatalf("%s: %v", shell, err)
    }
    if script == "" {
      t.Fatalf("%s: empty script", shell)
    }
    if !containsAll(script, "pipebox", "__complete") {
      t.Fatalf("%s script missing protocol hooks", shell)
    }
  }

  if _, err := CompletionScript("powershell"); err == nil {
    t.Fatal("expected error for unsupported shell")
  }
}

func TestUsageIncludesCommands(t *testing.T) {
  t.Parallel()

  usage := Usage()
  if !containsAll(usage, "pipebox doctor", "pipebox run", "pipebox completion") {
    t.Fatalf("usage missing commands:\n%s", usage)
  }
  if containsAll(usage, "__complete") {
    t.Fatal("hidden __complete should not appear in usage")
  }
}

func assertContains(t *testing.T, values []string, want ...string) {
  t.Helper()
  set := make(map[string]struct{}, len(values))
  for _, value := range values {
    set[value] = struct{}{}
  }
  for _, item := range want {
    if _, ok := set[item]; !ok {
      t.Fatalf("missing %q in %#v", item, values)
    }
  }
}

func assertNotContains(t *testing.T, values []string, banned ...string) {
  t.Helper()
  set := make(map[string]struct{}, len(values))
  for _, value := range values {
    set[value] = struct{}{}
  }
  for _, item := range banned {
    if _, ok := set[item]; ok {
      t.Fatalf("unexpected %q in %#v", item, values)
    }
  }
}

func containsAll(text string, parts ...string) bool {
  for _, part := range parts {
    if !strings.Contains(text, part) {
      return false
    }
  }
  return true
}
