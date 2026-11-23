package jenkins

import (
  "context"
  "fmt"
  "io"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "strconv"
  "strings"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/envfile"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins/jfrimage"
)

const (
  jenkinsHTTPPort        = "8080"
  pipeboxJobName         = "pipebox"
  controllerReadyTimeout = 3 * time.Minute
  consolePollInterval    = 500 * time.Millisecond
)

type controllerSession struct {
  baseURL string
  runHome string
  client  *http.Client
  cleanup func()
}

func runController(ctx context.Context, options Options) (exitCode int, status string) {
  if _, err := exec.LookPath("docker"); err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: "docker not found on PATH"})
    return 1, "failed"
  }
  if err := jfrimage.Ensure(); err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("jenkins image: %v", err)})
    return 1, "failed"
  }
  // Image exists now; the controller has not started yet.
  doctor.Emit()

  session, err := startController(ctx, options)
  if session != nil && session.cleanup != nil {
    defer session.cleanup()
  }
  if err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: err.Error()})
    return 1, "failed"
  }

  if err := waitForJenkins(ctx, session); err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("jenkins ready: %v", err)})
    return 1, "failed"
  }

  script, err := loadPipelineScript(options)
  if err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: err.Error()})
    return 1, "failed"
  }

  if err := createOrUpdateJob(ctx, session, script, options.Inputs); err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("create job: %v", err)})
    return 1, "failed"
  }

  buildNumber, err := triggerBuild(ctx, session, script, options.Inputs)
  if err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("trigger build: %v", err)})
    return 1, "failed"
  }

  graph, err := Graph(options.Repo, options.Workflow)
  if err != nil {
    _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("job graph: %v", err)})
    return streamConsole(ctx, session, buildNumber)
  }
  return streamBuild(ctx, session, buildNumber, graph.Jobs)
}

func startController(ctx context.Context, options Options) (*controllerSession, error) {
  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return nil, fmt.Errorf("resolve repo path: %w", err)
  }

  runHome := options.RunHome
  cleanups := make([]func(), 0)
  cleanup := func() {
    for index := len(cleanups) - 1; index >= 0; index-- {
      cleanups[index]()
    }
  }

  if runHome == "" {
    dir, err := os.MkdirTemp("", "pipebox-jenkins-home-*")
    if err != nil {
      return nil, fmt.Errorf("create run home: %w", err)
    }
    runHome = dir
    cleanups = append(cleanups, func() { _ = os.RemoveAll(dir) })
  }

  secretFiles, err := envfile.Resolve(repoPath, options.SecretFile)
  if err != nil {
    cleanup()
    return nil, fmt.Errorf("resolve secret files: %w", err)
  }
  secretEnv, err := envfile.Load(secretFiles, options.Secrets)
  if err != nil {
    cleanup()
    return nil, err
  }
  if err := writeSecretsCasc(runHome, secretEnv); err != nil {
    cleanup()
    return nil, fmt.Errorf("write secrets: %w", err)
  }

  mounts, mountCleanup, err := prepareModuleMounts(repoPath)
  if mountCleanup != nil {
    cleanups = append(cleanups, mountCleanup)
  }
  if err != nil {
    cleanup()
    return nil, err
  }

  cmd := exec.CommandContext(ctx, "docker", controllerRunArgs(repoPath, runHome, jfrimage.Image, mounts)...)
  output, err := cmd.CombinedOutput()
  if err != nil {
    cleanup()
    return nil, fmt.Errorf("start jenkins: %w (%s)", err, strings.TrimSpace(string(output)))
  }
  containerID := strings.TrimSpace(string(output))
  if containerID == "" {
    cleanup()
    return nil, fmt.Errorf("start jenkins: empty container id")
  }

  cleanups = append(cleanups, func() {
    stop := exec.Command("docker", "rm", "-f", containerID)
    _ = stop.Run()
    // Durable tasks create this beside the workspace for the run.
    _ = os.RemoveAll(repoPath + "@tmp")
  })

  hostPort, err := inspectHostPort(containerID, jenkinsHTTPPort)
  if err != nil {
    cleanup()
    return nil, err
  }

  return &controllerSession{
    baseURL: "http://127.0.0.1:" + hostPort,
    runHome: runHome,
    client:  &http.Client{Timeout: 30 * time.Second},
    cleanup: cleanup,
  }, nil
}

func controllerRunArgs(repoPath, runHome, image string, moduleMounts []moduleMount) []string {
  // Durable tasks create <repo>@tmp beside the workspace, so the parent has to
  // be the real host directory. The entrypoint starts as root, then setpriv
  // continues as this uid with the socket group seen inside the container.
  // The host socket gid does not match that group on Docker Desktop.
  // Host node_modules are platform-specific. Empty mounts let pnpm install
  // Linux bindings without rewriting the checkout.
  args := []string{
    "run",
    "-d",
    "--rm",
    "-u", "0",
    "--hostname", "pipebox-jenkins",
    "-p", "127.0.0.1::" + jenkinsHTTPPort,
    "-v", runHome + ":/var/jenkins_home",
    "-v", filepath.Dir(repoPath) + ":" + filepath.Dir(repoPath),
    "-v", dockerimage.SocketPath() + ":/var/run/docker.sock",
  }
  for _, mount := range moduleMounts {
    args = append(args, "-v", mount.Host+":"+mount.Path)
  }
  if uid, gid := os.Getuid(), os.Getgid(); uid >= 0 && gid >= 0 {
    args = append(args, "-e", "PIPEBOX_UID="+strconv.Itoa(uid), "-e", "PIPEBOX_GID="+strconv.Itoa(gid))
  }
  args = append(args, image)
  return args
}

func inspectHostPort(containerID, containerPort string) (string, error) {
  format := fmt.Sprintf(`{{(index (index .NetworkSettings.Ports "%s/tcp") 0).HostPort}}`, containerPort)
  cmd := exec.Command("docker", "inspect", "-f", format, containerID)
  output, err := cmd.CombinedOutput()
  if err != nil {
    return "", fmt.Errorf("inspect port: %w (%s)", err, strings.TrimSpace(string(output)))
  }
  port := strings.TrimSpace(string(output))
  if port == "" || port == "<no value>" {
    return "", fmt.Errorf("inspect port: no host port for %s", containerPort)
  }
  return port, nil
}

func waitForJenkins(ctx context.Context, session *controllerSession) error {
  deadline := time.Now().Add(controllerReadyTimeout)
  var lastErr error
  for time.Now().Before(deadline) {
    if err := ctx.Err(); err != nil {
      return err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, session.baseURL+"/api/json?tree=mode", nil)
    if err != nil {
      return err
    }
    resp, err := session.client.Do(req)
    if err == nil {
      jenkinsHeader := resp.Header.Get("X-Jenkins")
      _, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
      _ = resp.Body.Close()
      if resp.StatusCode == http.StatusOK && jenkinsHeader != "" {
        return nil
      }
      lastErr = fmt.Errorf("status %d", resp.StatusCode)
    } else {
      lastErr = err
    }
    time.Sleep(time.Second)
  }
  return fmt.Errorf("timeout waiting for jenkins: %v", lastErr)
}
