package actparse

import (
  "regexp"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  jobPrefixPattern = regexp.MustCompile(`^\[([^/\]]+)/([^\]]+)\]\s*(.*)$`)
  // Anchored: npm logs often contain "Run `npm …`" mid-line and must not become steps.
  stepStartPattern = regexp.MustCompile(`(?i)^(?:⭐\s*)?Run\s+(.+)$`)
  stepSuccessPattern = regexp.MustCompile(`(?i)^(?:✅\s*)?Success\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  stepFailurePattern = regexp.MustCompile(`(?i)^(?:❌\s*)?Failure\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  stepSkippedPattern = regexp.MustCompile(`(?i)^(?:🚧\s*)?Skipped\s+-\s+(.+?)(?:\s+\[([^\]]+)\])?\s*$`)
  // act prints these without a [workflow/job] prefix after another job already finished.
  globalActErrorPattern = regexp.MustCompile(`(?i)^Error:\s+Job\s+'.+'\s+failed`)
)

type Parser struct {
  currentJob  string
  currentStep string
}

func NewParser() *Parser {
  return &Parser{}
}

func (parser *Parser) HandleLine(rawLine string) []any {
  line := strings.TrimSpace(StripANSI(rawLine))
  if line == "" {
    return nil
  }

  if isGlobalActError(line) {
    return []any{
      event.Error{
        Type:    event.TypeError,
        Message: line,
      },
    }
  }

  job, rest, hasJobPrefix := splitJobPrefix(line)
  if !hasJobPrefix {
    // act often omits [workflow/job] on `| …` continuations; keep those on the
    // active step. Drop other unprefixed lines so parallel jobs don't cross-talk.
    if strings.HasPrefix(line, "|") && parser.currentJob != "" && parser.currentStep != "" {
      return []any{
        event.Log{
          Type: event.TypeLog,
          Job:  parser.currentJob,
          Step: parser.currentStep,
          Line: line,
        },
      }
    }
    return nil
  }
  if job != parser.currentJob {
    parser.currentStep = ""
  }
  parser.currentJob = job
  message := rest

  if step, ok := matchStepStart(message); ok {
    parser.currentStep = step
    return []any{
      event.StepStarted{
        Type: event.TypeStepStarted,
        Job:  parser.currentJob,
        Step: step,
      },
    }
  }

  if step, status, durationMs, ok := matchStepFinished(message); ok {
    parser.currentStep = step
    return []any{
      event.StepFinished{
        Type:       event.TypeStepFinished,
        Job:        parser.currentJob,
        Step:       step,
        Status:     status,
        DurationMs: durationMs,
      },
    }
  }

  if parser.currentStep == "" {
    return nil
  }

  return []any{
    event.Log{
      Type: event.TypeLog,
      Job:  parser.currentJob,
      Step: parser.currentStep,
      Line: line,
    },
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
