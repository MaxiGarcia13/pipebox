package dockerimage

import (
  "os"
  "path/filepath"
  "strings"
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

func TestBuildCommandArgsIncludesBuildArgs(t *testing.T) {
  args := buildCommandArgs(Config{
    Image:     "pipebox-jenkins:latest",
    BuildArgs: []string{"PIPEBOX_JENKINS_REVISION=ephemeral-3"},
  }, "/docker/jenkins")
  joined := strings.Join(args, " ")
  if joined != "build -t pipebox-jenkins:latest --build-arg PIPEBOX_JENKINS_REVISION=ephemeral-3 /docker/jenkins" {
    t.Fatalf("got %s", joined)
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

func TestResolveDockerfileDirEmbeddedAct(t *testing.T) {
  cacheDir := t.TempDir()
  t.Setenv("PIPEBOX_DOCKER_CACHE_DIR", cacheDir)
  t.Setenv("PIPEBOX_ACT_DOCKERFILE_DIR", "")
  chdirTemp(t)

  got, err := ResolveDockerfileDir(Config{
    Image:        "pipebox-act:latest",
    EnvVar:       "PIPEBOX_ACT_DOCKERFILE_DIR",
    DockerSubdir: "act",
    Label:        "act",
  })
  if err != nil {
    t.Fatalf("ResolveDockerfileDir: %v", err)
  }
  want := filepath.Join(cacheDir, "act")
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
  if !hasDockerfile(got) {
    t.Fatalf("missing Dockerfile in %s", got)
  }
  if _, err := os.Stat(filepath.Join(got, "actrc")); err != nil {
    t.Fatalf("missing actrc: %v", err)
  }
}

func TestResolveDockerfileDirEmbeddedJenkins(t *testing.T) {
  cacheDir := t.TempDir()
  t.Setenv("PIPEBOX_DOCKER_CACHE_DIR", cacheDir)
  t.Setenv("PIPEBOX_JENKINS_DOCKERFILE_DIR", "")
  chdirTemp(t)

  got, err := ResolveDockerfileDir(Config{
    Image:        "pipebox-jenkins:latest",
    EnvVar:       "PIPEBOX_JENKINS_DOCKERFILE_DIR",
    DockerSubdir: "jenkins",
    Label:        "jenkins",
  })
  if err != nil {
    t.Fatalf("ResolveDockerfileDir: %v", err)
  }
  want := filepath.Join(cacheDir, "jenkins")
  if got != want {
    t.Fatalf("got %q want %q", got, want)
  }
  for _, name := range []string{"Dockerfile", "casc.yaml", "entrypoint.sh", "plugins.txt"} {
    if _, err := os.Stat(filepath.Join(got, name)); err != nil {
      t.Fatalf("missing %s: %v", name, err)
    }
  }
}

func chdirTemp(t *testing.T) {
  t.Helper()
  previous, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  dir := t.TempDir()
  if err := os.Chdir(dir); err != nil {
    t.Fatal(err)
  }
  t.Cleanup(func() {
    _ = os.Chdir(previous)
  })
}
