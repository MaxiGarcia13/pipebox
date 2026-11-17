package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
)

func runLifecycle(command string, args []string) {
  switch command {
  case "install":
    result, err := install.Run(version)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    emitInstallLog(fmt.Sprintf("cli installed to %s", result.CLIPath))
    emitInstallLog(fmt.Sprintf("desktop installed to %s", result.DesktopPath))
    emitInstallLog("shell completions installed; open a new terminal tab to use them")
  case "install-desktop":
    desktopPath, err := install.InstallDesktop(version)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    emitInstallLog(fmt.Sprintf("desktop installed to %s", desktopPath))
  case "update":
    runUpdate(args)
  case "uninstall":
    runUninstall()
  }
}

func runUpdate(args []string) {
  result, err := install.Update(install.UpdateOptions{
    CurrentVersion: version,
    TargetVersion:  cli.FlagValue(args, "--version"),
    CheckOnly:      cli.HasFlag(args, "--check"),
  })
  if err != nil {
    emitError(err)
    os.Exit(1)
  }
  if result.UpToDate {
    emitInstallLog(fmt.Sprintf("already up to date (%s)", result.CurrentVersion))
    return
  }
  if cli.HasFlag(args, "--check") {
    emitInstallLog(fmt.Sprintf(
      "update available: %s → %s",
      result.CurrentVersion,
      result.TargetVersion,
    ))
    return
  }
  emitInstallLog(fmt.Sprintf("cli updated to %s (%s)", result.TargetVersion, result.CLIPath))
  emitInstallLog(fmt.Sprintf("desktop installed to %s", result.DesktopPath))
  emitInstallLog("shell completions updated; open a new terminal tab to use them")
}

func runUninstall() {
  cliPaths, desktopPath, err := install.Uninstall()
  if err != nil {
    emitError(err)
    os.Exit(1)
  }
  if len(cliPaths) == 0 {
    emitInstallLog("cli not installed")
  } else {
    for _, cliPath := range cliPaths {
      emitInstallLog(fmt.Sprintf("uninstalled cli %s", cliPath))
    }
  }
  if desktopPath == "" {
    emitInstallLog("desktop not installed")
  } else {
    emitInstallLog(fmt.Sprintf("uninstalled desktop %s", desktopPath))
  }
}
