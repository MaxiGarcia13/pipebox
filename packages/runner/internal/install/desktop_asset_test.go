package install

import (
  "runtime"
  "strings"
  "testing"
)

func TestDesktopAssetName(t *testing.T) {
  name, err := desktopAssetName()
  if err != nil {
    t.Fatalf("desktopAssetName: %v", err)
  }
  if !strings.HasPrefix(name, "PipeBox-") || !strings.HasSuffix(name, ".zip") {
    t.Fatalf("unexpected asset name %q", name)
  }
  switch runtime.GOOS {
  case "darwin":
    if !strings.Contains(name, "darwin") {
      t.Fatalf("expected darwin asset, got %q", name)
    }
  case "linux":
    if name != "PipeBox-linux-amd64.zip" && runtime.GOARCH == "amd64" {
      t.Fatalf("expected linux amd64 zip, got %q", name)
    }
  case "windows":
    if name != "PipeBox-windows-x64.zip" && runtime.GOARCH == "amd64" {
      t.Fatalf("expected windows x64 zip, got %q", name)
    }
  }
}
