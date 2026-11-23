package jenkins

import (
  "fmt"
  "os"
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/dockerimage"
)

func TestPipelineJobConfigXMLEscapesScript(t *testing.T) {
  xml := pipelineJobConfigXML(`pipeline { agent any; stages { stage('A') { steps { echo '<hi>' } } } }`, nil)
  if !strings.Contains(xml, "&lt;hi&gt;") {
    t.Fatalf("expected escaped script, got %s", xml)
  }
  if !strings.Contains(xml, "flow-definition") || !strings.Contains(xml, "CpsFlowDefinition") {
    t.Fatalf("missing job shape: %s", xml)
  }
  if !strings.Contains(xml, "<sandbox>false</sandbox>") {
    t.Fatalf("sandbox should be off: %s", xml)
  }
}

func TestWorkspaceWorkflowPathRelative(t *testing.T) {
  got, err := workspaceWorkflowPath("/repo", "Jenkinsfile")
  if err != nil {
    t.Fatal(err)
  }
  if got != "Jenkinsfile" {
    t.Fatalf("got %q", got)
  }
}

func TestWorkspaceWorkflowPathAbsolute(t *testing.T) {
  got, err := workspaceWorkflowPath("/Users/me/repo", "/Users/me/repo/Jenkinsfile")
  if err != nil {
    t.Fatal(err)
  }
  if got != "Jenkinsfile" {
    t.Fatalf("got %q", got)
  }
}

func TestControllerRunArgs(t *testing.T) {
  args := controllerRunArgs("/Users/me/repo", "/tmp/home", "pipebox-jenkins:latest", []moduleMount{{
    Host: "/tmp/empty-modules",
    Path: "/Users/me/repo/node_modules",
  }})
  joined := strings.Join(args, "\n")
  for _, want := range []string{
    "-d",
    "--rm",
    "-p\n127.0.0.1::8080",
    "-v\n/tmp/home:/var/jenkins_home",
    "-v\n/Users/me:/Users/me",
    "-v\n" + dockerimage.SocketPath() + ":/var/run/docker.sock",
    "-v\n/tmp/empty-modules:/Users/me/repo/node_modules",
    "pipebox-jenkins:latest",
  } {
    if !strings.Contains(joined, want) {
      t.Fatalf("missing %q in\n%s", want, joined)
    }
  }
  for _, want := range []string{
    "-u\n0",
    fmt.Sprintf("PIPEBOX_UID=%d", os.Getuid()),
    fmt.Sprintf("PIPEBOX_GID=%d", os.Getgid()),
  } {
    if os.Getuid() < 0 {
      break
    }
    if !strings.Contains(joined, want) {
      t.Fatalf("missing %q in\n%s", want, joined)
    }
  }
  if strings.Contains(joined, "PIPEBOX_HOST_WORKSPACE") || strings.Contains(joined, "--group-add") {
    t.Fatalf("unexpected arg in\n%s", joined)
  }
}
