package doctor

import (
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestAllChecksOK(t *testing.T) {
  t.Parallel()

  if allChecksOK(nil) {
    t.Fatal("empty checks should fail")
  }

  passed := []event.DoctorCheck{
    {Name: "docker", OK: true},
    {Name: "act_image", OK: true},
  }
  if !allChecksOK(passed) {
    t.Fatal("expected ok")
  }

  missingImage := []event.DoctorCheck{
    {Name: "docker", OK: true},
    {Name: "docker_daemon", OK: true},
    {Name: "act_image", OK: false},
  }
  if allChecksOK(missingImage) {
    t.Fatal("a missing image should fail doctor")
  }
}

func TestHasProvider(t *testing.T) {
  t.Parallel()

  if hasProvider(nil, event.ProviderGitHub) {
    t.Fatal("nil providers should not match")
  }
  if hasProvider([]string{}, event.ProviderGitHub) {
    t.Fatal("empty providers should not match")
  }
  if !hasProvider([]string{event.ProviderGitHub}, event.ProviderGitHub) {
    t.Fatal("expected github match")
  }
  if hasProvider([]string{event.ProviderGitHub}, event.ProviderJenkins) {
    t.Fatal("github-only should not match jenkins")
  }
}

func TestRunSkipsIrrelevantImages(t *testing.T) {
  t.Parallel()

  none := Run("test", nil)
  assertNoCheck(t, none, "act_image")
  assertNoCheck(t, none, "jenkins_image")

  githubOnly := Run("test", []string{event.ProviderGitHub})
  assertNoCheck(t, githubOnly, "jenkins_image")
  if daemonReachable(githubOnly) {
    assertHasCheck(t, githubOnly, "act_image")
  }

  jenkinsOnly := Run("test", []string{event.ProviderJenkins})
  assertNoCheck(t, jenkinsOnly, "act_image")
  if daemonReachable(jenkinsOnly) {
    assertHasCheck(t, jenkinsOnly, "jenkins_image")
  }
}

func daemonReachable(result event.Doctor) bool {
  for _, check := range result.Checks {
    if check.Name == "docker_daemon" {
      return check.OK
    }
  }
  return false
}

func assertHasCheck(t *testing.T, result event.Doctor, name string) {
  t.Helper()
  for _, check := range result.Checks {
    if check.Name == name {
      return
    }
  }
  t.Fatalf("expected %s in %#v", name, result.Checks)
}

func assertNoCheck(t *testing.T, result event.Doctor, name string) {
  t.Helper()
  for _, check := range result.Checks {
    if check.Name == name {
      t.Fatalf("unexpected %s in %#v", name, result.Checks)
    }
  }
}
