package jenkins

import (
  "context"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/envfile"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins/jfrimage"
)

const (
  workspaceMount          = "/workspace"
  runHomeMount            = "/jenkinsHome"
  materializedFileMount   = "/tmp/pipebox-Jenkinsfile"
)

func buildJFRCommand(ctx context.Context, options Options) (*exec.Cmd, func(), error) {
  if _, err := exec.LookPath("docker"); err != nil {
    return nil, nil, fmt.Errorf("docker not found on PATH")
  }
  if err := jfrimage.Ensure(); err != nil {
    return nil, nil, fmt.Errorf("jenkins image: %w", err)
  }

  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return nil, nil, fmt.Errorf("resolve repo path: %w", err)
  }

  args, cleanup, err := buildJFRArgs(repoPath, options)
  if err != nil {
    if cleanup != nil {
      cleanup()
    }
    return nil, nil, err
  }

  cmd := exec.CommandContext(ctx, "docker", args...)
  cmd.Env = os.Environ()
  return cmd, cleanup, nil
}

func buildJFRArgs(repoPath string, options Options) ([]string, func(), error) {
  cleanups := make([]func(), 0)
  cleanup := func() {
    for index := len(cleanups) - 1; index >= 0; index-- {
      cleanups[index]()
    }
  }

  runHome := options.RunHome
  if runHome == "" {
    dir, err := os.MkdirTemp("", "pipebox-jenkins-home-*")
    if err != nil {
      return nil, nil, fmt.Errorf("create run home: %w", err)
    }
    runHome = dir
    cleanups = append(cleanups, func() { _ = os.RemoveAll(dir) })
  }
  options.RunHome = runHome

  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    cleanup()
    return nil, nil, err
  }
  absWorkflow := filepath.Join(repoPath, filepath.FromSlash(workflowPath))
  source, err := os.ReadFile(absWorkflow)
  if err != nil {
    cleanup()
    return nil, nil, fmt.Errorf("read Jenkinsfile: %w", err)
  }

  jenkinsfileContainerPath := workspaceMount + "/" + filepath.ToSlash(workflowPath)
  if options.Job != "" {
    filtered, err := materializeJobFilter(string(source), options.Job)
    if err != nil {
      cleanup()
      return nil, nil, err
    }
    materialized, err := writeMaterializedJenkinsfile(filtered)
    if err != nil {
      cleanup()
      return nil, nil, err
    }
    cleanups = append(cleanups, materialized.Cleanup)
    jenkinsfileContainerPath = materializedFileMount
    options.materializedHostPath = materialized.Path
  }

  secretFiles, err := envfile.Resolve(repoPath, options.SecretFile)
  if err != nil {
    cleanup()
    return nil, nil, fmt.Errorf("resolve secret files: %w", err)
  }
  secretEnv, err := envfile.Load(secretFiles, options.Secrets)
  if err != nil {
    cleanup()
    return nil, nil, err
  }

  args := []string{
    "run",
    "--rm",
    "-t",
    "-v", fmt.Sprintf("%s:%s", repoPath, workspaceMount),
    "-v", fmt.Sprintf("%s:%s", runHome, runHomeMount),
    "-v", fmt.Sprintf("%s:/var/run/docker.sock", dockerimage.SocketPath()),
    "-w", workspaceMount,
  }
  if options.materializedHostPath != "" {
    args = append(args, "-v", fmt.Sprintf("%s:%s:ro", options.materializedHostPath, materializedFileMount))
  }
  for _, pair := range secretEnv {
    args = append(args, "-e", pair)
  }

  args = append(args, jfrimage.Image,
    "--runHome", runHomeMount,
    "--runWorkspace", workspaceMount,
    "-f", jenkinsfileContainerPath,
  )
  for _, input := range options.Inputs {
    args = append(args, "-a", input)
  }

  return args, cleanup, nil
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
