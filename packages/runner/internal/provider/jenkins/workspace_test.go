package jenkins

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestApplyWorkspaceAgentAnyFixture(t *testing.T) {
  source := readFixture(t, "jenkins-agent-any")
  got := applyWorkspace(source, "/work/repo")
  if strings.Contains(got, "agent any") {
    t.Fatalf("agent any should be rewritten:\n%s", got)
  }
  if !strings.Contains(got, "label 'built-in'") || !strings.Contains(got, "customWorkspace '/work/repo'") {
    t.Fatalf("missing built-in workspace:\n%s", got)
  }
  if !strings.Contains(got, "go version") {
    t.Fatalf("steps were dropped:\n%s", got)
  }
}

func TestApplyWorkspaceDockerAgentFixture(t *testing.T) {
  source := readFixture(t, "jenkins-docker-agent")
  got := applyWorkspace(source, "/work/repo")
  if !strings.Contains(got, "agent none") {
    t.Fatalf("agent none should stay:\n%s", got)
  }
  if !strings.Contains(got, "image 'alpine:3.21'") || !strings.Contains(got, "customWorkspace '/work/repo'") {
    t.Fatalf("docker agent should keep its image and gain a workspace:\n%s", got)
  }
}

func TestApplyWorkspaceReuseNodeKeepsParent(t *testing.T) {
  source := `
pipeline {
  agent any
  stages {
    stage('Inner') {
      agent {
        docker {
          image 'alpine'
          reuseNode true
        }
      }
      steps { sh 'true' }
    }
  }
}
`
  got := applyWorkspace(source, "/work/repo")
  if strings.Count(got, "customWorkspace") != 1 {
    t.Fatalf("expected one workspace injection:\n%s", got)
  }
  if !strings.Contains(got, "reuseNode true") || strings.Contains(got, "agent any") {
    t.Fatalf("got:\n%s", got)
  }
}

func TestApplyWorkspaceLabelAndExistingWorkspace(t *testing.T) {
  labeled := applyWorkspace(`pipeline {
  agent { label 'linux && docker' }
  stages { stage('A') { steps { echo 'x' } } }
}`, "/work/repo")
  if !strings.Contains(labeled, "label 'linux && docker'") || !strings.Contains(labeled, "customWorkspace '/work/repo'") {
    t.Fatalf("label agent:\n%s", labeled)
  }

  existing := `pipeline {
  agent {
    node {
      label 'built-in'
      customWorkspace '/keep'
    }
  }
  stages { stage('A') { steps { echo 'x' } } }
}`
  if applyWorkspace(existing, "/work/repo") != existing {
    t.Fatal("existing customWorkspace should stay")
  }
}

func TestApplyWorkspaceQuotesPathAndSkipsScripted(t *testing.T) {
  got := applyWorkspace("pipeline { agent any; stages { stage('A') { steps { echo 'x' } } } }", "/tmp/o'reilly")
  if !strings.Contains(got, `customWorkspace '/tmp/o\'reilly'`) {
    t.Fatalf("got:\n%s", got)
  }
  scripted := "node { sh 'echo agent any' }"
  if applyWorkspace(scripted, "/work/repo") != scripted {
    t.Fatal("scripted pipeline should stay unchanged")
  }
}

func readFixture(t *testing.T, name string) string {
  t.Helper()
  path := filepath.Join("..", "..", "..", "testdata", "repos", name, "Jenkinsfile")
  data, err := os.ReadFile(path)
  if err != nil {
    t.Fatal(err)
  }
  return string(data)
}
