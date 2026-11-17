package termui

import (
  "fmt"
  "io"
  "strings"
  "sync"
  "time"
)

func (u *UI) Track(label string, total int64, r io.Reader) io.Reader {
  if r == nil || !u.errTTY {
    return r
  }
  return &trackReader{
    ui:    u,
    label: label,
    total: total,
    inner: r,
  }
}

type trackReader struct {
  ui    *UI
  label string
  total int64
  inner io.Reader
  read  int64
  spin  int
  last  time.Time
  done  bool
  mu    sync.Mutex
}

var spinnerFrames = []rune{'|', '/', '-', '\\'}

func (t *trackReader) Read(p []byte) (int, error) {
  n, err := t.inner.Read(p)
  t.mu.Lock()
  defer t.mu.Unlock()
  if n > 0 {
    t.read += int64(n)
    t.render(false)
  }
  if err != nil {
    t.finish()
  }
  return n, err
}

func (t *trackReader) render(force bool) {
  if t.done {
    return
  }
  now := time.Now()
  if !force && !t.last.IsZero() && now.Sub(t.last) < 50*time.Millisecond {
    return
  }
  t.last = now

  if t.total > 0 {
    pct := int(t.read * 100 / t.total)
    if pct > 100 {
      pct = 100
    }
    const width = 28
    filled := width * pct / 100
    bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
    _, _ = fmt.Fprintf(t.ui.errOut, "\r%s [%s] %3d%%", t.label, bar, pct)
    return
  }

  frame := spinnerFrames[t.spin%len(spinnerFrames)]
  t.spin++
  _, _ = fmt.Fprintf(t.ui.errOut, "\r%s %c %s", t.label, frame, formatBytes(t.read))
}

func (t *trackReader) finish() {
  if t.done {
    return
  }
  t.done = true
  t.render(true)
  _, _ = fmt.Fprint(t.ui.errOut, "\r\033[K")
}

func formatBytes(n int64) string {
  const (
    kb = 1024
    mb = 1024 * kb
  )
  switch {
  case n >= mb:
    return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
  case n >= kb:
    return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
  default:
    return fmt.Sprintf("%d B", n)
  }
}
