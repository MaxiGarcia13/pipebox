package main

import (
  "fmt"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
)

func runCompletion(command string, args []string) {
  switch command {
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
  }
}
