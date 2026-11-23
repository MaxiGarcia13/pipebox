package install

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
)

const mcpReleaseTagPrefix = "packages/mcp/v"

func InstallMCP(version string, progress Progress) (string, error) {
  progress = resolveProgress(progress)
  version = normalizeVersion(version)
  if version == "" {
    return "", fmt.Errorf("missing version for mcp download")
  }

  asset, err := mcpAssetName(version)
  if err != nil {
    return "", err
  }

  target, err := mcpTargetPath()
  if err != nil {
    return "", err
  }

  tmpDir, err := os.MkdirTemp("", "pipebox-mcp-*")
  if err != nil {
    return "", fmt.Errorf("temp dir: %w", err)
  }
  defer func() { _ = os.RemoveAll(tmpDir) }()

  progress.Step("Downloading MCP…")
  tmpPath := filepath.Join(tmpDir, asset)
  if err := downloadFile(mcpReleaseAssetURL(version, asset), tmpPath, "mcp", progress); err != nil {
    return "", err
  }

  progress.Step("Installing MCP…")
  if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
    return "", fmt.Errorf("create bin dir: %w", err)
  }
  if err := copyFile(tmpPath, target); err != nil {
    return "", err
  }
  if err := os.Chmod(target, 0o755); err != nil {
    return "", fmt.Errorf("chmod: %w", err)
  }
  return target, nil
}

func UninstallMCP() (string, error) {
  path, err := mcpTargetPath()
  if err != nil {
    return "", err
  }
  if _, err := os.Stat(path); err != nil {
    if os.IsNotExist(err) {
      return "", nil
    }
    return path, fmt.Errorf("stat mcp: %w", err)
  }
  if err := os.Remove(path); err != nil {
    return path, fmt.Errorf("remove mcp: %w", err)
  }
  return path, nil
}

func MCPInstalled() bool {
  path, err := mcpTargetPath()
  if err != nil {
    return false
  }
  info, err := os.Stat(path)
  return err == nil && !info.IsDir()
}

func ResolveMCP() (string, error) {
  if envPath := os.Getenv("PIPEBOX_MCP"); envPath != "" {
    if abs, err := filepath.Abs(envPath); err == nil {
      envPath = abs
    }
    if isExecutableFile(envPath) {
      return envPath, nil
    }
    return "", fmt.Errorf("PIPEBOX_MCP not found or not executable: %s", envPath)
  }

  if path, err := mcpTargetPath(); err == nil && isExecutableFile(path) {
    return path, nil
  }

  if path, err := exec.LookPath(mcpBinaryName()); err == nil {
    return path, nil
  }

  for _, candidate := range mcpMonorepoCandidates() {
    if isExecutableFile(candidate) {
      return candidate, nil
    }
  }

  return "", fmt.Errorf(
    "mcp server not found; run: pipebox install-mcp (or set PIPEBOX_MCP)",
  )
}

func MCPConfigSnippet() string {
  return `{
  "mcpServers": {
    "pipebox": {
      "command": "pipebox-mcp"
    }
  }
}`
}

func mcpBinaryName() string {
  if runtime.GOOS == "windows" {
    return "pipebox-mcp.exe"
  }
  return "pipebox-mcp"
}

func mcpTargetPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, ".local", "bin", mcpBinaryName()), nil
}

func mcpAssetName(version string) (string, error) {
  return mcpAssetNameFor(version, runtime.GOOS, runtime.GOARCH)
}

func mcpAssetNameFor(version, goos, goarch string) (string, error) {
  version = normalizeVersion(version)
  if version == "" {
    return "", fmt.Errorf("missing version for mcp download")
  }

  switch goos {
  case "darwin", "linux":
    switch goarch {
    case "amd64", "arm64":
      return fmt.Sprintf("pipebox-mcp_v%s_%s_%s", version, goos, goarch), nil
    default:
      return "", fmt.Errorf("unsupported %s arch %s", goos, goarch)
    }
  case "windows":
    if goarch != "amd64" {
      return "", fmt.Errorf("unsupported windows arch %s (need amd64)", goarch)
    }
    return fmt.Sprintf("pipebox-mcp_v%s_windows_amd64.exe", version), nil
  default:
    return "", fmt.Errorf("unsupported OS %s", goos)
  }
}

func mcpReleaseAssetURL(version, asset string) string {
  version = normalizeVersion(version)
  return fmt.Sprintf(
    "https://github.com/%s/%s/releases/download/%s%s/%s",
    releaseOwner,
    releaseRepo,
    mcpReleaseTagPrefix,
    version,
    asset,
  )
}

func mcpMonorepoCandidates() []string {
  executablePath, err := os.Executable()
  if err != nil {
    return nil
  }
  executablePath, err = filepath.EvalSymlinks(executablePath)
  if err != nil {
    return nil
  }

  runnerDir := filepath.Dir(executablePath)
  return []string{
    filepath.Clean(filepath.Join(runnerDir, "..", "mcp", mcpBinaryName())),
    filepath.Clean(filepath.Join(runnerDir, mcpBinaryName())),
  }
}

func isExecutableFile(path string) bool {
  info, err := os.Stat(path)
  if err != nil || info.IsDir() {
    return false
  }
  if runtime.GOOS == "windows" {
    return true
  }
  return info.Mode()&0o111 != 0
}
