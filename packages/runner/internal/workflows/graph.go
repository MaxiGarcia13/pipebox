package workflows

import (
  "fmt"
  "os"

  "go.yaml.in/yaml/v4"

  "pipebox/internal/event"
)

type workflowJob struct {
  Name  string    `yaml:"name"`
  Needs needsList `yaml:"needs"`
}

type needsList []string

func (list *needsList) UnmarshalYAML(value *yaml.Node) error {
  if value == nil || value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
    *list = []string{}
    return nil
  }
  switch value.Kind {
  case yaml.ScalarNode:
    if value.Value == "" {
      *list = []string{}
      return nil
    }
    *list = []string{value.Value}
    return nil
  case yaml.SequenceNode:
    items := make([]string, 0, len(value.Content))
    for _, node := range value.Content {
      var item string
      if err := node.Decode(&item); err != nil {
        return err
      }
      if item != "" {
        items = append(items, item)
      }
    }
    *list = items
    return nil
  case yaml.AliasNode:
    return list.UnmarshalYAML(value.Alias)
  default:
    return fmt.Errorf("needs must be a string or list of strings")
  }
}

func LoadJobGraph(workflowPath string) (event.JobGraph, error) {
  data, err := os.ReadFile(workflowPath)
  if err != nil {
    return event.JobGraph{}, fmt.Errorf("read workflow: %w", err)
  }

  var root yaml.Node
  if err := yaml.Unmarshal(data, &root); err != nil {
    return event.JobGraph{}, fmt.Errorf("parse workflow: %w", err)
  }
  if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
    return event.JobGraph{Type: event.TypeJobGraph, Jobs: []event.JobNode{}}, nil
  }

  jobsNode := mapValue(root.Content[0], "jobs")
  if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
    return event.JobGraph{Type: event.TypeJobGraph, Jobs: []event.JobNode{}}, nil
  }

  jobs := make([]event.JobNode, 0, len(jobsNode.Content)/2)
  for index := 0; index+1 < len(jobsNode.Content); index += 2 {
    id := jobsNode.Content[index].Value
    var job workflowJob
    if err := jobsNode.Content[index+1].Decode(&job); err != nil {
      return event.JobGraph{}, fmt.Errorf("parse job %q: %w", id, err)
    }
    name := job.Name
    if name == "" {
      name = id
    }
    needs := []string(job.Needs)
    if needs == nil {
      needs = []string{}
    }
    jobs = append(jobs, event.JobNode{
      ID:    id,
      Name:  name,
      Needs: needs,
    })
  }

  return event.JobGraph{
    Type: event.TypeJobGraph,
    Jobs: jobs,
  }, nil
}

func mapValue(node *yaml.Node, key string) *yaml.Node {
  if node == nil || node.Kind != yaml.MappingNode {
    return nil
  }
  for index := 0; index+1 < len(node.Content); index += 2 {
    if node.Content[index].Value == key {
      return node.Content[index+1]
    }
  }
  return nil
}
