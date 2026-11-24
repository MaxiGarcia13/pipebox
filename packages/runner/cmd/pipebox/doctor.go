package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/termui"
)

func runDoctor() {
  result := doctor.Run(version)
  if termui.IsTTY(os.Stdout) {
    printDoctor(termui.New(os.Stdout, os.Stderr), result)
  } else if err := event.EmitStdout(result); err != nil {
    fail(err)
  }
  if !result.OK {
    os.Exit(1)
  }
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
