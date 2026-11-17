package doctor

import (
  "os/exec"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actimage"
)

func Run(version string) event.Doctor {
  checks := []event.DoctorCheck{
    lookPathCheck("docker"),
  }

  ok := checks[0].OK
  if checks[0].OK {
    daemonCheck := dockerInfoCheck()
    checks = append(checks, daemonCheck)
    if !daemonCheck.OK {
      ok = false
    } else {
      checks = append(checks, actImageCheck())
    }
  }

  return event.Doctor{
    Type:    event.TypeDoctor,
    OK:      ok,
    Version: version,
    Checks:  checks,
  }
}

func lookPathCheck(name string) event.DoctorCheck {
  path, err := exec.LookPath(name)
  if err != nil {
    return event.DoctorCheck{
      Name:   name,
      OK:     false,
      Detail: "not found on PATH",
    }
  }
  return event.DoctorCheck{
    Name: name,
    OK:   true,
    Path: path,
  }
}

func dockerInfoCheck() event.DoctorCheck {
  cmd := exec.Command("docker", "info")
  if err := cmd.Run(); err != nil {
    return event.DoctorCheck{
      Name:   "docker_daemon",
      OK:     false,
      Detail: "docker found but daemon not reachable",
    }
  }
  return event.DoctorCheck{
    Name:   "docker_daemon",
    OK:     true,
    Detail: "reachable",
  }
}

func actImageCheck() event.DoctorCheck {
  if !actimage.Exists() {
    return event.DoctorCheck{
      Name:   "act_image",
      OK:     false,
      Detail: actimage.Image + " missing (built on first run)",
    }
  }
  return event.DoctorCheck{
    Name:   "act_image",
    OK:     true,
    Detail: actimage.Image,
  }
}
