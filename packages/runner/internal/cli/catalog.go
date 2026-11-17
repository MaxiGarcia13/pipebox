package cli

type ValueKind int

const (
  ValueNone ValueKind = iota
  ValuePath
  ValueWorkflow
  ValueJob
  ValueProvider
  ValueVersion
  ValueOpaque
)

type Flag struct {
  Names      []string
  ValueKind  ValueKind
  Repeatable bool
  Usage      string
}

type Command struct {
  Name        string
  Usage       string
  Flags       []Flag
  Hidden      bool
  TakesNoArgs bool
}

type Catalog struct {
  RootFlags []Flag
  Commands  []Command
  Notes     []string
}

func DefaultCatalog() Catalog {
  repoFlag := Flag{
    Names:     []string{"--repo"},
    ValueKind: ValuePath,
    Usage:     "repository path",
  }
  workflowFlag := Flag{
    Names:     []string{"--workflow"},
    ValueKind: ValueWorkflow,
    Usage:     "workflow file",
  }
  providerFlag := Flag{
    Names:     []string{"--provider"},
    ValueKind: ValueProvider,
    Usage:     "provider id",
  }

  return Catalog{
    RootFlags: []Flag{
      {
        Names:     []string{"--repo"},
        ValueKind: ValuePath,
        Usage:     "open desktop UI for this repository",
      },
      {
        Names:     []string{"-n", "--new"},
        ValueKind: ValueNone,
        Usage:     "open a new desktop window",
      },
      {
        Names:     []string{"-h", "--help", "help"},
        ValueKind: ValueNone,
        Usage:     "show help",
      },
      {
        Names:     []string{"-v", "--version"},
        ValueKind: ValueNone,
        Usage:     "show version",
      },
    },
    Commands: []Command{
      {
        Name:        "doctor",
        Usage:       "check local dependencies",
        TakesNoArgs: true,
      },
      {
        Name:  "list",
        Usage: "list workflows in a repository",
        Flags: []Flag{repoFlag},
      },
      {
        Name:  "graph",
        Usage: "print the job graph for a workflow",
        Flags: []Flag{repoFlag, workflowFlag, providerFlag},
      },
      {
        Name:  "run",
        Usage: "run a workflow locally",
        Flags: []Flag{
          repoFlag,
          workflowFlag,
          providerFlag,
          {
            Names:     []string{"--job"},
            ValueKind: ValueJob,
            Usage:     "job id",
          },
          {
            Names:      []string{"--input"},
            ValueKind:  ValueOpaque,
            Repeatable: true,
            Usage:      "workflow_dispatch input KEY=VALUE",
          },
          {
            Names:      []string{"--secret"},
            ValueKind:  ValueOpaque,
            Repeatable: true,
            Usage:      "secret KEY=VALUE",
          },
          {
            Names:     []string{"--secret-file"},
            ValueKind: ValuePath,
            Usage:     "env file with secrets",
          },
        },
      },
      {
        Name:        "install",
        Usage:       "install CLI to ~/.local/bin and fetch the desktop app",
        TakesNoArgs: true,
      },
      {
        Name:        "install-desktop",
        Usage:       "fetch the desktop app only",
        TakesNoArgs: true,
      },
      {
        Name:  "update",
        Usage: "update managed CLI and desktop app",
        Flags: []Flag{
          {
            Names:     []string{"--check"},
            ValueKind: ValueNone,
            Usage:     "report whether an update is available",
          },
          {
            Names:     []string{"--version"},
            ValueKind: ValueVersion,
            Usage:     "install a specific version",
          },
        },
      },
      {
        Name:        "uninstall",
        Usage:       "remove managed CLI and desktop app",
        TakesNoArgs: true,
      },
      {
        Name:        "version",
        Usage:       "print version",
        TakesNoArgs: true,
      },
      {
        Name:  "completion",
        Usage: "print shell completion script (zsh|bash|fish)",
      },
      {
        Name:   "__complete",
        Usage:  "internal completion helper",
        Hidden: true,
      },
    },
    Notes: []string{
      "Desktop:",
      "  bare pipebox (or pipebox --repo <path>) opens the desktop UI (reuses a running app)",
      "  -n / --new opens a new window instead of reusing",
      "  pipebox install / install-desktop fetches the UI from GitHub Releases",
      "  pipebox update replaces managed CLI installs (~/.local/bin and go install bin) + desktop app",
      "  set PIPEBOX_APP to override the desktop path",
      "",
      "Secrets:",
      "  run loads .env and .env.* from the repo when --secret-file is omitted",
      "",
      "Inputs:",
      "  pass workflow_dispatch inputs with --input KEY=VALUE (repeatable)",
      "",
      "Completions:",
      "  add to your shell rc after install: eval \"$(pipebox completion zsh)\"  # or bash",
      "  fish: pipebox completion fish > ~/.config/fish/completions/pipebox.fish",
    },
  }
}

func (c Catalog) Command(name string) (Command, bool) {
  for _, command := range c.Commands {
    if command.Name == name {
      return command, true
    }
  }
  return Command{}, false
}

func (c Catalog) VisibleCommands() []Command {
  visible := make([]Command, 0, len(c.Commands))
  for _, command := range c.Commands {
    if command.Hidden {
      continue
    }
    visible = append(visible, command)
  }
  return visible
}

func (c Catalog) IsSubcommand(name string) bool {
  _, ok := c.Command(name)
  return ok
}

func flagHasName(flag Flag, name string) bool {
  for _, candidate := range flag.Names {
    if candidate == name {
      return true
    }
  }
  return false
}

func findFlag(flags []Flag, name string) (Flag, bool) {
  for _, flag := range flags {
    if flagHasName(flag, name) {
      return flag, true
    }
  }
  return Flag{}, false
}
