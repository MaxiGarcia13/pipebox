package jenkins

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func emitRunHomeArtifacts(runHome string) error {
  artifacts, err := listRunHomeArtifacts(runHome)
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

func listRunHomeArtifacts(runHome string) ([]event.Artifact, error) {
  jobsDir := filepath.Join(runHome, "jobs")
  jobEntries, err := os.ReadDir(jobsDir)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, err
  }

  artifacts := make([]event.Artifact, 0)
  for _, jobEntry := range jobEntries {
    if !jobEntry.IsDir() {
      continue
    }
    buildsDir := filepath.Join(jobsDir, jobEntry.Name(), "builds")
    buildEntries, err := os.ReadDir(buildsDir)
    if err != nil {
      if os.IsNotExist(err) {
        continue
      }
      return nil, err
    }
    for _, buildEntry := range buildEntries {
      if !buildEntry.IsDir() {
        continue
      }
      archiveDir := filepath.Join(buildsDir, buildEntry.Name(), "archive")
      found, err := collectArchiveFiles(archiveDir)
      if err != nil {
        return nil, err
      }
      artifacts = append(artifacts, found...)
    }
  }
  return artifacts, nil
}

func collectArchiveFiles(archiveDir string) ([]event.Artifact, error) {
  info, err := os.Stat(archiveDir)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, err
  }
  if !info.IsDir() {
    return nil, nil
  }

  artifacts := make([]event.Artifact, 0)
  err = filepath.Walk(archiveDir, func(path string, info os.FileInfo, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    if info.IsDir() {
      return nil
    }
    relative, err := filepath.Rel(archiveDir, path)
    if err != nil {
      return err
    }
    artifacts = append(artifacts, event.Artifact{
      Name: filepath.ToSlash(relative),
      Path: path,
      Size: info.Size(),
    })
    return nil
  })
  if err != nil {
    return nil, fmt.Errorf("walk archive: %w", err)
  }
  return artifacts, nil
}
