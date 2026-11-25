package main

import (
  "encoding/json"
  "fmt"
  "os"
  "strconv"
  "time"
)

func main() {
  if len(os.Args) < 2 {
    fmt.Fprintln(os.Stderr, "usage: fake-pipebox <command>")
    os.Exit(2)
  }

  switch os.Args[1] {
  case "list":
    emit(map[string]any{
      "type":      "workflow_list",
      "providers": []string{"github"},
      "workflows": []map[string]any{
        {
          "id":       "ci",
          "name":     "CI",
          "path":     ".github/workflows/ci.yml",
          "provider": "github",
        },
      },
    })
  case "run":
    if ms := os.Getenv("FAKE_PIPEBOX_SLEEP_MS"); ms != "" {
      if n, err := strconv.Atoi(ms); err == nil && n > 0 {
        time.Sleep(time.Duration(n) * time.Millisecond)
      }
    }
    emit(map[string]any{
      "type":     "run_started",
      "repo":     flagValue("--repo"),
      "workflow": flagValue("--workflow"),
    })
    for i := 0; i < 200; i++ {
      emit(map[string]any{
        "type": "log",
        "job":  "build",
        "step": "test",
        "line": fmt.Sprintf("noise-line-%d", i),
      })
    }
    emit(map[string]any{
      "type":       "step_finished",
      "job":        "build",
      "step":       "test",
      "status":     "failed",
      "durationMs": 10,
    })
    emit(map[string]any{
      "type":     "run_finished",
      "status":   "failed",
      "exitCode": 1,
    })
    os.Exit(1)
  default:
    fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
    os.Exit(2)
  }
}

func emit(value any) {
  if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
    fmt.Fprintf(os.Stderr, "encode: %v\n", err)
    os.Exit(1)
  }
}

func flagValue(name string) string {
  for i := 2; i < len(os.Args)-1; i++ {
    if os.Args[i] == name {
      return os.Args[i+1]
    }
  }
  return ""
}
