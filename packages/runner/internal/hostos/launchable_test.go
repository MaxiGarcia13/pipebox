package hostos

import (
  "os"
  "path/filepath"
  "runtime"
  "testing"
)

func TestWindowsIsLaunchableAcceptsExeWithoutExecuteBits(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  exePath := filepath.Join(dir, "PipeBox.exe")
  if err := os.WriteFile(exePath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake exe: %v", err)
  }

  if !(windowsPlatform{}).IsLaunchable(exePath) {
    t.Fatalf("expected %s to be launchable with mode 0666", exePath)
  }
}

func TestWindowsIsLaunchableRejectsNonExe(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  dllPath := filepath.Join(dir, "PipeBox.dll")
  if err := os.WriteFile(dllPath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake dll: %v", err)
  }

  if (windowsPlatform{}).IsLaunchable(dllPath) {
    t.Fatalf("expected %s not to be launchable", dllPath)
  }
}

func TestLinuxIsLaunchableRequiresExecuteBits(t *testing.T) {
  t.Parallel()
  if runtime.GOOS == "windows" {
    t.Skip("unix execute bits are not preserved on windows filesystems")
  }

  dir := t.TempDir()
  binaryPath := filepath.Join(dir, "PipeBox")
  if err := os.WriteFile(binaryPath, []byte("fake"), 0o666); err != nil {
    t.Fatalf("write fake binary: %v", err)
  }

  plat := linuxPlatform{}
  if plat.IsLaunchable(binaryPath) {
    t.Fatalf("expected %s not to be launchable without execute bits", binaryPath)
  }

  if err := os.Chmod(binaryPath, 0o755); err != nil {
    t.Fatalf("chmod: %v", err)
  }
  if !plat.IsLaunchable(binaryPath) {
    t.Fatalf("expected %s to be launchable with execute bits", binaryPath)
  }
}
