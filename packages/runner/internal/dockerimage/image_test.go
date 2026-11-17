package dockerimage

import (
  "os"
  "path/filepath"
  "testing"
)

func TestResolveDockerfileDirFromEnv(t *testing.T) {
  dir := t.TempDir()
  if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
    t.Fatal(err)
  }
  t.Setenv("PIPEBOX_TEST_DOCKERFILE_DIR", dir)

  got, err := ResolveDockerfileDir(Config{
    Image:        "pipebox-test:latest",
    EnvVar:       "PIPEBOX_TEST_DOCKERFILE_DIR",
    DockerSubdir: "test",
    Label:        "test",
  })
  if err != nil {
    t.Fatalf("ResolveDockerfileDir: %v", err)
  }
  if got != dir {
    t.Fatalf("got %q want %q", got, dir)
  }
}

func TestResolveDockerfileDirEnvMissingDockerfile(t *testing.T) {
  dir := t.TempDir()
  t.Setenv("PIPEBOX_TEST_DOCKERFILE_DIR", dir)
  if _, err := ResolveDockerfileDir(Config{
    Image:        "pipebox-test:latest",
    EnvVar:       "PIPEBOX_TEST_DOCKERFILE_DIR",
    DockerSubdir: "test",
    Label:        "test",
  }); err == nil {
    t.Fatal("expected error")
  }
}
