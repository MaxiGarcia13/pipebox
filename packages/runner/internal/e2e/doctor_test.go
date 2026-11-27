//go:build e2e

package e2e_test

import (
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestDoctorEmitsDoctorEvent(t *testing.T) {
  stdout, stderr, code := runPipebox(t, "doctor", "--repo", repoPath(t, "mixed"))

  doctor := decodeLine[event.Doctor](t, stdout)
  if doctor.Type != event.TypeDoctor {
    t.Fatalf("type: got %q", doctor.Type)
  }
  if doctor.Version == "" {
    t.Fatal("expected version")
  }
  if len(doctor.Checks) == 0 {
    t.Fatal("expected checks")
  }
  foundDocker := false
  daemonOK := false
  foundActImage := false
  foundJenkinsImage := false
  for _, check := range doctor.Checks {
    switch check.Name {
    case "docker":
      foundDocker = true
    case "docker_daemon":
      daemonOK = check.OK
    case "act_image":
      foundActImage = true
    case "jenkins_image":
      foundJenkinsImage = true
    }
  }
  if !foundDocker {
    t.Fatalf("expected docker check, got %#v", doctor.Checks)
  }
  if daemonOK && !foundActImage {
    t.Fatalf("expected act_image check for mixed repo when daemon is up, got %#v", doctor.Checks)
  }
  if daemonOK && !foundJenkinsImage {
    t.Fatalf("expected jenkins_image check for mixed repo when daemon is up, got %#v", doctor.Checks)
  }
  anyFailed := false
  for _, check := range doctor.Checks {
    if !check.OK {
      anyFailed = true
    }
  }
  if doctor.OK == anyFailed {
    t.Fatalf("ok=%v does not match checks %#v", doctor.OK, doctor.Checks)
  }
  wantCode := 0
  if !doctor.OK {
    wantCode = 1
  }
  if code != wantCode {
    t.Fatalf("exit %d want %d stderr=%q stdout=%q", code, wantCode, stderr, stdout)
  }
}

func TestDoctorScopesImagesToRepoProviders(t *testing.T) {
  stdout, _, _ := runPipebox(t, "doctor", "--repo", repoPath(t, "github-only"))
  doctor := decodeLine[event.Doctor](t, stdout)
  for _, check := range doctor.Checks {
    if check.Name == "jenkins_image" {
      t.Fatalf("github-only repo should skip jenkins_image: %#v", doctor.Checks)
    }
  }

  stdout, _, _ = runPipebox(t, "doctor", "--repo", repoPath(t, "jenkins-only"))
  doctor = decodeLine[event.Doctor](t, stdout)
  for _, check := range doctor.Checks {
    if check.Name == "act_image" {
      t.Fatalf("jenkins-only repo should skip act_image: %#v", doctor.Checks)
    }
  }

  stdout, _, _ = runPipebox(t, "doctor", "--repo", repoPath(t, "empty"))
  doctor = decodeLine[event.Doctor](t, stdout)
  for _, check := range doctor.Checks {
    if check.Name == "act_image" || check.Name == "jenkins_image" {
      t.Fatalf("empty repo should skip provider images: %#v", doctor.Checks)
    }
  }
}
