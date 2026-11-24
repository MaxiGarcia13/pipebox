package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/termui"
)

func runDoctor(args []string) {
  repoPath := cli.FlagValue(args, "--repo")
  if repoPath == "" {
    cwd, err := os.Getwd()
    if err != nil {
      fail(err)
    }
    repoPath = cwd
  }

  providers := discoveredProviderIDs(repoPath)
  result := doctor.Run(version, providers)
  if termui.IsTTY(os.Stdout) {
    printDoctor(termui.New(os.Stdout, os.Stderr), result)
  } else if err := event.EmitStdout(result); err != nil {
    fail(err)
  }
  if !result.OK {
    os.Exit(1)
  }
}

func discoveredProviderIDs(repoPath string) []string {
  discovered := provider.Discover(repoPath)
  ids := make([]string, 0, len(discovered))
  for _, item := range discovered {
    ids = append(ids, item.ID())
  }
  return ids
}

func printDoctor(ui *termui.UI, result event.Doctor) {
  ui.Info(fmt.Sprintf("pipebox %s", result.Version))
  ui.Info("")
  for _, check := range result.Checks {
    line := doctorCheckLine(check)
    if check.OK {
      ui.Ok(line)
    } else {
      ui.Fail(line)
    }
  }
  ui.Info("")
  if result.OK {
    ui.Info("doctor: ok")
  } else {
    ui.Fail("doctor: failed")
  }
}

func doctorCheckLine(check event.DoctorCheck) string {
  detail := check.Detail
  if detail == "" {
    detail = check.Path
  }
  if detail == "" {
    return check.Name
  }
  return fmt.Sprintf("%-14s %s", check.Name, detail)
}
