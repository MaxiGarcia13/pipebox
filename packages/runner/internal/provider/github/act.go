package github

import (
  "context"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "strconv"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/envfile"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actimage"
)

const materializedWorkflowMount = "/tmp/pipebox-workflow.yml"

func buildActCommand(ctx context.Context, options Options) (*exec.Cmd, func(), error) {
  if _, err := exec.LookPath("docker"); err != nil {
    return nil, nil, fmt.Errorf("docker not found on PATH")
  }

  if err := actimage.Ensure(); err != nil {
    return nil, nil, fmt.Errorf("act image: %w", err)
  }

  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return nil, nil, fmt.Errorf("resolve repo path: %w", err)
  }

  args, cleanup, err := buildActArgs(repoPath, options)
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

func buildActArgs(repoPath string, options Options) ([]string, func(), error) {
  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    return nil, nil, err
  }

  absWorkflow := workflowPath
  if !filepath.IsAbs(absWorkflow) {
    absWorkflow = filepath.Join(repoPath, filepath.FromSlash(workflowPath))
  }

  materialized, err := MaterializeLocalReusables(repoPath, absWorkflow)
  if err != nil {
    return nil, nil, fmt.Errorf("materialize reusable workflows: %w", err)
  }
  cleanup := materialized.Cleanup
  if cleanup == nil {
    cleanup = func() {}
  }

  jobID := strings.TrimSpace(options.Job)
  // act -j still runs needs; isolate the target job so sibling deps are not re-executed.
  if jobID != "" {
    isolated, isolateErr := IsolateJob(materialized.Path, jobID)
    if isolateErr != nil {
      cleanup()
      return nil, nil, fmt.Errorf("isolate job %q: %w", jobID, isolateErr)
    }
    previousCleanup := cleanup
    cleanup = func() {
      isolated.Cleanup()
      previousCleanup()
    }
    materialized = isolated.materializedWorkflow
    jobID = isolated.JobID
  }

  secretFiles, err := envfile.Resolve(repoPath, options.SecretFile)
  if err != nil {
    cleanup()
    return nil, nil, fmt.Errorf("resolve secret files: %w", err)
  }

  actEvent, err := resolveWorkflowActEvent(repoPath, options.Workflow)
  if err != nil {
    cleanup()
    return nil, nil, err
  }

  dockerSock := dockerimage.SocketPath()
  artifactDir, err := resolveArtifactDir(options.ArtifactDir)
  if err != nil {
    cleanup()
    return nil, nil, err
  }
  if options.ArtifactPort <= 0 {
    cleanup()
    return nil, nil, fmt.Errorf("artifact port is required")
  }
  artifactPort := strconv.Itoa(options.ArtifactPort)
  // Job containers are siblings of the act container (docker.sock). Publish the
  // artifact server and point ACTIONS_RUNTIME_URL at host.docker.internal so
  // upload-artifact can reach it; 127.0.0.1 only works when act runs on the host.
  runtimeURL := fmt.Sprintf("http://host.docker.internal:%s/", artifactPort)

  cacheDir, err := resolveCacheMountDir(options.CacheDir)
  if err != nil {
    cleanup()
    return nil, nil, err
  }
  if options.CachePort <= 0 {
    cleanup()
    return nil, nil, fmt.Errorf("cache port is required")
  }
  cachePort := strconv.Itoa(options.CachePort)
  // Same sibling-container networking for actions/cache; act injects
  // ACTIONS_CACHE_URL from --cache-server-external-url (no trailing slash).
  cacheExternalURL := fmt.Sprintf("http://host.docker.internal:%s", cachePort)

  actWorkflowPath, workflowMountArgs, err := resolveActWorkflowPath(repoPath, materialized.Path)
  if err != nil {
    cleanup()
    return nil, nil, err
  }

  args := []string{
    "run",
    "--rm",
    "-t",
    "-p", fmt.Sprintf("%s:%s", artifactPort, artifactPort),
    "-p", fmt.Sprintf("%s:%s", cachePort, cachePort),
    "--add-host", "host.docker.internal:host-gateway",
    "-e", "ACTIONS_RUNTIME_URL=" + runtimeURL,
    "-e", "ACT_DISABLE_VERSION_CHECK=1",
    "-v", fmt.Sprintf("%s:/workspace", repoPath),
    "-v", fmt.Sprintf("%s:/var/run/docker.sock", dockerSock),
    "-v", fmt.Sprintf("%s:%s", artifactDir, artifactMountPath),
    "-v", fmt.Sprintf("%s:%s", cacheDir, cacheMountPath),
    "-w", "/workspace",
  }
  args = append(args, workflowMountArgs...)

  actSecretArgs := make([]string, 0, len(secretFiles)*2+len(options.Secrets)*2)
  for index, secretFile := range secretFiles {
    workspacePath, inRepo, err := envfile.WorkspacePath(repoPath, secretFile)
    if err != nil {
      cleanup()
      return nil, nil, fmt.Errorf("secret file path: %w", err)
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
  // --concurrent-jobs=1: one job at a time so act stdout is not interleaved; actparse
  // cannot reliably attribute unprefixed "| …" lines across parallel jobs.
  args = append(args,
    "-P", "ubuntu-latest=catthehacker/ubuntu:act-latest",
    "-P", "ubuntu-24.04=catthehacker/ubuntu:act-24.04",
    "-P", "ubuntu-22.04=catthehacker/ubuntu:act-22.04",
    "-P", "ubuntu-20.04=catthehacker/ubuntu:act-20.04",
    "-P", "ubuntu-18.04=catthehacker/ubuntu:act-18.04",
    "--use-new-action-cache",
    "--concurrent-jobs", "1",
    "--artifact-server-path", artifactMountPath,
    "--artifact-server-addr", artifactBindAddr,
    "--artifact-server-port", artifactPort,
    "--cache-server-path", cacheMountPath,
    "--cache-server-addr", cacheBindAddr,
    "--cache-server-port", cachePort,
    "--cache-server-external-url", cacheExternalURL,
    "--container-options", "--add-host=host.docker.internal:host-gateway",
    "-W", actWorkflowPath,
  )
  if jobID != "" {
    args = append(args, "-j", jobID)
  }
  args = append(args, actSecretArgs...)
  for _, secret := range options.Secrets {
    args = append(args, "--secret", secret)
  }
  for _, input := range options.Inputs {
    args = append(args, "--input", input)
  }

  return args, cleanup, nil
}

func resolveActWorkflowPath(repoPath, materializedPath string) (containerPath string, mountArgs []string, err error) {
  absMaterialized, err := filepath.Abs(materializedPath)
  if err != nil {
    return "", nil, err
  }
  rel, relErr := filepath.Rel(repoPath, absMaterialized)
  if relErr == nil && rel != ".." && !filepath.IsAbs(rel) && !startsWithDotDot(rel) {
    return filepath.ToSlash(rel), nil, nil
  }
  return materializedWorkflowMount, []string{
    "-v", fmt.Sprintf("%s:%s:ro", absMaterialized, materializedWorkflowMount),
  }, nil
}

func startsWithDotDot(relative string) bool {
  return relative == ".." || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator)
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
  meta, err := LoadDispatch(filepath.Join(repoPath, filepath.FromSlash(relativePath)))
  if err != nil {
    return "", fmt.Errorf("resolve workflow event: %w", err)
  }
  return ResolveActEvent(meta.Events), nil
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

func resolveCacheMountDir(cacheDir string) (string, error) {
  if cacheDir == "" {
    return "", fmt.Errorf("cache dir is required")
  }
  absolutePath, err := filepath.Abs(cacheDir)
  if err != nil {
    return "", fmt.Errorf("resolve cache dir: %w", err)
  }
  if err := os.MkdirAll(absolutePath, 0o700); err != nil {
    return "", fmt.Errorf("create cache dir: %w", err)
  }
  return absolutePath, nil
}
