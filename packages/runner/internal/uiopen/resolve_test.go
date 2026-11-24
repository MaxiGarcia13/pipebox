package uiopen

import (
  "os"
  "path/filepath"
  "runtime"
  "testing"
)

func TestIsLaunchableOnWindowsAcceptsExeWithoutExecuteBits(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  exePath := filepath.Join(dir, "PipeBox.exe")
  if err := os.WriteFile(exePath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake exe: %v", err)
  }

  if !isLaunchableOn(exePath, "windows") {
    t.Fatalf("expected %s to be launchable on windows with mode 0666", exePath)
  }
}

func TestIsLaunchableOnWindowsRejectsNonExe(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  dllPath := filepath.Join(dir, "PipeBox.dll")
  if err := os.WriteFile(dllPath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake dll: %v", err)
  }

  if isLaunchableOn(dllPath, "windows") {
    t.Fatalf("expected %s not to be launchable on windows", dllPath)
  }
}

func TestIsLaunchableOnUnixRequiresExecuteBits(t *testing.T) {
  t.Parallel()
  if runtime.GOOS == "windows" {
    t.Skip("unix execute bits are not preserved on windows filesystems")
  }

  dir := t.TempDir()
  binaryPath := filepath.Join(dir, "PipeBox")
  if err := os.WriteFile(binaryPath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake binary: %v", err)
  }

  if isLaunchableOn(binaryPath, "linux") {
    t.Fatalf("expected %s not to be launchable on linux without execute bits", binaryPath)
  }

  if err := os.Chmod(binaryPath, 0o755); err != nil {
    t.Fatalf("chmod: %v", err)
  }
  if !isLaunchableOn(binaryPath, "linux") {
    t.Fatalf("expected %s to be launchable on linux with execute bits", binaryPath)
  }
}
