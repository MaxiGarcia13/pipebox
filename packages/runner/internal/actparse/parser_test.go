package actparse

import (
  "testing"

  "pipebox/internal/event"
)

func TestStripANSI(t *testing.T) {
  input := "\x1b[36m[ci/build]\x1b[0m hello"
  got := StripANSI(input)
  if got != "[ci/build] hello" {
    t.Fatalf("got %q", got)
  }
}

func TestParserStepLifecycle(t *testing.T) {
  parser := NewParser()

  started := parser.HandleLine("[ci/build] ⭐ Run Main Install")
  if len(started) != 1 {
    t.Fatalf("expected 1 event, got %d", len(started))
  }
  stepStarted, ok := started[0].(event.StepStarted)
  if !ok || stepStarted.Job != "build" || stepStarted.Step != "Main Install" {
    t.Fatalf("unexpected start event: %#v", started[0])
  }

  logEvents := parser.HandleLine("[ci/build] | npm ci")
  if len(logEvents) != 1 {
    t.Fatalf("expected log event")
  }
  logEvent, ok := logEvents[0].(event.Log)
  if !ok || logEvent.Step != "Main Install" || logEvent.Line == "" {
    t.Fatalf("unexpected log: %#v", logEvents[0])
  }

  finished := parser.HandleLine("[ci/build]   ✅  Success - Main Install [1.2s]")
  stepFinished, ok := finished[0].(event.StepFinished)
  if !ok || stepFinished.Status != "passed" || stepFinished.Step != "Main Install" {
    t.Fatalf("unexpected finish: %#v", finished[0])
  }
}

func TestParserFailureAndSkip(t *testing.T) {
  parser := NewParser()
  failed := parser.HandleLine("[ci/test]   ❌  Failure - Main Test")
  stepFailed := failed[0].(event.StepFinished)
  if stepFailed.Status != "failed" {
    t.Fatalf("expected failed, got %s", stepFailed.Status)
  }

  skipped := parser.HandleLine("[ci/test]   🚧  Skipped - Main Deploy")
  stepSkipped := skipped[0].(event.StepFinished)
  if stepSkipped.Status != "skipped" {
    t.Fatalf("expected skipped, got %s", stepSkipped.Status)
  }
}
