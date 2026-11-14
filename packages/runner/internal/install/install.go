package install

import (
  "fmt"
  "io"
  "os"
  "path/filepath"
)

type Result struct {
  CLIPath     string
  DesktopPath string
}

func Run(version string) (Result, error) {
  cliPath, err := installCLI()
  if err != nil {
    return Result{}, err
  }

  desktopPath, err := InstallDesktop(version)
  if err != nil {
    return Result{CLIPath: cliPath}, fmt.Errorf("cli installed to %s, but desktop failed: %w", cliPath, err)
  }

  return Result{CLIPath: cliPath, DesktopPath: desktopPath}, nil
}

func installCLI() (string, error) {
  executablePath, err := os.Executable()
  if err != nil {
    return "", fmt.Errorf("resolve executable: %w", err)
  }

  path, err := cliTargetPath()
  if err != nil {
    return "", err
  }

  if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
    return "", fmt.Errorf("create bin dir: %w", err)
  }

  if err := copyFile(executablePath, path); err != nil {
    return "", err
  }
  if err := os.Chmod(path, 0o755); err != nil {
    return "", fmt.Errorf("chmod: %w", err)
  }

  return path, nil
}

func Uninstall() (cliPath string, desktopPath string, err error) {
  cliPath, err = cliTargetPath()
  if err != nil {
    return "", "", err
  }

  if _, statErr := os.Stat(cliPath); statErr == nil {
    if removeErr := os.Remove(cliPath); removeErr != nil {
      return cliPath, "", fmt.Errorf("remove cli: %w", removeErr)
    }
  } else if !os.IsNotExist(statErr) {
    return cliPath, "", fmt.Errorf("stat cli: %w", statErr)
  } else {
    cliPath = ""
  }

  desktopPath, err = UninstallDesktop()
  if err != nil {
    return cliPath, desktopPath, err
  }
  return cliPath, desktopPath, nil
}

func cliTargetPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", "pipebox"), nil
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
