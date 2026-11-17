package jenkins

import (
  "testing"
)

func TestLoadParams(t *testing.T) {
  source := `
pipeline {
  agent any
  parameters {
    string(name: 'BRANCH', defaultValue: 'main', description: 'Git branch')
    booleanParam(name: 'DEPLOY', defaultValue: true, description: 'Deploy?')
    choice(name: 'ENV', choices: ['dev', 'prod'], description: 'Target')
    text(name: 'NOTES', defaultValue: 'none')
  }
  stages {
    stage('Build') {
      steps { echo 'ok' }
    }
  }
}
`
  inputs := LoadParams(source)
  if len(inputs) != 4 {
    t.Fatalf("inputs: %#v", inputs)
  }

  if inputs[0].ID != "BRANCH" || inputs[0].Type != "string" || inputs[0].Default != "main" {
    t.Fatalf("string: %#v", inputs[0])
  }
  if inputs[0].Description != "Git branch" {
    t.Fatalf("description: %#v", inputs[0])
  }
  if inputs[1].ID != "DEPLOY" || inputs[1].Type != "boolean" || inputs[1].Default != "true" {
    t.Fatalf("boolean: %#v", inputs[1])
  }
  if inputs[2].ID != "ENV" || inputs[2].Type != "choice" {
    t.Fatalf("choice: %#v", inputs[2])
  }
  if len(inputs[2].Options) != 2 || inputs[2].Options[0] != "dev" || inputs[2].Options[1] != "prod" {
    t.Fatalf("options: %#v", inputs[2].Options)
  }
  if inputs[3].ID != "NOTES" || inputs[3].Type != "string" || inputs[3].Default != "none" {
    t.Fatalf("text: %#v", inputs[3])
  }
}

func TestLoadParamsEmpty(t *testing.T) {
  inputs := LoadParams(`pipeline { agent any stages { stage('A') { steps { echo 'x' } } } }`)
  if len(inputs) != 0 {
    t.Fatalf("expected no inputs, got %#v", inputs)
  }
}
