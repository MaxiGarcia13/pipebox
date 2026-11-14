package install

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

type UpdateOptions struct {
  CurrentVersion string
  TargetVersion  string
  CheckOnly      bool
}

type UpdateResult struct {
  CurrentVersion string
  TargetVersion  string
  UpToDate       bool
  CLIPath        string
  DesktopPath    string
}

func Update(options UpdateOptions) (UpdateResult, error) {
  current := normalizeVersion(options.CurrentVersion)
  target := normalizeVersion(options.TargetVersion)
  if target == "" {
    latest, err := LatestVersion()
    if err != nil {
      return UpdateResult{}, err
    }
    target = latest
  }

  result := UpdateResult{
    CurrentVersion: current,
    TargetVersion:  target,
    UpToDate:       current != "" && current == target,
  }
  if options.CheckOnly || result.UpToDate {
    return result, nil
  }

  cliPath, err := installCLIFromRelease(target)
  if err != nil {
    return result, err
  }
  result.CLIPath = cliPath

  desktopPath, err := InstallDesktop(target)
  if err != nil {
    return result, fmt.Errorf("cli updated to %s, but desktop failed: %w", cliPath, err)
  }
  result.DesktopPath = desktopPath
  return result, nil
}

func installCLIFromRelease(version string) (string, error) {
  asset, err := cliAssetName(version)
  if err != nil {
    return "", err
  }

  targets, err := cliUpdateTargets()
  if err != nil {
    return "", err
  }

  tmpDir, err := os.MkdirTemp("", "pipebox-cli-*")
  if err != nil {
    return "", fmt.Errorf("temp dir: %w", err)
  }
  defer func() { _ = os.RemoveAll(tmpDir) }()

  tmpPath := filepath.Join(tmpDir, asset)
  if err := downloadFile(releaseAssetURL(version, asset), tmpPath); err != nil {
    return "", err
  }

  updated := make([]string, 0, len(targets))
  for _, path := range targets {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
      return "", fmt.Errorf("create bin dir: %w", err)
    }
    if err := copyFile(tmpPath, path); err != nil {
      return "", err
    }
    if err := os.Chmod(path, 0o755); err != nil {
      return "", fmt.Errorf("chmod: %w", err)
    }
    updated = append(updated, path)
  }
  return strings.Join(updated, ", "), nil
}
