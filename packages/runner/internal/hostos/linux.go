package hostos

import (
  "fmt"
  "os"
  "path/filepath"
)

type linuxPlatform struct{}

func (linuxPlatform) ID() string { return "linux" }

func (linuxPlatform) CLIBinaryName() string { return "pipebox" }

func (p linuxPlatform) CLILocalBinPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", p.CLIBinaryName()), nil
}

func (p linuxPlatform) CLIGoBinPath() (string, error) {
  return cliGoBinPath(p.CLIBinaryName())
}

func (linuxPlatform) DesktopDir() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "share", "pipebox-app"), nil
}

func (linuxPlatform) DesktopLaunchCandidates() []string {
  candidates := []string{
    "/opt/pipebox-app/PipeBox",
    "/usr/bin/pipebox-app",
    "/usr/bin/PipeBox",
  }
  if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
    candidates = append([]string{
      filepath.Join(homeDir, ".local", "share", "pipebox-app", "PipeBox"),
    }, candidates...)
  }
  return candidates
}

func (linuxPlatform) IsLaunchable(path string) bool {
  info, err := os.Stat(path)
  if err != nil || info.IsDir() {
    return false
  }
  return info.Mode()&0o111 != 0
}

func (linuxPlatform) ReplaceFile(source, target string) error {
  return unixReplaceFile(source, target)
}

func (linuxPlatform) RemoveFile(path string) error {
  return unixRemoveFile(path)
}

func (linuxPlatform) RemoveAll(path string) error {
  return unixRemoveAll(path)
}

func (linuxPlatform) PathEqualKey(path string) string {
  return filepath.Clean(path)
}
