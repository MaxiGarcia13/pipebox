package hostcache

import (
  "fmt"
  "os"
  "path/filepath"
)

func Root() (string, error) {
  base, err := os.UserCacheDir()
  if err != nil {
    return "", fmt.Errorf("cache dir: %w", err)
  }
  return filepath.Join(base, "pipebox"), nil
}

func Path(elem ...string) (string, error) {
  root, err := Root()
  if err != nil {
    return "", err
  }
  parts := make([]string, 0, 1+len(elem))
  parts = append(parts, root)
  parts = append(parts, elem...)
  return filepath.Join(parts...), nil
}
