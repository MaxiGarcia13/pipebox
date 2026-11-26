package main

import (
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/uiopen"
)

const version = "0.24.0"

func main() {
  parsed := cli.Parse(os.Args[1:])
  switch parsed.Kind {
  case cli.KindHelp:
    printUsage()
  case cli.KindVersion:
    fmt.Println(version)
  case cli.KindOpenUI:
    suggestUpdateIfNeeded()
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
  if command != "__complete" && cli.WantsHelp(args) {
    printCommandUsage(command)
    return
  }

  switch command {
  case "doctor":
    runDoctor(args)
  case "list":
    runList(args)
  case "graph":
    runGraph(args)
  case "run":
    runWorkflow(args)
  case "install", "install-desktop", "install-mcp", "update", "uninstall":
    runLifecycle(command, args)
  case "mcp":
    runMCP(args)
  case "completion", "__complete":
    runCompletion(command, args)
  default:
    fail(fmt.Errorf("unknown command %q", command))
  }
}
