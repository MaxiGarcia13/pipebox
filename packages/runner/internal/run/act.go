package run

import (
  "context"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strconv"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/actimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/workflows"
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

  args, err := buildActArgs(repoPath, options)
  if err != nil {
    return nil, err
  }

  cmd := exec.CommandContext(ctx, "docker", args...)
  cmd.Env = os.Environ()
  return cmd, nil
}

func buildActArgs(repoPath string, options Options) ([]string, error) {
  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    return nil, err
  }

  secretFiles, err := resolveSecretFiles(repoPath, options.SecretFile)
  if err != nil {
    return nil, fmt.Errorf("resolve secret files: %w", err)
  }

  actEvent, err := resolveWorkflowActEvent(repoPath, options.Workflow)
  if err != nil {
    return nil, err
  }

  dockerSock := dockerSocketPath()
  artifactDir, err := resolveArtifactDir(options.ArtifactDir)
  if err != nil {
    return nil, err
  }
  if options.ArtifactPort <= 0 {
    return nil, fmt.Errorf("artifact port is required")
  }
  artifactPort := strconv.Itoa(options.ArtifactPort)
  // Job containers are siblings of the act container (docker.sock). Publish the
  // artifact server and point ACTIONS_RUNTIME_URL at host.docker.internal so
  // upload-artifact can reach it; 127.0.0.1 only works when act runs on the host.
  runtimeURL := fmt.Sprintf("http://host.docker.internal:%s/", artifactPort)

  args := []string{
    "run",
    "--rm",
    "-t",
    "-p", fmt.Sprintf("%s:%s", artifactPort, artifactPort),
    "--add-host", "host.docker.internal:host-gateway",
    "-e", "ACTIONS_RUNTIME_URL="+runtimeURL,
    "-v", fmt.Sprintf("%s:/workspace", repoPath),
    "-v", fmt.Sprintf("%s:/var/run/docker.sock", dockerSock),
    "-v", fmt.Sprintf("%s:%s", artifactDir, artifactMountPath),
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

  args = append(args, actimage.Image, actEvent)
  // -P skips act's interactive first-run image survey (Medium defaults).
  // --use-new-action-cache: avoid shared worktree races between parallel jobs (nektos/act#6180).
  args = append(args,
    "-P", "ubuntu-latest=catthehacker/ubuntu:act-latest",
    "-P", "ubuntu-24.04=catthehacker/ubuntu:act-24.04",
    "-P", "ubuntu-22.04=catthehacker/ubuntu:act-22.04",
    "-P", "ubuntu-20.04=catthehacker/ubuntu:act-20.04",
    "-P", "ubuntu-18.04=catthehacker/ubuntu:act-18.04",
    "--use-new-action-cache",
    "--artifact-server-path", artifactMountPath,
    "--artifact-server-addr", artifactBindAddr,
    "--artifact-server-port", artifactPort,
    "--container-options", "--add-host=host.docker.internal:host-gateway",
    "-W", workflowPath,
  )
  if options.Job != "" {
    args = append(args, "-j", options.Job)
  }
  args = append(args, actSecretArgs...)
  for _, secret := range options.Secrets {
    args = append(args, "--secret", secret)
  }
  for _, input := range options.Inputs {
    args = append(args, "--input", input)
  }

  return args, nil
}

func resolveWorkflowActEvent(repoPath, workflow string) (string, error) {
  relativePath := workflow
  if filepath.IsAbs(workflow) {
    var err error
    relativePath, err = filepath.Rel(repoPath, workflow)
    if err != nil {
      return "", fmt.Errorf("workflow must be inside the repo: %w", err)
    }
  }
  meta, err := workflows.LoadDispatch(filepath.Join(repoPath, filepath.FromSlash(relativePath)))
  if err != nil {
    return "", fmt.Errorf("resolve workflow event: %w", err)
  }
  return workflows.ResolveActEvent(meta.Events), nil
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

func resolveArtifactDir(artifactDir string) (string, error) {
  if artifactDir == "" {
    return "", fmt.Errorf("artifact dir is required")
  }
  absolutePath, err := filepath.Abs(artifactDir)
  if err != nil {
    return "", fmt.Errorf("resolve artifact dir: %w", err)
  }
  if err := os.MkdirAll(absolutePath, 0o700); err != nil {
    return "", fmt.Errorf("create artifact dir: %w", err)
  }
  return absolutePath, nil
}
