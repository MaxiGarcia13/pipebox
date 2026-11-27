package actparse

import (
  "regexp"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/ansi"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  jobPrefixPattern = regexp.MustCompile(`^\[([^/\]]+)/([^\]]+)\]\s*(.*)$`)
  // Anchored: npm logs often contain "Run `npm …`" mid-line and must not become steps.
  stepStartPattern   = regexp.MustCompile(`(?i)^(?:⭐\s*)?Run\s+(.+)$`)
  stepSuccessPattern = regexp.MustCompile(`(?i)^(?:✅\s*)?Success\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  stepFailurePattern = regexp.MustCompile(`(?i)^(?:❌\s*)?Failure\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  stepSkippedPattern = regexp.MustCompile(`(?i)^(?:🚧\s*)?Skipped\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  // act prints these without a [workflow/job] prefix after another job already finished.
  globalActErrorPattern = regexp.MustCompile(`(?i)^Error:\s+Job\s+'.+'\s+failed`)
)

type Parser struct {
  currentJob  string
  currentStep string
  raw         bool
}

func NewParser() *Parser {
  return &Parser{}
}

func NewRawParser() *Parser {
  return &Parser{raw: true}
}

func (parser *Parser) HandleLine(rawLine string) []any {
  line := strings.TrimSpace(ansi.Strip(rawLine))
  if line == "" {
    return nil
  }

  if parser.raw {
    return parser.handleRawLine(line)
  }
  return parser.handleAttributedLine(line)
}

func (parser *Parser) handleRawLine(line string) []any {
  events := make([]any, 0, 2)

  if isGlobalActError(line) {
    events = append(events, globalErrorEvent(line))
  }

  job, rest, hasJobPrefix := splitJobPrefix(line)
  if !hasJobPrefix {
    // Unprefixed: never sticky-attribute across parallel jobs.
    events = append(events, unscopedLog(line))
    return events
  }

  parser.setJob(job)
  if stepEvents := parser.matchPrefixedStep(rest); len(stepEvents) > 0 {
    events = append(events, stepEvents...)
  }
  events = append(events, parser.logForJob(line))
  return events
}

func (parser *Parser) handleAttributedLine(line string) []any {
  if isGlobalActError(line) {
    return []any{globalErrorEvent(line)}
  }

  job, rest, hasJobPrefix := splitJobPrefix(line)
  if !hasJobPrefix {
    // act often omits [workflow/job] on `| …` continuations; keep those on the
    // active step. Drop other unprefixed lines so parallel jobs don't cross-talk.
    if strings.HasPrefix(line, "|") && parser.currentJob != "" && parser.currentStep != "" {
      return []any{parser.logForJob(line)}
    }
    return nil
  }

  parser.setJob(job)
  if stepEvents := parser.matchPrefixedStep(rest); len(stepEvents) > 0 {
    return stepEvents
  }
  if parser.currentStep == "" {
    return nil
  }
  return []any{parser.logForJob(line)}
}

func (parser *Parser) setJob(job string) {
  if job != parser.currentJob {
    parser.currentStep = ""
  }
  parser.currentJob = job
}

func (parser *Parser) matchPrefixedStep(message string) []any {
  if step, ok := matchStepStart(message); ok {
    return []any{parser.noteStepStart(step)}
  }
  if step, status, durationMs, ok := matchStepFinished(message); ok {
    return []any{parser.noteStepFinished(step, status, durationMs)}
  }
  return nil
}

func (parser *Parser) noteStepStart(step string) event.StepStarted {
  parser.currentStep = step
  return event.StepStarted{
    Type: event.TypeStepStarted,
    Job:  parser.currentJob,
    Step: step,
  }
}

func (parser *Parser) noteStepFinished(step, status string, durationMs int64) event.StepFinished {
  parser.currentStep = step
  return event.StepFinished{
    Type:       event.TypeStepFinished,
    Job:        parser.currentJob,
    Step:       step,
    Status:     status,
    DurationMs: durationMs,
  }
}

func (parser *Parser) logForJob(line string) event.Log {
  logEvent := event.Log{
    Type: event.TypeLog,
    Job:  parser.currentJob,
    Line: line,
  }
  if parser.currentStep != "" {
    logEvent.Step = parser.currentStep
  }
  return logEvent
}

func globalErrorEvent(line string) event.Error {
  return event.Error{
    Type:    event.TypeError,
    Message: line,
  }
}

func unscopedLog(line string) event.Log {
  return event.Log{
    Type: event.TypeLog,
    Line: line,
  }
}

func isGlobalActError(line string) bool {
  if globalActErrorPattern.MatchString(line) {
    return true
  }
  return strings.HasPrefix(line, "Error: Could not find any stages")
}

func splitJobPrefix(line string) (job string, rest string, ok bool) {
  matches := jobPrefixPattern.FindStringSubmatch(line)
  if matches == nil {
    return "", line, false
  }

  return strings.TrimSpace(matches[2]), matches[3], true
}

func matchStepStart(message string) (string, bool) {
  message = strings.TrimSpace(message)
  if message == "" || strings.HasPrefix(message, "|") {
    return "", false
  }
  if !strings.Contains(message, "Run ") && !strings.Contains(message, "⭐") {
    return "", false
  }
  // Avoid treating "docker run" style noise as step starts.
  if strings.Contains(message, "docker ") {
    return "", false
  }
  matches := stepStartPattern.FindStringSubmatch(message)
  if matches == nil {
    return "", false
  }
  step := strings.TrimSpace(matches[1])
  // npm hints: Run `npm audit` / Run 'npm fund'
  if step == "" || strings.HasPrefix(step, "`") || strings.HasPrefix(step, "'") {
    return "", false
  }
  return step, true
}

func matchStepFinished(message string) (step string, status string, durationMs int64, ok bool) {
  message = strings.TrimSpace(message)
  if message == "" || strings.HasPrefix(message, "|") {
    return "", "", 0, false
  }
  if matches := stepSuccessPattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "passed", parseActDurationMs(matches[2]), true
  }
  if matches := stepFailurePattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "failed", parseActDurationMs(matches[2]), true
  }
  if matches := stepSkippedPattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "skipped", parseActDurationMs(matches[2]), true
  }
  return "", "", 0, false
}
