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
    Usage:     "Repository path",
  }
  workflowFlag := Flag{
    Names:     []string{"--workflow"},
    ValueKind: ValueWorkflow,
    Usage:     "Workflow file",
  }
  providerFlag := Flag{
    Names:     []string{"--provider"},
    ValueKind: ValueProvider,
    Usage:     "Provider id",
  }

  return Catalog{
    RootFlags: []Flag{
      {
        Names:     []string{"-h", "--help", "help"},
        ValueKind: ValueNone,
        Usage:     "Show help",
      },
      {
        Names:     []string{"-v", "--version"},
        ValueKind: ValueNone,
        Usage:     "Print version",
      },
      {
        Names:     []string{"-n", "--new"},
        ValueKind: ValueNone,
        Usage:     "Open a new desktop window",
      },
      {
        Names:     []string{"--repo"},
        ValueKind: ValuePath,
        Usage:     "Repository path",
      },
    },
    Commands: []Command{
      {
        Name:  "doctor",
        Usage: "Run environment diagnostics",
        Flags: []Flag{repoFlag},
      },
      {
        Name:  "list",
        Usage: "List workflows in a repo",
        Flags: []Flag{repoFlag},
      },
      {
        Name:  "graph",
        Usage: "Show workflow graph",
        Flags: []Flag{repoFlag, workflowFlag, providerFlag},
      },
      {
        Name:  "run",
        Usage: "Run a workflow",
        Flags: []Flag{
          repoFlag,
          workflowFlag,
          providerFlag,
          {
            Names:     []string{"--job"},
            ValueKind: ValueJob,
            Usage:     "Job id",
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
            Usage:      "Secret KEY=VALUE",
          },
          {
            Names:     []string{"--secret-file"},
            ValueKind: ValuePath,
            Usage:     "Secrets file path",
          },
        },
      },
      {
        Name:        "install",
        Usage:       "Install managed CLI + desktop app",
        TakesNoArgs: true,
      },
      {
        Name:        "install-desktop",
        Usage:       "Install desktop app only",
        TakesNoArgs: true,
      },
      {
        Name:  "update",
        Usage: "Update CLI and desktop app",
        Flags: []Flag{
          {
            Names:     []string{"--check"},
            ValueKind: ValueNone,
            Usage:     "Report whether an update is available",
          },
          {
            Names:     []string{"--version"},
            ValueKind: ValueVersion,
            Usage:     "Install a specific version",
          },
        },
      },
      {
        Name:        "uninstall",
        Usage:       "Remove managed installs",
        TakesNoArgs: true,
      },
      {
        Name:        "version",
        Usage:       "Print version",
        TakesNoArgs: true,
      },
      {
        Name:  "completion",
        Usage: "Generate shell completion scripts",
      },
      {
        Name:   "__complete",
        Usage:  "internal completion helper",
        Hidden: true,
      },
    },
    Notes: []string{
      "Desktop:",
      "  Bare `pipebox` (or `pipebox --repo <path>`) opens the desktop UI.",
      "  `-n` / `--new` opens a new window instead of reusing.",
      "  `install` / `install-desktop` fetch the UI from GitHub Releases.",
      "  `update` replaces managed CLI installs + desktop app.",
      "  Set `PIPEBOX_APP` to override the desktop path.",
      "",
      "Secrets:",
      "  `run` loads `.env` and `.env.*` from the repo when `--secret-file` is omitted.",
      "",
      "Inputs:",
      "  Pass workflow_dispatch inputs with `--input KEY=VALUE` (repeatable).",
      "",
      "Examples:",
      "  pipebox doctor",
      "  pipebox doctor --repo <path>",
      "  pipebox list --repo <path>",
      "  pipebox run --repo <path> --workflow <file>",
      "  pipebox completion zsh",
      "",
      "Completions:",
      "  Add to your shell rc after install: eval \"$(pipebox completion zsh)\"  # or bash",
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
