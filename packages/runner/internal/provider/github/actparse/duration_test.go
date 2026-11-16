package actparse

import "testing"

func TestParseActDurationMs(t *testing.T) {
  t.Parallel()

  cases := []struct {
    raw  string
    want int64
  }{
    {"1.2s", 1200},
    {"500ms", 500},
    {"45s", 45000},
    {"1m2s", 62000},
    {"", 0},
    {"not-a-duration", 0},
  }
  for _, testCase := range cases {
    got := parseActDurationMs(testCase.raw)
    if got != testCase.want {
      t.Fatalf("%q: got %d want %d", testCase.raw, got, testCase.want)
    }
  }
}
