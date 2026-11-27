//go:build e2e

package e2e_test

import (
  "os"
  "path/filepath"
  "testing"
)

func TestRunGitHubCacheAndArtifacts(t *testing.T) {
  requireHostDockerE2E(t)

  repo := repoPath(t, "github-cache-artifacts")
  // Shared workspace leftovers must not fake cache/artifact sharing.
  for _, name := range []string{".pipebox-cache-demo", "demo-artifact.txt", "downloaded"} {
    _ = os.RemoveAll(filepath.Join(repo, name))
  }

  stdout, stderr, code := runPipebox(
    t,
    "run",
    "--repo", repo,
    "--workflow", "cache-artifacts",
    "--provider", "github",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  lines := ndjsonLines(stdout)
  if !logContains(lines, "CACHE_OK") {
    t.Fatalf("missing CACHE_OK in logs; stdout=%s", stdout)
  }
  if !logContains(lines, "ARTIFACT_OK") {
    t.Fatalf("missing ARTIFACT_OK in logs; stdout=%s", stdout)
  }

  artifacts := lastArtifacts(t, lines)
  if len(artifacts.Artifacts) == 0 {
    t.Fatal("expected artifacts event")
  }
  found := false
  for _, artifact := range artifacts.Artifacts {
    if artifact.Name == "demo-out" && artifact.Size > 0 {
      found = true
      break
    }
  }
  if !found {
    t.Fatalf("expected demo-out artifact, got %#v", artifacts.Artifacts)
  }

  finished := lastRunFinished(t, lines)
  if finished.Status != "passed" || finished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", finished)
  }
}

func TestRunGitHubParallelJobsKeepLogJobTags(t *testing.T) {
  requireHostDockerE2E(t)

  stdout, stderr, code := runPipebox(
    t,
    "run",
    "--repo", repoPath(t, "github-parallel-jobs"),
    "--workflow", "parallel",
    "--provider", "github",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  lines := ndjsonLines(stdout)
  assertJobMarkerLogs(t, lines, "alpha", "ALPHA_")
  assertJobMarkerLogs(t, lines, "beta", "BETA_")

  finished := lastRunFinished(t, lines)
  if finished.Status != "passed" || finished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", finished)
  }
}
