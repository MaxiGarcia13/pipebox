package install

import (
  "net/http"
  "net/http/httptest"
  "testing"
)

func TestNormalizeVersion(t *testing.T) {
  if got := normalizeVersion("v1.2.3"); got != "1.2.3" {
    t.Fatalf("got %q", got)
  }
  if got := normalizeVersion(" 1.2.3 "); got != "1.2.3" {
    t.Fatalf("got %q", got)
  }
}

func TestParseReleaseTag(t *testing.T) {
  version, err := parseReleaseTag("packages/runner/v0.4.1")
  if err != nil {
    t.Fatalf("parseReleaseTag: %v", err)
  }
  if version != "0.4.1" {
    t.Fatalf("got %q", version)
  }

  if _, err := parseReleaseTag("v0.4.1"); err == nil {
    t.Fatal("expected error for non-prefixed tag")
  }
}

func TestParseReleaseTagWithPrefix(t *testing.T) {
  version, err := parseReleaseTagWithPrefix("packages/mcp/v0.1.0", mcpReleaseTagPrefix)
  if err != nil {
    t.Fatalf("parseReleaseTagWithPrefix: %v", err)
  }
  if version != "0.1.0" {
    t.Fatalf("got %q", version)
  }

  if _, err := parseReleaseTagWithPrefix("packages/runner/v0.1.0", mcpReleaseTagPrefix); err == nil {
    t.Fatal("expected error for runner tag with mcp prefix")
  }
}

func TestLatestVersionWithPrefix(t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("User-Agent") == "" {
      t.Fatal("missing User-Agent")
    }
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write([]byte(`[
      {"tag_name":"packages/mcp/v0.5.0"},
      {"tag_name":"packages/runner/v0.4.1"},
      {"tag_name":"packages/mcp/v0.4.1"}
    ]`))
  }))
  defer server.Close()

  runnerVersion, err := latestVersionWithPrefix(server.URL, releaseTagPrefix)
  if err != nil {
    t.Fatalf("runner latest: %v", err)
  }
  if runnerVersion != "0.4.1" {
    t.Fatalf("runner got %q", runnerVersion)
  }

  mcpVersion, err := latestVersionWithPrefix(server.URL, mcpReleaseTagPrefix)
  if err != nil {
    t.Fatalf("mcp latest: %v", err)
  }
  if mcpVersion != "0.5.0" {
    t.Fatalf("mcp got %q", mcpVersion)
  }
}

func TestLatestVersionWithPrefixRejectsNoMatch(t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
    _, _ = w.Write([]byte(`[{"tag_name":"other/v1.0.0"}]`))
  }))
  defer server.Close()

  if _, err := latestVersionWithPrefix(server.URL, releaseTagPrefix); err == nil {
    t.Fatal("expected error")
  }
}

func TestUpdateCheckOnly(t *testing.T) {
  result, err := Update(UpdateOptions{
    CurrentVersion: "0.4.1",
    TargetVersion:  "0.5.0",
    CheckOnly:      true,
  })
  if err != nil {
    t.Fatalf("Update: %v", err)
  }
  if result.UpToDate {
    t.Fatal("expected update available")
  }
  if result.TargetVersion != "0.5.0" || result.CLIPath != "" {
    t.Fatalf("got %+v", result)
  }
}

func TestUpdateAlreadyCurrent(t *testing.T) {
  result, err := Update(UpdateOptions{
    CurrentVersion: "0.4.1",
    TargetVersion:  "v0.4.1",
  })
  if err != nil {
    t.Fatalf("Update: %v", err)
  }
  if !result.UpToDate {
    t.Fatalf("expected up to date, got %+v", result)
  }
  if result.CLIPath != "" || result.DesktopPath != "" {
    t.Fatalf("should not download when current: %+v", result)
  }
}
