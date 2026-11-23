package jenkins

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider/jenkins/jfrparse"
)

type consoleBuffer struct {
  pending string
}

func (buffer *consoleBuffer) push(chunk string) []string {
  if chunk == "" {
    return nil
  }
  text := buffer.pending + chunk
  parts := strings.Split(text, "\n")
  buffer.pending = parts[len(parts)-1]
  return parts[:len(parts)-1]
}

func (buffer *consoleBuffer) flush() string {
  line := buffer.pending
  buffer.pending = ""
  return line
}

func streamBuild(ctx context.Context, session *controllerSession, buildNumber int, jobs []event.JobNode) (int, string) {
  if !isScriptedGraph(jobs) && session.runHome != "" {
    exitCode, status, fallback := streamStages(ctx, session, buildNumber, jobs)
    if !fallback {
      return exitCode, status
    }
  }
  return streamConsole(ctx, session, buildNumber)
}

func isScriptedGraph(jobs []event.JobNode) bool {
  return len(jobs) == 1 && jobs[0].ID == "pipeline"
}

func streamConsole(ctx context.Context, session *controllerSession, buildNumber int) (int, string) {
  parser := jfrparse.NewParser()
  var buffer consoleBuffer
  start := int64(0)
  for {
    if err := ctx.Err(); err != nil {
      return 130, "cancelled"
    }

    moreData, textSize, err := fetchConsole(ctx, session, buildNumber, start)
    if err != nil {
      _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("console: %v", err)})
      return 1, "failed"
    }
    if moreData != "" {
      emitConsoleLines(parser, buffer.push(moreData))
      start = textSize
    }

    done, result, err := buildFinished(ctx, session, buildNumber)
    if err != nil {
      _ = event.EmitStdout(event.Error{Type: event.TypeError, Message: fmt.Sprintf("build poll: %v", err)})
      return 1, "failed"
    }
    if done {
      if leftover, _, err := fetchConsole(ctx, session, buildNumber, start); err == nil && leftover != "" {
        emitConsoleLines(parser, buffer.push(leftover))
      }
      if tail := buffer.flush(); tail != "" {
        emitConsoleLines(parser, []string{tail})
      }
      emitConsoleLines(parser, []string{"Finished: " + result})
      switch strings.ToUpper(result) {
      case "SUCCESS":
        return 0, "passed"
      case "ABORTED":
        return 130, "cancelled"
      default:
        return 1, "failed"
      }
    }

    time.Sleep(consolePollInterval)
  }
}

func emitConsoleLines(parser *jfrparse.Parser, lines []string) {
  for _, line := range lines {
    for _, item := range parser.HandleLine(line) {
      _ = event.EmitStdout(item)
    }
  }
}

func fetchConsole(ctx context.Context, session *controllerSession, buildNumber int, start int64) (string, int64, error) {
  target := fmt.Sprintf("%s/job/%s/%d/logText/progressiveText?start=%d", session.baseURL, pipeboxJobName, buildNumber, start)
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
  if err != nil {
    return "", start, err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return "", start, err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode == http.StatusNotFound {
    return "", start, nil
  }
  if resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
    return "", start, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
  }
  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return "", start, err
  }
  textSize := start + int64(len(body))
  if raw := resp.Header.Get("X-Text-Size"); raw != "" {
    if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
      textSize = parsed
    }
  }
  return string(body), textSize, nil
}

func buildFinished(ctx context.Context, session *controllerSession, buildNumber int) (done bool, result string, err error) {
  target := fmt.Sprintf("%s/job/%s/%d/api/json?tree=building,result", session.baseURL, pipeboxJobName, buildNumber)
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
  if err != nil {
    return false, "", err
  }
  resp, err := session.client.Do(req)
  if err != nil {
    return false, "", err
  }
  defer func() { _ = resp.Body.Close() }()
  if resp.StatusCode == http.StatusNotFound {
    return false, "", nil
  }
  if resp.StatusCode >= 300 {
    return false, "", fmt.Errorf("status %d", resp.StatusCode)
  }
  var payload struct {
    Building bool   `json:"building"`
    Result   string `json:"result"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
    return false, "", err
  }
  if payload.Building || payload.Result == "" {
    return false, "", nil
  }
  return true, payload.Result, nil
}
