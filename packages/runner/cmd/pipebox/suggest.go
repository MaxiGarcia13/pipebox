package main

import (
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/termui"
)

func suggestUpdateIfNeeded() {
  install.MaybeSuggestUpdate(version, os.Stderr, termui.IsTTY(os.Stderr))
}
