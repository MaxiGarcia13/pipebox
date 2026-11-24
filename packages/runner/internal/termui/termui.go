package termui

import (
  "fmt"
  "io"
  "os"
)

type UI struct {
  out    io.Writer
  errOut io.Writer
  outTTY bool
  errTTY bool
}

func New(out, errOut *os.File) *UI {
  return newUI(out, errOut, IsTTY(out), IsTTY(errOut))
}

func newUI(out, errOut io.Writer, outTTY, errTTY bool) *UI {
  return &UI{
    out:    out,
    errOut: errOut,
    outTTY: outTTY,
    errTTY: errTTY,
  }
}

func IsTTY(f *os.File) bool {
  if f == nil {
    return false
  }
  info, err := f.Stat()
  if err != nil {
    return false
  }
  return info.Mode()&os.ModeCharDevice != 0
}

func (u *UI) Info(msg string) {
  _, _ = fmt.Fprintln(u.out, msg)
}

func (u *UI) Ok(msg string) {
  if u.outTTY {
    _, _ = fmt.Fprintf(u.out, "✓ %s\n", msg)
    return
  }
  _, _ = fmt.Fprintln(u.out, msg)
}

func (u *UI) Fail(msg string) {
  if u.outTTY {
    _, _ = fmt.Fprintf(u.out, "✗ %s\n", msg)
    return
  }
  _, _ = fmt.Fprintln(u.out, msg)
}

func (u *UI) Err(err error) {
  if err == nil {
    return
  }
  _, _ = fmt.Fprintf(u.errOut, "error: %s\n", err.Error())
}

func (u *UI) Step(message string) {
  _, _ = fmt.Fprintln(u.errOut, message)
}
