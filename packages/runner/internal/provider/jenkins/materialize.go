package jenkins

import (
  "fmt"
  "os"
  "strings"
)

type materializedJenkinsfile struct {
  Path    string
  Cleanup func()
}

func materializeJobFilter(source, job string) (string, error) {
  job = strings.TrimSpace(job)
  if job == "" {
    return source, nil
  }

  cleaned := stripJenkinsNoise(source)
  pipelineBody, ok := extractFirstBlock(cleaned, "pipeline")
  if !ok {
    return "", fmt.Errorf("jenkins --job requires Declarative Pipeline")
  }

  stageText, stageName, ok := findStageSource(cleaned, pipelineBody, job)
  if !ok {
    return "", fmt.Errorf("jenkins stage %q not found", job)
  }
  _ = stageName

  var builder strings.Builder
  builder.WriteString("pipeline {\n")
  if agent := extractAgentClause(pipelineBody); agent != "" {
    builder.WriteString("  ")
    builder.WriteString(agent)
    builder.WriteString("\n")
  } else {
    builder.WriteString("  agent any\n")
  }
  if parameters, ok := extractFirstBlock(pipelineBody, "parameters"); ok {
    builder.WriteString("  parameters {\n")
    builder.WriteString(parameters)
    builder.WriteString("\n  }\n")
  }
  if environment, ok := extractFirstBlock(pipelineBody, "environment"); ok {
    builder.WriteString("  environment {\n")
    builder.WriteString(environment)
    builder.WriteString("\n  }\n")
  }
  builder.WriteString("  stages {\n")
  builder.WriteString(indentBlock(stageText, "    "))
  builder.WriteString("\n  }\n")
  builder.WriteString("}\n")
  return builder.String(), nil
}

func writeMaterializedJenkinsfile(content string) (materializedJenkinsfile, error) {
  file, err := os.CreateTemp("", "pipebox-jenkinsfile-*.groovy")
  if err != nil {
    return materializedJenkinsfile{}, err
  }
  path := file.Name()
  if _, err := file.WriteString(content); err != nil {
    _ = file.Close()
    _ = os.Remove(path)
    return materializedJenkinsfile{}, err
  }
  if err := file.Close(); err != nil {
    _ = os.Remove(path)
    return materializedJenkinsfile{}, err
  }
  return materializedJenkinsfile{
    Path: path,
    Cleanup: func() {
      _ = os.Remove(path)
    },
  }, nil
}

func findStageSource(fullSource, pipelineBody, job string) (stageText string, stageName string, ok bool) {
  stagesBody, ok := extractFirstBlock(pipelineBody, "stages")
  if !ok {
    return "", "", false
  }

  index := 0
  for index < len(stagesBody) {
    index = skipSpace(stagesBody, index)
    if index >= len(stagesBody) {
      break
    }
    if hasIdentAt(stagesBody, index, "parallel") {
      cursor := skipSpace(stagesBody, index+len("parallel"))
      if cursor < len(stagesBody) && stagesBody[cursor] == '{' {
        body, end, bodyOK := balancedBody(stagesBody, cursor)
        if !bodyOK {
          break
        }
        if text, name, found := findStageInBody(fullSource, body, job); found {
          return text, name, true
        }
        index = end
        continue
      }
    }
    if hasIdentAt(stagesBody, index, "stage") {
      text, name, end, found := readStageSource(stagesBody, index)
      if !found {
        index++
        continue
      }
      if stageMatches(name, job) {
        return text, name, true
      }
      index = end
      continue
    }
    index++
  }
  return "", "", false
}

func findStageInBody(fullSource, body, job string) (string, string, bool) {
  _ = fullSource
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
    text, name, end, found := readStageSource(body, index)
    if !found {
      index++
      continue
    }
    if stageMatches(name, job) {
      return text, name, true
    }
    index = end
  }
  return "", "", false
}

func readStageSource(source string, from int) (text string, name string, end int, ok bool) {
  match := stageNamePattern.FindStringSubmatch(source[from:])
  if match == nil {
    return "", "", from, false
  }
  openBrace := from + strings.Index(source[from:], "{")
  if openBrace < from {
    return "", "", from, false
  }
  _, end, ok = balancedBody(source, openBrace)
  if !ok {
    return "", "", from, false
  }
  return strings.TrimSpace(source[from:end]), match[1], end, true
}

func stageMatches(stageName, job string) bool {
  if stageName == job {
    return true
  }
  return slugify(stageName) == slugify(job)
}

func extractAgentClause(pipelineBody string) string {
  found := indexOfIdent(pipelineBody, "agent", 0)
  if found < 0 {
    return ""
  }
  cursor := skipSpace(pipelineBody, found+len("agent"))
  if cursor >= len(pipelineBody) {
    return ""
  }
  if pipelineBody[cursor] == '{' {
    _, end, ok := balancedBody(pipelineBody, cursor)
    if !ok {
      return ""
    }
    return strings.TrimSpace(pipelineBody[found:end])
  }
  end := cursor
  for end < len(pipelineBody) && !isSpaceOrBrace(pipelineBody[end]) {
    end++
  }
  return strings.TrimSpace(pipelineBody[found:end])
}

func isSpaceOrBrace(char byte) bool {
  return char == ' ' || char == '\t' || char == '\n' || char == '\r' || char == '{' || char == '}'
}

func indentBlock(block, prefix string) string {
  lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
  for index, line := range lines {
    if strings.TrimSpace(line) == "" {
      lines[index] = ""
      continue
    }
    lines[index] = prefix + line
  }
  return strings.Join(lines, "\n")
}
