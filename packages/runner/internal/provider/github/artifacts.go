package github

import (
  "fmt"
  "net"
  "os"
  "path/filepath"
  "strconv"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

const (
  artifactMountPath = "/tmp/pipebox-artifacts"
  artifactBindAddr  = "0.0.0.0"
)

type artifactRuntime struct {
  Dir  string
  Port int
}

func prepareArtifactRuntime() (artifactRuntime, error) {
  dir := filepath.Join(os.TempDir(), "pipebox-artifacts", strconv.FormatInt(time.Now().UnixNano(), 10))
  if err := os.MkdirAll(dir, 0o700); err != nil {
    return artifactRuntime{}, fmt.Errorf("create artifact dir: %w", err)
  }
  port, err := allocatePort()
  if err != nil {
    return artifactRuntime{}, err
  }
  return artifactRuntime{Dir: dir, Port: port}, nil
}

func allocatePort() (int, error) {
  listener, err := net.Listen("tcp", "127.0.0.1:0")
  if err != nil {
    return 0, fmt.Errorf("allocate artifact port: %w", err)
  }
  address, ok := listener.Addr().(*net.TCPAddr)
  if !ok {
    _ = listener.Close()
    return 0, fmt.Errorf("allocate artifact port: unexpected address type")
  }
  port := address.Port
  if err := listener.Close(); err != nil {
    return 0, fmt.Errorf("allocate artifact port: %w", err)
  }
  return port, nil
}

func emitArtifacts(dir string) error {
  artifacts, err := listArtifacts(dir)
  if err != nil {
    return err
  }
  if len(artifacts) == 0 {
    return nil
  }
  return event.EmitStdout(event.Artifacts{
    Type:      event.TypeArtifacts,
    Artifacts: artifacts,
  })
}

func listArtifacts(baseDir string) ([]event.Artifact, error) {
  runEntries, err := os.ReadDir(baseDir)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, err
  }

  artifacts := make([]event.Artifact, 0)
  for _, runEntry := range runEntries {
    if !runEntry.IsDir() {
      continue
    }
    runPath := filepath.Join(baseDir, runEntry.Name())
    nameEntries, err := os.ReadDir(runPath)
    if err != nil {
      return nil, err
    }
    for _, nameEntry := range nameEntries {
      if !nameEntry.IsDir() {
        continue
      }
      name := nameEntry.Name()
      artifactDir := filepath.Join(runPath, name)
      zipPath := filepath.Join(artifactDir, name+".zip")
      if info, err := os.Stat(zipPath); err == nil && !info.IsDir() {
        artifacts = append(artifacts, event.Artifact{
          Name: name,
          Path: zipPath,
          Size: info.Size(),
        })
        continue
      }
      size, err := directorySize(artifactDir)
      if err != nil {
        return nil, err
      }
      artifacts = append(artifacts, event.Artifact{
        Name: name,
        Path: artifactDir,
        Size: size,
      })
    }
  }
  return artifacts, nil
}

func directorySize(root string) (int64, error) {
  var total int64
  err := filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if !info.IsDir() {
      total += info.Size()
    }
    return nil
  })
  return total, err
}
