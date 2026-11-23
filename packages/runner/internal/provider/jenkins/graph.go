package jenkins

import (
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "strings"
  "unicode"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var stageNamePattern = regexp.MustCompile(`(?is)^stage\s*\(\s*['"]([^'"]+)['"]\s*\)\s*\{`)

type stageNode struct {
  Name  string
  Steps []string
}

type stageGroup struct {
  Parallel bool
  Stages   []stageNode
}

func Graph(repoPath, workflow string) (event.JobGraph, error) {
  repoAbs, err := filepath.Abs(repoPath)
  if err != nil {
    return event.JobGraph{}, err
  }

  relativePath := workflow
  if filepath.IsAbs(workflow) {
    relativePath, err = filepath.Rel(repoAbs, workflow)
    if err != nil {
      return event.JobGraph{}, fmt.Errorf("workflow must be inside the repo: %w", err)
    }
  }

  absWorkflow := filepath.Join(repoAbs, filepath.FromSlash(relativePath))
  data, err := os.ReadFile(absWorkflow)
  if err != nil {
    return event.JobGraph{}, fmt.Errorf("read Jenkinsfile: %w", err)
  }
  return ParseJobGraph(string(data))
}

func ParseJobGraph(source string) (event.JobGraph, error) {
  cleaned := stripJenkinsNoise(source)
  pipelineBody, ok := extractFirstBlock(cleaned, "pipeline")
  if !ok {
    return scriptedGraph(), nil
  }

  stagesBody, ok := extractFirstBlock(pipelineBody, "stages")
  if !ok {
    return scriptedGraph(), nil
  }

  groups := parseStageGroups(stagesBody)
  if len(groups) == 0 {
    return scriptedGraph(), nil
  }

  jobs := make([]event.JobNode, 0)
  previous := []string{}
  usedIDs := map[string]int{}

  for _, group := range groups {
    groupIDs := make([]string, 0, len(group.Stages))
    for _, stage := range group.Stages {
      id := uniqueStageID(stage.Name, usedIDs)
      needs := append([]string{}, previous...)
      jobs = append(jobs, event.JobNode{
        ID:    id,
        Name:  stage.Name,
        Needs: needs,
        Steps: stage.Steps,
      })
      groupIDs = append(groupIDs, id)
    }
    if group.Parallel {
      previous = groupIDs
    } else if len(groupIDs) == 1 {
      previous = []string{groupIDs[0]}
    }
  }

  return event.JobGraph{
    Type: event.TypeJobGraph,
    Jobs: jobs,
  }, nil
}

func scriptedGraph() event.JobGraph {
  return event.JobGraph{
    Type: event.TypeJobGraph,
    Jobs: []event.JobNode{{
      ID:    "pipeline",
      Name:  "pipeline",
      Needs: []string{},
      Steps: []string{"Jenkinsfile"},
    }},
  }
}

func parseStageGroups(stagesBody string) []stageGroup {
  groups := make([]stageGroup, 0)
  index := 0
  for index < len(stagesBody) {
    index = skipSpace(stagesBody, index)
    if index >= len(stagesBody) {
      break
    }

    if hasIdentAt(stagesBody, index, "parallel") {
      cursor := skipSpace(stagesBody, index+len("parallel"))
      if cursor < len(stagesBody) && stagesBody[cursor] == '{' {
        body, end, ok := balancedBody(stagesBody, cursor)
        if !ok {
          break
        }
        stages := parseStagesInBody(body)
        if len(stages) > 0 {
          groups = append(groups, stageGroup{Parallel: true, Stages: stages})
        }
        index = end
        continue
      }
    }

    if hasIdentAt(stagesBody, index, "stage") {
      group, end, ok := parseStageGroupAt(stagesBody, index)
      if !ok {
        index++
        continue
      }
      groups = append(groups, group)
      index = end
      continue
    }

    index++
  }
  return groups
}

func parseStageGroupAt(source string, from int) (stageGroup, int, bool) {
  name, body, end, ok := readStageBody(source, from)
  if !ok {
    return stageGroup{}, from, false
  }
  // stage('CI') { parallel { stage('test')… } } → flatten to parallel jobs (no empty wrapper).
  if parallelBody, parallelOK := extractFirstBlock(body, "parallel"); parallelOK {
    stages := parseStagesInBody(parallelBody)
    if len(stages) > 0 {
      return stageGroup{Parallel: true, Stages: stages}, end, true
    }
  }
  return stageGroup{
    Parallel: false,
    Stages:   []stageNode{{Name: name, Steps: parseStepLabels(body)}},
  }, end, true
}

func parseStagesInBody(body string) []stageNode {
  stages := make([]stageNode, 0)
  index := 0
  for index < len(body) {
    index = skipSpace(body, index)
    if index >= len(body) {
      break
    }
    if !hasIdentAt(body, index, "stage") {
      index++
      continue
    }
    name, stageBody, end, ok := readStageBody(body, index)
    if !ok {
      index++
      continue
    }
    if parallelBody, parallelOK := extractFirstBlock(stageBody, "parallel"); parallelOK {
      nested := parseStagesInBody(parallelBody)
      if len(nested) > 0 {
        stages = append(stages, nested...)
        index = end
        continue
      }
    }
    stages = append(stages, stageNode{Name: name, Steps: parseStepLabels(stageBody)})
    index = end
  }
  return stages
}

func readStageBody(source string, from int) (name string, body string, end int, ok bool) {
  match := stageNamePattern.FindStringSubmatch(source[from:])
  if match == nil {
    return "", "", from, false
  }
  openBrace := from + strings.Index(source[from:], "{")
  if openBrace < from {
    return "", "", from, false
  }
  body, end, ok = balancedBody(source, openBrace)
  if !ok {
    return "", "", from, false
  }
  return match[1], body, end, true
}

func parseStepLabels(stageBody string) []string {
  stepsBody, ok := extractFirstBlock(stageBody, "steps")
  if !ok {
    return []string{}
  }
  labels := make([]string, 0)
  for _, call := range splitTopLevelCalls(stepsBody) {
    label := stepLabel(call)
    if label != "" {
      labels = append(labels, label)
    }
  }
  return labels
}

func stepLabel(call string) string {
  trimmed := strings.TrimSpace(call)
  if trimmed == "" {
    return ""
  }
  // Keyword only — matches Jenkins "[Pipeline] <step>" console tokens.
  var builder strings.Builder
  for _, char := range trimmed {
    if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == '$' {
      builder.WriteRune(char)
      continue
    }
    break
  }
  return builder.String()
}

func uniqueStageID(name string, used map[string]int) string {
  base := slugify(name)
  if base == "" {
    base = "stage"
  }
  count := used[base]
  used[base] = count + 1
  if count == 0 {
    return base
  }
  return fmt.Sprintf("%s-%d", base, count+1)
}

func slugify(name string) string {
  var builder strings.Builder
  lastDash := false
  for _, char := range strings.ToLower(name) {
    if unicode.IsLetter(char) || unicode.IsDigit(char) {
      builder.WriteRune(char)
      lastDash = false
      continue
    }
    if !lastDash && builder.Len() > 0 {
      builder.WriteByte('-')
      lastDash = true
    }
  }
  return strings.Trim(builder.String(), "-")
}
