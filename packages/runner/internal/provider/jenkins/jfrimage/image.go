package jfrimage

import "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"

const Image = "pipebox-jenkins:latest"

var config = dockerimage.Config{
  Image:        Image,
  EnvVar:       "PIPEBOX_JENKINS_DOCKERFILE_DIR",
  DockerSubdir: "jenkins",
  Label:        "jenkins",
}

func Exists() bool {
  return dockerimage.Exists(Image)
}

func Ensure() error {
  return dockerimage.Ensure(config)
}

func Build() error {
  return dockerimage.Build(config)
}
