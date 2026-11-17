package jenkins

// Tracks whether the scanner is inside a ' or " literal (with escapes).
type quoteTracker struct {
  single bool
  double bool
}

func (tracker quoteTracker) active() bool {
  return tracker.single || tracker.double
}

// Consumes one character (or an escape pair) and returns how many bytes were used.
func (tracker *quoteTracker) advance(source string, index int) int {
  if tracker.single {
    if source[index] == '\\' && index+1 < len(source) {
      return 2
    }
    if source[index] == '\'' {
      tracker.single = false
    }
    return 1
  }
  if tracker.double {
    if source[index] == '\\' && index+1 < len(source) {
      return 2
    }
    if source[index] == '"' {
      tracker.double = false
    }
    return 1
  }
  switch source[index] {
  case '\'':
    tracker.single = true
  case '"':
    tracker.double = true
  }
  return 1
}

func inQuotes(source string, at int) bool {
  var tracker quoteTracker
  for index := 0; index < at && index < len(source); {
    index += tracker.advance(source, index)
  }
  return tracker.active()
}
