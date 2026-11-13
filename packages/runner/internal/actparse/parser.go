package actparse

import (
  "regexp"
  "strings"

  "pipebox/internal/event"
)

var (
  jobPrefixPattern = regexp.MustCompile(`^\[([^/\]]+)/([^\]]+)\]\s*(.*)$`)
  stepStartPattern = regexp.MustCompile(`(?i)(?:⭐\s*)?Run\s+(.+)$`)
  stepSuccessPattern = regexp.MustCompile(`(?i)(?:✅\s*)?Success\s+-\s+(.+?)(?:\s+\[[^\]]+\])?\s*$`)
  stepFailurePattern = regexp.MustCompile(`(?i)(?:❌\s*)?Failure\s+-\s+(.+?)(?:\s+\[[^\]]+\])?\s*$`)
  stepSkippedPattern = regexp.MustCompile(`(?i)(?:🚧\s*)?Skipped\s+-\s+(.+?)(?:\s+\[[^\]]+\])?\s*$`)
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

  job, rest := splitJobPrefix(line)
  if job != "" {
    parser.currentJob = job
  }
  message := rest
  if message == "" {
    message = line
  }

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

  if step, status, ok := matchStepFinished(message); ok {
    parser.currentStep = step
    return []any{
      event.StepFinished{
        Type:   event.TypeStepFinished,
        Job:    parser.currentJob,
        Step:   step,
        Status: status,
      },
    }
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

func splitJobPrefix(line string) (job string, rest string) {
  matches := jobPrefixPattern.FindStringSubmatch(line)
  if matches == nil {
    return "", line
  }
  return matches[2], matches[3]
}

func matchStepStart(message string) (string, bool) {
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
  if step == "" {
    return "", false
  }
  return step, true
}

func matchStepFinished(message string) (step string, status string, ok bool) {
  if matches := stepSuccessPattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "passed", true
  }
  if matches := stepFailurePattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "failed", true
  }
  if matches := stepSkippedPattern.FindStringSubmatch(message); matches != nil {
    return strings.TrimSpace(matches[1]), "skipped", true
  }
  return "", "", false
}
