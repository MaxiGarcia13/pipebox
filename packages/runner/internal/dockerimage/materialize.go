package dockerimage

import (
  "fmt"
  "io/fs"
  "os"
  "path"
  "path/filepath"
  "strings"

  dockerassets "github.com/MaxiGarcia13/pipebox/packages/runner/docker"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostcache"
)

func materializeEmbedded(subdir string) (string, error) {
  cleaned := path.Clean(subdir)
  if cleaned == "." || cleaned == ".." || path.IsAbs(cleaned) || cleaned != subdir {
    return "", fmt.Errorf("invalid docker subdir %q", subdir)
  }

  root, err := cacheRoot()
  if err != nil {
    return "", err
  }
  dest := filepath.Join(root, filepath.FromSlash(cleaned))
  if err := extractEmbedded(cleaned, dest); err != nil {
    return "", err
  }
  if !hasDockerfile(dest) {
    return "", fmt.Errorf("embedded docker/%s has no Dockerfile", cleaned)
  }
  return dest, nil
}

func cacheRoot() (string, error) {
  if override := os.Getenv("PIPEBOX_DOCKER_CACHE_DIR"); override != "" {
    return override, nil
  }
  return hostcache.Path("docker")
}

func extractEmbedded(subdir, dest string) error {
  if err := os.MkdirAll(dest, 0o755); err != nil {
    return fmt.Errorf("create docker cache: %w", err)
  }

  return fs.WalkDir(dockerassets.FS, subdir, func(entryPath string, entry fs.DirEntry, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    relative := strings.TrimPrefix(entryPath, subdir)
    relative = strings.TrimPrefix(relative, "/")
    if relative == "" {
      return nil
    }
    target := filepath.Join(dest, filepath.FromSlash(relative))
    if entry.IsDir() {
      return os.MkdirAll(target, 0o755)
    }
    data, err := dockerassets.FS.ReadFile(entryPath)
    if err != nil {
      return err
    }
    if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
      return err
    }
    return os.WriteFile(target, data, 0o644)
  })
}
