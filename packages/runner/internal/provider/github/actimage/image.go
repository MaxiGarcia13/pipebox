package actimage

import "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"

const Image = "pipebox-act:latest"

var config = dockerimage.Config{
  Image:        Image,
  EnvVar:       "PIPEBOX_ACT_DOCKERFILE_DIR",
  DockerSubdir: "act",
  Label:        "act",
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
