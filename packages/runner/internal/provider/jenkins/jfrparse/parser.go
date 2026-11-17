package jfrparse

import (
  "regexp"
  "strings"
  "unicode"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/ansi"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  pipelinePrefixPattern = regexp.MustCompile(`^\[Pipeline\]\s*(.*)$`)
  stageOpenPattern      = regexp.MustCompile(`^\{\s*\((.+)\)\s*$`)
  finishedPattern       = regexp.MustCompile(`(?i)^Finished:\s+(SUCCESS|FAILURE|ABORTED|UNSTABLE)\s*$`)
)

var metaPipelineTokens = map[string]struct{}{
  "stage": {},
  "node":  {},
  "parallel": {},
  "ws":    {},
  "dir":   {},
  "withEnv": {},
  "withCredentials": {},
  "script": {},
  "catchError": {},
  "retry": {},
  "timeout": {},
  "waitUntil": {},
  "End of Pipeline": {},
}

type Parser struct {
  currentJob  string
  currentStep string
}

func NewParser() *Parser {
  return &Parser{}
}

func (parser *Parser) HandleLine(rawLine string) []any {
  line := strings.TrimSpace(ansi.Strip(rawLine))
  if line == "" {
    return nil
  }

  if finishedPattern.MatchString(line) {
    return parser.finishOpenStep()
  }

  rest, ok := matchPipelinePrefix(line)
  if !ok {
    if parser.currentJob != "" && parser.currentStep != "" {
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

  if stageName, stageOK := matchStageOpen(rest); stageOK {
    events := parser.finishOpenStep()
    parser.currentJob = slugify(stageName)
    parser.currentStep = ""
    return events
  }

  if rest == "}" || strings.HasPrefix(rest, "//") {
    return parser.finishOpenStep()
  }

  if isMetaPipelineToken(rest) {
    return nil
  }

  events := parser.finishOpenStep()
  parser.currentStep = rest
  if parser.currentJob == "" {
    parser.currentJob = "pipeline"
  }
  events = append(events, event.StepStarted{
    Type: event.TypeStepStarted,
    Job:  parser.currentJob,
    Step: parser.currentStep,
  })
  return events
}

func (parser *Parser) finishOpenStep() []any {
  if parser.currentJob == "" || parser.currentStep == "" {
    return nil
  }
  step := parser.currentStep
  parser.currentStep = ""
  return []any{
    event.StepFinished{
      Type:   event.TypeStepFinished,
      Job:    parser.currentJob,
      Step:   step,
      Status: "passed",
    },
  }
}

func matchPipelinePrefix(line string) (string, bool) {
  matches := pipelinePrefixPattern.FindStringSubmatch(line)
  if matches == nil {
    return "", false
  }
  return strings.TrimSpace(matches[1]), true
}

func matchStageOpen(rest string) (string, bool) {
  matches := stageOpenPattern.FindStringSubmatch(rest)
  if matches == nil {
    return "", false
  }
  name := strings.TrimSpace(matches[1])
  name = strings.TrimPrefix(name, "Branch: ")
  return strings.TrimSpace(name), name != ""
}

func isMetaPipelineToken(rest string) bool {
  if _, ok := metaPipelineTokens[rest]; ok {
    return true
  }
  if strings.HasPrefix(rest, "//") {
    return true
  }
  return false
}

func slugify(name string) string {
  var builder strings.Builder
  lastDash := false
  for _, char := range strings.ToLower(name) {
    if unicode.IsLetter(char) || unicode.IsDigit(char) {
      builder.WriteRune(char)
      lastDash = false
      continue
    }
    if !lastDash && builder.Len() > 0 {
      builder.WriteByte('-')
      lastDash = true
    }
  }
  return strings.Trim(builder.String(), "-")
}
