package jfrimage

import (
  "os"
  "path/filepath"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
)

func TestResolveDockerfileDirFromEnv(t *testing.T) {
  dir := t.TempDir()
  if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
    t.Fatal(err)
  }
  t.Setenv("PIPEBOX_JENKINS_DOCKERFILE_DIR", dir)

  got, err := dockerimage.ResolveDockerfileDir(config)
  if err != nil {
    t.Fatalf("ResolveDockerfileDir: %v", err)
  }
  if got != dir {
    t.Fatalf("got %q want %q", got, dir)
  }
}

func TestResolveDockerfileDirEnvMissingDockerfile(t *testing.T) {
  dir := t.TempDir()
  t.Setenv("PIPEBOX_JENKINS_DOCKERFILE_DIR", dir)
  if _, err := dockerimage.ResolveDockerfileDir(config); err == nil {
    t.Fatal("expected error")
  }
}
