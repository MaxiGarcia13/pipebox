package uiopen

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
)

func Open(repoPath string) error {
  appPath, err := ResolveApp()
  if err != nil {
    return err
  }

  args := []string{}
  if repoPath != "" {
    absRepo, absErr := filepath.Abs(repoPath)
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
    args = append(args, "--repo", absRepo)
  }

  return startApp(appPath, args)
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
