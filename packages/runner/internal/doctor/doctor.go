package doctor

import (
  "os/exec"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/github/actimage"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins/jfrimage"
)

func Run(version string, providers []string) event.Doctor {
  checks := []event.DoctorCheck{
    lookPathCheck("docker"),
  }

  if checks[0].OK {
    daemonCheck := dockerInfoCheck()
    checks = append(checks, daemonCheck)
    if daemonCheck.OK {
      if hasProvider(providers, event.ProviderGitHub) {
        checks = append(checks, actImageCheck())
      }
      if hasProvider(providers, event.ProviderJenkins) {
        checks = append(checks, jenkinsImageCheck())
      }
    }
  }

  return event.Doctor{
    Type:    event.TypeDoctor,
    OK:      allChecksOK(checks),
    Version: version,
    Checks:  checks,
  }
}

func Emit(providers ...string) {
  _ = event.EmitStdout(Run("", providers))
}

func hasProvider(providers []string, id string) bool {
  for _, provider := range providers {
    if provider == id {
      return true
    }
  }
  return false
}

func allChecksOK(checks []event.DoctorCheck) bool {
  if len(checks) == 0 {
    return false
  }
  for _, check := range checks {
    if !check.OK {
      return false
    }
  }
  return true
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
      Detail: missingImageDetail(actimage.Image+" missing", actimage.DockerfileDir),
    }
  }
  return event.DoctorCheck{
    Name:   "act_image",
    OK:     true,
    Detail: actimage.Image,
  }
}

func jenkinsImageCheck() event.DoctorCheck {
  if !jfrimage.Exists() {
    return event.DoctorCheck{
      Name:   "jenkins_image",
      OK:     false,
      Detail: missingImageDetail(jfrimage.Image+" missing or outdated", jfrimage.DockerfileDir),
    }
  }
  return event.DoctorCheck{
    Name:   "jenkins_image",
    OK:     true,
    Detail: jfrimage.Image,
  }
}

func missingImageDetail(prefix string, resolveDockerfileDir func() (string, error)) string {
  if _, err := resolveDockerfileDir(); err != nil {
    return prefix + "; cannot build: " + err.Error()
  }
  return prefix + " (built on first run)"
}
