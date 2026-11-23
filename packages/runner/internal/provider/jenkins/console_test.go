package jenkins

import "testing"

func TestConsoleBufferKeepsPartialLine(t *testing.T) {
  var buffer consoleBuffer
  lines := buffer.push("hello\nwor")
  if len(lines) != 1 || lines[0] != "hello" {
    t.Fatalf("lines: %#v", lines)
  }
  if buffer.pending != "wor" {
    t.Fatalf("pending: %q", buffer.pending)
  }
  lines = buffer.push("ld\nnext")
  if len(lines) != 1 || lines[0] != "world" {
    t.Fatalf("lines: %#v", lines)
  }
  if buffer.flush() != "next" || buffer.pending != "" {
    t.Fatalf("flush pending: %q", buffer.pending)
  }
}
