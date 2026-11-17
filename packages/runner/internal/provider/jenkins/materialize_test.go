package jenkins

import (
  "strings"
  "testing"
)

func TestMaterializeJobFilter(t *testing.T) {
  source := `
pipeline {
  agent any
  parameters {
    string(name: 'BRANCH', defaultValue: 'main')
  }
  stages {
    stage('Build') {
      steps { echo 'build' }
    }
    stage('Test') {
      steps { echo 'test' }
    }
  }
}
`
  filtered, err := materializeJobFilter(source, "test")
  if err != nil {
    t.Fatalf("materializeJobFilter: %v", err)
  }
  if !strings.Contains(filtered, "stage('Test')") {
    t.Fatalf("missing Test stage: %s", filtered)
  }
  if strings.Contains(filtered, "stage('Build')") {
    t.Fatalf("Build stage should be removed: %s", filtered)
  }
  if !strings.Contains(filtered, "parameters") || !strings.Contains(filtered, "BRANCH") {
    t.Fatalf("parameters should be kept: %s", filtered)
  }
}

func TestMaterializeJobFilterScriptedFails(t *testing.T) {
  _, err := materializeJobFilter("node { echo 'x' }", "build")
  if err == nil {
    t.Fatal("expected error")
  }
}

func TestMaterializeJobFilterMissingStage(t *testing.T) {
  source := `pipeline { agent any stages { stage('Build') { steps { echo 'x' } } } }`
  _, err := materializeJobFilter(source, "missing")
  if err == nil {
    t.Fatal("expected error")
  }
}
