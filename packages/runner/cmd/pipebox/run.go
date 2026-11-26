package main

import (
  "context"
  "fmt"
  "os"

  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/provider"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/termui"
)

func runList(args []string) {
  repoPath, err := cli.RepoFromArgs(args)
  if err != nil {
    fail(err)
  }
  result, err := provider.List(repoPath)
  if err != nil {
    emitError(err)
    os.Exit(1)
  }
  if termui.IsTTY(os.Stdout) {
    fmt.Println(formatWorkflowList(result))
  } else if err := event.EmitStdout(result); err != nil {
    fail(err)
  }
  suggestUpdateIfNeeded()
}

func runGraph(args []string) {
  repoPath, err := cli.RepoFromArgs(args)
  if err != nil {
    fail(err)
  }
  workflowRef := cli.FlagValue(args, "--workflow")
  providerHint := cli.FlagValue(args, "--provider")
  if workflowRef == "" {
    fail(fmt.Errorf("usage: pipebox graph --workflow <id|path> [--provider <id>]"))
  }
  workflowPath, err := provider.ResolveWorkflow(repoPath, workflowRef, providerHint)
  if err != nil {
    emitError(err)
    os.Exit(1)
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
  if termui.IsTTY(os.Stdout) {
    fmt.Println(formatJobGraph(result))
  } else if err := event.EmitStdout(result); err != nil {
    fail(err)
  }
}

func runWorkflow(args []string) {
  repoPath, err := cli.RepoFromArgs(args)
  if err != nil {
    fail(err)
  }
  workflowRef := cli.FlagValue(args, "--workflow")
  providerHint := cli.FlagValue(args, "--provider")
  job := cli.FlagValue(args, "--job")
  secretFile := cli.FlagValue(args, "--secret-file")
  secrets := cli.FlagValues(args, "--secret")
  inputs := cli.FlagValues(args, "--input")
  if workflowRef == "" {
    fail(fmt.Errorf("usage: pipebox run --workflow <id|path> [--provider <id>] [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]"))
  }
  workflowPath, err := provider.ResolveWorkflow(repoPath, workflowRef, providerHint)
  if err != nil {
    emitError(err)
    os.Exit(1)
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
  suggestUpdateIfNeeded()
  os.Exit(exitCode)
}
