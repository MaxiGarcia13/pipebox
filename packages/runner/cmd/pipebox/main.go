package main

import (
  "context"
  "fmt"
  "os"
  "strings"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/uiopen"
)

const version = "0.11.0"

func main() {
  parsed := cli.Parse(os.Args[1:])
  switch parsed.Kind {
  case cli.KindHelp:
    printUsage()
  case cli.KindVersion:
    fmt.Println(version)
  case cli.KindOpenUI:
    if err := uiopen.Open(parsed.Repo, parsed.NewWindow); err != nil {
      fmt.Fprintf(os.Stderr, "pipebox: %v\n", err)
      os.Exit(1)
    }
  case cli.KindError:
    fail(parsed.Err)
  case cli.KindSubcommand:
    runSubcommand(parsed.Command, parsed.CommandArgs)
  }
}

func runSubcommand(command string, args []string) {
  switch command {
  case "doctor":
    if err := event.EmitStdout(doctor.Run(version)); err != nil {
      fail(err)
    }
  case "list":
    repoPath, err := flagValue(args, "--repo")
    if err != nil || repoPath == "" {
      fail(fmt.Errorf("usage: pipebox list --repo <path>"))
    }
    result, err := provider.List(repoPath)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    if err := event.EmitStdout(result); err != nil {
      fail(err)
    }
  case "graph":
    repoPath, _ := flagValue(args, "--repo")
    workflowPath, _ := flagValue(args, "--workflow")
    providerHint, _ := flagValue(args, "--provider")
    if repoPath == "" || workflowPath == "" {
      fail(fmt.Errorf("usage: pipebox graph --repo <path> --workflow <file> [--provider <id>]"))
    }
    resolved, err := provider.Resolve(repoPath, workflowPath, providerHint)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    result, err := resolved.Graph(repoPath, workflowPath)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    if err := event.EmitStdout(result); err != nil {
      fail(err)
    }
  case "run":
    repoPath, _ := flagValue(args, "--repo")
    workflowPath, _ := flagValue(args, "--workflow")
    providerHint, _ := flagValue(args, "--provider")
    job, _ := flagValue(args, "--job")
    secretFile, _ := flagValue(args, "--secret-file")
    secrets := flagValues(args, "--secret")
    inputs := flagValues(args, "--input")
    if repoPath == "" || workflowPath == "" {
      fail(fmt.Errorf("usage: pipebox run --repo <path> --workflow <file> [--provider <id>] [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]"))
    }
    resolved, err := provider.Resolve(repoPath, workflowPath, providerHint)
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    exitCode, err := resolved.Run(context.Background(), provider.RunOptions{
      Repo:       repoPath,
      Workflow:   workflowPath,
      Job:        job,
      Secrets:    secrets,
      SecretFile: secretFile,
      Inputs:     inputs,
    })
    if err != nil {
      emitError(err)
      if exitCode == 0 {
        exitCode = 1
      }
    }
    os.Exit(exitCode)
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
    checkOnly := hasFlag(args, "--check")
    targetVersion, _ := flagValue(args, "--version")
    result, err := install.Update(install.UpdateOptions{
      CurrentVersion: version,
      TargetVersion:  targetVersion,
      CheckOnly:      checkOnly,
    })
    if err != nil {
      emitError(err)
      os.Exit(1)
    }
    if result.UpToDate {
      emitInstallLog(fmt.Sprintf("already up to date (%s)", result.CurrentVersion))
      return
    }
    if checkOnly {
      emitInstallLog(fmt.Sprintf("update available: %s → %s", result.CurrentVersion, result.TargetVersion))
      return
    }
    emitInstallLog(fmt.Sprintf("cli updated to %s (%s)", result.TargetVersion, result.CLIPath))
    emitInstallLog(fmt.Sprintf("desktop installed to %s", result.DesktopPath))
    emitInstallLog("shell completions updated; open a new terminal tab to use them")
  case "uninstall":
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
  case "completion":
    if len(args) != 1 {
      fail(fmt.Errorf("usage: pipebox completion <zsh|bash|fish>"))
    }
    script, err := cli.CompletionScript(args[0])
    if err != nil {
      fail(err)
    }
    fmt.Print(script)
  case "__complete":
    completeArgs := args
    if len(completeArgs) > 0 && completeArgs[0] == "--" {
      completeArgs = completeArgs[1:]
    }
    for _, candidate := range cli.Complete(completeArgs, cli.CompleteOptions{
      Catalog:  cli.DefaultCatalog(),
      Resolver: cli.ProviderResolver{},
    }) {
      fmt.Println(candidate)
    }
  default:
    fail(fmt.Errorf("unknown command %q", command))
  }
}

func hasFlag(args []string, name string) bool {
  for _, arg := range args {
    if arg == name {
      return true
    }
  }
  return false
}

func flagValue(args []string, name string) (string, error) {
  values := flagValues(args, name)
  if len(values) == 0 {
    return "", nil
  }
  return values[0], nil
}

func flagValues(args []string, name string) []string {
  values := []string{}
  prefix := name + "="
  for index := 0; index < len(args); index++ {
    arg := args[index]
    if arg == name {
      if index+1 >= len(args) {
        continue
      }
      values = append(values, args[index+1])
      index++
      continue
    }
    if len(arg) > len(prefix) && arg[:len(prefix)] == prefix {
      values = append(values, arg[len(prefix):])
    }
  }
  return values
}

func emitInstallLog(line string) {
  if err := event.EmitStdout(event.Log{
    Type: event.TypeLog,
    Line: line,
  }); err != nil {
    fail(err)
  }
}

func emitError(err error) {
  _ = event.EmitStdout(event.Error{
    Type:    event.TypeError,
    Message: err.Error(),
  })
}

func fail(err error) {
  emitError(err)
  printUsage()
  os.Exit(2)
}

func printUsage() {
  fmt.Fprintln(os.Stderr, strings.TrimRight(cli.Usage(), "\n"))
}
