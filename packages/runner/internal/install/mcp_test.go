package install

import (
  "runtime"
  "strings"
  "testing"
)

func TestMCPAssetName(t *testing.T) {
  name, err := mcpAssetName("0.1.0")
  if err != nil {
    t.Fatalf("mcpAssetName: %v", err)
  }
  if !strings.HasPrefix(name, "pipebox-mcp_v0.1.0_") {
    t.Fatalf("unexpected asset name %q", name)
  }

  switch runtime.GOOS {
  case "darwin", "linux":
    if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
      t.Skip("unsupported arch for this OS")
    }
    want := "pipebox-mcp_v0.1.0_" + runtime.GOOS + "_" + runtime.GOARCH
    if name != want {
      t.Fatalf("got %q want %q", name, want)
    }
  case "windows":
    if runtime.GOARCH != "amd64" {
      t.Skip("unsupported windows arch")
    }
    if name != "pipebox-mcp_v0.1.0_windows_amd64.exe" {
      t.Fatalf("got %q", name)
    }
  }
}

func TestMCPAssetNameNormalizesVPrefix(t *testing.T) {
  withV, err := mcpAssetName("v1.2.3")
  if err != nil {
    t.Fatalf("mcpAssetName: %v", err)
  }
  withoutV, err := mcpAssetName("1.2.3")
  if err != nil {
    t.Fatalf("mcpAssetName: %v", err)
  }
  if withV != withoutV {
    t.Fatalf("prefix mismatch: %q vs %q", withV, withoutV)
  }
}

func TestMCPReleaseAssetURL(t *testing.T) {
  url := mcpReleaseAssetURL("0.1.0", "pipebox-mcp_v0.1.0_darwin_arm64")
  want := "https://github.com/MaxiGarcia13/pipebox/releases/download/packages/mcp/v0.1.0/pipebox-mcp_v0.1.0_darwin_arm64"
  if url != want {
    t.Fatalf("got %q want %q", url, want)
  }
}

func TestMCPConfigSnippet(t *testing.T) {
  snippet := MCPConfigSnippet()
  if !strings.Contains(snippet, `"command": "pipebox-mcp"`) {
    t.Fatalf("unexpected snippet: %s", snippet)
  }
}
