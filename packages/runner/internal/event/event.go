package event

import (
  "encoding/json"
  "io"
  "os"
)

const (
  TypeWorkflowList = "workflow_list"
  TypeRunStarted   = "run_started"
  TypeRunFinished  = "run_finished"
  TypeJobGraph     = "job_graph"
  TypeStepStarted  = "step_started"
  TypeStepFinished = "step_finished"
  TypeLog          = "log"
  TypeError        = "error"
  TypeDoctor       = "doctor"
)

type Workflow struct {
  ID   string `json:"id"`
  Name string `json:"name"`
  Path string `json:"path"`
}

type WorkflowList struct {
  Type      string     `json:"type"`
  Workflows []Workflow `json:"workflows"`
}

type RunStarted struct {
  Type     string `json:"type"`
  Repo     string `json:"repo"`
  Workflow string `json:"workflow"`
  Job      string `json:"job,omitempty"`
}

type RunFinished struct {
  Type     string `json:"type"`
  Status   string `json:"status"`
  ExitCode int    `json:"exitCode"`
}

type StepStarted struct {
  Type string `json:"type"`
  Job  string `json:"job"`
  Step string `json:"step"`
}

type JobNode struct {
  ID    string   `json:"id"`
  Name  string   `json:"name"`
  Needs []string `json:"needs"`
}

type JobGraph struct {
  Type string    `json:"type"`
  Jobs []JobNode `json:"jobs"`
}

type StepFinished struct {
  Type       string `json:"type"`
  Job        string `json:"job"`
  Step       string `json:"step"`
  Status     string `json:"status"`
  DurationMs int64  `json:"durationMs,omitempty"`
}

type Log struct {
  Type string `json:"type"`
  Job  string `json:"job,omitempty"`
  Step string `json:"step,omitempty"`
  Line string `json:"line"`
}

type Error struct {
  Type    string `json:"type"`
  Message string `json:"message"`
}

type DoctorCheck struct {
  Name   string `json:"name"`
  OK     bool   `json:"ok"`
  Detail string `json:"detail,omitempty"`
  Path   string `json:"path,omitempty"`
}

type Doctor struct {
  Type   string        `json:"type"`
  OK     bool          `json:"ok"`
  Checks []DoctorCheck `json:"checks"`
}

func Emit(w io.Writer, value any) error {
  if w == nil {
    w = os.Stdout
  }
  encoder := json.NewEncoder(w)
  return encoder.Encode(value)
}

func EmitStdout(value any) error {
  return Emit(os.Stdout, value)
}
