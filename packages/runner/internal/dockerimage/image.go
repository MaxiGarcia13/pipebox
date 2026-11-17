package dockerimage

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
)

type Config struct {
  Image        string
  EnvVar       string
  DockerSubdir string
  Label        string
}

func Exists(image string) bool {
  cmd := exec.Command("docker", "image", "inspect", image)
  return cmd.Run() == nil
}

func Ensure(config Config) error {
  if Exists(config.Image) {
    return nil
  }
  return Build(config)
}

func Build(config Config) error {
  dockerfileDir, err := ResolveDockerfileDir(config)
  if err != nil {
    return err
  }

  cmd := exec.Command(
    "docker",
    "build",
    "-t", config.Image,
    dockerfileDir,
  )
  cmd.Stdout = os.Stderr
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    return fmt.Errorf("build %s: %w", config.Image, err)
  }
  return nil
}

func ResolveDockerfileDir(config Config) (string, error) {
  if config.EnvVar != "" {
    if envPath := os.Getenv(config.EnvVar); envPath != "" {
      if hasDockerfile(envPath) {
        return envPath, nil
      }
      return "", fmt.Errorf("%s has no Dockerfile: %s", config.EnvVar, envPath)
    }
  }

  executablePath, err := os.Executable()
  if err == nil {
    candidate := filepath.Join(filepath.Dir(executablePath), "docker", config.DockerSubdir)
    if hasDockerfile(candidate) {
      return candidate, nil
    }
  }

  workingDir, err := os.Getwd()
  if err == nil {
    for _, relativePath := range []string{
      filepath.Join("docker", config.DockerSubdir),
      filepath.Join("packages", "runner", "docker", config.DockerSubdir),
    } {
      candidate := filepath.Join(workingDir, relativePath)
      if hasDockerfile(candidate) {
        return candidate, nil
      }
    }
  }

  label := config.Label
  if label == "" {
    label = config.DockerSubdir
  }
  return "", fmt.Errorf("%s Dockerfile not found (expected docker/%s next to pipebox)", label, config.DockerSubdir)
}

func SocketPath() string {
  if runtime.GOOS == "windows" {
    return "//var/run/docker.sock"
  }
  return "/var/run/docker.sock"
}

func hasDockerfile(dirPath string) bool {
  info, err := os.Stat(filepath.Join(dirPath, "Dockerfile"))
  return err == nil && !info.IsDir()
}
