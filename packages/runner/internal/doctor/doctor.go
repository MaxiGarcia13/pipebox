package doctor

import (
  "os/exec"

  "pipebox/internal/event"
)

func Run() event.Doctor {
  checks := []event.DoctorCheck{
    lookPathCheck("act"),
    lookPathCheck("docker"),
  }

  ok := true
  for _, check := range checks {
    if !check.OK {
      ok = false
    }
  }

  if dockerOK := checks[1].OK; dockerOK {
    checks = append(checks, dockerInfoCheck())
    if !checks[len(checks)-1].OK {
      ok = false
    }
  }

  return event.Doctor{
    Type:   event.TypeDoctor,
    OK:     ok,
    Checks: checks,
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
