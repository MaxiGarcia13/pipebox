package runcache

import (
  "encoding/json"
  "fmt"
  "sync"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/event"
)

type FailedStep struct {
  Job        string `json:"job"`
  Step       string `json:"step"`
  DurationMs int64  `json:"durationMs,omitempty"`
}

type Summary struct {
  RunID         string       `json:"runId"`
  Repo          string       `json:"repo,omitempty"`
  Workflow      string       `json:"workflow,omitempty"`
  Job           string       `json:"job,omitempty"`
  Status        string       `json:"status,omitempty"`
  ExitCode      int          `json:"exitCode"`
  FailedSteps   []FailedStep `json:"failedSteps"`
  ArtifactCount int          `json:"artifactCount"`
}

type stepKey struct {
  Job  string
  Step string
}

type Record struct {
  Summary Summary
  logs    map[stepKey][]string
}

type Cache struct {
  mu      sync.RWMutex
  nextID  int
  latest  string
  records map[string]*Record
  limit   int
}

func New() *Cache {
  return &Cache{
    records: make(map[string]*Record),
    limit:   8,
  }
}

func (c *Cache) Begin(repo, workflow, job string) string {
  c.mu.Lock()
  defer c.mu.Unlock()

  c.nextID++
  id := fmt.Sprintf("run-%d", c.nextID)
  c.latest = id
  c.records[id] = &Record{
    Summary: Summary{
      RunID:       id,
      Repo:        repo,
      Workflow:    workflow,
      Job:         job,
      FailedSteps: []FailedStep{},
    },
    logs: make(map[stepKey][]string),
  }
  c.evictLocked()
  return id
}

func (c *Cache) Ingest(runID string, raw json.RawMessage) error {
  var head struct {
    Type string `json:"type"`
  }
  if err := json.Unmarshal(raw, &head); err != nil {
    return err
  }

  c.mu.Lock()
  defer c.mu.Unlock()

  rec := c.records[runID]
  if rec == nil {
    return fmt.Errorf("unknown run %q", runID)
  }

  switch head.Type {
  case event.TypeRunStarted:
    var ev event.RunStarted
    if err := json.Unmarshal(raw, &ev); err != nil {
      return err
    }
    if ev.Repo != "" {
      rec.Summary.Repo = ev.Repo
    }
    if ev.Workflow != "" {
      rec.Summary.Workflow = ev.Workflow
    }
    if ev.Job != "" {
      rec.Summary.Job = ev.Job
    }
  case event.TypeRunFinished:
    var ev event.RunFinished
    if err := json.Unmarshal(raw, &ev); err != nil {
      return err
    }
    rec.Summary.Status = ev.Status
    rec.Summary.ExitCode = ev.ExitCode
  case event.TypeStepFinished:
    var ev event.StepFinished
    if err := json.Unmarshal(raw, &ev); err != nil {
      return err
    }
    if ev.Status == "failed" {
      rec.Summary.FailedSteps = append(rec.Summary.FailedSteps, FailedStep{
        Job:        ev.Job,
        Step:       ev.Step,
        DurationMs: ev.DurationMs,
      })
    }
  case event.TypeLog:
    var ev event.Log
    if err := json.Unmarshal(raw, &ev); err != nil {
      return err
    }
    key := stepKey{Job: ev.Job, Step: ev.Step}
    rec.logs[key] = append(rec.logs[key], ev.Line)
  case event.TypeArtifacts:
    var ev event.Artifacts
    if err := json.Unmarshal(raw, &ev); err != nil {
      return err
    }
    rec.Summary.ArtifactCount = len(ev.Artifacts)
  }
  return nil
}

func (c *Cache) Finish(runID string, exitCode int) Summary {
  c.mu.Lock()
  defer c.mu.Unlock()

  rec := c.records[runID]
  if rec == nil {
    return Summary{RunID: runID, ExitCode: exitCode, FailedSteps: []FailedStep{}}
  }
  if rec.Summary.Status == "" {
    rec.Summary.ExitCode = exitCode
    if exitCode == 0 {
      rec.Summary.Status = "passed"
    } else {
      rec.Summary.Status = "failed"
    }
  }
  return cloneSummary(rec.Summary)
}

func (c *Cache) Summary(runID string) (Summary, error) {
  c.mu.RLock()
  defer c.mu.RUnlock()

  id := runID
  if id == "" {
    id = c.latest
  }
  rec := c.records[id]
  if rec == nil {
    if id == "" {
      return Summary{}, fmt.Errorf("no runs cached")
    }
    return Summary{}, fmt.Errorf("unknown run %q", id)
  }
  return cloneSummary(rec.Summary), nil
}

func (c *Cache) StepLogs(runID, job, step string, tail int) ([]string, error) {
  c.mu.RLock()
  defer c.mu.RUnlock()

  id := runID
  if id == "" {
    id = c.latest
  }
  rec := c.records[id]
  if rec == nil {
    if id == "" {
      return nil, fmt.Errorf("no runs cached")
    }
    return nil, fmt.Errorf("unknown run %q", id)
  }

  lines := rec.logs[stepKey{Job: job, Step: step}]
  if tail > 0 && len(lines) > tail {
    lines = lines[len(lines)-tail:]
  }
  out := make([]string, len(lines))
  copy(out, lines)
  return out, nil
}

func (c *Cache) LatestID() string {
  c.mu.RLock()
  defer c.mu.RUnlock()
  return c.latest
}

func (c *Cache) evictLocked() {
  for len(c.records) > c.limit {
    oldest := ""
    oldestNum := c.nextID + 1
    for id := range c.records {
      var n int
      if _, err := fmt.Sscanf(id, "run-%d", &n); err != nil {
        continue
      }
      if n < oldestNum {
        oldestNum = n
        oldest = id
      }
    }
    if oldest == "" || oldest == c.latest {
      return
    }
    delete(c.records, oldest)
  }
}

func cloneSummary(in Summary) Summary {
  out := in
  if in.FailedSteps == nil {
    out.FailedSteps = []FailedStep{}
    return out
  }
  out.FailedSteps = make([]FailedStep, len(in.FailedSteps))
  copy(out.FailedSteps, in.FailedSteps)
  return out
}
