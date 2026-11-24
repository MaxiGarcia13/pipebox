package uiopen

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostos"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
)

func Open(repoPath string, newWindow bool) error {
  appPath, err := ResolveApp()
  if err != nil {
    return err
  }

  repoPath = implicitRepo(repoPath)
  absRepo := ""
  if repoPath != "" {
    var absErr error
    absRepo, absErr = filepath.Abs(repoPath)
    if absErr != nil {
      return fmt.Errorf("resolve --repo: %w", absErr)
    }
    info, statErr := os.Stat(absRepo)
    if statErr != nil {
      return fmt.Errorf("resolve --repo: %w", statErr)
    }
    if !info.IsDir() {
      return fmt.Errorf("--repo is not a directory: %s", absRepo)
    }
  }

  return startApp(appPath, launchArgs(absRepo, newWindow))
}

func implicitRepo(repoPath string) string {
  if repoPath != "" {
    return repoPath
  }
  cwd, err := os.Getwd()
  if err != nil {
    return ""
  }
  if len(provider.Discover(cwd)) == 0 {
    return ""
  }
  return cwd
}

// Chromium keeps --name=value as one switch; the two-token form --name value is
// split on second-instance and breaks Electron argv parsing.
func launchArgs(absRepo string, newWindow bool) []string {
  args := make([]string, 0, 2)
  if absRepo != "" {
    args = append(args, "--repo="+absRepo)
  }
  if newWindow {
    args = append(args, "--new")
  }
  return args
}

func startApp(appPath string, args []string) error {
  binaryPath := appPath
  if runtime.GOOS == "darwin" && strings.HasSuffix(appPath, ".app") {
    binaryPath = hostos.MacOSAppExecutable(appPath)
  }

  cmd := exec.Command(binaryPath, args...)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Start(); err != nil {
    return fmt.Errorf("launch desktop app: %w", err)
  }
  return nil
}
