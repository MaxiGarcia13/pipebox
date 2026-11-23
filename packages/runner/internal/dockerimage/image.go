package dockerimage

import (
  "bufio"
  "fmt"
  "io"
  "io/fs"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
  "sync"

  dockerassets "github.com/MaxiGarcia13/pipebox/packages/runner/docker"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type Config struct {
  Image        string
  EnvVar       string
  DockerSubdir string
  Label        string
  BuildArgs    []string
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

  cmd := exec.Command("docker", buildCommandArgs(config, dockerfileDir)...)
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return fmt.Errorf("stdout pipe: %w", err)
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return fmt.Errorf("stderr pipe: %w", err)
  }

  if err := cmd.Start(); err != nil {
    return fmt.Errorf("start docker build: %w", err)
  }

  var waitGroup sync.WaitGroup
  waitGroup.Add(2)
  go streamBuildLines(&waitGroup, stdout)
  go streamBuildLines(&waitGroup, stderr)

  waitErr := cmd.Wait()
  waitGroup.Wait()
  if waitErr != nil {
    return fmt.Errorf("build %s: %w", config.Image, waitErr)
  }
  return nil
}

func buildCommandArgs(config Config, dockerfileDir string) []string {
  args := []string{"build", "-t", config.Image}
  for _, buildArg := range config.BuildArgs {
    args = append(args, "--build-arg", buildArg)
  }
  args = append(args, dockerfileDir)
  return args
}

func streamBuildLines(waitGroup *sync.WaitGroup, reader io.Reader) {
  defer waitGroup.Done()
  scanner := bufio.NewScanner(reader)
  scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for scanner.Scan() {
    line := strings.TrimRight(scanner.Text(), "\r")
    if strings.TrimSpace(line) == "" {
      continue
    }
    _ = event.EmitStdout(event.Log{
      Type: event.TypeLog,
      Line: line,
    })
  }
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

  if hasEmbedded(config.DockerSubdir) {
    return materializeEmbedded(config.DockerSubdir)
  }

  label := config.Label
  if label == "" {
    label = config.DockerSubdir
  }
  return "", fmt.Errorf("%s Dockerfile not found (expected docker/%s next to pipebox)", label, config.DockerSubdir)
}

func hasEmbedded(subdir string) bool {
  if subdir == "" {
    return false
  }
  info, err := fs.Stat(dockerassets.FS, subdir)
  return err == nil && info.IsDir()
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
