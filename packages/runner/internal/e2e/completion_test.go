//go:build e2e

package e2e_test

import (
  "path/filepath"
  "strings"
  "testing"
)

func TestCompletionScript(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "completion", "zsh")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  if !strings.Contains(stdout, "pipebox __complete") {
    t.Fatalf("expected __complete hook, got %q", stdout)
  }
}

func TestCompleteRootCommands(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "__complete", "--", "")
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  lines := strings.Split(strings.TrimSpace(stdout), "\n")
  found := map[string]bool{}
  for _, line := range lines {
    found[line] = true
  }
  for _, want := range []string{"doctor", "list", "run", "completion"} {
    if !found[want] {
      t.Fatalf("missing %q in %#v", want, lines)
    }
  }
  if found["__complete"] {
    t.Fatal("hidden __complete should not be suggested")
  }
}

func TestCompleteWorkflows(t *testing.T) {
  repo := repoPath(t, "github-only")
  stdout, stderr, code := runPipebox(
    t,
    "__complete",
    "--",
    "run",
    "--repo",
    repo,
    "--workflow",
    "",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }
  want := filepath.ToSlash(filepath.Join(".github", "workflows", "ci.yml"))
  if !strings.Contains(filepath.ToSlash(stdout), want) {
    t.Fatalf("expected workflow %q in %q", want, stdout)
  }
}
