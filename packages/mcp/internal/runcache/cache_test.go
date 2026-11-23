package runcache

import (
  "encoding/json"
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

  summary := cache.Finish(runID, 1)
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
  summary := cache.Finish(runID, 0)
  if summary.Status != "passed" || summary.ExitCode != 0 {
    t.Fatalf("got %#v", summary)
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
