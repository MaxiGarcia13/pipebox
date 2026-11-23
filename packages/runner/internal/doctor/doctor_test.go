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
