package server

import (
  "context"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "testing"
  "time"

  "github.com/mark3labs/mcp-go/mcp"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runcache"
  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/runner"
)

func TestRunWorkflowReturnsRunningThenCompletes(t *testing.T) {
  bin := buildFakePipebox(t)
  t.Setenv("FAKE_PIPEBOX_SLEEP_MS", "300")

  srv := New(&runner.Client{Bin: bin})
  req := mcp.CallToolRequest{}
  req.Params.Arguments = map[string]any{
    "repo":     "/repo",
    "workflow": "ci.yml",
  }

  started := time.Now()
  result, err := srv.runWorkflow(context.Background(), req)
  if err != nil {
    t.Fatal(err)
  }
  if time.Since(started) > 200*time.Millisecond {
    t.Fatalf("run_workflow blocked too long: %s", time.Since(started))
  }
  if result == nil || result.IsError {
    t.Fatalf("unexpected result: %#v", result)
  }

  text := toolText(t, result)
  if !strings.Contains(text, `"status":"running"`) || !strings.Contains(text, `"runId":"run-1"`) {
    t.Fatalf("expected running summary, got %s", text)
  }

  deadline := time.Now().Add(3 * time.Second)
  var summary runcache.Summary
  for {
    summary, err = srv.Cache.Summary("run-1")
    if err != nil {
      t.Fatal(err)
    }
    if summary.Status != runcache.StatusRunning {
      break
    }
    if time.Now().After(deadline) {
      t.Fatal("run stayed running too long")
    }
    time.Sleep(25 * time.Millisecond)
  }
  srv.Wait()
  if summary.Status != runcache.StatusFailed {
    t.Fatalf("final status: %#v", summary)
  }
  if len(summary.FailedSteps) != 1 {
    t.Fatalf("failed steps: %#v", summary.FailedSteps)
  }
}

func toolText(t *testing.T, result *mcp.CallToolResult) string {
  t.Helper()
  if len(result.Content) == 0 {
    t.Fatal("empty tool content")
  }
  text, ok := mcp.AsTextContent(result.Content[0])
  if !ok {
    t.Fatalf("content: %#v", result.Content[0])
  }
  return text.Text
}

func buildFakePipebox(t *testing.T) string {
  t.Helper()
  dir := t.TempDir()
  out := filepath.Join(dir, "fake-pipebox")
  cmd := exec.Command("go", "build", "-o", out, "./testdata/fake-pipebox")
  cmd.Dir = filepath.Join(moduleDir(t), "internal", "runner")
  cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
  if output, err := cmd.CombinedOutput(); err != nil {
    t.Fatalf("build fake-pipebox: %v\n%s", err, output)
  }
  return out
}

func moduleDir(t *testing.T) string {
  t.Helper()
  dir, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  for {
    if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
      return dir
    }
    parent := filepath.Dir(dir)
    if parent == dir {
      t.Fatal("go.mod not found")
    }
    dir = parent
  }
}
