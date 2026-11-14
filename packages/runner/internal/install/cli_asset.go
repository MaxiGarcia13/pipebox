package install

import (
  "fmt"
  "runtime"
)

func cliAssetName(version string) (string, error) {
  version = normalizeVersion(version)
  if version == "" {
    return "", fmt.Errorf("missing version for cli download")
  }

  switch runtime.GOOS {
  case "darwin", "linux":
    switch runtime.GOARCH {
    case "amd64", "arm64":
      return fmt.Sprintf("pipebox_v%s_%s_%s", version, runtime.GOOS, runtime.GOARCH), nil
    default:
      return "", fmt.Errorf("unsupported %s arch %s", runtime.GOOS, runtime.GOARCH)
    }
  case "windows":
    if runtime.GOARCH != "amd64" {
      return "", fmt.Errorf("unsupported windows arch %s (need amd64)", runtime.GOARCH)
    }
    return fmt.Sprintf("pipebox_v%s_windows_amd64.exe", version), nil
  default:
    return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
  }
}

func releaseAssetURL(version, asset string) string {
  version = normalizeVersion(version)
  return fmt.Sprintf(
    "https://github.com/%s/%s/releases/download/%s%s/%s",
    releaseOwner,
    releaseRepo,
    releaseTagPrefix,
    version,
    asset,
  )
}
