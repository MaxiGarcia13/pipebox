package hostos

import (
  "os"
  "path/filepath"
  "testing"
)

func TestWindowsReplaceFileOverwritesExisting(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  target := filepath.Join(dir, "pipebox.exe")
  source := filepath.Join(dir, "new.exe")
  if err := os.WriteFile(target, []byte("old"), 0o666); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(source, []byte("new"), 0o666); err != nil {
    t.Fatal(err)
  }

  if err := (windowsPlatform{}).ReplaceFile(source, target); err != nil {
    t.Fatalf("ReplaceFile: %v", err)
  }

  got, err := os.ReadFile(target)
  if err != nil {
    t.Fatal(err)
  }
  if string(got) != "new" {
    t.Fatalf("got %q want %q", got, "new")
  }
  if _, err := os.Stat(oldPath(target)); !os.IsNotExist(err) {
    t.Fatalf("expected .old cleaned up, stat err=%v", err)
  }
}

func TestWindowsReplaceFileCreatesMissingTarget(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  target := filepath.Join(dir, "pipebox.exe")
  source := filepath.Join(dir, "new.exe")
  if err := os.WriteFile(source, []byte("fresh"), 0o666); err != nil {
    t.Fatal(err)
  }

  if err := (windowsPlatform{}).ReplaceFile(source, target); err != nil {
    t.Fatalf("ReplaceFile: %v", err)
  }

  got, err := os.ReadFile(target)
  if err != nil {
    t.Fatal(err)
  }
  if string(got) != "fresh" {
    t.Fatalf("got %q want %q", got, "fresh")
  }
}

func TestWindowsRemoveFile(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  path := filepath.Join(dir, "pipebox.exe")
  if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
    t.Fatal(err)
  }

  if err := (windowsPlatform{}).RemoveFile(path); err != nil {
    t.Fatalf("RemoveFile: %v", err)
  }
  if _, err := os.Stat(path); !os.IsNotExist(err) {
    t.Fatalf("expected removed, stat err=%v", err)
  }
}

func TestWindowsPathEqualKeyCaseFolds(t *testing.T) {
  t.Parallel()

  plat := windowsPlatform{}
  a := plat.PathEqualKey(`C:\Users\Maxi\go\bin\pipebox.exe`)
  b := plat.PathEqualKey(`c:\users\maxi\go\bin\PIPEBOX.EXE`)
  if a != b {
    t.Fatalf("PathEqualKey mismatch: %q vs %q", a, b)
  }
}

func TestIsAccessDenied(t *testing.T) {
  t.Parallel()

  if !isAccessDenied(errString("Access is denied.")) {
    t.Fatal("expected access denied match")
  }
  if !isAccessDenied(errString("The process cannot access the file because it is being used by another process.")) {
    t.Fatal("expected in-use match")
  }
  if !isAccessDenied(errString("permission denied")) {
    t.Fatal("expected permission denied match")
  }
  if isAccessDenied(errString("no such file")) {
    t.Fatal("did not expect match")
  }
}

type errString string

func (e errString) Error() string { return string(e) }
