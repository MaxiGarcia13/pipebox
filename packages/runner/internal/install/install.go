package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func targetPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(homeDir, ".local", "bin", "pipebox"), nil
}

func Run() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}

	path, err := targetPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create bin dir: %w", err)
	}

	if err := copyFile(executablePath, path); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("chmod: %w", err)
	}

	return path, nil
}

func Uninstall() (path string, removed bool, err error) {
	path, err = targetPath()
	if err != nil {
		return "", false, err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return path, false, nil
		}
		return "", false, fmt.Errorf("stat install: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return "", false, fmt.Errorf("remove install: %w", err)
	}
	return path, true, nil
}

func copyFile(sourcePath, targetPath string) (err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() {
		if closeErr := source.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close source: %w", closeErr)
		}
	}()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("open target: %w", err)
	}
	defer func() {
		if closeErr := target.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close target: %w", closeErr)
		}
	}()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}
	return nil
}
