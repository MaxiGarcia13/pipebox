package install

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostos"
)

func cliBinaryName() string {
  return hostos.Current().CLIBinaryName()
}

func cliTargetPath() (string, error) {
  return hostos.Current().CLILocalBinPath()
}

func cliGoBinPath() (string, error) {
  return hostos.Current().CLIGoBinPath()
}

func managedCLIPaths() ([]string, error) {
  localPath, err := cliTargetPath()
  if err != nil {
    return nil, err
  }
  goPath, err := cliGoBinPath()
  if err != nil {
    return nil, err
  }
  return uniquePaths(localPath, goPath), nil
}

func existingManagedCLIPaths() ([]string, error) {
  paths, err := managedCLIPaths()
  if err != nil {
    return nil, err
  }
  existing := make([]string, 0, len(paths))
  for _, path := range paths {
    if _, err := os.Stat(path); err == nil {
      existing = append(existing, path)
      continue
    } else if !os.IsNotExist(err) {
      return nil, fmt.Errorf("stat cli %s: %w", path, err)
    }
  }
  return existing, nil
}

func cliUpdateTargets() ([]string, error) {
  existing, err := existingManagedCLIPaths()
  if err != nil {
    return nil, err
  }
  if len(existing) > 0 {
    return existing, nil
  }

  localPath, err := cliTargetPath()
  if err != nil {
    return nil, err
  }
  return []string{localPath}, nil
}

func uniquePaths(paths ...string) []string {
  plat := hostos.Current()
  seen := make(map[string]struct{}, len(paths))
  out := make([]string, 0, len(paths))
  for _, path := range paths {
    if path == "" {
      continue
    }
    clean := filepath.Clean(path)
    key := plat.PathEqualKey(clean)
    if _, ok := seen[key]; ok {
      continue
    }
    seen[key] = struct{}{}
    out = append(out, clean)
  }
  return out
}
