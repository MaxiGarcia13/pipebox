package jfrparse

import (
  "regexp"
  "sort"
  "strings"
  "unicode"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/ansi"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  pipelinePrefixPattern = regexp.MustCompile(`^\[Pipeline\]\s*(.*)$`)
  stageOpenPattern      = regexp.MustCompile(`^\{\s*\((.+)\)\s*$`)
  finishedPattern       = regexp.MustCompile(`(?i)^Finished:\s+(SUCCESS|FAILURE|ABORTED|UNSTABLE)\s*$`)
  failedBranchPattern   = regexp.MustCompile(`(?i)^Failed in branch\s+(.+)$`)
  scriptErrorPattern    = regexp.MustCompile(`(?i)^ERROR:\s+script returned exit code\b`)
  jenkinsLoggerPattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} .*\[id=\d+\]`)
  // Jenkins ConsoleNote: ESC[8mha:////<base64>ESC[0m — after ansi.Strip the payload remains.
  consoleNotePattern = regexp.MustCompile(`ha:////[A-Za-z0-9+/=]+`)
)

var metaPipelineTokens = map[string]struct{}{
  "stage":              {},
  "node":               {},
  "parallel":           {},
  "ws":                 {},
  "dir":                {},
  "withEnv":            {},
  "withCredentials":    {},
  "script":             {},
  "catchError":         {},
  "retry":              {},
  "timeout":            {},
  "waitUntil":          {},
  "getContext":         {},
  "isUnix":             {},
  "Start of Pipeline":  {},
  "End of Pipeline":    {},
}

type branchState struct {
  step string
}

type Parser struct {
  branches map[string]*branchState
  // active owns following plain log lines; "}" closes its step when present.
  active   string
  pending  []string
  lastJob  string
  lastStep string
}

func NewParser() *Parser {
  return &Parser{
    branches: map[string]*branchState{},
  }
}

func (parser *Parser) HandleLine(rawLine string) []any {
  line := normalizeConsoleLine(rawLine)
  if line == "" {
    return nil
  }

  if matches := finishedPattern.FindStringSubmatch(line); matches != nil {
    return parser.handleFinished(finishedStatus(matches[1]))
  }

  if matches := failedBranchPattern.FindStringSubmatch(line); matches != nil {
    return parser.handleFailedBranch(slugify(matches[1]), line)
  }

  if scriptErrorPattern.MatchString(line) {
    return parser.handleScriptError(line)
  }

  rest, ok := matchPipelinePrefix(line)
  if !ok {
    return parser.handlePlainLine(line)
  }

  if stageName, stageOK := matchStageOpen(rest); stageOK {
    parser.openStage(slugify(stageName))
    return nil
  }

  if rest == "}" {
    return parser.finishBranchStep(parser.closeJob(), "passed")
  }

  // Anonymous block open — not a user step.
  if rest == "{" {
    return nil
  }

  if strings.HasPrefix(rest, "//") {
    return nil
  }

  if isMetaPipelineToken(rest) {
    return nil
  }

  return parser.startStep(rest)
}

func (parser *Parser) handleFinished(status string) []any {
  events := parser.finishAllOpenSteps(status)
  // Jenkins closes steps with "}" before "Finished: FAILURE"; correct the last step.
  if status != "passed" && len(events) == 0 && parser.lastJob != "" && parser.lastStep != "" {
    events = append(events, event.StepFinished{
      Type:   event.TypeStepFinished,
      Job:    parser.lastJob,
      Step:   parser.lastStep,
      Status: status,
    })
  }
  return events
}

func (parser *Parser) handleFailedBranch(job, line string) []any {
  events := parser.failJobStep(job)
  step := parser.stepLabelFor(job, events)
  parser.active = job
  return append(events, event.Log{
    Type: event.TypeLog,
    Job:  job,
    Step: step,
    Line: line,
  })
}

func (parser *Parser) handleScriptError(line string) []any {
  job := parser.closeJob()
  if job == "" {
    job = parser.lastJob
  }
  events := parser.failJobStep(job)
  step := parser.stepLabelFor(job, events)
  logEvent := event.Log{
    Type: event.TypeLog,
    Line: line,
  }
  if job != "" {
    logEvent.Job = job
    logEvent.Step = step
  }
  return append(events, logEvent)
}

func (parser *Parser) failJobStep(job string) []any {
  events := parser.finishBranchStep(job, "failed")
  if len(events) > 0 {
    return events
  }
  if job != "" && parser.lastJob == job && parser.lastStep != "" {
    return []any{
      event.StepFinished{
        Type:   event.TypeStepFinished,
        Job:    parser.lastJob,
        Step:   parser.lastStep,
        Status: "failed",
      },
    }
  }
  return nil
}

func (parser *Parser) stepLabelFor(job string, events []any) string {
  if finished, ok := eventsFirstFinished(events); ok {
    return finished.Step
  }
  if step := parser.stepForLog(job); step != "" {
    return step
  }
  if parser.lastJob == job {
    return parser.lastStep
  }
  return ""
}

func (parser *Parser) handlePlainLine(line string) []any {
  job := parser.logJob()
  if job != "" {
    step := parser.stepForLog(job)
    if step != "" {
      return []any{
        event.Log{
          Type: event.TypeLog,
          Job:  job,
          Step: step,
          Line: line,
        },
      }
    }
  }

  if jenkinsLoggerPattern.MatchString(line) {
    return nil
  }
  return []any{
    event.Log{
      Type: event.TypeLog,
      Line: line,
    },
  }
}

func (parser *Parser) openStage(job string) {
  if job == "" {
    return
  }
  if _, ok := parser.branches[job]; !ok {
    parser.branches[job] = &branchState{}
    parser.pending = append(parser.pending, job)
  }
  parser.active = job
}

func (parser *Parser) startStep(stepName string) []any {
  job := parser.pickStepJob()
  if job == "" {
    job = "pipeline"
    parser.openStage(job)
  }

  events := parser.finishBranchStep(job, "passed")
  branch := parser.branches[job]
  branch.step = stepName
  parser.active = job
  parser.removePending(job)

  events = append(events, event.StepStarted{
    Type: event.TypeStepStarted,
    Job:  job,
    Step: stepName,
  })
  return events
}

func (parser *Parser) pickStepJob() string {
  // Prefer an open branch that does not already have a running step (parallel sh/sh).
  for _, job := range parser.pending {
    branch := parser.branches[job]
    if branch != nil && branch.step == "" {
      return job
    }
  }
  for job, branch := range parser.branches {
    if branch != nil && branch.step == "" {
      return job
    }
  }
  if parser.active != "" {
    return parser.active
  }
  return ""
}

func (parser *Parser) closeJob() string {
  if parser.active != "" {
    branch := parser.branches[parser.active]
    if branch != nil && branch.step != "" {
      return parser.active
    }
  }
  return parser.soleOpenStepJob()
}

func (parser *Parser) soleOpenStepJob() string {
  var found string
  for job, branch := range parser.branches {
    if branch == nil || branch.step == "" {
      continue
    }
    if found != "" {
      return ""
    }
    found = job
  }
  return found
}

func (parser *Parser) finishBranchStep(job, status string) []any {
  if job == "" {
    return nil
  }
  branch := parser.branches[job]
  if branch == nil || branch.step == "" {
    return nil
  }
  step := branch.step
  branch.step = ""
  parser.lastJob = job
  parser.lastStep = step
  parser.retargetActive()
  return []any{
    event.StepFinished{
      Type:   event.TypeStepFinished,
      Job:    job,
      Step:   step,
      Status: status,
    },
  }
}

func (parser *Parser) finishAllOpenSteps(status string) []any {
  jobs := make([]string, 0, len(parser.branches))
  for job := range parser.branches {
    jobs = append(jobs, job)
  }
  sort.Strings(jobs)
  events := make([]any, 0)
  for _, job := range jobs {
    events = append(events, parser.finishBranchStep(job, status)...)
  }
  return events
}

func (parser *Parser) retargetActive() {
  if parser.active != "" {
    branch := parser.branches[parser.active]
    if branch != nil && branch.step != "" {
      return
    }
  }
  if job := parser.soleOpenStepJob(); job != "" {
    parser.active = job
  }
}

func (parser *Parser) logJob() string {
  if parser.active != "" {
    branch := parser.branches[parser.active]
    if branch != nil && branch.step != "" {
      return parser.active
    }
  }
  if job := parser.soleOpenStepJob(); job != "" {
    return job
  }
  return parser.active
}

func (parser *Parser) stepForLog(job string) string {
  branch := parser.branches[job]
  if branch == nil {
    return ""
  }
  if branch.step != "" {
    return branch.step
  }
  if parser.lastJob == job {
    return parser.lastStep
  }
  return ""
}

func (parser *Parser) removePending(job string) {
  filtered := parser.pending[:0]
  for _, candidate := range parser.pending {
    if candidate != job {
      filtered = append(filtered, candidate)
    }
  }
  parser.pending = filtered
}

func eventsFirstFinished(events []any) (event.StepFinished, bool) {
  for _, item := range events {
    if finished, ok := item.(event.StepFinished); ok {
      return finished, true
    }
  }
  return event.StepFinished{}, false
}

func normalizeConsoleLine(rawLine string) string {
  line := ansi.Strip(rawLine)
  line = consoleNotePattern.ReplaceAllString(line, "")
  return strings.TrimSpace(line)
}

func finishedStatus(result string) string {
  switch strings.ToUpper(strings.TrimSpace(result)) {
  case "SUCCESS":
    return "passed"
  case "ABORTED":
    return "cancelled"
  default:
    return "failed"
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
