package jenkins

import (
  "regexp"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

var (
  paramNamePattern        = regexp.MustCompile(`(?i)\bname\s*:\s*['"]([^'"]+)['"]`)
  paramDefaultPattern     = regexp.MustCompile(`(?i)\bdefaultValue\s*:\s*(?:'''([\s\S]*?)'''|"""([\s\S]*?)"""|['"]([^'"]*)['"]|(true|false|-?\d+(?:\.\d+)?))`)
  paramDescriptionPattern = regexp.MustCompile(`(?i)\bdescription\s*:\s*['"]([^'"]*)['"]`)
  paramChoicesPattern     = regexp.MustCompile(`(?i)\bchoices\s*:\s*\[([^\]]*)\]`)
  choiceItemPattern       = regexp.MustCompile(`['"]([^'"]+)['"]`)
)

func LoadParams(source string) []event.WorkflowInput {
  cleaned := stripJenkinsNoise(source)
  body, ok := extractFirstBlock(cleaned, "parameters")
  if !ok {
    return []event.WorkflowInput{}
  }

  inputs := make([]event.WorkflowInput, 0)
  for _, call := range splitTopLevelCalls(body) {
    input, ok := parseParameterCall(call)
    if ok {
      inputs = append(inputs, input)
    }
  }
  return inputs
}

func parseParameterCall(call string) (event.WorkflowInput, bool) {
  trimmed := strings.TrimSpace(call)
  if trimmed == "" {
    return event.WorkflowInput{}, false
  }

  kind := paramKind(trimmed)
  if kind == "" {
    return event.WorkflowInput{}, false
  }

  nameMatch := paramNamePattern.FindStringSubmatch(trimmed)
  if nameMatch == nil {
    return event.WorkflowInput{}, false
  }

  input := event.WorkflowInput{
    ID:          nameMatch[1],
    Type:        kind,
    Description: firstSubmatch(paramDescriptionPattern, trimmed),
    Default:     parseDefaultValue(trimmed),
  }

  if kind == "choice" {
    choicesMatch := paramChoicesPattern.FindStringSubmatch(trimmed)
    if choicesMatch != nil {
      items := choiceItemPattern.FindAllStringSubmatch(choicesMatch[1], -1)
      options := make([]string, 0, len(items))
      for _, item := range items {
        options = append(options, item[1])
      }
      input.Options = options
    }
  }

  if kind == "boolean" {
    lower := strings.ToLower(input.Default)
    if lower != "true" && lower != "false" && input.Default != "" {
      input.Default = "false"
    }
  }

  return input, true
}

func paramKind(call string) string {
  lower := strings.ToLower(call)
  switch {
  case strings.HasPrefix(lower, "string("):
    return "string"
  case strings.HasPrefix(lower, "text("):
    return "string"
  case strings.HasPrefix(lower, "password("):
    return "string"
  case strings.HasPrefix(lower, "booleanparam("):
    return "boolean"
  case strings.HasPrefix(lower, "choice("):
    return "choice"
  default:
    return ""
  }
}

func parseDefaultValue(call string) string {
  match := paramDefaultPattern.FindStringSubmatch(call)
  if match == nil {
    return ""
  }
  for _, group := range match[1:] {
    if group != "" {
      return group
    }
  }
  return ""
}

func firstSubmatch(pattern *regexp.Regexp, source string) string {
  match := pattern.FindStringSubmatch(source)
  if match == nil {
    return ""
  }
  return match[1]
}
