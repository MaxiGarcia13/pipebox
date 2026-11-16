package workflows

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"

  "go.yaml.in/yaml/v4"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

const maxReusableExpandDepth = 8

type workflowJob struct {
  Name  string         `yaml:"name"`
  Needs needsList      `yaml:"needs"`
  Uses  string         `yaml:"uses"`
  Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
  Name string `yaml:"name"`
  Uses string `yaml:"uses"`
  Run  string `yaml:"run"`
}

type jobDraft struct {
  ID    string
  Name  string
  Needs []string
  Steps []string
  Uses  string
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

  return LoadJobGraph(repoAbs, filepath.Join(repoAbs, filepath.FromSlash(relativePath)))
}

func LoadJobGraph(repoRoot, workflowPath string) (event.JobGraph, error) {
  repoAbs, err := filepath.Abs(repoRoot)
  if err != nil {
    return event.JobGraph{}, err
  }
  absWorkflow, err := filepath.Abs(workflowPath)
  if err != nil {
    return event.JobGraph{}, err
  }

  drafts, err := parseJobs(absWorkflow)
  if err != nil {
    return event.JobGraph{}, err
  }

  stack := map[string]struct{}{absWorkflow: {}}
  jobs, err := expandJobs(repoAbs, drafts, stack, 0)
  if err != nil {
    return event.JobGraph{}, err
  }

  return event.JobGraph{
    Type: event.TypeJobGraph,
    Jobs: jobs,
  }, nil
}

func parseJobs(workflowPath string) ([]jobDraft, error) {
  data, err := os.ReadFile(workflowPath)
  if err != nil {
    return nil, fmt.Errorf("read workflow: %w", err)
  }

  var root yaml.Node
  if err := yaml.Unmarshal(data, &root); err != nil {
    return nil, fmt.Errorf("parse workflow: %w", err)
  }
  if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
    return []jobDraft{}, nil
  }

  jobsNode := mapValue(root.Content[0], "jobs")
  if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
    return []jobDraft{}, nil
  }

  drafts := make([]jobDraft, 0, len(jobsNode.Content)/2)
  for index := 0; index+1 < len(jobsNode.Content); index += 2 {
    id := jobsNode.Content[index].Value
    var job workflowJob
    if err := jobsNode.Content[index+1].Decode(&job); err != nil {
      return nil, fmt.Errorf("parse job %q: %w", id, err)
    }
    name := job.Name
    if name == "" {
      name = id
    }
    needs := []string(job.Needs)
    if needs == nil {
      needs = []string{}
    }
    steps := make([]string, 0, len(job.Steps))
    for _, step := range job.Steps {
      label := stepLabel(step)
      if label != "" {
        steps = append(steps, label)
      }
    }
    drafts = append(drafts, jobDraft{
      ID:    id,
      Name:  name,
      Needs: needs,
      Steps: steps,
      Uses:  strings.TrimSpace(job.Uses),
    })
  }
  return drafts, nil
}

func expandJobs(
  repoRoot string,
  drafts []jobDraft,
  stack map[string]struct{},
  depth int,
) ([]event.JobNode, error) {
  if depth > maxReusableExpandDepth {
    return nil, fmt.Errorf("reusable workflow expansion exceeded depth %d", maxReusableExpandDepth)
  }

  expansionMap := make(map[string][]string, len(drafts))
  result := make([]event.JobNode, 0, len(drafts))
  seenIDs := make(map[string]struct{}, len(drafts))

  for _, draft := range drafts {
    if draft.Uses == "" {
      node := jobNodeFromDraft(draft)
      if err := appendUniqueJob(&result, seenIDs, node); err != nil {
        return nil, err
      }
      expansionMap[draft.ID] = []string{draft.ID}
      continue
    }

    calleePath, local, err := resolveLocalReusable(repoRoot, draft.Uses)
    if err != nil {
      return nil, err
    }
    if !local {
      node := opaqueReusableJob(draft)
      if err := appendUniqueJob(&result, seenIDs, node); err != nil {
        return nil, err
      }
      expansionMap[draft.ID] = []string{draft.ID}
      continue
    }

    if _, cycling := stack[calleePath]; cycling {
      return nil, fmt.Errorf("reusable workflow cycle detected: %s", draft.Uses)
    }

    calleeDrafts, err := parseJobs(calleePath)
    if err != nil {
      node := opaqueReusableJob(draft)
      if err := appendUniqueJob(&result, seenIDs, node); err != nil {
        return nil, err
      }
      expansionMap[draft.ID] = []string{draft.ID}
      continue
    }

    stack[calleePath] = struct{}{}
    expanded, err := expandJobs(repoRoot, calleeDrafts, stack, depth+1)
    delete(stack, calleePath)
    if err != nil {
      return nil, err
    }

    for index := range expanded {
      if len(expanded[index].Needs) == 0 {
        expanded[index].Needs = append([]string{}, draft.Needs...)
      }
      if err := appendUniqueJob(&result, seenIDs, expanded[index]); err != nil {
        return nil, err
      }
    }
    expansionMap[draft.ID] = leafJobIDs(expanded)
  }

  for index := range result {
    result[index].Needs = remapNeeds(result[index].Needs, expansionMap)
  }
  return result, nil
}

func jobNodeFromDraft(draft jobDraft) event.JobNode {
  steps := draft.Steps
  if steps == nil {
    steps = []string{}
  }
  needs := draft.Needs
  if needs == nil {
    needs = []string{}
  }
  return event.JobNode{
    ID:    draft.ID,
    Name:  draft.Name,
    Needs: needs,
    Steps: steps,
  }
}

func opaqueReusableJob(draft jobDraft) event.JobNode {
  node := jobNodeFromDraft(draft)
  if len(node.Steps) == 0 {
    node.Steps = []string{draft.Uses}
  }
  return node
}

func appendUniqueJob(result *[]event.JobNode, seenIDs map[string]struct{}, node event.JobNode) error {
  if _, exists := seenIDs[node.ID]; exists {
    return fmt.Errorf("duplicate job id %q after reusable workflow expansion", node.ID)
  }
  seenIDs[node.ID] = struct{}{}
  *result = append(*result, node)
  return nil
}

func leafJobIDs(jobs []event.JobNode) []string {
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

func remapNeeds(needs []string, expansionMap map[string][]string) []string {
  if len(needs) == 0 {
    return []string{}
  }
  remapped := make([]string, 0, len(needs))
  seen := make(map[string]struct{}, len(needs))
  for _, need := range needs {
    replacements, ok := expansionMap[need]
    if !ok {
      replacements = []string{need}
    }
    for _, id := range replacements {
      if _, exists := seen[id]; exists {
        continue
      }
      seen[id] = struct{}{}
      remapped = append(remapped, id)
    }
  }
  return remapped
}

func resolveLocalReusable(repoRoot, uses string) (absPath string, local bool, err error) {
  if !strings.HasPrefix(uses, "./") && !strings.HasPrefix(uses, ".github/") {
    return "", false, nil
  }

  relative := strings.TrimPrefix(uses, "./")
  cleaned := filepath.Clean(filepath.FromSlash(relative))
  if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || filepath.IsAbs(cleaned) {
    return "", false, fmt.Errorf("reusable workflow path escapes repo: %s", uses)
  }

  absPath = filepath.Join(repoRoot, cleaned)
  relToRepo, relErr := filepath.Rel(repoRoot, absPath)
  if relErr != nil || relToRepo == ".." || strings.HasPrefix(relToRepo, ".."+string(filepath.Separator)) {
    return "", false, fmt.Errorf("reusable workflow path escapes repo: %s", uses)
  }
  return absPath, true, nil
}

func stepLabel(step workflowStep) string {
  if step.Name != "" {
    return step.Name
  }
  if step.Uses != "" {
    return step.Uses
  }
  run := strings.TrimSpace(step.Run)
  if run == "" {
    return ""
  }
  if line, _, ok := strings.Cut(run, "\n"); ok {
    run = strings.TrimSpace(line)
  }
  return run
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
