package uiopen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func ResolveApp() (string, error) {
	if envPath := os.Getenv("PIPEBOX_APP"); envPath != "" {
		if abs, err := filepath.Abs(envPath); err == nil {
			envPath = abs
		}
		if isLaunchable(envPath) {
			return envPath, nil
		}
		return "", fmt.Errorf("PIPEBOX_APP not found or not launchable: %s", envPath)
	}

	for _, candidate := range platformDefaults() {
		if isLaunchable(candidate) {
			return candidate, nil
		}
	}

	for _, candidate := range monorepoCandidates() {
		if isLaunchable(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf(
		"desktop app not found; install PipeBox, run pnpm --filter @pipebox/app package, or set PIPEBOX_APP",
	)
}

func platformDefaults() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/PipeBox.app",
			"/Applications/@pipebox-app.app",
		}
	case "linux":
		return []string{
			"/opt/PipeBox/pipebox-app",
			"/usr/bin/pipebox-app",
			"/usr/local/bin/pipebox-app",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return nil
		}
		return []string{
			filepath.Join(localAppData, "PipeBox", "PipeBox.exe"),
			filepath.Join(localAppData, "@pipebox-app", "@pipebox-app.exe"),
		}
	default:
		return nil
	}
}

func monorepoCandidates() []string {
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
				info, statErr := os.Stat(match)
				if statErr != nil || info.IsDir() {
					continue
				}
				if info.Mode()&0o111 != 0 {
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

func isLaunchable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if runtime.GOOS == "darwin" && info.IsDir() && filepath.Ext(path) == ".app" {
		executable := macOSExecutable(path)
		execInfo, execErr := os.Stat(executable)
		if execErr != nil || execInfo.IsDir() || execInfo.Mode()&0o111 == 0 {
			return false
		}
		// Incomplete Forge packages may ship a stub binary without Electron Framework.
		framework := filepath.Join(path, "Contents", "Frameworks", "Electron Framework.framework")
		_, frameworkErr := os.Stat(framework)
		return frameworkErr == nil
	}
	if info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}
