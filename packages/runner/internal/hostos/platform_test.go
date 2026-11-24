package hostos

import (
  "runtime"
  "testing"
)

func TestCurrentMatchesGOOS(t *testing.T) {
  t.Parallel()

  got := Current().ID()
  switch runtime.GOOS {
  case "darwin", "windows":
    if got != runtime.GOOS {
      t.Fatalf("Current().ID()=%q want %q", got, runtime.GOOS)
    }
  default:
    if got != "linux" {
      t.Fatalf("Current().ID()=%q want linux", got)
    }
  }
}
