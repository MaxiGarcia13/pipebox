package jenkins

import (
  "context"
  "net/http"
  "net/http/httptest"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "testing"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func TestStageWatchKeepsParallelBranchesAndLaterStage(t *testing.T) {
  logBody := "[Pipeline] sh\n+ echo TEST_LINE\nTEST_LINE\n[Pipeline] sh\n+ echo LINT_LINE\nLINT_LINE\nFailed in branch lint\n"
  lintAt := strings.Index(logBody, "[Pipeline] sh\n+ echo LINT_LINE")
  branchAt := strings.Index(logBody, "Failed in branch lint")

  home := t.TempDir()
  buildDir := filepath.Join(home, "jobs", pipeboxJobName, "builds", "1")
  if err := os.MkdirAll(buildDir, 0o755); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filepath.Join(buildDir, "log"), []byte(logBody), 0o644); err != nil {
    t.Fatal(err)
  }
  index := "0 7\n" + strconv.Itoa(lintAt) + " 9\n" + strconv.Itoa(branchAt) + " 15\n"
  if err := os.WriteFile(filepath.Join(buildDir, "log-index"), []byte(index), 0o644); err != nil {
    t.Fatal(err)
  }

  phase := "running"
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    switch r.URL.Path {
    case "/job/pipebox/1/wfapi/describe":
      if phase == "running" {
        _, _ = w.Write([]byte(`{"status":"IN_PROGRESS","stages":[
          {"id":"6","name":"test","status":"IN_PROGRESS","durationMillis":5},
          {"id":"8","name":"lint","status":"IN_PROGRESS","durationMillis":5},
          {"id":"20","name":"e2e","status":"NOT_EXECUTED","durationMillis":0}
        ]}`))
        return
      }
      _, _ = w.Write([]byte(`{"status":"SUCCESS","stages":[
        {"id":"6","name":"test","status":"SUCCESS","durationMillis":15},
        {"id":"8","name":"lint","status":"SUCCESS","durationMillis":12},
        {"id":"20","name":"e2e","status":"SUCCESS","durationMillis":9}
      ]}`))
    case "/job/pipebox/1/execution/node/6/wfapi/describe":
      _, _ = w.Write([]byte(atomDescribe("6", "test", phase)))
    case "/job/pipebox/1/execution/node/8/wfapi/describe":
      _, _ = w.Write([]byte(atomDescribe("8", "lint", phase)))
    case "/job/pipebox/1/execution/node/20/wfapi/describe":
      _, _ = w.Write([]byte(atomDescribe("20", "e2e", phase)))
    default:
      http.NotFound(w, r)
    }
  }))
  defer server.Close()

  watch := newStageWatch(&controllerSession{
    baseURL: server.URL,
    runHome: home,
    client:  server.Client(),
  }, 1, []event.JobNode{
    {ID: "test", Name: "test", Steps: []string{"sh"}},
    {ID: "lint", Name: "lint", Steps: []string{"sh"}},
    {ID: "e2e", Name: "e2e", Steps: []string{"sh"}},
  })

  running, fallback, err := watch.poll(context.Background(), false)
  if err != nil || fallback {
    t.Fatalf("poll: fallback %v err %v", fallback, err)
  }
  started := jobsWith(running, event.TypeStepStarted)
  if !started["test"] || !started["lint"] || started["e2e"] {
    t.Fatalf("started while running: %#v\n%#v", started, running)
  }
  if !hasLine(running, "test", "TEST_LINE") || !hasLine(running, "lint", "LINT_LINE") {
    t.Fatalf("logs: %#v", running)
  }
  if !hasLine(running, "lint", "Failed in branch lint") {
    t.Fatalf("branch failure log: %#v", running)
  }
  if hasFinished(running, "e2e") {
    t.Fatalf("e2e finished early: %#v", running)
  }

  e2eAt := len(logBody)
  extra := "[Pipeline] sh\n+ echo E2E_LINE\nE2E_LINE\n"
  file, err := os.OpenFile(filepath.Join(buildDir, "log"), os.O_APPEND|os.O_WRONLY, 0o644)
  if err != nil {
    t.Fatal(err)
  }
  if _, err := file.WriteString(extra); err != nil {
    t.Fatal(err)
  }
  _ = file.Close()
  indexFile, err := os.OpenFile(filepath.Join(buildDir, "log-index"), os.O_APPEND|os.O_WRONLY, 0o644)
  if err != nil {
    t.Fatal(err)
  }
  if _, err := indexFile.WriteString(strconv.Itoa(e2eAt) + " 21\n"); err != nil {
    t.Fatal(err)
  }
  _ = indexFile.Close()
  phase = "done"

  finished, fallback, err := watch.poll(context.Background(), true)
  if err != nil || fallback {
    t.Fatalf("final poll: fallback %v err %v", fallback, err)
  }
  if !hasLine(finished, "e2e", "E2E_LINE") {
    t.Fatalf("e2e log: %#v", finished)
  }
  if !hasStatus(finished, "test", "passed") || !hasStatus(finished, "lint", "passed") || !hasStatus(finished, "e2e", "passed") {
    t.Fatalf("finished: %#v", finished)
  }
}

func TestAtomStatusFinishesLintWhileE2EIsStillPending(t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    switch r.URL.Path {
    case "/job/pipebox/1/wfapi/describe":
      _, _ = w.Write([]byte(`{"status":"IN_PROGRESS","stages":[
        {"id":"8","name":"lint","status":"IN_PROGRESS","durationMillis":40},
        {"id":"20","name":"e2e","status":"IN_PROGRESS","durationMillis":1}
      ]}`))
    case "/job/pipebox/1/execution/node/8/wfapi/describe":
      _, _ = w.Write([]byte(`{"id":"8","name":"lint","status":"IN_PROGRESS","stageFlowNodes":[{"id":"9","name":"Shell Script","status":"SUCCESS","durationMillis":40}]}`))
    case "/job/pipebox/1/execution/node/20/wfapi/describe":
      _, _ = w.Write([]byte(`{"id":"20","name":"e2e","status":"IN_PROGRESS","stageFlowNodes":[{"id":"21","name":"Shell Script","status":"NOT_EXECUTED"}]}`))
    default:
      http.NotFound(w, r)
    }
  }))
  defer server.Close()

  watch := newStageWatch(&controllerSession{
    baseURL: server.URL,
    runHome: t.TempDir(),
    client:  server.Client(),
  }, 1, []event.JobNode{
    {ID: "lint", Name: "lint", Steps: []string{"sh"}},
    {ID: "e2e", Name: "e2e", Steps: []string{"sh"}},
  })
  events, fallback, err := watch.poll(context.Background(), false)
  if err != nil || fallback {
    t.Fatalf("poll: fallback %v err %v", fallback, err)
  }
  if !hasStatus(events, "lint", "passed") {
    t.Fatalf("lint should finish from its sh atom: %#v", events)
  }
  if jobsWith(events, event.TypeStepStarted)["e2e"] || hasFinished(events, "e2e") {
    t.Fatalf("e2e started from the stage chunk: %#v", events)
  }
}

func atomDescribe(id, name, phase string) string {
  status := "IN_PROGRESS"
  if phase == "done" {
    status = "SUCCESS"
  }
  if name == "e2e" && phase != "done" {
    status = "NOT_EXECUTED"
  }
  atomID := "7"
  switch name {
  case "lint":
    atomID = "9"
  case "e2e":
    atomID = "21"
  }
  return `{"id":"` + id + `","name":"` + name + `","status":"` + status + `","stageFlowNodes":[{"id":"` + atomID + `","name":"Shell Script","status":"` + status + `","durationMillis":10}]}`
}

func TestNegativeStageDurationIsOmitted(t *testing.T) {
  watch := newStageWatch(&controllerSession{}, 1, nil)
  tracked := watch.track("8", event.JobNode{ID: "lint", Name: "lint", Steps: []string{"sh"}})
  tracked.duration = -11
  tracked.running["sh"] = true
  events := watch.finishStep(tracked, "sh", "passed")
  finished, ok := events[0].(event.StepFinished)
  if !ok || finished.Status != "passed" || finished.DurationMs != 0 {
    t.Fatalf("finished: %#v", events)
  }
}

func TestDescribeMissFallsBackBeforeAnyStage(t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.NotFound(w, r)
  }))
  defer server.Close()
  watch := newStageWatch(&controllerSession{baseURL: server.URL, runHome: t.TempDir(), client: server.Client()}, 1, []event.JobNode{
    {ID: "test", Name: "test", Steps: []string{"sh"}},
  })
  _, fallback, err := watch.poll(context.Background(), false)
  if err != nil || !fallback {
    t.Fatalf("fallback %v err %v", fallback, err)
  }
}

func jobsWith(events []any, kind string) map[string]bool {
  found := map[string]bool{}
  for _, item := range events {
    switch value := item.(type) {
    case event.StepStarted:
      if value.Type == kind {
        found[value.Job] = true
      }
    }
  }
  return found
}

func hasLine(events []any, job, line string) bool {
  for _, item := range events {
    logEvent, ok := item.(event.Log)
    if ok && logEvent.Job == job && logEvent.Line == line && logEvent.Step == "sh" {
      return true
    }
  }
  return false
}

func hasStatus(events []any, job, status string) bool {
  for _, item := range events {
    finished, ok := item.(event.StepFinished)
    if ok && finished.Job == job && finished.Status == status {
      return true
    }
  }
  return false
}

func hasFinished(events []any, job string) bool {
  for _, item := range events {
    finished, ok := item.(event.StepFinished)
    if ok && finished.Job == job {
      return true
    }
  }
  return false
}
