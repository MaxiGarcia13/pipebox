package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/termui"
)

func runLifecycle(command string, args []string) {
  ui := termui.New(os.Stdout, os.Stderr)
  switch command {
  case "install":
    result, err := install.Run(version, ui)
    if err != nil {
      ui.Err(err)
      os.Exit(1)
    }
    ui.Ok(fmt.Sprintf("CLI installed (%s)", result.CLIPath))
    ui.Ok(fmt.Sprintf("Desktop installed (%s)", result.DesktopPath))
    ui.Ok("Shell completions installed")
    ui.Info("Open a new terminal tab to use completions.")
  case "install-desktop":
    desktopPath, err := install.InstallDesktop(version, ui)
    if err != nil {
      ui.Err(err)
      os.Exit(1)
    }
    ui.Ok(fmt.Sprintf("Desktop installed (%s)", desktopPath))
  case "update":
    runUpdate(ui, args)
  case "uninstall":
    runUninstall(ui)
  }
}

func runUpdate(ui *termui.UI, args []string) {
  result, err := install.Update(install.UpdateOptions{
    CurrentVersion: version,
    TargetVersion:  cli.FlagValue(args, "--version"),
    CheckOnly:      cli.HasFlag(args, "--check"),
    Progress:       ui,
  })
  if err != nil {
    ui.Err(err)
    os.Exit(1)
  }
  if result.UpToDate {
    ui.Info(fmt.Sprintf("Already up to date (%s)", result.CurrentVersion))
    return
  }
  if cli.HasFlag(args, "--check") {
    ui.Info(fmt.Sprintf(
      "Update available: %s → %s",
      result.CurrentVersion,
      result.TargetVersion,
    ))
    return
  }
  ui.Ok(fmt.Sprintf("CLI updated to %s (%s)", result.TargetVersion, result.CLIPath))
  ui.Ok(fmt.Sprintf("Desktop installed (%s)", result.DesktopPath))
  ui.Ok("Shell completions updated")
  ui.Info("Open a new terminal tab to use completions.")
}

func runUninstall(ui *termui.UI) {
  cliPaths, desktopPath, err := install.Uninstall()
  if err != nil {
    ui.Err(err)
    os.Exit(1)
  }
  if len(cliPaths) == 0 {
    ui.Info("CLI not installed")
  } else {
    for _, cliPath := range cliPaths {
      ui.Ok(fmt.Sprintf("Uninstalled CLI (%s)", cliPath))
    }
  }
  if desktopPath == "" {
    ui.Info("Desktop not installed")
  } else {
    ui.Ok(fmt.Sprintf("Uninstalled desktop (%s)", desktopPath))
  }
}
