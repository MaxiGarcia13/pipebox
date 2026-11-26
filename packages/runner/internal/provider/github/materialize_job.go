package github

import (
  "fmt"
  "os"
  "strings"

  "go.yaml.in/yaml/v4"
)

type isolatedJob struct {
  materializedWorkflow
  JobID string
}

func IsolateJob(workflowPath, job string) (isolatedJob, error) {
  job = strings.TrimSpace(job)
  if job == "" {
    return isolatedJob{
      materializedWorkflow: materializedWorkflow{Path: workflowPath, Cleanup: func() {}},
    }, nil
  }

  root, err := readWorkflowDocument(workflowPath)
  if err != nil {
    return isolatedJob{}, err
  }
  if root == nil {
    return isolatedJob{}, fmt.Errorf("workflow %q is empty", workflowPath)
  }

  jobsNode := mapValue(root, "jobs")
  if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
    return isolatedJob{}, fmt.Errorf("workflow %q has no jobs", workflowPath)
  }

  jobID, jobNode, ok := findJobNode(jobsNode, job)
  if !ok {
    return isolatedJob{}, fmt.Errorf("job %q not found", job)
  }

  isolated := cloneYAMLNode(jobNode)
  setJobNeeds(isolated, nil)

  jobsMapping := &yaml.Node{Kind: yaml.MappingNode}
  jobsMapping.Content = append(jobsMapping.Content,
    &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: jobID},
    isolated,
  )
  setMapValue(root, "jobs", jobsMapping)

  data, err := yaml.Marshal(root)
  if err != nil {
    return isolatedJob{}, fmt.Errorf("marshal isolated workflow: %w", err)
  }

  tempFile, err := os.CreateTemp("", "pipebox-job-*.yml")
  if err != nil {
    return isolatedJob{}, fmt.Errorf("create isolated workflow: %w", err)
  }
  tempPath := tempFile.Name()
  if _, err := tempFile.Write(data); err != nil {
    _ = tempFile.Close()
    _ = os.Remove(tempPath)
    return isolatedJob{}, fmt.Errorf("write isolated workflow: %w", err)
  }
  if err := tempFile.Close(); err != nil {
    _ = os.Remove(tempPath)
    return isolatedJob{}, fmt.Errorf("close isolated workflow: %w", err)
  }

  return isolatedJob{
    materializedWorkflow: materializedWorkflow{
      Path: tempPath,
      Cleanup: func() {
        _ = os.Remove(tempPath)
      },
    },
    JobID: jobID,
  }, nil
}

func findJobNode(jobsNode *yaml.Node, job string) (id string, node *yaml.Node, ok bool) {
  for index := 0; index+1 < len(jobsNode.Content); index += 2 {
    jobID := jobsNode.Content[index].Value
    jobNode := jobsNode.Content[index+1]
    if jobID == job {
      return jobID, jobNode, true
    }
    name := scalarMapValue(jobNode, "name")
    if name != "" && name == job {
      return jobID, jobNode, true
    }
  }
  return "", nil, false
}
