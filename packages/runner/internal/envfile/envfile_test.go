package envfile

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestResolveFindsEnvVariants(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  for _, name := range []string{".env", ".env.local", ".env.development", "README.md"} {
    if err := os.WriteFile(filepath.Join(repoPath, name), []byte("A=1\n"), 0o600); err != nil {
      t.Fatal(err)
    }
  }
  if err := os.Mkdir(filepath.Join(repoPath, ".env.dir"), 0o700); err != nil {
    t.Fatal(err)
  }

  files, err := Resolve(repoPath, "")
  if err != nil {
    t.Fatal(err)
  }
  got := make([]string, 0, len(files))
  for _, file := range files {
    got = append(got, filepath.Base(file))
  }
  want := []string{".env", ".env.development", ".env.local"}
  if len(got) != len(want) {
    t.Fatalf("got %#v, want %#v", got, want)
  }
  for index := range want {
    if got[index] != want[index] {
      t.Fatalf("got %#v, want %#v", got, want)
    }
  }
}

func TestResolveExplicitOverride(t *testing.T) {
  t.Parallel()

  repoPath := t.TempDir()
  if err := os.WriteFile(filepath.Join(repoPath, ".env"), []byte("A=1\n"), 0o600); err != nil {
    t.Fatal(err)
  }
  explicit := filepath.Join(repoPath, "custom.env")
  if err := os.WriteFile(explicit, []byte("B=1\n"), 0o600); err != nil {
    t.Fatal(err)
  }

  files, err := Resolve(repoPath, explicit)
  if err != nil {
    t.Fatal(err)
  }
  if len(files) != 1 || files[0] != explicit {
    t.Fatalf("got %#v, want [%q]", files, explicit)
  }
}

func TestIsName(t *testing.T) {
  t.Parallel()

  cases := map[string]bool{
    ".env":            true,
    ".env.local":      true,
    ".env.production": true,
    ".environment":    false,
    "env":             false,
    ".envrc":          false,
  }
  for name, want := range cases {
    if got := IsName(name); got != want {
      t.Fatalf("%q: got %v, want %v", name, got, want)
    }
  }
}

func TestLoad(t *testing.T) {
  t.Parallel()

  repo := t.TempDir()
  path := filepath.Join(repo, ".env")
  if err := os.WriteFile(path, []byte("A=1\nB=2\n"), 0o644); err != nil {
    t.Fatal(err)
  }
  pairs, err := Load([]string{path}, []string{"C=3"})
  if err != nil {
    t.Fatal(err)
  }
  joined := strings.Join(pairs, ",")
  if !strings.Contains(joined, "A=1") || !strings.Contains(joined, "C=3") {
    t.Fatalf("pairs: %#v", pairs)
  }
}
