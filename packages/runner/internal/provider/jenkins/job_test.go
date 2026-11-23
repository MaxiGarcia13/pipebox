package jenkins

import (
  "net/url"
  "strings"
  "testing"
)

func TestPipelineJobConfigXMLParameters(t *testing.T) {
  script := `
pipeline {
  agent any
  parameters {
    string(name: 'BRANCH', defaultValue: 'main', description: 'git ref')
    booleanParam(name: 'DEBUG', defaultValue: true)
    choice(name: 'ENV', choices: ['dev', 'prod'])
  }
  stages { stage('A') { steps { echo 'x' } } }
}
`
  params, err := resolveJobParameters(script, []string{"BRANCH=dev", "NOTE=a b"})
  if err != nil {
    t.Fatal(err)
  }
  xml := pipelineJobConfigXML(script, params)
  if !strings.Contains(xml, "<name>BRANCH</name>") || !strings.Contains(xml, "<defaultValue>main</defaultValue>") {
    t.Fatalf("jenkinsfile default should stay in the job config: %s", xml)
  }
  if !strings.Contains(xml, "<description>git ref</description>") {
    t.Fatalf("missing description: %s", xml)
  }
  if !strings.Contains(xml, "BooleanParameterDefinition") || !strings.Contains(xml, "<defaultValue>true</defaultValue>") {
    t.Fatalf("missing boolean parameter: %s", xml)
  }
  if !strings.Contains(xml, "<string>dev</string>") || !strings.Contains(xml, "<string>prod</string>") {
    t.Fatalf("missing choices: %s", xml)
  }
  if !strings.Contains(xml, "<name>NOTE</name>") || !strings.Contains(xml, "<defaultValue>a b</defaultValue>") {
    t.Fatalf("missing cli parameter: %s", xml)
  }

  endpoint, err := buildEndpoint("http://jenkins", script, []string{"BRANCH=dev", "NOTE=a b"})
  if err != nil {
    t.Fatal(err)
  }
  if !strings.Contains(endpoint, "/job/pipebox/buildWithParameters?") {
    t.Fatalf("endpoint: %s", endpoint)
  }
  query, err := url.ParseQuery(strings.TrimPrefix(endpoint, "http://jenkins/job/pipebox/buildWithParameters?"))
  if err != nil {
    t.Fatal(err)
  }
  if query.Get("BRANCH") != "dev" || query.Get("NOTE") != "a b" || query.Get("DEBUG") != "true" || query.Get("ENV") != "dev" {
    t.Fatalf("query: %v", query)
  }
}

func TestBuildEndpointWithoutParameters(t *testing.T) {
  script := "pipeline { agent any; stages { stage('A') { steps { echo 'x' } } } }"
  endpoint, err := buildEndpoint("http://jenkins", script, nil)
  if err != nil {
    t.Fatal(err)
  }
  if endpoint != "http://jenkins/job/pipebox/build" {
    t.Fatalf("got %s", endpoint)
  }
}

func TestResolveJobParametersRejectsInvalidInput(t *testing.T) {
  if _, err := resolveJobParameters("pipeline { agent any }", []string{"NOPE"}); err == nil {
    t.Fatal("expected error")
  }
}
