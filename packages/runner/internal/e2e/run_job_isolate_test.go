package e2e_test

import (
  "testing"
)

func TestRunGitHubJobIgnoresNeedsDeps(t *testing.T) {
  requireHostDockerE2E(t)

  stdout, stderr, code := runPipebox(
    t,
    "run",
    "--repo", repoPath(t, "github-needs-jobs"),
    "--workflow", "needs",
    "--provider", "github",
    "--job", "e2e",
    "--input", "fail_e2e=false",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  lines := ndjsonLines(stdout)
  assertHasJobSteps(t, lines, "e2e")
  assertNoJobSteps(t, lines, "lint", "test")
  assertNoLogMarkers(t, lines, "LINT_OK", "TEST_OK")
  if !logContains(lines, "E2E_OK") {
    t.Fatalf("missing E2E_OK; stdout=%s", stdout)
  }

  finished := lastRunFinished(t, lines)
  if finished.Status != "passed" || finished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", finished)
  }
}

func TestRunGitHubJobRerunAfterFailure(t *testing.T) {
  requireHostDockerE2E(t)

  repo := repoPath(t, "github-needs-jobs")
  failOut, failErr, failCode := runPipebox(
    t,
    "run",
    "--repo", repo,
    "--workflow", "needs",
    "--provider", "github",
    "--job", "e2e",
    "--input", "fail_e2e=true",
  )
  if failCode == 0 {
    t.Fatalf("expected failing run, got exit 0 stderr=%q stdout=%q", failErr, failOut)
  }
  failLines := ndjsonLines(failOut)
  assertHasJobSteps(t, failLines, "e2e")
  assertNoJobSteps(t, failLines, "lint", "test")
  if !logContains(failLines, "E2E_FAIL") {
    t.Fatalf("missing E2E_FAIL; stdout=%s", failOut)
  }
  failFinished := lastRunFinished(t, failLines)
  if failFinished.Status == "passed" {
    t.Fatalf("expected failed run_finished: %#v", failFinished)
  }

  passOut, passErr, passCode := runPipebox(
    t,
    "run",
    "--repo", repo,
    "--workflow", "needs",
    "--provider", "github",
    "--job", "e2e",
    "--input", "fail_e2e=false",
  )
  if passCode != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", passCode, passErr, passOut)
  }
  passLines := ndjsonLines(passOut)
  assertHasJobSteps(t, passLines, "e2e")
  assertNoJobSteps(t, passLines, "lint", "test")
  assertNoLogMarkers(t, passLines, "LINT_OK", "TEST_OK")
  if !logContains(passLines, "E2E_OK") {
    t.Fatalf("missing E2E_OK; stdout=%s", passOut)
  }
  passFinished := lastRunFinished(t, passLines)
  if passFinished.Status != "passed" || passFinished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", passFinished)
  }
}

func TestRunJenkinsJobIgnoresPriorStages(t *testing.T) {
  requireHostDockerE2E(t)

  stdout, stderr, code := runPipebox(
    t,
    "run",
    "--repo", repoPath(t, "jenkins-needs-stages"),
    "--workflow", "Jenkinsfile",
    "--provider", "jenkins",
    "--job", "e2e",
    "--input", "FAIL_E2E=false",
  )
  if code != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
  }

  lines := ndjsonLines(stdout)
  assertHasJobSteps(t, lines, "e2e")
  assertNoJobSteps(t, lines, "lint", "test")
  assertNoLogMarkers(t, lines, "LINT_OK", "TEST_OK")
  if !logContains(lines, "E2E_OK") {
    t.Fatalf("missing E2E_OK; stdout=%s", stdout)
  }

  finished := lastRunFinished(t, lines)
  if finished.Status != "passed" || finished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", finished)
  }
}

func TestRunJenkinsJobRerunAfterFailure(t *testing.T) {
  requireHostDockerE2E(t)

  repo := repoPath(t, "jenkins-needs-stages")
  failOut, failErr, failCode := runPipebox(
    t,
    "run",
    "--repo", repo,
    "--workflow", "Jenkinsfile",
    "--provider", "jenkins",
    "--job", "e2e",
    "--input", "FAIL_E2E=true",
  )
  if failCode == 0 {
    t.Fatalf("expected failing run, got exit 0 stderr=%q stdout=%q", failErr, failOut)
  }
  failLines := ndjsonLines(failOut)
  assertHasJobSteps(t, failLines, "e2e")
  assertNoJobSteps(t, failLines, "lint", "test")
  if !logContains(failLines, "E2E_FAIL") {
    t.Fatalf("missing E2E_FAIL; stdout=%s", failOut)
  }
  failFinished := lastRunFinished(t, failLines)
  if failFinished.Status == "passed" {
    t.Fatalf("expected failed run_finished: %#v", failFinished)
  }

  passOut, passErr, passCode := runPipebox(
    t,
    "run",
    "--repo", repo,
    "--workflow", "Jenkinsfile",
    "--provider", "jenkins",
    "--job", "e2e",
    "--input", "FAIL_E2E=false",
  )
  if passCode != 0 {
    t.Fatalf("exit %d stderr=%q stdout=%q", passCode, passErr, passOut)
  }
  passLines := ndjsonLines(passOut)
  assertHasJobSteps(t, passLines, "e2e")
  assertNoJobSteps(t, passLines, "lint", "test")
  assertNoLogMarkers(t, passLines, "LINT_OK", "TEST_OK")
  if !logContains(passLines, "E2E_OK") {
    t.Fatalf("missing E2E_OK; stdout=%s", passOut)
  }
  passFinished := lastRunFinished(t, passLines)
  if passFinished.Status != "passed" || passFinished.ExitCode != 0 {
    t.Fatalf("run_finished: %#v", passFinished)
  }
}
