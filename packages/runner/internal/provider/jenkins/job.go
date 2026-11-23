package jenkins

import (
  "bytes"
  "context"
  "encoding/json"
  "encoding/xml"
  "fmt"
  "io"
  "net/http"
  "net/url"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

type jobParameter struct {
  Name         string
  Type         string
  Description  string
  DefaultValue string
  Value        string
  Options      []string
}

func loadPipelineScript(options Options) (string, error) {
  repoPath, err := filepath.Abs(options.Repo)
  if err != nil {
    return "", err
  }
  workflowPath, err := workspaceWorkflowPath(repoPath, options.Workflow)
  if err != nil {
    return "", err
  }
  absWorkflow := filepath.Join(repoPath, filepath.FromSlash(workflowPath))
  data, err := os.ReadFile(absWorkflow)
  if err != nil {
    return "", fmt.Errorf("read Jenkinsfile: %w", err)
  }
  script := string(data)
  if options.Job != "" {
    filtered, err := materializeJobFilter(script, options.Job)
    if err != nil {
      return "", err
    }
    script = filtered
  }
  return applyWorkspace(script, repoPath), nil
}

func createOrUpdateJob(ctx context.Context, session *controllerSession, script string, inputs []string) error {
  params, err := resolveJobParameters(script, inputs)
  if err != nil {
    return err
  }
  config := pipelineJobConfigXML(script, params)
  exists, err := jobExists(ctx, session)
  if err != nil {
    return err
  }
  if exists {
    return postXML(ctx, session, session.baseURL+"/job/"+pipeboxJobName+"/config.xml", config)
  }
  return postXML(ctx, session, session.baseURL+"/createItem?name="+pipeboxJobName, config)
}

func jobExists(ctx context.Context, session *controllerSession) (bool, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, session.baseURL+"/job/"+pipeboxJobName+"/api/json", nil)
  if err != nil {
    return false, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return false, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode == http.StatusNotFound {
    return false, nil
  }
  if resp.StatusCode >= 300 {
    body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
    return false, fmt.Errorf("job check: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
  }
  return true, nil
}

func postXML(ctx context.Context, session *controllerSession, target, body string) error {
  req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(body))
  if err != nil {
    return err
  }
  req.Header.Set("Content-Type", "application/xml;charset=UTF-8")
  resp, err := session.client.Do(req)
  if err != nil {
    return err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
    return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
  }
  return nil
}

func pipelineJobConfigXML(script string, params []jobParameter) string {
  var buf bytes.Buffer
  buf.WriteString(`<?xml version='1.1' encoding='UTF-8'?>`)
  buf.WriteString(`<flow-definition plugin="workflow-job">`)
  buf.WriteString(`<description>PipeBox ephemeral pipeline</description>`)
  buf.WriteString(`<keepDependencies>false</keepDependencies>`)
  writeParameterProperties(&buf, params)
  buf.WriteString(`<definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">`)
  buf.WriteString(`<script>`)
  _ = xml.EscapeText(&buf, []byte(script))
  buf.WriteString(`</script>`)
  buf.WriteString(`<sandbox>false</sandbox>`)
  buf.WriteString(`</definition>`)
  buf.WriteString(`<triggers/>`)
  buf.WriteString(`<disabled>false</disabled>`)
  buf.WriteString(`</flow-definition>`)
  return buf.String()
}

func writeParameterProperties(buf *bytes.Buffer, params []jobParameter) {
  if len(params) == 0 {
    return
  }
  buf.WriteString(`<properties>`)
  buf.WriteString(`<hudson.model.ParametersDefinitionProperty>`)
  buf.WriteString(`<parameterDefinitions>`)
  for _, param := range params {
    writeParameterDefinition(buf, param)
  }
  buf.WriteString(`</parameterDefinitions>`)
  buf.WriteString(`</hudson.model.ParametersDefinitionProperty>`)
  buf.WriteString(`</properties>`)
}

func writeParameterDefinition(buf *bytes.Buffer, param jobParameter) {
  switch param.Type {
  case "boolean":
    buf.WriteString(`<hudson.model.BooleanParameterDefinition>`)
    writeParameterName(buf, param)
    buf.WriteString(`<defaultValue>`)
    buf.WriteString(booleanText(param.DefaultValue))
    buf.WriteString(`</defaultValue>`)
    buf.WriteString(`</hudson.model.BooleanParameterDefinition>`)
  case "choice":
    buf.WriteString(`<hudson.model.ChoiceParameterDefinition>`)
    writeParameterName(buf, param)
    buf.WriteString(`<choices class="java.util.Arrays$ArrayList"><a class="string-array">`)
    for _, option := range param.Options {
      buf.WriteString(`<string>`)
      _ = xml.EscapeText(buf, []byte(option))
      buf.WriteString(`</string>`)
    }
    buf.WriteString(`</a></choices>`)
    buf.WriteString(`</hudson.model.ChoiceParameterDefinition>`)
  default:
    buf.WriteString(`<hudson.model.StringParameterDefinition>`)
    writeParameterName(buf, param)
    buf.WriteString(`<defaultValue>`)
    _ = xml.EscapeText(buf, []byte(param.DefaultValue))
    buf.WriteString(`</defaultValue>`)
    buf.WriteString(`<trim>false</trim>`)
    buf.WriteString(`</hudson.model.StringParameterDefinition>`)
  }
}

func writeParameterName(buf *bytes.Buffer, param jobParameter) {
  buf.WriteString(`<name>`)
  _ = xml.EscapeText(buf, []byte(param.Name))
  buf.WriteString(`</name>`)
  buf.WriteString(`<description>`)
  _ = xml.EscapeText(buf, []byte(param.Description))
  buf.WriteString(`</description>`)
}

func resolveJobParameters(script string, inputs []string) ([]jobParameter, error) {
  loaded := LoadParams(script)
  params := make([]jobParameter, 0, len(loaded)+len(inputs))
  indexByName := map[string]int{}
  for _, input := range loaded {
    param := jobParameter{
      Name:         input.ID,
      Type:         input.Type,
      Description:  input.Description,
      DefaultValue: input.Default,
      Value:        parameterValue(input),
      Options:      input.Options,
    }
    indexByName[param.Name] = len(params)
    params = append(params, param)
  }
  for _, raw := range inputs {
    key, value, ok := strings.Cut(raw, "=")
    key = strings.TrimSpace(key)
    if !ok || key == "" {
      return nil, fmt.Errorf("invalid input %q (want KEY=VALUE)", raw)
    }
    if index, found := indexByName[key]; found {
      params[index].Value = value
      if params[index].Type == "boolean" {
        params[index].Value = booleanText(value)
      }
      continue
    }
    params = append(params, jobParameter{
      Name:         key,
      Type:         "string",
      DefaultValue: value,
      Value:        value,
    })
    indexByName[key] = len(params) - 1
  }
  return params, nil
}

func parameterValue(input event.WorkflowInput) string {
  if input.Type == "boolean" {
    return booleanText(input.Default)
  }
  if input.Default != "" {
    return input.Default
  }
  if input.Type == "choice" && len(input.Options) > 0 {
    return input.Options[0]
  }
  return ""
}

func booleanText(value string) string {
  if strings.EqualFold(strings.TrimSpace(value), "true") {
    return "true"
  }
  return "false"
}

func buildEndpoint(base, script string, inputs []string) (string, error) {
  params, err := resolveJobParameters(script, inputs)
  if err != nil {
    return "", err
  }
  if len(params) == 0 {
    return base + "/job/" + pipeboxJobName + "/build", nil
  }
  values := url.Values{}
  for _, param := range params {
    values.Set(param.Name, param.Value)
  }
  return base + "/job/" + pipeboxJobName + "/buildWithParameters?" + values.Encode(), nil
}

func triggerBuild(ctx context.Context, session *controllerSession, script string, inputs []string) (int, error) {
  next, err := nextBuildNumber(ctx, session)
  if err != nil {
    return 0, err
  }

  endpoint, err := buildEndpoint(session.baseURL, script, inputs)
  if err != nil {
    return 0, err
  }
  req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
  if err != nil {
    return 0, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return 0, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode >= 400 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
    return 0, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
  }

  deadline := time.Now().Add(2 * time.Minute)
  for time.Now().Before(deadline) {
    if err := ctx.Err(); err != nil {
      return 0, err
    }
    ready, err := buildStarted(ctx, session, next)
    if err != nil {
      return 0, err
    }
    if ready {
      return next, nil
    }
    time.Sleep(500 * time.Millisecond)
  }
  return 0, fmt.Errorf("build %d did not start", next)
}

func nextBuildNumber(ctx context.Context, session *controllerSession) (int, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, session.baseURL+"/job/"+pipeboxJobName+"/api/json?tree=nextBuildNumber", nil)
  if err != nil {
    return 0, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return 0, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
    return 0, fmt.Errorf("next build: status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
  }
  var payload struct {
    NextBuildNumber int `json:"nextBuildNumber"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
    return 0, err
  }
  if payload.NextBuildNumber < 1 {
    return 1, nil
  }
  return payload.NextBuildNumber, nil
}

func buildStarted(ctx context.Context, session *controllerSession, number int) (bool, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/job/%s/%d/api/json?tree=building", session.baseURL, pipeboxJobName, number), nil)
  if err != nil {
    return false, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return false, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode == http.StatusNotFound {
    return false, nil
  }
  if resp.StatusCode >= 300 {
    return false, fmt.Errorf("build status %d", resp.StatusCode)
  }
  return true, nil
}

func workspaceWorkflowPath(repoPath, workflow string) (string, error) {
  if filepath.IsAbs(workflow) {
    relativePath, err := filepath.Rel(repoPath, workflow)
    if err != nil {
      return "", fmt.Errorf("workflow must be inside the repo: %w", err)
    }
    return filepath.ToSlash(relativePath), nil
  }
  return filepath.ToSlash(workflow), nil
}
