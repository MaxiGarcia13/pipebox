package hostos

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

type darwinPlatform struct{}

func (darwinPlatform) ID() string { return "darwin" }

func (darwinPlatform) CLIBinaryName() string { return "pipebox" }

func (p darwinPlatform) CLILocalBinPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", p.CLIBinaryName()), nil
}

func (p darwinPlatform) CLIGoBinPath() (string, error) {
  return cliGoBinPath(p.CLIBinaryName())
}

func (darwinPlatform) DesktopDir() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, "Applications", "PipeBox.app"), nil
}

func (darwinPlatform) DesktopLaunchCandidates() []string {
  candidates := []string{"/Applications/PipeBox.app"}
  if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
    candidates = append(candidates, filepath.Join(homeDir, "Applications", "PipeBox.app"))
  }
  return candidates
}

func (darwinPlatform) IsLaunchable(path string) bool {
  info, err := os.Stat(path)
  if err != nil {
    return false
  }
  if info.IsDir() && filepath.Ext(path) == ".app" {
    executable := MacOSAppExecutable(path)
    execInfo, execErr := os.Stat(executable)
    if execErr != nil || execInfo.IsDir() || execInfo.Mode()&0o111 == 0 {
      return false
    }
    // Incomplete Forge packages may ship a stub binary without Electron Framework.
    framework := filepath.Join(path, "Contents", "Frameworks", "Electron Framework.framework")
    _, frameworkErr := os.Stat(framework)
    return frameworkErr == nil
  }
  if info.IsDir() {
    return false
  }
  return info.Mode()&0o111 != 0
}

func (darwinPlatform) ReplaceFile(source, target string) error {
  return unixReplaceFile(source, target)
}

func (darwinPlatform) RemoveFile(path string) error {
  return unixRemoveFile(path)
}

func (darwinPlatform) RemoveAll(path string) error {
  return unixRemoveAll(path)
}

func (darwinPlatform) PathEqualKey(path string) string {
  return filepath.Clean(path)
}

func MacOSAppExecutable(appPath string) string {
  base := strings.TrimSuffix(filepath.Base(appPath), ".app")
  return filepath.Join(appPath, "Contents", "MacOS", base)
}
