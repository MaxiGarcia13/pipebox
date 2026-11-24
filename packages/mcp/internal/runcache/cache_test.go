package runcache

import (
  "encoding/json"
  "errors"
  "fmt"
  "testing"
)

func TestCacheFailedStepsAndTail(t *testing.T) {
  cache := New()
  runID := cache.Begin("/repo", "ci.yml", "")

  mustIngest(t, cache, runID, map[string]any{
    "type": "run_started", "repo": "/repo", "workflow": "ci.yml",
  })
  mustIngest(t, cache, runID, map[string]any{
    "type": "step_finished", "job": "build", "step": "lint", "status": "passed",
  })
  mustIngest(t, cache, runID, map[string]any{
    "type": "step_finished", "job": "build", "step": "test",
    "status": "failed", "durationMs": 42,
  })
  for i := 0; i < 5; i++ {
    mustIngest(t, cache, runID, map[string]any{
      "type": "log", "job": "build", "step": "test", "line": line(i),
    })
  }
  mustIngest(t, cache, runID, map[string]any{
    "type": "run_finished", "status": "failed", "exitCode": 1,
  })
  mustIngest(t, cache, runID, map[string]any{
    "type": "artifacts",
    "artifacts": []map[string]any{
      {"name": "a", "path": "/a", "size": 1},
      {"name": "b", "path": "/b", "size": 2},
    },
  })

  summary := cache.Finish(runID, 1, nil)
  if summary.Status != "failed" || summary.ExitCode != 1 {
    t.Fatalf("unexpected summary: %#v", summary)
  }
  if len(summary.FailedSteps) != 1 {
    t.Fatalf("failed steps: %#v", summary.FailedSteps)
  }
  if summary.FailedSteps[0].Step != "test" || summary.FailedSteps[0].DurationMs != 42 {
    t.Fatalf("failed step: %#v", summary.FailedSteps[0])
  }
  if summary.ArtifactCount != 2 {
    t.Fatalf("artifact count: %d", summary.ArtifactCount)
  }
  if len(summary.SetupLogTail) != 0 {
    t.Fatalf("setupLogTail should be empty when failedSteps present: %#v", summary.SetupLogTail)
  }

  lines, err := cache.StepLogs(runID, "build", "test", 2)
  if err != nil {
    t.Fatal(err)
  }
  if len(lines) != 2 || lines[0] != line(3) || lines[1] != line(4) {
    t.Fatalf("tail lines: %#v", lines)
  }

  latest, err := cache.Summary("")
  if err != nil {
    t.Fatal(err)
  }
  if latest.RunID != runID {
    t.Fatalf("latest run id: %s", latest.RunID)
  }
}

func TestCacheFinishInfersStatus(t *testing.T) {
  cache := New()
  runID := cache.Begin("/repo", "ci.yml", "")
  started, err := cache.Summary(runID)
  if err != nil {
    t.Fatal(err)
  }
  if started.Status != StatusRunning {
    t.Fatalf("begin status: %#v", started)
  }
  summary := cache.Finish(runID, 0, nil)
  if summary.Status != StatusPassed || summary.ExitCode != 0 {
    t.Fatalf("got %#v", summary)
  }
}

func TestCacheSkipsEvictingRunning(t *testing.T) {
  cache := New()
  cache.limit = 1
  first := cache.Begin("/repo", "a.yml", "")
  _ = cache.Begin("/repo", "b.yml", "")
  if _, err := cache.Summary(first); err != nil {
    t.Fatalf("running run should not be evicted: %v", err)
  }
}

func TestCacheErrorsAndSetupLogTail(t *testing.T) {
  cache := New()
  runID := cache.Begin("/repo", "ci", "")

  mustIngest(t, cache, runID, map[string]any{
    "type": "run_started", "repo": "/repo", "workflow": ".github/workflows/ci.yml",
  })
  for i := 0; i < 25; i++ {
    mustIngest(t, cache, runID, map[string]any{
      "type": "log", "line": fmt.Sprintf("setup-%d", i),
    })
  }
  mustIngest(t, cache, runID, map[string]any{
    "type": "error", "message": "act image: build failed",
  })
  mustIngest(t, cache, runID, map[string]any{
    "type": "error", "message": "act image: build failed",
  })
  mustIngest(t, cache, runID, map[string]any{
    "type": "run_finished", "status": "failed", "exitCode": 1,
  })

  summary := cache.Finish(runID, 1, errors.New("act image: build failed"))
  if summary.Status != "failed" {
    t.Fatalf("status: %q", summary.Status)
  }
  if len(summary.FailedSteps) != 0 {
    t.Fatalf("failedSteps: %#v", summary.FailedSteps)
  }
  if len(summary.Errors) != 1 || summary.Errors[0] != "act image: build failed" {
    t.Fatalf("errors: %#v", summary.Errors)
  }
  if len(summary.SetupLogTail) != setupLogTailLines {
    t.Fatalf("setupLogTail len: %d", len(summary.SetupLogTail))
  }
  if summary.SetupLogTail[0] != "setup-5" || summary.SetupLogTail[19] != "setup-24" {
    t.Fatalf("setupLogTail: %#v", summary.SetupLogTail)
  }
  if summary.Workflow != ".github/workflows/ci.yml" {
    t.Fatalf("workflow: %q", summary.Workflow)
  }

  lines, err := cache.StepLogs(runID, "", "", 3)
  if err != nil {
    t.Fatal(err)
  }
  if len(lines) != 3 || lines[0] != "setup-22" {
    t.Fatalf("setup step logs: %#v", lines)
  }
}

func TestCacheFinishAppendsRunErr(t *testing.T) {
  cache := New()
  runID := cache.Begin("/repo", "ci.yml", "")
  summary := cache.Finish(runID, 1, errors.New("spawn failed"))
  if summary.Status != "failed" {
    t.Fatalf("status: %q", summary.Status)
  }
  if len(summary.Errors) != 1 || summary.Errors[0] != "spawn failed" {
    t.Fatalf("errors: %#v", summary.Errors)
  }
}

func mustIngest(t *testing.T, cache *Cache, runID string, payload map[string]any) {
  t.Helper()
  raw, err := json.Marshal(payload)
  if err != nil {
    t.Fatal(err)
  }
  if err := cache.Ingest(runID, raw); err != nil {
    t.Fatal(err)
  }
}

func line(i int) string {
  return fmt.Sprintf("log-line-%d", i)
}
