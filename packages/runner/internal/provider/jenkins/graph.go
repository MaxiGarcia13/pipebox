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
      stage, end, ok := parseStageAt(stagesBody, index)
      if !ok {
        index++
        continue
      }
      groups = append(groups, stageGroup{Parallel: false, Stages: []stageNode{stage}})
      index = end
      continue
    }

    index++
  }
  return groups
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
    stage, end, ok := parseStageAt(body, index)
    if !ok {
      index++
      continue
    }
    stages = append(stages, stage)
    index = end
  }
  return stages
}

func parseStageAt(source string, from int) (stageNode, int, bool) {
  match := stageNamePattern.FindStringSubmatch(source[from:])
  if match == nil {
    return stageNode{}, from, false
  }
  openBrace := from + strings.Index(source[from:], "{")
  if openBrace < from {
    return stageNode{}, from, false
  }
  body, end, ok := balancedBody(source, openBrace)
  if !ok {
    return stageNode{}, from, false
  }
  steps := parseStepLabels(body)
  return stageNode{Name: match[1], Steps: steps}, end, true
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
  line := trimmed
  if newline := strings.IndexByte(line, '\n'); newline >= 0 {
    line = strings.TrimSpace(line[:newline])
  }
  line = strings.TrimSuffix(line, "{")
  line = strings.TrimSpace(line)
  if line == "" {
    return ""
  }
  return line
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
