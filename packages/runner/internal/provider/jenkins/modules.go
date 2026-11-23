package jenkins

import (
  "os"
  "path/filepath"
)

type moduleMount struct {
  Host string
  Path string
}

func prepareModuleMounts(repoPath string) ([]moduleMount, func(), error) {
  dirs := workspaceModuleDirs(repoPath)
  mounts := make([]moduleMount, 0, len(dirs))
  created := make([]string, 0, len(dirs))
  cleanup := func() {
    for _, dir := range created {
      _ = os.RemoveAll(dir)
    }
  }
  for _, dir := range dirs {
    host, err := os.MkdirTemp("", "pipebox-modules-*")
    if err != nil {
      cleanup()
      return nil, nil, err
    }
    created = append(created, host)
    mounts = append(mounts, moduleMount{Host: host, Path: dir})
  }
  return mounts, cleanup, nil
}

func workspaceModuleDirs(repoPath string) []string {
  dirs := make([]string, 0)
  consider := func(dir string) {
    modules := filepath.Join(dir, "node_modules")
    info, err := os.Stat(modules)
    if err == nil && info.IsDir() {
      dirs = append(dirs, modules)
    }
  }
  consider(repoPath)
  for _, name := range []string{"packages", "apps", "services"} {
    entries, err := os.ReadDir(filepath.Join(repoPath, name))
    if err != nil {
      continue
    }
    for _, entry := range entries {
      if entry.IsDir() {
        consider(filepath.Join(repoPath, name, entry.Name()))
      }
    }
  }
  return dirs
}
