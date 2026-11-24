package jfrimage

import (
  "os/exec"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
)

const (
  Image = "pipebox-jenkins:latest"
  // Bump when Dockerfile / Casc / plugins / entrypoint change so Ensure rebuilds.
  Revision = "ephemeral-5"
)

var config = dockerimage.Config{
  Image:        Image,
  EnvVar:       "PIPEBOX_JENKINS_DOCKERFILE_DIR",
  DockerSubdir: "jenkins",
  Label:        "jenkins",
  BuildArgs:    []string{"PIPEBOX_JENKINS_REVISION=" + Revision},
}

func Exists() bool {
  if !dockerimage.Exists(Image) {
    return false
  }
  output, err := exec.Command(
    "docker", "image", "inspect",
    "-f", `{{index .Config.Labels "pipebox.jenkins.revision"}}`,
    Image,
  ).Output()
  if err != nil {
    return false
  }
  return strings.TrimSpace(string(output)) == Revision
}

func DockerfileDir() (string, error) {
  return dockerimage.ResolveDockerfileDir(config)
}

func Ensure() error {
  if Exists() {
    return nil
  }
  return Build()
}

func Build() error {
  return dockerimage.Build(config)
}
