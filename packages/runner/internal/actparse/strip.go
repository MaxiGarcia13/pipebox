package actparse

import "regexp"

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StripANSI(line string) string {
  return ansiEscape.ReplaceAllString(line, "")
}
