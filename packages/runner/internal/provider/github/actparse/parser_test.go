package actparse

import (
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
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
  if stepFinished.DurationMs != 1200 {
    t.Fatalf("duration: got %d want 1200", stepFinished.DurationMs)
  }
}

func TestParserIgnoresNpmRunHints(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[ci/build] ⭐ Run Main Install")

  for _, line := range []string{
    "[ci/build] | Run `npm audit` for details.",
    "[ci/build] | Run `npm install-scripts ls` to review.",
    "[ci/build] Run `npm fund` for details.",
  } {
    events := parser.HandleLine(line)
    if len(events) != 1 {
      t.Fatalf("%q: expected 1 event, got %d", line, len(events))
    }
    if _, ok := events[0].(event.Log); !ok {
      t.Fatalf("%q: expected log, got %#v", line, events[0])
    }
  }
}

func TestParserTrimsPaddedJobName(t *testing.T) {
  parser := NewParser()

  started := parser.HandleLine("[Validation/Lint     ] ⭐ Run Main actions/checkout@v4")
  if len(started) != 1 {
    t.Fatalf("expected 1 event, got %d", len(started))
  }
  stepStarted, ok := started[0].(event.StepStarted)
  if !ok || stepStarted.Job != "Lint" {
    t.Fatalf("expected job Lint, got %#v", started[0])
  }

  finished := parser.HandleLine("[Validation/Lint     ]   ✅  Success - Main actions/checkout@v4 [1s]")
  stepFinished, ok := finished[0].(event.StepFinished)
  if !ok || stepFinished.Job != "Lint" || stepFinished.Status != "passed" {
    t.Fatalf("unexpected finish: %#v", finished[0])
  }
}

func TestParserFailureAndSkip(t *testing.T) {
  parser := NewParser()
  failed := parser.HandleLine("[ci/test]   ❌  Failure - Main Test [500ms]")
  stepFailed := failed[0].(event.StepFinished)
  if stepFailed.Status != "failed" {
    t.Fatalf("expected failed, got %s", stepFailed.Status)
  }
  if stepFailed.DurationMs != 500 {
    t.Fatalf("duration: got %d want 500", stepFailed.DurationMs)
  }

  skipped := parser.HandleLine("[ci/test]   🚧  Skipped - Main Deploy")
  stepSkipped := skipped[0].(event.StepFinished)
  if stepSkipped.Status != "skipped" {
    t.Fatalf("expected skipped, got %s", stepSkipped.Status)
  }
  if stepSkipped.DurationMs != 0 {
    t.Fatalf("expected no duration, got %d", stepSkipped.DurationMs)
  }
}

func TestParserGlobalJobFailedError(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[CI/test] ⭐ Run Complete job")
  _ = parser.HandleLine("[CI/test]   ✅  Success - Complete job")

  items := parser.HandleLine("Error: Job 'lint' failed")
  if len(items) != 1 {
    t.Fatalf("items: %#v", items)
  }
  errEvent, ok := items[0].(event.Error)
  if !ok || errEvent.Message != "Error: Job 'lint' failed" {
    t.Fatalf("expected global error, got %#v", items[0])
  }
}

func TestParserUnprefixedPipeContinuesActiveStep(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[CI/lint] ⭐ Run Main Install")

  items := parser.HandleLine("| Packages: +578")
  logEvent, ok := items[0].(event.Log)
  if !ok || logEvent.Job != "lint" || logEvent.Step != "Main Install" {
    t.Fatalf("expected continuation log, got %#v", items)
  }

  if dropped := parser.HandleLine("Downloading something"); len(dropped) != 0 {
    t.Fatalf("expected drop non-pipe unprefixed, got %#v", dropped)
  }
}

func TestParserClearsStepOnJobSwitch(t *testing.T) {
  parser := NewParser()
  _ = parser.HandleLine("[CI/test] ⭐ Run Main Install")
  _ = parser.HandleLine("[CI/test] | npm ci")

  // Prefixed lint line without a new Run must not keep test's Install step.
  if items := parser.HandleLine("[CI/lint] | stray"); len(items) != 0 {
    t.Fatalf("expected drop until lint step starts, got %#v", items)
  }

  started := parser.HandleLine("[CI/lint] ⭐ Run Main Install")
  if len(started) != 1 {
    t.Fatalf("started: %#v", started)
  }
  logEvents := parser.HandleLine("[CI/lint] | pnpm install")
  logEvent, ok := logEvents[0].(event.Log)
  if !ok || logEvent.Job != "lint" || logEvent.Step != "Main Install" {
    t.Fatalf("log: %#v", logEvents)
  }
}
