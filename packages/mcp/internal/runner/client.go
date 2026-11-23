package runner

import (
  "bufio"
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "io"
  "os"
  "os/exec"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/mcp/internal/event"
)

type Client struct {
  Bin string
}

func Default() *Client {
  return &Client{Bin: resolveBin()}
}

func resolveBin() string {
  if bin := strings.TrimSpace(os.Getenv("PIPEBOX_BIN")); bin != "" {
    return bin
  }
  if path, err := exec.LookPath("pipebox"); err == nil {
    return path
  }
  return "pipebox"
}

type ListOptions struct {
  Repo string
}

type GraphOptions struct {
  Repo     string
  Workflow string
  Provider string
}

type RunOptions struct {
  Repo     string
  Workflow string
  Provider string
  Job      string
  Inputs   []string
}

func (c *Client) Doctor(ctx context.Context) (*event.Doctor, error) {
  raw, err := c.runCollect(ctx, nil, "doctor")
  if err != nil {
    return nil, err
  }
  var out event.Doctor
  if err := decodeLastOfType(raw, event.TypeDoctor, &out); err != nil {
    return nil, err
  }
  return &out, nil
}

func (c *Client) List(ctx context.Context, opts ListOptions) (*event.WorkflowList, error) {
  if opts.Repo == "" {
    return nil, fmt.Errorf("repo is required")
  }
  raw, err := c.runCollect(ctx, nil, "list", "--repo", opts.Repo)
  if err != nil {
    return nil, err
  }
  var out event.WorkflowList
  if err := decodeLastOfType(raw, event.TypeWorkflowList, &out); err != nil {
    return nil, err
  }
  return &out, nil
}

func (c *Client) Graph(ctx context.Context, opts GraphOptions) (*event.JobGraph, error) {
  if opts.Repo == "" || opts.Workflow == "" {
    return nil, fmt.Errorf("repo and workflow are required")
  }
  args := []string{"graph", "--repo", opts.Repo, "--workflow", opts.Workflow}
  if opts.Provider != "" {
    args = append(args, "--provider", opts.Provider)
  }
  raw, err := c.runCollect(ctx, nil, args...)
  if err != nil {
    return nil, err
  }
  var out event.JobGraph
  if err := decodeLastOfType(raw, event.TypeJobGraph, &out); err != nil {
    return nil, err
  }
  return &out, nil
}

func (c *Client) Run(ctx context.Context, opts RunOptions, onEvent func(json.RawMessage) error) (int, error) {
  if opts.Repo == "" || opts.Workflow == "" {
    return 1, fmt.Errorf("repo and workflow are required")
  }
  args := []string{"run", "--repo", opts.Repo, "--workflow", opts.Workflow}
  if opts.Provider != "" {
    args = append(args, "--provider", opts.Provider)
  }
  if opts.Job != "" {
    args = append(args, "--job", opts.Job)
  }
  for _, input := range opts.Inputs {
    args = append(args, "--input", input)
  }
  return c.runStream(ctx, onEvent, args...)
}

func (c *Client) runCollect(ctx context.Context, onEvent func(json.RawMessage) error, args ...string) ([]byte, error) {
  var buf bytes.Buffer
  exitCode, err := c.runStream(ctx, func(raw json.RawMessage) error {
    buf.Write(raw)
    buf.WriteByte('\n')
    if onEvent != nil {
      return onEvent(raw)
    }
    return nil
  }, args...)
  if err != nil && exitCode == 0 {
    return buf.Bytes(), err
  }
  if err != nil {
    return buf.Bytes(), err
  }
  if exitCode != 0 {
    return buf.Bytes(), fmt.Errorf("pipebox %s: exit %d", args[0], exitCode)
  }
  return buf.Bytes(), nil
}

func (c *Client) runStream(ctx context.Context, onEvent func(json.RawMessage) error, args ...string) (int, error) {
  cmd := exec.CommandContext(ctx, c.Bin, args...)
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return 1, err
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return 1, err
  }

  if err := cmd.Start(); err != nil {
    return 1, fmt.Errorf("start pipebox: %w", err)
  }

  errCh := make(chan error, 2)
  go func() {
    errCh <- scanNDJSON(stdout, onEvent)
  }()
  go func() {
    errCh <- drainStderr(stderr)
  }()

  var scanErr error
  for i := 0; i < 2; i++ {
    if err := <-errCh; err != nil && scanErr == nil {
      scanErr = err
    }
  }

  waitErr := cmd.Wait()
  exitCode := 0
  if waitErr != nil {
    if exitErr, ok := waitErr.(*exec.ExitError); ok {
      exitCode = exitErr.ExitCode()
    } else {
      return 1, waitErr
    }
  }
  if scanErr != nil {
    return exitCode, scanErr
  }
  return exitCode, nil
}

func scanNDJSON(r io.Reader, onEvent func(json.RawMessage) error) error {
  scanner := bufio.NewScanner(r)
  scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for scanner.Scan() {
    line := bytes.TrimSpace(scanner.Bytes())
    if len(line) == 0 {
      continue
    }
    if !json.Valid(line) {
      continue
    }
    raw := append(json.RawMessage(nil), line...)
    if onEvent != nil {
      if err := onEvent(raw); err != nil {
        return err
      }
    }
  }
  return scanner.Err()
}

func drainStderr(r io.Reader) error {
  _, err := io.Copy(io.Discard, r)
  return err
}

func decodeLastOfType(raw []byte, typ string, dest any) error {
  scanner := bufio.NewScanner(bytes.NewReader(raw))
  var match json.RawMessage
  for scanner.Scan() {
    line := bytes.TrimSpace(scanner.Bytes())
    if len(line) == 0 {
      continue
    }
    var head struct {
      Type string `json:"type"`
    }
    if err := json.Unmarshal(line, &head); err != nil {
      continue
    }
    if head.Type == typ {
      match = append(json.RawMessage(nil), line...)
    }
  }
  if err := scanner.Err(); err != nil {
    return err
  }
  if match == nil {
    return fmt.Errorf("missing %s event in pipebox output", typ)
  }
  return json.Unmarshal(match, dest)
}
