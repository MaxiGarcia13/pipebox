package envfile

import (
  "bufio"
  "fmt"
  "os"
  "path/filepath"
  "sort"
  "strings"
)

func Resolve(repoPath, explicit string) ([]string, error) {
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
    if !IsName(name) {
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

func IsName(name string) bool {
  return name == ".env" || strings.HasPrefix(name, ".env.")
}

func WorkspacePath(repoPath, secretFile string) (string, bool, error) {
  relativePath, err := filepath.Rel(repoPath, secretFile)
  if err != nil {
    return "", false, err
  }
  if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
    return "", false, nil
  }
  return filepath.ToSlash(relativePath), true, nil
}

func SourceNames(repoPath string, secretFiles []string) []string {
  names := make([]string, 0, len(secretFiles))
  for _, secretFile := range secretFiles {
    if relativePath, inRepo, err := WorkspacePath(repoPath, secretFile); err == nil && inRepo {
      names = append(names, relativePath)
      continue
    }
    names = append(names, filepath.Base(secretFile))
  }
  return names
}

func Load(secretFiles []string, explicit []string) ([]string, error) {
  pairs := make([]string, 0)
  seen := map[string]struct{}{}

  for _, secretFile := range secretFiles {
    filePairs, err := Parse(secretFile)
    if err != nil {
      return nil, err
    }
    for _, pair := range filePairs {
      key := pairKey(pair)
      if key == "" {
        continue
      }
      if _, exists := seen[key]; exists {
        continue
      }
      seen[key] = struct{}{}
      pairs = append(pairs, pair)
    }
  }

  for _, secret := range explicit {
    key, value, ok := strings.Cut(secret, "=")
    if !ok || strings.TrimSpace(key) == "" {
      return nil, fmt.Errorf("invalid secret %q (want KEY=VALUE)", secret)
    }
    key = strings.TrimSpace(key)
    pairs = append(pairs, key+"="+value)
    seen[key] = struct{}{}
  }

  return pairs, nil
}

func Parse(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer func() { _ = file.Close() }()

  pairs := make([]string, 0)
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())
    if line == "" || strings.HasPrefix(line, "#") {
      continue
    }
    line = strings.TrimPrefix(line, "export ")
    key, value, ok := strings.Cut(line, "=")
    if !ok {
      continue
    }
    key = strings.TrimSpace(key)
    if key == "" {
      continue
    }
    value = strings.TrimSpace(value)
    value = strings.Trim(value, `"'`)
    pairs = append(pairs, key+"="+value)
  }
  if err := scanner.Err(); err != nil {
    return nil, err
  }
  return pairs, nil
}

func pairKey(pair string) string {
  key, _, ok := strings.Cut(pair, "=")
  if !ok {
    return ""
  }
  return key
}
