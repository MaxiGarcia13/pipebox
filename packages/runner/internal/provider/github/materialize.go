package github

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"

  "go.yaml.in/yaml/v4"
)

type materializedWorkflow struct {
  Path    string
  Cleanup func()
}

type yamlJob struct {
  ID    string
  Node  *yaml.Node
  Uses  string
  Needs []string
}

func MaterializeLocalReusables(repoRoot, workflowPath string) (materializedWorkflow, error) {
  repoAbs, err := filepath.Abs(repoRoot)
  if err != nil {
    return materializedWorkflow{}, err
  }
  absWorkflow, err := filepath.Abs(workflowPath)
  if err != nil {
    return materializedWorkflow{}, err
  }

  root, err := readWorkflowDocument(absWorkflow)
  if err != nil {
    return materializedWorkflow{}, err
  }
  if root == nil {
    return materializedWorkflow{Path: absWorkflow, Cleanup: func() {}}, nil
  }

  jobsNode := mapValue(root, "jobs")
  if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
    return materializedWorkflow{Path: absWorkflow, Cleanup: func() {}}, nil
  }

  drafts, err := yamlJobsFromMapping(jobsNode)
  if err != nil {
    return materializedWorkflow{}, err
  }
  if !hasLocalReusable(drafts) {
    return materializedWorkflow{Path: absWorkflow, Cleanup: func() {}}, nil
  }

  stack := map[string]struct{}{absWorkflow: {}}
  expanded, err := expandYAMLJobs(repoAbs, drafts, stack, 0)
  if err != nil {
    return materializedWorkflow{}, err
  }

  jobsMapping := &yaml.Node{Kind: yaml.MappingNode}
  for _, job := range expanded {
    jobsMapping.Content = append(jobsMapping.Content,
      &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: job.ID},
      job.Node,
    )
  }
  setMapValue(root, "jobs", jobsMapping)

  data, err := yaml.Marshal(root)
  if err != nil {
    return materializedWorkflow{}, fmt.Errorf("marshal materialized workflow: %w", err)
  }

  tempFile, err := os.CreateTemp("", "pipebox-workflow-*.yml")
  if err != nil {
    return materializedWorkflow{}, fmt.Errorf("create materialized workflow: %w", err)
  }
  tempPath := tempFile.Name()
  if _, err := tempFile.Write(data); err != nil {
    _ = tempFile.Close()
    _ = os.Remove(tempPath)
    return materializedWorkflow{}, fmt.Errorf("write materialized workflow: %w", err)
  }
  if err := tempFile.Close(); err != nil {
    _ = os.Remove(tempPath)
    return materializedWorkflow{}, fmt.Errorf("close materialized workflow: %w", err)
  }

  return materializedWorkflow{
    Path: tempPath,
    Cleanup: func() {
      _ = os.Remove(tempPath)
    },
  }, nil
}

func hasLocalReusable(jobs []yamlJob) bool {
  for _, job := range jobs {
    if strings.HasPrefix(job.Uses, "./") || strings.HasPrefix(job.Uses, ".github/") {
      return true
    }
  }
  return false
}

func readWorkflowDocument(path string) (*yaml.Node, error) {
  data, err := os.ReadFile(path)
  if err != nil {
    return nil, fmt.Errorf("read workflow: %w", err)
  }
  var document yaml.Node
  if err := yaml.Unmarshal(data, &document); err != nil {
    return nil, fmt.Errorf("parse workflow: %w", err)
  }
  if document.Kind != yaml.DocumentNode || len(document.Content) == 0 {
    return nil, nil
  }
  root := document.Content[0]
  if root == nil || root.Kind != yaml.MappingNode {
    return nil, nil
  }
  return cloneYAMLNode(root), nil
}

func yamlJobsFromMapping(jobsNode *yaml.Node) ([]yamlJob, error) {
  jobs := make([]yamlJob, 0, len(jobsNode.Content)/2)
  for index := 0; index+1 < len(jobsNode.Content); index += 2 {
    id := jobsNode.Content[index].Value
    node := cloneYAMLNode(jobsNode.Content[index+1])
    uses := strings.TrimSpace(scalarMapValue(node, "uses"))
    needs, err := needsFromJobNode(node)
    if err != nil {
      return nil, fmt.Errorf("parse job %q needs: %w", id, err)
    }
    jobs = append(jobs, yamlJob{
      ID:    id,
      Node:  node,
      Uses:  uses,
      Needs: needs,
    })
  }
  return jobs, nil
}

func expandYAMLJobs(
  repoRoot string,
  jobs []yamlJob,
  stack map[string]struct{},
  depth int,
) ([]yamlJob, error) {
  if depth > maxReusableExpandDepth {
    return nil, fmt.Errorf("reusable workflow expansion exceeded depth %d", maxReusableExpandDepth)
  }

  expansionMap := make(map[string][]string, len(jobs))
  result := make([]yamlJob, 0, len(jobs))
  seenIDs := make(map[string]struct{}, len(jobs))

  for _, job := range jobs {
    if job.Uses == "" {
      if err := appendUniqueYAMLJob(&result, seenIDs, job); err != nil {
        return nil, err
      }
      expansionMap[job.ID] = []string{job.ID}
      continue
    }

    calleePath, local, err := resolveLocalReusable(repoRoot, job.Uses)
    if err != nil {
      return nil, err
    }
    if !local {
      if err := appendUniqueYAMLJob(&result, seenIDs, job); err != nil {
        return nil, err
      }
      expansionMap[job.ID] = []string{job.ID}
      continue
    }

    if _, cycling := stack[calleePath]; cycling {
      return nil, fmt.Errorf("reusable workflow cycle detected: %s", job.Uses)
    }

    calleeRoot, err := readWorkflowDocument(calleePath)
    if err != nil {
      if err := appendUniqueYAMLJob(&result, seenIDs, job); err != nil {
        return nil, err
      }
      expansionMap[job.ID] = []string{job.ID}
      continue
    }
    calleeJobsNode := mapValue(calleeRoot, "jobs")
    if calleeJobsNode == nil || calleeJobsNode.Kind != yaml.MappingNode {
      if err := appendUniqueYAMLJob(&result, seenIDs, job); err != nil {
        return nil, err
      }
      expansionMap[job.ID] = []string{job.ID}
      continue
    }
    calleeJobs, err := yamlJobsFromMapping(calleeJobsNode)
    if err != nil {
      return nil, err
    }

    stack[calleePath] = struct{}{}
    expanded, err := expandYAMLJobs(repoRoot, calleeJobs, stack, depth+1)
    delete(stack, calleePath)
    if err != nil {
      return nil, err
    }

    for index := range expanded {
      if len(expanded[index].Needs) == 0 {
        expanded[index].Needs = append([]string{}, job.Needs...)
        setJobNeeds(expanded[index].Node, expanded[index].Needs)
      }
      if err := appendUniqueYAMLJob(&result, seenIDs, expanded[index]); err != nil {
        return nil, err
      }
    }
    expansionMap[job.ID] = leafYAMLJobIDs(expanded)
  }

  for index := range result {
    remapped := remapNeeds(result[index].Needs, expansionMap)
    result[index].Needs = remapped
    setJobNeeds(result[index].Node, remapped)
  }
  return result, nil
}

func appendUniqueYAMLJob(result *[]yamlJob, seenIDs map[string]struct{}, job yamlJob) error {
  if _, exists := seenIDs[job.ID]; exists {
    return fmt.Errorf("duplicate job id %q after reusable workflow expansion", job.ID)
  }
  seenIDs[job.ID] = struct{}{}
  *result = append(*result, job)
  return nil
}

func leafYAMLJobIDs(jobs []yamlJob) []string {
  needed := make(map[string]struct{}, len(jobs))
  for _, job := range jobs {
    for _, need := range job.Needs {
      needed[need] = struct{}{}
    }
  }
  leaves := make([]string, 0, len(jobs))
  for _, job := range jobs {
    if _, isNeeded := needed[job.ID]; !isNeeded {
      leaves = append(leaves, job.ID)
    }
  }
  return leaves
}

func needsFromJobNode(jobNode *yaml.Node) ([]string, error) {
  needsNode := mapValue(jobNode, "needs")
  if needsNode == nil {
    return []string{}, nil
  }
  var needs needsList
  if err := needs.UnmarshalYAML(needsNode); err != nil {
    return nil, err
  }
  if needs == nil {
    return []string{}, nil
  }
  return []string(needs), nil
}

func setJobNeeds(jobNode *yaml.Node, needs []string) {
  if jobNode == nil || jobNode.Kind != yaml.MappingNode {
    return
  }
  if len(needs) == 0 {
    deleteMapKey(jobNode, "needs")
    return
  }
  if len(needs) == 1 {
    setMapValue(jobNode, "needs", &yaml.Node{
      Kind:  yaml.ScalarNode,
      Tag:   "!!str",
      Value: needs[0],
    })
    return
  }
  sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
  for _, need := range needs {
    sequence.Content = append(sequence.Content, &yaml.Node{
      Kind:  yaml.ScalarNode,
      Tag:   "!!str",
      Value: need,
    })
  }
  setMapValue(jobNode, "needs", sequence)
}

func scalarMapValue(node *yaml.Node, key string) string {
  value := mapValue(node, key)
  if value == nil || value.Kind != yaml.ScalarNode {
    return ""
  }
  return value.Value
}

func setMapValue(node *yaml.Node, key string, value *yaml.Node) {
  if node == nil || node.Kind != yaml.MappingNode {
    return
  }
  for index := 0; index+1 < len(node.Content); index += 2 {
    if node.Content[index].Value == key {
      node.Content[index+1] = value
      return
    }
  }
  node.Content = append(node.Content,
    &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
    value,
  )
}

func deleteMapKey(node *yaml.Node, key string) {
  if node == nil || node.Kind != yaml.MappingNode {
    return
  }
  for index := 0; index+1 < len(node.Content); index += 2 {
    if node.Content[index].Value != key {
      continue
    }
    node.Content = append(node.Content[:index], node.Content[index+2:]...)
    return
  }
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
  if node == nil {
    return nil
  }
  cloned := *node
  if len(node.Content) > 0 {
    cloned.Content = make([]*yaml.Node, len(node.Content))
    for index, child := range node.Content {
      cloned.Content[index] = cloneYAMLNode(child)
    }
  }
  return &cloned
}
