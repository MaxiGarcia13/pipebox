package install

import (
  "fmt"
  "os"
  "path/filepath"
  "runtime"
  "strings"
)

func cliBinaryName() string {
  if runtime.GOOS == "windows" {
    return "pipebox.exe"
  }
  return "pipebox"
}

func cliTargetPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", cliBinaryName()), nil
}

func cliGoBinPath() (string, error) {
  if gobin := strings.TrimSpace(os.Getenv("GOBIN")); gobin != "" {
    return filepath.Join(gobin, cliBinaryName()), nil
  }

  if gopath := strings.TrimSpace(os.Getenv("GOPATH")); gopath != "" {
    first := strings.Split(gopath, string(os.PathListSeparator))[0]
    first = strings.TrimSpace(first)
    if first != "" {
      return filepath.Join(first, "bin", cliBinaryName()), nil
    }
  }

  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, "go", "bin", cliBinaryName()), nil
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
  seen := make(map[string]struct{}, len(paths))
  out := make([]string, 0, len(paths))
  for _, path := range paths {
    if path == "" {
      continue
    }
    clean := filepath.Clean(path)
    key := clean
    if runtime.GOOS == "windows" {
      key = strings.ToLower(clean)
    }
    if _, ok := seen[key]; ok {
      continue
    }
    seen[key] = struct{}{}
    out = append(out, clean)
  }
  return out
}
