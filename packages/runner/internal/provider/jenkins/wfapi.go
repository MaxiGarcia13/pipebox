package jenkins

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "regexp"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/ansi"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  pipelineMetaPattern = regexp.MustCompile(`^\[Pipeline\]\s`)
  buildFinishedPattern = regexp.MustCompile(`(?i)^Finished:\s+(SUCCESS|FAILURE|ABORTED|UNSTABLE)\s*$`)
  branchFailedPattern  = regexp.MustCompile(`(?i)^Failed in branch\s+(.+)$`)
  consoleNotePattern   = regexp.MustCompile(`ha:////[A-Za-z0-9+/=]+`)
)

type wfRun struct {
  Status string    `json:"status"`
  Stages []wfStage `json:"stages"`
}

type wfStage struct {
  ID             string   `json:"id"`
  Name           string   `json:"name"`
  Status         string   `json:"status"`
  DurationMillis int64    `json:"durationMillis"`
  Atoms          []wfAtom `json:"stageFlowNodes"`
}

type wfAtom struct {
  ID             string `json:"id"`
  Name           string `json:"name"`
  Status         string `json:"status"`
  DurationMillis int64  `json:"durationMillis"`
}

func fetchJSON(ctx context.Context, session *controllerSession, target string, dest any) (int, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
  if err != nil {
    return 0, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return 0, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode == http.StatusNotFound {
    return resp.StatusCode, nil
  }
  if resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
    return resp.StatusCode, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
  }
  if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
    return resp.StatusCode, err
  }
  return resp.StatusCode, nil
}

func stageStatus(jenkins string) (status string, terminal bool) {
  switch jenkins {
  case "SUCCESS":
    return "passed", true
  case "UNSTABLE", "FAILED":
    return "failed", true
  case "ABORTED":
    return "cancelled", true
  case "NOT_EXECUTED":
    return "skipped", true
  default:
    return "running", false
  }
}

func displayLine(raw string) (string, bool) {
  line := ansi.Strip(raw)
  line = consoleNotePattern.ReplaceAllString(line, "")
  line = strings.TrimSpace(line)
  if line == "" || pipelineMetaPattern.MatchString(line) || buildFinishedPattern.MatchString(line) {
    return "", false
  }
  return line, true
}

func matchJob(name string, jobs []event.JobNode) (event.JobNode, bool) {
  candidates := []string{slugify(name)}
  if trimmed := slugify(strings.TrimPrefix(name, "Branch: ")); trimmed != "" && trimmed != candidates[0] {
    candidates = append(candidates, trimmed)
  }
  for _, job := range jobs {
    for _, candidate := range candidates {
      if job.ID == candidate || slugify(job.Name) == candidate {
        return job, true
      }
    }
  }
  return event.JobNode{}, false
}

func graphStep(job event.JobNode, index int) string {
  if index >= 0 && index < len(job.Steps) {
    return job.Steps[index]
  }
  if len(job.Steps) > 0 {
    return job.Steps[len(job.Steps)-1]
  }
  return "sh"
}
