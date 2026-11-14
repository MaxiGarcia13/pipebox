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
    Inputs: []event.WorkflowInput{},
  }

  onNode := mapValue(doc, "on")
  if onNode == nil {
    onNode = mapValue(doc, "true")
  }
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
