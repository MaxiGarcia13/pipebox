package main

import (
  "context"
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
)

func runList(args []string) {
  repoPath := cli.FlagValue(args, "--repo")
  if repoPath == "" {
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
}

func runGraph(args []string) {
  repoPath := cli.FlagValue(args, "--repo")
  workflowPath := cli.FlagValue(args, "--workflow")
  providerHint := cli.FlagValue(args, "--provider")
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
}

func runWorkflow(args []string) {
  repoPath := cli.FlagValue(args, "--repo")
  workflowPath := cli.FlagValue(args, "--workflow")
  providerHint := cli.FlagValue(args, "--provider")
  job := cli.FlagValue(args, "--job")
  secretFile := cli.FlagValue(args, "--secret-file")
  secrets := cli.FlagValues(args, "--secret")
  inputs := cli.FlagValues(args, "--input")
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
}
