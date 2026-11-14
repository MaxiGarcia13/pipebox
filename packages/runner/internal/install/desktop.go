package install

import (
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "runtime"
  "time"
)

const (
  releaseOwner     = "MaxiGarcia13"
  releaseRepo      = "pipebox"
  releaseTagPrefix = "packages/runner/v"
)

func InstallDesktop(version string) (string, error) {
  version = normalizeVersion(version)
  if version == "" {
    return "", fmt.Errorf("missing version for desktop download")
  }

  asset, err := desktopAssetName()
  if err != nil {
    return "", err
  }

  url := releaseAssetURL(version, asset)

  tmpDir, err := os.MkdirTemp("", "pipebox-desktop-*")
  if err != nil {
    return "", fmt.Errorf("temp dir: %w", err)
  }
  defer func() { _ = os.RemoveAll(tmpDir) }()

  zipPath := filepath.Join(tmpDir, asset)
  if err := downloadFile(url, zipPath); err != nil {
    return "", err
  }

  extractDir := filepath.Join(tmpDir, "extract")
  if err := os.MkdirAll(extractDir, 0o755); err != nil {
    return "", fmt.Errorf("extract dir: %w", err)
  }
  if err := unzip(zipPath, extractDir); err != nil {
    return "", err
  }

  return placeDesktop(extractDir)
}

func UninstallDesktop() (string, error) {
  path, err := userDesktopPath()
  if err != nil {
    return "", err
  }
  if _, err := os.Stat(path); err != nil {
    if os.IsNotExist(err) {
      return "", nil
    }
    return path, fmt.Errorf("stat desktop: %w", err)
  }
  if err := os.RemoveAll(path); err != nil {
    return path, fmt.Errorf("remove desktop: %w", err)
  }
  return path, nil
}

func desktopAssetName() (string, error) {
  switch runtime.GOOS {
  case "darwin":
    arch := "x64"
    if runtime.GOARCH == "arm64" {
      arch = "arm64"
    } else if runtime.GOARCH != "amd64" {
      return "", fmt.Errorf("unsupported darwin arch %s", runtime.GOARCH)
    }
    return fmt.Sprintf("PipeBox-darwin-%s.zip", arch), nil
  case "linux":
    if runtime.GOARCH != "amd64" {
      return "", fmt.Errorf("unsupported linux arch %s (need amd64)", runtime.GOARCH)
    }
    return "PipeBox-linux-amd64.zip", nil
  case "windows":
    if runtime.GOARCH != "amd64" {
      return "", fmt.Errorf("unsupported windows arch %s (need amd64)", runtime.GOARCH)
    }
    return "PipeBox-windows-x64.zip", nil
  default:
    return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
  }
}

func downloadFile(url, dest string) (err error) {
  client := &http.Client{Timeout: 5 * time.Minute}
  resp, err := client.Get(url)
  if err != nil {
    return fmt.Errorf("download: %w", err)
  }
  defer func() {
    if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close download body: %w", closeErr)
    }
  }()

  if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("download: %s (%s)", resp.Status, url)
  }

  out, err := os.Create(dest)
  if err != nil {
    return fmt.Errorf("create download: %w", err)
  }
  defer func() {
    if closeErr := out.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close download file: %w", closeErr)
    }
  }()

  if _, err := io.Copy(out, resp.Body); err != nil {
    return fmt.Errorf("write download: %w", err)
  }
  return nil
}
