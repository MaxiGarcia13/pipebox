package jfrparse

import (
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
