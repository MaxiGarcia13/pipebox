package termui

import (
  "bytes"
  "errors"
  "io"
  "strings"
  "testing"
)

func TestPrinterOkUsesCheckmarkOnTTY(t *testing.T) {
  var out bytes.Buffer
  ui := newUI(&out, io.Discard, true, false)
  ui.Ok("CLI updated")
  if got := out.String(); got != "✓ CLI updated\n" {
    t.Fatalf("got %q", got)
  }
}

func TestPrinterOkPlainWithoutTTY(t *testing.T) {
  var out bytes.Buffer
  ui := newUI(&out, io.Discard, false, false)
  ui.Ok("CLI updated")
  if got := out.String(); got != "CLI updated\n" {
    t.Fatalf("got %q", got)
  }
}

func TestPrinterErr(t *testing.T) {
  var errOut bytes.Buffer
  ui := newUI(io.Discard, &errOut, false, false)
  ui.Err(errors.New("boom"))
  if got := errOut.String(); got != "error: boom\n" {
    t.Fatalf("got %q", got)
  }
}

func TestTrackPassthroughWithoutTTY(t *testing.T) {
  var errOut bytes.Buffer
  ui := newUI(io.Discard, &errOut, false, false)
  src := strings.NewReader("hello")
  wrapped := ui.Track("cli", 5, src)
  data, err := io.ReadAll(wrapped)
  if err != nil {
    t.Fatalf("ReadAll: %v", err)
  }
  if string(data) != "hello" {
    t.Fatalf("got %q", data)
  }
  if errOut.Len() != 0 {
    t.Fatalf("expected no progress output, got %q", errOut.String())
  }
}

func TestTrackRendersAndClearsOnTTY(t *testing.T) {
  var errOut bytes.Buffer
  ui := newUI(io.Discard, &errOut, false, true)
  src := strings.NewReader(strings.Repeat("a", 100))
  wrapped := ui.Track("cli", 100, src)
  data, err := io.ReadAll(wrapped)
  if err != nil {
    t.Fatalf("ReadAll: %v", err)
  }
  if len(data) != 100 {
    t.Fatalf("got %d bytes", len(data))
  }
  out := errOut.String()
  if !strings.Contains(out, "cli") {
    t.Fatalf("expected label in output: %q", out)
  }
  if !strings.HasSuffix(out, "\r\033[K") {
    t.Fatalf("expected cleared line suffix, got %q", out)
  }
}
