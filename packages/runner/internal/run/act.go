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

  if err := actimage.Ensure(); err != nil {
    return nil, fmt.Errorf("act image: %w", err)
  }

  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return nil, fmt.Errorf("resolve repo path: %w", err)
  }

  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    return nil, err
  }

  secretFiles, err := resolveSecretFiles(repoPath, options.SecretFile)
  if err != nil {
    return nil, fmt.Errorf("resolve secret files: %w", err)
  }

  dockerSock := dockerSocketPath()

  args := []string{
    "run",
    "--rm",
    "-t",
    "-v", fmt.Sprintf("%s:/workspace", repoPath),
    "-v", fmt.Sprintf("%s:/var/run/docker.sock", dockerSock),
    "-w", "/workspace",
  }

  actSecretArgs := make([]string, 0, len(secretFiles)*2+len(options.Secrets)*2)
  for index, secretFile := range secretFiles {
    workspacePath, inRepo, err := workspaceSecretPath(repoPath, secretFile)
    if err != nil {
      return nil, fmt.Errorf("secret file path: %w", err)
    }
    if inRepo {
      actSecretArgs = append(actSecretArgs, "--secret-file", "/workspace/"+workspacePath)
      continue
    }
    mountPath := fmt.Sprintf("/tmp/pipebox-secrets-%d", index)
    args = append(args, "-v", fmt.Sprintf("%s:%s:ro", secretFile, mountPath))
    actSecretArgs = append(actSecretArgs, "--secret-file", mountPath)
  }

  // -P skips act's interactive first-run image survey (Medium defaults).
  // --use-new-action-cache: avoid shared worktree races between parallel jobs (nektos/act#6180).
  args = append(args, actimage.Image,
    "-P", "ubuntu-latest=catthehacker/ubuntu:act-latest",
    "-P", "ubuntu-24.04=catthehacker/ubuntu:act-24.04",
    "-P", "ubuntu-22.04=catthehacker/ubuntu:act-22.04",
    "-P", "ubuntu-20.04=catthehacker/ubuntu:act-20.04",
    "-P", "ubuntu-18.04=catthehacker/ubuntu:act-18.04",
    "--use-new-action-cache",
    "-W", workflowPath,
  )
  if options.Job != "" {
    args = append(args, "-j", options.Job)
  }
  args = append(args, actSecretArgs...)
  for _, secret := range options.Secrets {
    args = append(args, "--secret", secret)
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
