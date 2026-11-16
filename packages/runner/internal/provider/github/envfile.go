package github

import (
  "os"
  "path/filepath"
  "sort"
  "strings"
)

func resolveSecretFiles(repoPath, explicit string) ([]string, error) {
  if explicit != "" {
    absolutePath, err := filepath.Abs(explicit)
    if err != nil {
      return nil, err
    }
    if _, err := os.Stat(absolutePath); err != nil {
      return nil, err
    }
    return []string{absolutePath}, nil
  }

  entries, err := os.ReadDir(repoPath)
  if err != nil {
    return nil, err
  }

  names := make([]string, 0)
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    name := entry.Name()
    if !isEnvFileName(name) {
      continue
    }
    names = append(names, name)
  }

  sort.SliceStable(names, func(left, right int) bool {
    if names[left] == ".env" {
      return true
    }
    if names[right] == ".env" {
      return false
    }
    return names[left] < names[right]
  })

  files := make([]string, 0, len(names))
  for _, name := range names {
    files = append(files, filepath.Join(repoPath, name))
  }
  return files, nil
}

func isEnvFileName(name string) bool {
  return name == ".env" || strings.HasPrefix(name, ".env.")
}

func workspaceSecretPath(repoPath, secretFile string) (string, bool, error) {
  relativePath, err := filepath.Rel(repoPath, secretFile)
  if err != nil {
    return "", false, err
  }
  if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
    return "", false, nil
  }
  return filepath.ToSlash(relativePath), true, nil
}

func secretSourceNames(repoPath string, secretFiles []string) []string {
  names := make([]string, 0, len(secretFiles))
  for _, secretFile := range secretFiles {
    if relativePath, inRepo, err := workspaceSecretPath(repoPath, secretFile); err == nil && inRepo {
      names = append(names, relativePath)
      continue
    }
    names = append(names, filepath.Base(secretFile))
  }
  return names
}
