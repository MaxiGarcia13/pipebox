package jenkins

import (
  "context"
  "fmt"
  "path/filepath"
  "strconv"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type stepRef struct {
  Job  string
  Step string
}

type stepAtom struct {
  status   string
  duration int64
}

type trackedStage struct {
  job      event.JobNode
  duration int64
  running  map[string]bool
  done     map[string]bool
  atoms    map[string]stepAtom
}

type stageWatch struct {
  session *controllerSession
  build   int
  jobs    []event.JobNode
  nodes   map[string]stepRef
  stages  map[string]*trackedStage
  byJob   map[string]*trackedStage
  queued  []indexedLine
  log       logTail
  saw       bool
  emptyDone int
}

func streamStages(ctx context.Context, session *controllerSession, buildNumber int, jobs []event.JobNode) (int, string, bool) {
  watch := newStageWatch(session, buildNumber, jobs)
  for {
    if err := ctx.Err(); err != nil {
      return 130, "cancelled", false
    }
    events, fallback, err := watch.poll(ctx, false)
    emitStageEvents(events)
    if err != nil {
      _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("stages: %v", err)})
      return 1, "failed", false
    }
    if fallback {
      return 0, "", true
    }

    finished, result, err := buildFinished(ctx, session, buildNumber)
    if err != nil {
      _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("build poll: %v", err)})
      return 1, "failed", false
    }
    if finished {
      events, _, err = watch.poll(ctx, true)
      emitStageEvents(events)
      if err != nil {
        _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("stages: %v", err)})
        return 1, "failed", false
      }
      return resultCode(result)
    }

    timer := time.NewTimer(consolePollInterval)
    select {
    case <-ctx.Done():
      timer.Stop()
      return 130, "cancelled", false
    case <-timer.C:
    }
  }
}

func newStageWatch(session *controllerSession, buildNumber int, jobs []event.JobNode) *stageWatch {
  return &stageWatch{
    session: session,
    build:   buildNumber,
    jobs:    jobs,
    nodes:   map[string]stepRef{},
    stages:  map[string]*trackedStage{},
    byJob:   map[string]*trackedStage{},
  }
}

func emitStageEvents(events []any) {
  for _, item := range events {
    _ = event.EmitStdout(item)
  }
}

func (watch *stageWatch) poll(ctx context.Context, final bool) ([]any, bool, error) {
  var described wfRun
  status, err := fetchJSON(ctx, watch.session, watch.runURL("/wfapi/describe"), &described)
  if err != nil {
    return nil, false, err
  }
  if status == 404 {
    return nil, !watch.saw, nil
  }
  if len(described.Stages) == 0 {
    if !final {
      finished, _, finishedErr := buildFinished(ctx, watch.session, watch.build)
      if finishedErr != nil {
        return nil, false, finishedErr
      }
      final = finished
    }
    if final {
      watch.emptyDone++
    }
    // A finished declarative run still has stages. Give the controller a moment
    // before treating an empty list as "no pipeline API".
    if final && !watch.saw && watch.emptyDone >= 3 {
      return nil, true, nil
    }
    return nil, false, nil
  }
  watch.saw = true

  for _, stage := range described.Stages {
    job, ok := matchJob(stage.Name, watch.jobs)
    if !ok {
      continue
    }
    watch.track(stage.ID, job)
    if err := watch.loadAtoms(ctx, stage.ID); err != nil {
      return nil, false, err
    }
  }
  events, err := watch.readLogs()
  if err != nil {
    return nil, false, err
  }
  events = append(events, watch.settle(described.Stages, final)...)
  if final {
    events = append(events, watch.flushQueue(true)...)
  }
  return events, false, nil
}

func (watch *stageWatch) loadAtoms(ctx context.Context, stageID string) error {
  tracked := watch.stages[stageID]
  if tracked == nil {
    return nil
  }
  var described wfStage
  status, err := fetchJSON(ctx, watch.session, watch.nodeURL(stageID), &described)
  if err != nil || status == 404 {
    return err
  }
  if len(described.Atoms) == 0 {
    return nil
  }
  if tracked.atoms == nil {
    tracked.atoms = map[string]stepAtom{}
  }
  for index, atom := range described.Atoms {
    step := graphStep(tracked.job, index)
    tracked.atoms[step] = stepAtom{status: atom.Status, duration: atom.DurationMillis}
    watch.nodes[atom.ID] = stepRef{Job: tracked.job.ID, Step: step}
  }
  watch.nodes[stageID] = stepRef{Job: tracked.job.ID, Step: graphStep(tracked.job, 0)}
  return nil
}

func (watch *stageWatch) settle(stages []wfStage, final bool) []any {
  events := make([]any, 0)
  for _, stage := range stages {
    job, ok := matchJob(stage.Name, watch.jobs)
    if !ok {
      continue
    }
    tracked := watch.track(stage.ID, job)
    if stage.DurationMillis > 0 {
      tracked.duration = stage.DurationMillis
    }
    if len(tracked.atoms) > 0 {
      events = append(events, watch.settleAtoms(tracked, final)...)
      continue
    }
    // No sh step yet. The stage chunk stays in progress across the next stage,
    // so it must not light the card.
    mapped, terminal := stageStatus(stage.Status)
    if !terminal || (mapped == "skipped" && !final) {
      continue
    }
    events = append(events, watch.finishJob(tracked, mapped)...)
  }
  return events
}

func (watch *stageWatch) settleAtoms(tracked *trackedStage, final bool) []any {
  events := make([]any, 0)
  steps := tracked.job.Steps
  if len(steps) == 0 {
    steps = []string{"sh"}
  }
  for _, step := range steps {
    atom, ok := tracked.atoms[step]
    if !ok || atom.status == "" || atom.status == "NOT_EXECUTED" {
      if final && ok && atom.status == "NOT_EXECUTED" {
        events = append(events, watch.finishStep(tracked, step, "skipped")...)
      }
      continue
    }
    mapped, terminal := stageStatus(atom.status)
    if !terminal {
      events = append(events, watch.startStep(tracked, step)...)
      continue
    }
    if atom.duration > 0 {
      tracked.duration = atom.duration
    }
    events = append(events, watch.finishStep(tracked, step, mapped)...)
  }
  return events
}

func (watch *stageWatch) readLogs() ([]any, error) {
  logPath, indexPath, ok := watch.logPaths()
  if !ok {
    return nil, nil
  }
  lines, err := watch.log.read(logPath, indexPath)
  if err != nil {
    return nil, err
  }
  watch.queued = append(watch.queued, lines...)
  return watch.flushQueue(false), nil
}

func (watch *stageWatch) flushQueue(force bool) []any {
  events := make([]any, 0)
  pending := make([]indexedLine, 0)
  for _, line := range watch.queued {
    text, keep := displayLine(line.Text)
    if !keep {
      continue
    }
    if ref, ok := watch.nodes[line.NodeID]; ok {
      logged, hold := watch.logEvent(ref, text, force)
      if hold {
        pending = append(pending, line)
        continue
      }
      events = append(events, logged...)
      continue
    }
    if jobName, failed := failedBranchJob(text); failed {
        if tracked := watch.byJob[jobName]; tracked != nil {
          logged, hold := watch.logEvent(stepRef{Job: tracked.job.ID, Step: graphStep(tracked.job, 0)}, text, force)
          if hold {
            pending = append(pending, line)
            continue
          }
          events = append(events, logged...)
          continue
        }
    }
    if line.NodeID == "" || force {
      events = append(events, event.Log{Type: event.TypeLog, Line: text})
      continue
    }
    pending = append(pending, line)
  }
  watch.queued = pending
  return events
}

func (watch *stageWatch) logEvent(ref stepRef, text string, force bool) ([]any, bool) {
  tracked := watch.byJob[ref.Job]
  if tracked != nil {
    if atom, ok := tracked.atoms[ref.Step]; ok && (atom.status == "" || atom.status == "NOT_EXECUTED") && !force {
      return nil, true
    }
  }
  events := make([]any, 0)
  if tracked != nil {
    if atom, ok := tracked.atoms[ref.Step]; !ok || !atomTerminal(atom.status) {
      events = append(events, watch.startStep(tracked, ref.Step)...)
    }
  }
  events = append(events, event.Log{
    Type: event.TypeLog,
    Job:  ref.Job,
    Step: ref.Step,
    Line: text,
  })
  return events, false
}

func atomTerminal(status string) bool {
  _, terminal := stageStatus(status)
  return terminal && status != ""
}

func (watch *stageWatch) track(stageID string, job event.JobNode) *trackedStage {
  if tracked := watch.stages[stageID]; tracked != nil {
    return tracked
  }
  tracked := &trackedStage{
    job:     job,
    running: map[string]bool{},
    done:    map[string]bool{},
  }
  watch.stages[stageID] = tracked
  watch.byJob[job.ID] = tracked
  watch.nodes[stageID] = stepRef{Job: job.ID, Step: graphStep(job, 0)}
  return tracked
}

func (watch *stageWatch) startStep(tracked *trackedStage, step string) []any {
  if tracked.running[step] || tracked.done[step] {
    return nil
  }
  tracked.running[step] = true
  return []any{event.StepStarted{Type: event.TypeStepStarted, Job: tracked.job.ID, Step: step}}
}

func (watch *stageWatch) finishStep(tracked *trackedStage, step, status string) []any {
  if tracked.done[step] {
    return nil
  }
  tracked.done[step] = true
  tracked.running[step] = false
  finished := event.StepFinished{
    Type:   event.TypeStepFinished,
    Job:    tracked.job.ID,
    Step:   step,
    Status: status,
  }
  if (status == "passed" || status == "failed") && tracked.duration > 0 {
    finished.DurationMs = tracked.duration
  }
  return []any{finished}
}

func (watch *stageWatch) finishJob(tracked *trackedStage, status string) []any {
  steps := tracked.job.Steps
  if len(steps) == 0 {
    steps = []string{"sh"}
  }
  events := make([]any, 0)
  failedHead := status == "failed" || status == "cancelled"
  for index, step := range steps {
    if tracked.done[step] {
      continue
    }
    if status == "skipped" || (failedHead && !tracked.running[step] && index > 0) {
      events = append(events, watch.finishStep(tracked, step, "skipped")...)
      continue
    }
    events = append(events, watch.finishStep(tracked, step, status)...)
  }
  return events
}

func (watch *stageWatch) logPaths() (string, string, bool) {
  if watch.session.runHome == "" {
    return "", "", false
  }
  logPath := filepath.Join(watch.session.runHome, "jobs", pipeboxJobName, "builds", strconv.Itoa(watch.build), "log")
  return logPath, logPath + "-index", true
}

func (watch *stageWatch) runURL(suffix string) string {
  return fmt.Sprintf("%s/job/%s/%d%s", watch.session.baseURL, pipeboxJobName, watch.build, suffix)
}

func (watch *stageWatch) nodeURL(stageID string) string {
  return watch.runURL("/execution/node/" + stageID + "/wfapi/describe")
}

func failedBranchJob(line string) (string, bool) {
  matches := branchFailedPattern.FindStringSubmatch(line)
  if matches == nil {
    return "", false
  }
  return slugify(matches[1]), true
}

func resultCode(result string) (int, string, bool) {
  switch result {
  case "SUCCESS":
    return 0, "passed", false
  case "ABORTED":
    return 130, "cancelled", false
  default:
    return 1, "failed", false
  }
}
