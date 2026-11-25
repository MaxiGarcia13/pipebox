package github

import (
  "fmt"
  "os"
  "path/filepath"
  "strconv"
  "time"
)

const (
  cacheMountPath = "/tmp/pipebox-act-cache"
  cacheBindAddr  = "0.0.0.0"
)

type cacheRuntime struct {
  Dir  string
  Port int
}

func prepareCacheRuntime() (cacheRuntime, error) {
  dir := filepath.Join(os.TempDir(), "pipebox-act-cache", strconv.FormatInt(time.Now().UnixNano(), 10))
  if err := os.MkdirAll(dir, 0o700); err != nil {
    return cacheRuntime{}, fmt.Errorf("create cache dir: %w", err)
  }
  port, err := allocatePort()
  if err != nil {
    return cacheRuntime{}, err
  }
  return cacheRuntime{Dir: dir, Port: port}, nil
}
