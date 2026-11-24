package install

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostos"
)

type Result struct {
  CLIPath     string
  DesktopPath string
}

func Run(version string, progress Progress) (Result, error) {
  progress = resolveProgress(progress)

  progress.Step("Installing CLI…")
  cliPath, err := installCLI()
  if err != nil {
    return Result{}, err
  }

  progress.Step("Installing shell completions…")
  if _, err := InstallCompletions(); err != nil {
    return Result{CLIPath: cliPath}, fmt.Errorf("cli installed to %s, but completions failed: %w", cliPath, err)
  }

  desktopPath, err := InstallDesktop(version, progress)
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

  if err := hostos.Current().ReplaceFile(executablePath, path); err != nil {
    return "", err
  }
  return path, nil
}

func Uninstall() (cliPaths []string, desktopPath string, err error) {
  cliPaths, err = existingManagedCLIPaths()
  if err != nil {
    return nil, "", err
  }

  plat := hostos.Current()
  for _, path := range cliPaths {
    if removeErr := plat.RemoveFile(path); removeErr != nil {
      return cliPaths, "", fmt.Errorf("remove cli %s: %w", path, removeErr)
    }
  }

  if _, err := UninstallCompletions(); err != nil {
    return cliPaths, "", err
  }

  desktopPath, err = UninstallDesktop()
  if err != nil {
    return cliPaths, desktopPath, err
  }
  return cliPaths, desktopPath, nil
}
