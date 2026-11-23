package event

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
  TypeArtifacts    = "artifacts"
)

type WorkflowInput struct {
  ID          string   `json:"id"`
  Description string   `json:"description,omitempty"`
  Required    bool     `json:"required,omitempty"`
  Default     string   `json:"default,omitempty"`
  Type        string   `json:"type,omitempty"`
  Options     []string `json:"options,omitempty"`
}

type Workflow struct {
  ID       string          `json:"id"`
  Name     string          `json:"name"`
  Path     string          `json:"path"`
  Provider string          `json:"provider"`
  Callable bool            `json:"callable,omitempty"`
  Events   []string        `json:"events,omitempty"`
  Inputs   []WorkflowInput `json:"inputs,omitempty"`
}

type WorkflowList struct {
  Type      string     `json:"type"`
  Providers []string   `json:"providers,omitempty"`
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
  Steps []string `json:"steps"`
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
  Type    string        `json:"type"`
  OK      bool          `json:"ok"`
  Version string        `json:"version"`
  Checks  []DoctorCheck `json:"checks"`
}

type Artifact struct {
  Name string `json:"name"`
  Path string `json:"path"`
  Size int64  `json:"size"`
}

type Artifacts struct {
  Type      string     `json:"type"`
  Artifacts []Artifact `json:"artifacts"`
}
