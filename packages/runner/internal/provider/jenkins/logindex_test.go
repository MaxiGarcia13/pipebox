package jenkins

import (
  "os"
  "path/filepath"
  "testing"
)

func TestParseLogIndex(t *testing.T) {
  marks := parseLogIndex("0 7\n18\n24 9\n")
  if len(marks) != 3 || marks[0].NodeID != "7" || marks[1].NodeID != "" || marks[1].Offset != 18 || marks[2].NodeID != "9" {
    t.Fatalf("marks: %#v", marks)
  }
}

func TestLogTailSplitsNodesOnNewline(t *testing.T) {
  dir := t.TempDir()
  logPath := filepath.Join(dir, "log")
  indexPath := logPath + "-index"
  body := "aaa\nbbb\ncc"
  if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(indexPath, []byte("0 1\n4 2\n"), 0o644); err != nil {
    t.Fatal(err)
  }

  var tail logTail
  lines, err := tail.read(logPath, indexPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(lines) != 2 || lines[0].NodeID != "1" || lines[0].Text != "aaa" || lines[1].NodeID != "2" || lines[1].Text != "bbb" {
    t.Fatalf("lines: %#v", lines)
  }
  if tail.pending != "cc" || tail.pendingNode != "2" {
    t.Fatalf("pending: %q node %q", tail.pending, tail.pendingNode)
  }

  file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
  if err != nil {
    t.Fatal(err)
  }
  if _, err := file.WriteString("c\n"); err != nil {
    t.Fatal(err)
  }
  _ = file.Close()

  lines, err = tail.read(logPath, indexPath)
  if err != nil {
    t.Fatal(err)
  }
  if len(lines) != 1 || lines[0].NodeID != "2" || lines[0].Text != "ccc" {
    t.Fatalf("lines: %#v pending %q", lines, tail.pending)
  }
}
