package jfrparse

import (
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestParserStageAndStepLifecycle(t *testing.T) {
  parser := NewParser()
  lines := []string{
    "[Pipeline] stage",
    "[Pipeline] { (Build)",
    "[Pipeline] echo",
    "hello",
    "[Pipeline] }",
    "[Pipeline] // stage",
    "Finished: SUCCESS",
  }

  events := make([]any, 0)
  for _, line := range lines {
    events = append(events, parser.HandleLine(line)...)
  }

  if len(events) != 3 {
    t.Fatalf("events: %#v", events)
  }
  started, ok := events[0].(event.StepStarted)
  if !ok || started.Job != "build" || started.Step != "echo" {
    t.Fatalf("started: %#v", events[0])
  }
  logEvent, ok := events[1].(event.Log)
  if !ok || logEvent.Line != "hello" || logEvent.Job != "build" {
    t.Fatalf("log: %#v", events[1])
  }
  finished, ok := events[2].(event.StepFinished)
  if !ok || finished.Status != "passed" || finished.Step != "echo" {
    t.Fatalf("finished: %#v", events[2])
  }
}

func TestParserBranchStageName(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[Pipeline] { (Branch: Lint)")
  events := parser.HandleLine("[Pipeline] sh")
  if len(events) != 1 {
    t.Fatalf("events: %#v", events)
  }
  started := events[0].(event.StepStarted)
  if started.Job != "lint" {
    t.Fatalf("job: %#v", started)
  }
}

func TestParserUnscopedLauncherError(t *testing.T) {
  parser := NewParser()
  events := parser.HandleLine("option '--runWorkspace' (<runWorkspace>) should be specified only once")
  if len(events) != 1 {
    t.Fatalf("events: %#v", events)
  }
  logEvent, ok := events[0].(event.Log)
  if !ok || logEvent.Job != "" || logEvent.Step != "" {
    t.Fatalf("log: %#v", events[0])
  }
  if !strings.Contains(logEvent.Line, "runWorkspace") {
    t.Fatalf("line: %#v", logEvent)
  }
}

func TestParserDropsJenkinsBootstrapNoise(t *testing.T) {
  parser := NewParser()
  line := "2026-11-17 19:06:27.016+0000 [id=48]\tWARNING\thudson.ClassicPluginStrategy#createPluginWrapper: x"
  if events := parser.HandleLine(line); len(events) != 0 {
    t.Fatalf("events: %#v", events)
  }
}

func TestParserStripsConsoleNotes(t *testing.T) {
  parser := NewParser()
  note := "ha:////4IWeQcaIeaGOposT5OHXzewD3gt7bYixNseO0a3GDNpKAAAAoh+LCAAAAAAAAP9tjTEOwjAQBM8BClpKHuFItIiK1krDC0x8GCfWnbEdkooX8TX+gCESFVvtrLSa5wtWKcKBo5UdUu8otU4GP9jS5Mixv3geZcdn2TIl9igbHBs2eJyx4YwwR1SwULBGaj0nRzbDRnX6rmuvydanHMu2V1A5c4MHCFXMWcf8hSnC9jqYxPTz/BXAFEIGsfuclm8zQVqFvQAAAA=="
  // After ANSI strip, ConsoleNote payload sticks to [Pipeline] and used to break step detection.
  lines := []string{
    "\x1b[8m" + note + "\x1b[0m[Pipeline] { (test)",
    "\x1b[8m" + note + "\x1b[0m[Pipeline] sh",
    "+ pnpm test",
    "\x1b[8m" + note + "\x1b[0m[Pipeline] }",
  }

  events := make([]any, 0)
  for _, line := range lines {
    events = append(events, parser.HandleLine(line)...)
  }

  if len(events) != 3 {
    t.Fatalf("events: %#v", events)
  }
  started, ok := events[0].(event.StepStarted)
  if !ok || started.Job != "test" || started.Step != "sh" {
    t.Fatalf("started: %#v", events[0])
  }
  logEvent, ok := events[1].(event.Log)
  if !ok || logEvent.Line != "+ pnpm test" || logEvent.Job != "test" || logEvent.Step != "sh" {
    t.Fatalf("log: %#v", events[1])
  }
  finished, ok := events[2].(event.StepFinished)
  if !ok || finished.Status != "passed" || finished.Step != "sh" {
    t.Fatalf("finished: %#v", events[2])
  }
}

func TestParserFinishedFailureMarksStepFailed(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[Pipeline] { (lint)")
  _ = parser.HandleLine("[Pipeline] sh")
  events := parser.HandleLine("Finished: FAILURE")
  if len(events) != 1 {
    t.Fatalf("events: %#v", events)
  }
  finished, ok := events[0].(event.StepFinished)
  if !ok || finished.Status != "failed" || finished.Job != "lint" || finished.Step != "sh" {
    t.Fatalf("finished: %#v", events[0])
  }
}

func TestParserFinishedFailureCorrectsClosedStep(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[Pipeline] { (test)")
  _ = parser.HandleLine("[Pipeline] sh")
  closed := parser.HandleLine("[Pipeline] }")
  if len(closed) != 1 {
    t.Fatalf("closed: %#v", closed)
  }
  if finished := closed[0].(event.StepFinished); finished.Status != "passed" {
    t.Fatalf("closed status: %#v", finished)
  }
  events := parser.HandleLine("Finished: FAILURE")
  if len(events) != 1 {
    t.Fatalf("events: %#v", events)
  }
  finished, ok := events[0].(event.StepFinished)
  if !ok || finished.Status != "failed" || finished.Job != "test" || finished.Step != "sh" {
    t.Fatalf("finished: %#v", events[0])
  }
}

func TestParserDropsStandaloneConsoleNote(t *testing.T) {
  parser := NewParser()
  line := "\x1b[8mha:////AAAA==\x1b[0m"
  if events := parser.HandleLine(line); len(events) != 0 {
    t.Fatalf("events: %#v", events)
  }
}

func TestParserIgnoresGetContextAndAnonymousBlocks(t *testing.T) {
  parser := NewParser()
  lines := []string{
    "[Pipeline] { (lint)",
    "[Pipeline] sh",
    "+ pnpm lint",
    "[Pipeline] }",
    "[Pipeline] getContext",
    "[Pipeline] {",
    "[Pipeline] }",
    "ERROR: script returned exit code 1",
  }
  events := make([]any, 0)
  for _, line := range lines {
    events = append(events, parser.HandleLine(line)...)
  }

  for _, item := range events {
    if started, ok := item.(event.StepStarted); ok && started.Step == "getContext" {
      t.Fatalf("getContext should not be a step: %#v", events)
    }
    if started, ok := item.(event.StepStarted); ok && started.Step == "{" {
      t.Fatalf("anonymous block should not be a step: %#v", events)
    }
  }

  var sawFailedSh bool
  for _, item := range events {
    finished, ok := item.(event.StepFinished)
    if ok && finished.Step == "sh" && finished.Status == "failed" {
      sawFailedSh = true
    }
  }
  if !sawFailedSh {
    t.Fatalf("expected sh failed after script error: %#v", events)
  }
}

func TestParserParallelBranchesSeparateSteps(t *testing.T) {
  parser := NewParser()
  lines := []string{
    "[Pipeline] parallel",
    "[Pipeline] { (Branch: test)",
    "[Pipeline] { (Branch: lint)",
    "[Pipeline] stage",
    "[Pipeline] { (test)",
    "[Pipeline] stage",
    "[Pipeline] { (lint)",
    "[Pipeline] sh",
    "[Pipeline] sh",
    "+ pnpm test",
    "+ pnpm lint",
    "[Pipeline] }",
    "Failed in branch lint",
    "[Pipeline] }",
    "Failed in branch test",
    "Finished: FAILURE",
  }

  events := make([]any, 0)
  for _, line := range lines {
    events = append(events, parser.HandleLine(line)...)
  }

  startedJobs := map[string]string{}
  finished := map[string]string{}
  logsByJob := map[string][]string{}
  for _, item := range events {
    switch value := item.(type) {
    case event.StepStarted:
      startedJobs[value.Job] = value.Step
    case event.StepFinished:
      finished[value.Job] = value.Status
    case event.Log:
      if value.Job != "" {
        logsByJob[value.Job] = append(logsByJob[value.Job], value.Line)
      }
    }
  }

  if startedJobs["test"] != "sh" || startedJobs["lint"] != "sh" {
    t.Fatalf("started: %#v", startedJobs)
  }
  if finished["test"] != "failed" || finished["lint"] != "failed" {
    t.Fatalf("finished: %#v", finished)
  }
  if !containsLine(logsByJob["test"], "+ pnpm test") && !containsLine(logsByJob["lint"], "+ pnpm test") {
    t.Fatalf("expected test command in a branch log: %#v", logsByJob)
  }
  if !containsLine(logsByJob["lint"], "+ pnpm lint") && !containsLine(logsByJob["test"], "+ pnpm lint") {
    t.Fatalf("expected lint command in a branch log: %#v", logsByJob)
  }
}

func containsLine(lines []string, want string) bool {
  for _, line := range lines {
    if line == want {
      return true
    }
  }
  return false
}
