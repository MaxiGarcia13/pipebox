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

func parseReleaseTagWithPrefix(tag, prefix string) (string, error) {
  tag = strings.TrimSpace(tag)
  if !strings.HasPrefix(tag, prefix) {
    return "", fmt.Errorf("unexpected release tag %q", tag)
  }
  version := normalizeVersion(strings.TrimPrefix(tag, prefix))
  if version == "" {
    return "", fmt.Errorf("empty version in tag %q", tag)
  }
  return version, nil
}

func parseReleaseTag(tag string) (string, error) {
  return parseReleaseTagWithPrefix(tag, releaseTagPrefix)
}

func LatestVersion() (string, error) {
  return latestVersionWithPrefix(releaseListAPIURL(), releaseTagPrefix)
}

func LatestMCPVersion() (string, error) {
  return latestVersionWithPrefix(releaseListAPIURL(), mcpReleaseTagPrefix)
}

func releaseListAPIURL() string {
  return fmt.Sprintf(
    "https://api.github.com/repos/%s/%s/releases?per_page=30",
    releaseOwner,
    releaseRepo,
  )
}

func latestVersionWithPrefix(apiURL, prefix string) (string, error) {
  return latestVersionWithPrefixTimeout(apiURL, prefix, 30*time.Second)
}

func latestVersionWithPrefixTimeout(apiURL, prefix string, timeout time.Duration) (string, error) {
  req, err := http.NewRequest(http.MethodGet, apiURL, nil)
  if err != nil {
    return "", fmt.Errorf("latest release request: %w", err)
  }
  req.Header.Set("Accept", "application/vnd.github+json")
  req.Header.Set("User-Agent", "pipebox")

  if timeout <= 0 {
    timeout = 30 * time.Second
  }
  client := &http.Client{Timeout: timeout}
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

  var releases []struct {
    TagName string `json:"tag_name"`
  }
  if err := json.Unmarshal(body, &releases); err != nil {
    return "", fmt.Errorf("latest release json: %w", err)
  }
  for _, release := range releases {
    version, err := parseReleaseTagWithPrefix(release.TagName, prefix)
    if err != nil {
      continue
    }
    return version, nil
  }
  return "", fmt.Errorf("no release with prefix %q", prefix)
}
