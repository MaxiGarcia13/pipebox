package install

import (
  "fmt"
  "io"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/hostcache"
)

const (
  suggestTTL     = time.Hour
  suggestTimeout = 2 * time.Second
)

type suggestConfig struct {
  currentVersion string
  errOut         io.Writer
  errTTY         bool
  now            time.Time
  ttl            time.Duration
  cachePath      string
  fetchLatest    func() (string, error)
}

func MaybeSuggestUpdate(currentVersion string, errOut io.Writer, errTTY bool) {
  maybeSuggestUpdate(suggestConfig{
    currentVersion: currentVersion,
    errOut:         errOut,
    errTTY:         errTTY,
    now:            time.Now(),
    ttl:            suggestTTL,
    fetchLatest: func() (string, error) {
      return latestVersionWithPrefixTimeout(releaseListAPIURL(), releaseTagPrefix, suggestTimeout)
    },
  })
}

func maybeSuggestUpdate(cfg suggestConfig) {
  if !cfg.errTTY || cfg.errOut == nil {
    return
  }
  current := normalizeVersion(cfg.currentVersion)
  if current == "" {
    return
  }
  if cfg.now.IsZero() {
    cfg.now = time.Now()
  }
  if cfg.ttl <= 0 {
    cfg.ttl = suggestTTL
  }
  if cfg.fetchLatest == nil {
    return
  }

  cachePath := cfg.cachePath
  if cachePath == "" {
    var err error
    cachePath, err = hostcache.Path("update-check.json")
    if err != nil {
      return
    }
  }

  if _, ok := hostcache.LoadFresh[string](cachePath, cfg.now, cfg.ttl); ok {
    return
  }

  latest, err := cfg.fetchLatest()
  if err != nil {
    return
  }
  latest = normalizeVersion(latest)
  _ = hostcache.Store(cachePath, cfg.now, latest)
  if latest == "" || latest == current {
    return
  }

  _, _ = fmt.Fprintf(
    cfg.errOut,
    "Update available: %s → %s\n  Run `pipebox update` to install.\n",
    current,
    latest,
  )
}
