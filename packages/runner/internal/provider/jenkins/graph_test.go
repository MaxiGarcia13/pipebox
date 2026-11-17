package jenkins

import (
  "path/filepath"
  "testing"
)

func TestParseJobGraphSequential(t *testing.T) {
  source := `
pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        echo 'building'
        sh 'make'
      }
    }
    stage('Test') {
      steps {
        sh 'make test'
      }
    }
  }
}
`
  graph, err := ParseJobGraph(source)
  if err != nil {
    t.Fatalf("ParseJobGraph: %v", err)
  }
  if len(graph.Jobs) != 2 {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
  build := graph.Jobs[0]
  if build.ID != "build" || build.Name != "Build" {
    t.Fatalf("build: %#v", build)
  }
  if len(build.Needs) != 0 {
    t.Fatalf("build needs: %#v", build.Needs)
  }
  if len(build.Steps) != 2 || build.Steps[0] != "echo 'building'" || build.Steps[1] != "sh 'make'" {
    t.Fatalf("build steps: %#v", build.Steps)
  }
  testJob := graph.Jobs[1]
  if testJob.ID != "test" || testJob.Name != "Test" {
    t.Fatalf("test: %#v", testJob)
  }
  if len(testJob.Needs) != 1 || testJob.Needs[0] != "build" {
    t.Fatalf("test needs: %#v", testJob.Needs)
  }
}

func TestParseJobGraphParallel(t *testing.T) {
  source := `
pipeline {
  agent any
  stages {
    stage('Prep') {
      steps { echo 'prep' }
    }
    parallel {
      stage('Lint') {
        steps { sh 'lint' }
      }
      stage('Unit') {
        steps { sh 'test' }
      }
    }
    stage('Publish') {
      steps { echo 'done' }
    }
  }
}
`
  graph, err := ParseJobGraph(source)
  if err != nil {
    t.Fatalf("ParseJobGraph: %v", err)
  }
  if len(graph.Jobs) != 4 {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
  if graph.Jobs[1].ID != "lint" || len(graph.Jobs[1].Needs) != 1 || graph.Jobs[1].Needs[0] != "prep" {
    t.Fatalf("lint: %#v", graph.Jobs[1])
  }
  if graph.Jobs[2].ID != "unit" || len(graph.Jobs[2].Needs) != 1 || graph.Jobs[2].Needs[0] != "prep" {
    t.Fatalf("unit: %#v", graph.Jobs[2])
  }
  publish := graph.Jobs[3]
  if publish.ID != "publish" {
    t.Fatalf("publish: %#v", publish)
  }
  if len(publish.Needs) != 2 {
    t.Fatalf("publish needs: %#v", publish.Needs)
  }
  if publish.Needs[0] != "lint" || publish.Needs[1] != "unit" {
    t.Fatalf("publish needs order: %#v", publish.Needs)
  }
}

func TestParseJobGraphScriptedOpaque(t *testing.T) {
  source := `
node {
  stage('Build') {
    echo 'scripted'
  }
}
`
  graph, err := ParseJobGraph(source)
  if err != nil {
    t.Fatalf("ParseJobGraph: %v", err)
  }
  if len(graph.Jobs) != 1 {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
  job := graph.Jobs[0]
  if job.ID != "pipeline" || len(job.Steps) != 1 || job.Steps[0] != "Jenkinsfile" {
    t.Fatalf("job: %#v", job)
  }
}

func TestParseJobGraphDuplicateStageNames(t *testing.T) {
  source := `
pipeline {
  agent any
  stages {
    stage('Build') { steps { echo 'one' } }
    stage('Build') { steps { echo 'two' } }
  }
}
`
  graph, err := ParseJobGraph(source)
  if err != nil {
    t.Fatalf("ParseJobGraph: %v", err)
  }
  if graph.Jobs[0].ID != "build" || graph.Jobs[1].ID != "build-2" {
    t.Fatalf("ids: %#v %#v", graph.Jobs[0], graph.Jobs[1])
  }
}

func TestGraphFixtureJenkinsOnly(t *testing.T) {
  repo := filepath.Join("..", "..", "..", "testdata", "repos", "jenkins-only")
  graph, err := Graph(repo, "Jenkinsfile")
  if err != nil {
    t.Fatalf("Graph: %v", err)
  }
  if len(graph.Jobs) != 1 {
    t.Fatalf("jobs: %#v", graph.Jobs)
  }
  job := graph.Jobs[0]
  if job.ID != "build" || job.Name != "Build" {
    t.Fatalf("job: %#v", job)
  }
  if len(job.Steps) != 1 || job.Steps[0] != "echo 'jenkins-only fixture'" {
    t.Fatalf("steps: %#v", job.Steps)
  }
}

func TestListEnrichesCallableParams(t *testing.T) {
  dir := t.TempDir()
  content := `
pipeline {
  agent any
  parameters {
    string(name: 'FOO', defaultValue: 'bar')
  }
  stages {
    stage('Build') { steps { echo 'x' } }
  }
}
`
  if err := writeFile(dir, "Jenkinsfile", content); err != nil {
    t.Fatal(err)
  }
  list, err := List(dir)
  if err != nil {
    t.Fatalf("List: %v", err)
  }
  if len(list.Workflows) != 1 {
    t.Fatalf("workflows: %#v", list.Workflows)
  }
  workflow := list.Workflows[0]
  if !workflow.Callable {
    t.Fatal("expected callable")
  }
  if len(workflow.Inputs) != 1 || workflow.Inputs[0].ID != "FOO" {
    t.Fatalf("inputs: %#v", workflow.Inputs)
  }
}
