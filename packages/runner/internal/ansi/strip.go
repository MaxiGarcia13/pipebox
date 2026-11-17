package ansi

import "regexp"

var escape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func Strip(line string) string {
  return escape.ReplaceAllString(line, "")
}
