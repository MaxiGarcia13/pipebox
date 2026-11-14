package install

import (
  "runtime"
  "strings"
  "testing"
)

func TestCLIAssetName(t *testing.T) {
  name, err := cliAssetName("0.4.1")
  if err != nil {
    t.Fatalf("cliAssetName: %v", err)
  }
  if !strings.HasPrefix(name, "pipebox_v0.4.1_") {
    t.Fatalf("unexpected asset name %q", name)
  }

  switch runtime.GOOS {
  case "darwin", "linux":
    want := "pipebox_v0.4.1_" + runtime.GOOS + "_" + runtime.GOARCH
    if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
      t.Skip("unsupported arch for this OS")
    }
    if name != want {
      t.Fatalf("got %q want %q", name, want)
    }
  case "windows":
    if runtime.GOARCH != "amd64" {
      t.Skip("unsupported windows arch")
    }
    if name != "pipebox_v0.4.1_windows_amd64.exe" {
      t.Fatalf("got %q", name)
    }
  }
}

func TestCLIAssetNameNormalizesVPrefix(t *testing.T) {
  withV, err := cliAssetName("v1.2.3")
  if err != nil {
    t.Fatalf("cliAssetName: %v", err)
  }
  withoutV, err := cliAssetName("1.2.3")
  if err != nil {
    t.Fatalf("cliAssetName: %v", err)
  }
  if withV != withoutV {
    t.Fatalf("prefix mismatch: %q vs %q", withV, withoutV)
  }
}

func TestReleaseAssetURL(t *testing.T) {
  url := releaseAssetURL("0.4.1", "pipebox_v0.4.1_darwin_arm64")
  want := "https://github.com/MaxiGarcia13/pipebox/releases/download/packages/runner/v0.4.1/pipebox_v0.4.1_darwin_arm64"
  if url != want {
    t.Fatalf("got %q want %q", url, want)
  }
}
