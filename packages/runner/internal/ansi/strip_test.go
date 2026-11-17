package ansi

import "testing"

func TestStrip(t *testing.T) {
  input := "\x1b[36m[ci/build]\x1b[0m hello"
  got := Strip(input)
  if got != "[ci/build] hello" {
    t.Fatalf("got %q", got)
  }
}
