package main

import (
	"fmt"
	"os"

	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/cli"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/install"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/run"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/uiopen"
	"github.com/MaxiGarcia13/pipebox/packages/runner/internal/workflows"
)

const version = "0.1.0"

func main() {
	parsed := cli.Parse(os.Args[1:])
	switch parsed.Kind {
	case cli.KindHelp:
		printUsage()
	case cli.KindVersion:
		fmt.Println(version)
	case cli.KindOpenUI:
		if err := uiopen.Open(parsed.Repo); err != nil {
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
		result, err := workflows.List(repoPath)
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
		if repoPath == "" || workflowPath == "" {
			fail(fmt.Errorf("usage: pipebox graph --repo <path> --workflow <file>"))
		}
		result, err := workflows.Graph(repoPath, workflowPath)
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
		job, _ := flagValue(args, "--job")
		secretFile, _ := flagValue(args, "--secret-file")
		secrets := flagValues(args, "--secret")
		inputs := flagValues(args, "--input")
		if repoPath == "" || workflowPath == "" {
			fail(fmt.Errorf("usage: pipebox run --repo <path> --workflow <file> [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]"))
		}
		exitCode, err := run.Start(run.Options{
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
		targetPath, err := install.Run()
		if err != nil {
			emitError(err)
			os.Exit(1)
		}
		if err := event.EmitStdout(event.Log{
			Type: event.TypeLog,
			Line: fmt.Sprintf("installed to %s", targetPath),
		}); err != nil {
			fail(err)
		}
	case "uninstall":
		targetPath, removed, err := install.Uninstall()
		if err != nil {
			emitError(err)
			os.Exit(1)
		}
		line := fmt.Sprintf("not installed at %s", targetPath)
		if removed {
			line = fmt.Sprintf("uninstalled %s", targetPath)
		}
		if err := event.EmitStdout(event.Log{
			Type: event.TypeLog,
			Line: line,
		}); err != nil {
			fail(err)
		}
	default:
		fail(fmt.Errorf("unknown command %q", command))
	}
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
	fmt.Fprintln(os.Stderr, `pipebox — local CI runner CLI

Usage:
  pipebox
  pipebox --repo <path>
  pipebox --help | -h
  pipebox --version | -v
  pipebox version
  pipebox doctor
  pipebox list --repo <path>
  pipebox graph --repo <path> --workflow <file>
  pipebox run --repo <path> --workflow <file> [--job <id>] [--input KEY=VALUE] [--secret KEY=VALUE] [--secret-file <path>]
  pipebox install
  pipebox uninstall

Desktop:
  bare pipebox (or pipebox --repo <path>) opens the desktop UI
  set PIPEBOX_APP to the Electron app/.app/binary if it is not in a default location

Secrets:
  run loads .env and .env.* from the repo when --secret-file is omitted

Inputs:
  pass workflow_dispatch inputs with --input KEY=VALUE (repeatable)`)
}
