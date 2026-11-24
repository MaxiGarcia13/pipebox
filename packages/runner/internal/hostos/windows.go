package hostos

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

type windowsPlatform struct{}

func (windowsPlatform) ID() string { return "windows" }

func (windowsPlatform) CLIBinaryName() string { return "pipebox.exe" }

func (p windowsPlatform) CLILocalBinPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", p.CLIBinaryName()), nil
}

func (p windowsPlatform) CLIGoBinPath() (string, error) {
  return cliGoBinPath(p.CLIBinaryName())
}

func (windowsPlatform) DesktopDir() (string, error) {
  localAppData := os.Getenv("LOCALAPPDATA")
  if localAppData == "" {
    return "", fmt.Errorf("LOCALAPPDATA is not set")
  }
  return filepath.Join(localAppData, "PipeBox"), nil
}

func (p windowsPlatform) DesktopLaunchCandidates() []string {
  dir, err := p.DesktopDir()
  if err != nil {
    return nil
  }
  return []string{filepath.Join(dir, "PipeBox.exe")}
}

func (windowsPlatform) IsLaunchable(path string) bool {
  info, err := os.Stat(path)
  if err != nil || info.IsDir() {
    return false
  }
  return strings.EqualFold(filepath.Ext(path), ".exe")
}

func (windowsPlatform) ReplaceFile(source, target string) error {
  removeQuiet(oldPath(target))
  if _, err := os.Stat(target); err == nil {
    if renameErr := os.Rename(target, oldPath(target)); renameErr != nil {
      return fmt.Errorf("rename aside %s: %w", target, renameErr)
    }
  } else if !os.IsNotExist(err) {
    return fmt.Errorf("stat target: %w", err)
  }

  if err := CopyFile(source, target); err != nil {
    // Best-effort restore if the new write failed after rename.
    if _, statErr := os.Stat(oldPath(target)); statErr == nil {
      _ = os.Rename(oldPath(target), target)
    }
    return err
  }
  if err := chmodExecutable(target); err != nil {
    return err
  }
  removeQuiet(oldPath(target))
  return nil
}

func (windowsPlatform) RemoveFile(path string) error {
  removeQuiet(oldPath(path))
  removeErr := os.Remove(path)
  if removeErr == nil {
    return nil
  }
  if !isAccessDenied(removeErr) {
    return fmt.Errorf("remove %s: %w", path, removeErr)
  }

  if renameErr := os.Rename(path, oldPath(path)); renameErr != nil {
    return fmt.Errorf("remove %s: %w", path, removeErr)
  }
  removeQuiet(oldPath(path))
  return nil
}

func (windowsPlatform) RemoveAll(path string) error {
  if err := os.RemoveAll(path); err == nil {
    return nil
  } else if !isAccessDenied(err) {
    return fmt.Errorf("remove %s: %w", path, err)
  }

  if err := filepath.WalkDir(path, func(entryPath string, entry os.DirEntry, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    if entry.IsDir() {
      return nil
    }
    if !strings.EqualFold(filepath.Ext(entryPath), ".exe") {
      return nil
    }
    removeQuiet(oldPath(entryPath))
    if removeErr := os.Remove(entryPath); removeErr == nil {
      return nil
    } else if !isAccessDenied(removeErr) {
      return removeErr
    }
    if renameErr := os.Rename(entryPath, oldPath(entryPath)); renameErr != nil {
      return renameErr
    }
    removeQuiet(oldPath(entryPath))
    return nil
  }); err != nil {
    return fmt.Errorf("prepare remove %s: %w", path, err)
  }

  if err := os.RemoveAll(path); err != nil {
    return fmt.Errorf("remove %s: %w", path, err)
  }
  return nil
}

func (windowsPlatform) PathEqualKey(path string) string {
  return strings.ToLower(filepath.Clean(path))
}
