package jenkins

import (
  "os"
  "strconv"
  "strings"
)

// Jenkins FileLogStorage writes "<byte-offset> <flow-node-id>" when a step starts.
// A line with only an offset returns the log to the overall build.
type logMark struct {
  Offset int64
  NodeID string
}

type indexedLine struct {
  NodeID string
  Text   string
}

type logTail struct {
  pos         int64
  pending     string
  pendingNode string
}

func parseLogIndex(text string) []logMark {
  marks := make([]logMark, 0)
  for _, line := range strings.Split(text, "\n") {
    line = strings.TrimSpace(line)
    if line == "" {
      continue
    }
    offsetText := line
    nodeID := ""
    if space := strings.IndexByte(line, ' '); space >= 0 {
      offsetText = line[:space]
      nodeID = strings.TrimSpace(line[space+1:])
    }
    offset, err := strconv.ParseInt(offsetText, 10, 64)
    if err != nil {
      continue
    }
    marks = append(marks, logMark{Offset: offset, NodeID: nodeID})
  }
  return marks
}

func (tail *logTail) read(logPath, indexPath string) ([]indexedLine, error) {
  indexBytes, err := os.ReadFile(indexPath)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, err
  }
  marks := parseLogIndex(string(indexBytes))
  if len(marks) == 0 {
    return nil, nil
  }

  info, err := os.Stat(logPath)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, err
  }
  size := info.Size()
  if tail.pos > size {
    tail.pos = size
  }
  if tail.pos == size && tail.pending == "" {
    return nil, nil
  }

  file, err := os.Open(logPath)
  if err != nil {
    return nil, err
  }
  defer func() { _ = file.Close() }()

  lines := make([]indexedLine, 0)
  for _, span := range spansThrough(marks, size) {
    if span.End <= tail.pos {
      continue
    }
    from := span.Start
    if tail.pos > from {
      from = tail.pos
    }
    chunk := make([]byte, span.End-from)
    n, err := file.ReadAt(chunk, from)
    if err != nil && n == 0 {
      return lines, err
    }
    chunk = chunk[:n]
    tail.pos = from + int64(n)

    text := string(chunk)
    if tail.pending != "" && tail.pendingNode != span.NodeID {
      lines = append(lines, indexedLine{NodeID: tail.pendingNode, Text: tail.pending})
      tail.pending = ""
    }
    if tail.pending != "" && tail.pendingNode == span.NodeID {
      text = tail.pending + text
      tail.pending = ""
    }
    parts := strings.Split(text, "\n")
    closed := span.End < size || strings.HasSuffix(string(chunk), "\n")
    if !closed {
      tail.pending = parts[len(parts)-1]
      tail.pendingNode = span.NodeID
      parts = parts[:len(parts)-1]
    } else if len(parts) > 0 && parts[len(parts)-1] == "" {
      parts = parts[:len(parts)-1]
    }
    for _, part := range parts {
      lines = append(lines, indexedLine{NodeID: span.NodeID, Text: part})
    }
    if tail.pos >= size {
      break
    }
  }
  return lines, nil
}

type logSpan struct {
  Start  int64
  End    int64
  NodeID string
}

func spansThrough(marks []logMark, size int64) []logSpan {
  if len(marks) == 0 || size <= 0 {
    return nil
  }
  spans := make([]logSpan, 0, len(marks)+1)
  if marks[0].Offset > 0 {
    end := marks[0].Offset
    if end > size {
      end = size
    }
    spans = append(spans, logSpan{Start: 0, End: end})
  }
  for index, mark := range marks {
    if mark.Offset >= size {
      break
    }
    end := size
    if index+1 < len(marks) && marks[index+1].Offset < end {
      end = marks[index+1].Offset
    }
    if end < mark.Offset {
      continue
    }
    spans = append(spans, logSpan{Start: mark.Offset, End: end, NodeID: mark.NodeID})
  }
  return spans
}
