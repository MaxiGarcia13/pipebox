package install

import (
  "fmt"
  "io"
  "os"
  "path/filepath"
)

func Run() (string, error) {
  executablePath, err := os.Executable()
  if err != nil {
    return "", fmt.Errorf("resolve executable: %w", err)
  }

  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }

  binDir := filepath.Join(homeDir, ".local", "bin")
  if err := os.MkdirAll(binDir, 0o755); err != nil {
    return "", fmt.Errorf("create bin dir: %w", err)
  }

  targetPath := filepath.Join(binDir, "pipebox")
  if err := copyFile(executablePath, targetPath); err != nil {
    return "", err
  }
  if err := os.Chmod(targetPath, 0o755); err != nil {
    return "", fmt.Errorf("chmod: %w", err)
  }

  return targetPath, nil
}

func copyFile(sourcePath, targetPath string) (err error) {
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
