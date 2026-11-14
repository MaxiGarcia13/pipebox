package workflows

import (
  "fmt"
  "os"
  "strconv"
  "strings"

  "go.yaml.in/yaml/v4"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type dispatchMeta struct {
  Name     string
  Callable bool
  Events   []string
  Inputs   []event.WorkflowInput
}

func LoadDispatch(workflowPath string) (dispatchMeta, error) {
  data, err := os.ReadFile(workflowPath)
  if err != nil {
    return dispatchMeta{}, fmt.Errorf("read workflow: %w", err)
  }

  var root yaml.Node
  if err := yaml.Unmarshal(data, &root); err != nil {
    return dispatchMeta{}, fmt.Errorf("parse workflow: %w", err)
  }
  if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
    return dispatchMeta{}, nil
  }

  doc := root.Content[0]
  meta := dispatchMeta{
    Name:   scalarString(mapValue(doc, "name")),
    Events: []string{},
    Inputs: []event.WorkflowInput{},
  }

  onNode := mapValue(doc, "on")
  if onNode == nil {
    onNode = mapValue(doc, "true")
  }
  meta.Events = listOnEvents(onNode)

  dispatchNode := findWorkflowDispatch(onNode)
  if dispatchNode == nil {
    return meta, nil
  }

  meta.Callable = true
  if dispatchNode.Kind != yaml.MappingNode {
    return meta, nil
  }

  inputsNode := mapValue(dispatchNode, "inputs")
  if inputsNode == nil || inputsNode.Kind != yaml.MappingNode {
    return meta, nil
  }

  inputs := make([]event.WorkflowInput, 0, len(inputsNode.Content)/2)
  for index := 0; index+1 < len(inputsNode.Content); index += 2 {
    id := inputsNode.Content[index].Value
    if id == "" {
      continue
    }
    input, err := parseWorkflowInput(id, inputsNode.Content[index+1])
    if err != nil {
      return dispatchMeta{}, fmt.Errorf("parse input %q: %w", id, err)
    }
    inputs = append(inputs, input)
  }
  meta.Inputs = inputs
  return meta, nil
}

func listOnEvents(onNode *yaml.Node) []string {
  if onNode == nil {
    return []string{}
  }
  switch onNode.Kind {
  case yaml.ScalarNode:
    if onNode.Value == "" {
      return []string{}
    }
    return []string{onNode.Value}
  case yaml.SequenceNode:
    events := make([]string, 0, len(onNode.Content))
    for _, item := range onNode.Content {
      if item.Kind == yaml.AliasNode {
        item = item.Alias
      }
      if item != nil && item.Kind == yaml.ScalarNode && item.Value != "" {
        events = append(events, item.Value)
      }
    }
    return events
  case yaml.MappingNode:
    events := make([]string, 0, len(onNode.Content)/2)
    for index := 0; index+1 < len(onNode.Content); index += 2 {
      key := onNode.Content[index]
      if key != nil && key.Kind == yaml.ScalarNode && key.Value != "" {
        events = append(events, key.Value)
      }
    }
    return events
  case yaml.AliasNode:
    return listOnEvents(onNode.Alias)
  default:
    return []string{}
  }
}

func ResolveActEvent(events []string) string {
  if len(events) == 0 {
    return "push"
  }
  preferred := []string{
    "workflow_dispatch",
    "push",
    "pull_request",
    "workflow_call",
  }
  present := make(map[string]struct{}, len(events))
  for _, eventName := range events {
    present[eventName] = struct{}{}
  }
  for _, eventName := range preferred {
    if _, ok := present[eventName]; ok {
      return eventName
    }
  }
  return events[0]
}

func findWorkflowDispatch(onNode *yaml.Node) *yaml.Node {
  if onNode == nil {
    return nil
  }
  switch onNode.Kind {
  case yaml.ScalarNode:
    if onNode.Value == "workflow_dispatch" {
      return onNode
    }
    return nil
  case yaml.SequenceNode:
    for _, item := range onNode.Content {
      if item.Kind == yaml.ScalarNode && item.Value == "workflow_dispatch" {
        return item
      }
    }
    return nil
  case yaml.MappingNode:
    return mapValue(onNode, "workflow_dispatch")
  case yaml.AliasNode:
    return findWorkflowDispatch(onNode.Alias)
  default:
    return nil
  }
}

func parseWorkflowInput(id string, node *yaml.Node) (event.WorkflowInput, error) {
  input := event.WorkflowInput{ID: id, Options: []string{}}
  if node == nil || node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
    return input, nil
  }
  if node.Kind != yaml.MappingNode {
    return event.WorkflowInput{}, fmt.Errorf("input must be a mapping")
  }

  input.Description = scalarString(mapValue(node, "description"))
  input.Type = normalizeInputType(scalarString(mapValue(node, "type")))
  input.Default = scalarString(mapValue(node, "default"))
  input.Required = scalarBool(mapValue(node, "required"))

  optionsNode := mapValue(node, "options")
  if optionsNode != nil && optionsNode.Kind == yaml.SequenceNode {
    options := make([]string, 0, len(optionsNode.Content))
    for _, optionNode := range optionsNode.Content {
      value := scalarString(optionNode)
      if value != "" {
        options = append(options, value)
      }
    }
    input.Options = options
  }

  return input, nil
}

func normalizeInputType(value string) string {
  switch strings.ToLower(strings.TrimSpace(value)) {
  case "", "string":
    return "string"
  case "boolean":
    return "boolean"
  case "choice":
    return "choice"
  case "number":
    return "number"
  case "environment":
    return "string"
  default:
    return "string"
  }
}

func scalarString(node *yaml.Node) string {
  if node == nil {
    return ""
  }
  if node.Kind == yaml.AliasNode {
    return scalarString(node.Alias)
  }
  if node.Kind != yaml.ScalarNode {
    return ""
  }
  return node.Value
}

func scalarBool(node *yaml.Node) bool {
  value := strings.TrimSpace(strings.ToLower(scalarString(node)))
  if value == "" {
    return false
  }
  parsed, err := strconv.ParseBool(value)
  if err != nil {
    return false
  }
  return parsed
}
