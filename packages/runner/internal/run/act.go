package run

import (
  "context"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"

  "pipebox/internal/actimage"
)

func buildActCommand(ctx context.Context, options Options) (*exec.Cmd, error) {
  if _, err := exec.LookPath("docker"); err != nil {
    return nil, fmt.Errorf("docker not found on PATH")
  }

  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return nil, fmt.Errorf("resolve repo path: %w", err)
  }

  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    return nil, err
  }

  dockerSock := dockerSocketPath()
  args := []string{
    "run",
    "--rm",
    "-v", fmt.Sprintf("%s:/workspace", repoPath),
    "-v", fmt.Sprintf("%s:/var/run/docker.sock", dockerSock),
    "-w", "/workspace",
    actimage.Image,
    "-W", workflowPath,
  }
  if options.Job != "" {
    args = append(args, "-j", options.Job)
  }

  cmd := exec.CommandContext(ctx, "docker", args...)
  cmd.Env = os.Environ()
  return cmd, nil
}

func workspaceWorkflowPath(repoPath, workflow string) (string, error) {
  if filepath.IsAbs(workflow) {
    relativePath, err := filepath.Rel(repoPath, workflow)
    if err != nil {
      return "", fmt.Errorf("workflow must be inside the repo: %w", err)
    }
    return filepath.ToSlash(relativePath), nil
  }
  return filepath.ToSlash(workflow), nil
}

func dockerSocketPath() string {
  if runtime.GOOS == "windows" {
    return "//var/run/docker.sock"
  }
  return "/var/run/docker.sock"
}
