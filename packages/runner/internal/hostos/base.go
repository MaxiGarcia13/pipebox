package hostos

import (
  "fmt"
  "io"
  "os"
  "strings"
)

const oldSuffix = ".old"

func CopyFile(sourcePath, targetPath string) (err error) {
  source, err := os.Open(sourcePath)
  if err != nil {
    return fmt.Errorf("open source: %w", err)
  }
  defer func() {
    if closeErr := source.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close source: %w", closeErr)
    }
  }()

  target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
  if err != nil {
    return fmt.Errorf("open target: %w", err)
  }
  defer func() {
    if closeErr := target.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close target: %w", closeErr)
    }
  }()

  if _, err := io.Copy(target, source); err != nil {
    return fmt.Errorf("copy binary: %w", err)
  }
  return nil
}

func chmodExecutable(path string) error {
  if err := os.Chmod(path, 0o755); err != nil {
    return fmt.Errorf("chmod: %w", err)
  }
  return nil
}

func oldPath(path string) string {
  return path + oldSuffix
}

func removeQuiet(path string) {
  _ = os.Remove(path)
}

func isAccessDenied(err error) bool {
  if err == nil {
    return false
  }
  msg := strings.ToLower(err.Error())
  return strings.Contains(msg, "access is denied") ||
    strings.Contains(msg, "being used by another process") ||
    strings.Contains(msg, "permission denied")
}

func unixReplaceFile(source, target string) error {
  removeQuiet(oldPath(target))
  if err := CopyFile(source, target); err != nil {
    return err
  }
  return chmodExecutable(target)
}

func unixRemoveFile(path string) error {
  if err := os.Remove(path); err != nil {
    return fmt.Errorf("remove %s: %w", path, err)
  }
  return nil
}

func unixRemoveAll(path string) error {
  if err := os.RemoveAll(path); err != nil {
    return fmt.Errorf("remove %s: %w", path, err)
  }
  return nil
}
