package install

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strings"
  "time"
)

func normalizeVersion(version string) string {
  return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func parseReleaseTag(tag string) (string, error) {
  tag = strings.TrimSpace(tag)
  if !strings.HasPrefix(tag, releaseTagPrefix) {
    return "", fmt.Errorf("unexpected release tag %q", tag)
  }
  version := normalizeVersion(strings.TrimPrefix(tag, releaseTagPrefix))
  if version == "" {
    return "", fmt.Errorf("empty version in tag %q", tag)
  }
  return version, nil
}

func LatestVersion() (string, error) {
  return latestVersionFrom(releaseLatestAPIURL())
}

func releaseLatestAPIURL() string {
  return fmt.Sprintf(
    "https://api.github.com/repos/%s/%s/releases/latest",
    releaseOwner,
    releaseRepo,
  )
}

func latestVersionFrom(apiURL string) (string, error) {
  req, err := http.NewRequest(http.MethodGet, apiURL, nil)
  if err != nil {
    return "", fmt.Errorf("latest release request: %w", err)
  }
  req.Header.Set("Accept", "application/vnd.github+json")
  req.Header.Set("User-Agent", "pipebox")

  client := &http.Client{Timeout: 30 * time.Second}
  resp, err := client.Do(req)
  if err != nil {
    return "", fmt.Errorf("latest release: %w", err)
  }
  defer func() { _ = resp.Body.Close() }()

  body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
  if err != nil {
    return "", fmt.Errorf("latest release body: %w", err)
  }
  if resp.StatusCode != http.StatusOK {
    return "", fmt.Errorf("latest release: %s", resp.Status)
  }

  var release struct {
    TagName string `json:"tag_name"`
  }
  if err := json.Unmarshal(body, &release); err != nil {
    return "", fmt.Errorf("latest release json: %w", err)
  }
  return parseReleaseTag(release.TagName)
}
