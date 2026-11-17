package uiopen

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
)

func Open(repoPath string, newWindow bool) error {
  appPath, err := ResolveApp()
  if err != nil {
    return err
  }

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
    binaryPath = macOSExecutable(appPath)
  }

  cmd := exec.Command(binaryPath, args...)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Start(); err != nil {
    return fmt.Errorf("launch desktop app: %w", err)
  }
  return nil
}

func macOSExecutable(appPath string) string {
  base := strings.TrimSuffix(filepath.Base(appPath), ".app")
  return filepath.Join(appPath, "Contents", "MacOS", base)
}
