package install

import (
  "fmt"
  "os"
  "path/filepath"
  "runtime"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
)

func InstallCompletions() ([]string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return nil, fmt.Errorf("resolve home: %w", err)
  }

  targets := completionTargets(homeDir)
  installed := make([]string, 0, len(targets))
  for _, target := range targets {
    script, err := cli.CompletionScript(target.Shell)
    if err != nil {
      return installed, err
    }
    if err := os.MkdirAll(filepath.Dir(target.Path), 0o755); err != nil {
      return installed, fmt.Errorf("create completion dir: %w", err)
    }
    if err := os.WriteFile(target.Path, []byte(script), 0o644); err != nil {
      return installed, fmt.Errorf("write %s: %w", target.Path, err)
    }
    installed = append(installed, target.Path)
  }
  return installed, nil
}

func UninstallCompletions() ([]string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return nil, fmt.Errorf("resolve home: %w", err)
  }

  removed := make([]string, 0)
  for _, target := range completionTargets(homeDir) {
    if _, err := os.Stat(target.Path); err != nil {
      if os.IsNotExist(err) {
        continue
      }
      return removed, fmt.Errorf("stat %s: %w", target.Path, err)
    }
    if err := os.Remove(target.Path); err != nil {
      return removed, fmt.Errorf("remove %s: %w", target.Path, err)
    }
    removed = append(removed, target.Path)
  }
  return removed, nil
}

type completionTarget struct {
  Shell string
  Path  string
}

func completionTargets(homeDir string) []completionTarget {
  targets := []completionTarget{
    {
      Shell: "zsh",
      Path:  filepath.Join(homeDir, ".local", "share", "zsh", "site-functions", "_pipebox"),
    },
    {
      Shell: "bash",
      Path: filepath.Join(
        homeDir,
        ".local",
        "share",
        "bash-completion",
        "completions",
        "pipebox",
      ),
    },
    {
      Shell: "fish",
      Path: filepath.Join(
        homeDir,
        ".config",
        "fish",
        "completions",
        "pipebox.fish",
      ),
    },
  }
  if runtime.GOOS == "darwin" {
    targets = append(targets, completionTarget{
      Shell: "zsh",
      Path:  filepath.Join(homeDir, ".zsh", "completions", "_pipebox"),
    })
  }
  return targets
}
