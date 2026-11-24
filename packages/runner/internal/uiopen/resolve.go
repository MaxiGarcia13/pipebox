package uiopen

import (
  "fmt"
  "os"
  "path/filepath"
  "runtime"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostos"
)

func ResolveApp() (string, error) {
  plat := hostos.Current()

  if envPath := os.Getenv("PIPEBOX_APP"); envPath != "" {
    if abs, err := filepath.Abs(envPath); err == nil {
      envPath = abs
    }
    if plat.IsLaunchable(envPath) {
      return envPath, nil
    }
    return "", fmt.Errorf("PIPEBOX_APP not found or not launchable: %s", envPath)
  }

  for _, candidate := range plat.DesktopLaunchCandidates() {
    if plat.IsLaunchable(candidate) {
      return candidate, nil
    }
  }

  for _, candidate := range monorepoCandidates(plat) {
    if plat.IsLaunchable(candidate) {
      return candidate, nil
    }
  }

  return "", fmt.Errorf(
    "desktop app not found; run: pipebox install-desktop (or set PIPEBOX_APP)",
  )
}

func monorepoCandidates(plat hostos.Platform) []string {
  executablePath, err := os.Executable()
  if err != nil {
    return nil
  }
  executablePath, err = filepath.EvalSymlinks(executablePath)
  if err != nil {
    return nil
  }

  runnerDir := filepath.Dir(executablePath)
  appOut := filepath.Clean(filepath.Join(runnerDir, "..", "app", "out"))
  entries, err := os.ReadDir(appOut)
  if err != nil {
    return nil
  }

  var candidates []string
  for _, entry := range entries {
    if !entry.IsDir() {
      continue
    }
    platformDir := filepath.Join(appOut, entry.Name())
    switch runtime.GOOS {
    case "darwin":
      matches, _ := filepath.Glob(filepath.Join(platformDir, "*.app"))
      candidates = append(candidates, matches...)
    case "linux":
      matches, _ := filepath.Glob(filepath.Join(platformDir, "*"))
      for _, match := range matches {
        if plat.IsLaunchable(match) {
          candidates = append(candidates, match)
        }
      }
    case "windows":
      matches, _ := filepath.Glob(filepath.Join(platformDir, "*.exe"))
      candidates = append(candidates, matches...)
    }
  }
  return candidates
}
