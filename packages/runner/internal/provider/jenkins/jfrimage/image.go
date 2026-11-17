package jfrimage

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
)

const Image = "pipebox-jenkins:latest"

func Exists() bool {
  cmd := exec.Command("docker", "image", "inspect", Image)
  return cmd.Run() == nil
}

func Ensure() error {
  if Exists() {
    return nil
  }
  return Build()
}

func Build() error {
  dockerfileDir, err := resolveDockerfileDir()
  if err != nil {
    return err
  }

  cmd := exec.Command(
    "docker",
    "build",
    "-t", Image,
    dockerfileDir,
  )
  cmd.Stdout = os.Stderr
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    return fmt.Errorf("build %s: %w", Image, err)
  }
  return nil
}

func resolveDockerfileDir() (string, error) {
  if envPath := os.Getenv("PIPEBOX_JENKINS_DOCKERFILE_DIR"); envPath != "" {
    if hasDockerfile(envPath) {
      return envPath, nil
    }
    return "", fmt.Errorf("PIPEBOX_JENKINS_DOCKERFILE_DIR has no Dockerfile: %s", envPath)
  }

  executablePath, err := os.Executable()
  if err == nil {
    candidate := filepath.Join(filepath.Dir(executablePath), "docker", "jenkins")
    if hasDockerfile(candidate) {
      return candidate, nil
    }
  }

  workingDir, err := os.Getwd()
  if err == nil {
    for _, relativePath := range []string{
      filepath.Join("docker", "jenkins"),
      filepath.Join("packages", "runner", "docker", "jenkins"),
    } {
      candidate := filepath.Join(workingDir, relativePath)
      if hasDockerfile(candidate) {
        return candidate, nil
      }
    }
  }

  return "", fmt.Errorf("jenkins Dockerfile not found (expected docker/jenkins next to pipebox)")
}

func hasDockerfile(dirPath string) bool {
  info, err := os.Stat(filepath.Join(dirPath, "Dockerfile"))
  return err == nil && !info.IsDir()
}
