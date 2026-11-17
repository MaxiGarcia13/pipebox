package cli

import (
  "fmt"
  "strings"
)

func Usage() string {
  return UsageCatalog(DefaultCatalog())
}

func CommandUsage(name string) string {
  return CommandUsageCatalog(DefaultCatalog(), name)
}

func UsageCatalog(catalog Catalog) string {
  var builder strings.Builder
  builder.WriteString("pipebox — local CI runner CLI\n\n")
  builder.WriteString("Usage:\n")
  builder.WriteString("  pipebox [command] [flags]\n\n")

  commands := catalog.VisibleCommands()
  builder.WriteString("Commands:\n")
  writeCommandRows(&builder, commands)

  builder.WriteString("\nFlags:\n")
  writeFlagRows(&builder, catalog.RootFlags)

  if len(catalog.Notes) > 0 {
    builder.WriteByte('\n')
    for _, note := range catalog.Notes {
      builder.WriteString(note)
      builder.WriteByte('\n')
    }
  }

  return strings.TrimRight(builder.String(), "\n")
}

func CommandUsageCatalog(catalog Catalog, name string) string {
  command, ok := catalog.Command(name)
  if !ok || command.Hidden {
    return UsageCatalog(catalog)
  }

  var builder strings.Builder
  builder.WriteString("Usage:\n")
  builder.WriteString("  ")
  builder.WriteString(commandSynopsis(command))
  builder.WriteByte('\n')

  if len(command.Flags) > 0 {
    builder.WriteString("\nFlags:\n")
    writeFlagRows(&builder, command.Flags)
  }

  return strings.TrimRight(builder.String(), "\n")
}

func WantsHelp(args []string) bool {
  for _, arg := range args {
    if arg == "-h" || arg == "--help" || arg == "help" {
      return true
    }
  }
  return false
}

func writeCommandRows(builder *strings.Builder, commands []Command) {
  width := 0
  for _, command := range commands {
    if length := len(command.Name); length > width {
      width = length
    }
  }
  for _, command := range commands {
    fmt.Fprintf(builder, "  %-*s  %s\n", width, command.Name, command.Usage)
  }
}

func writeFlagRows(builder *strings.Builder, flags []Flag) {
  labels := make([]string, len(flags))
  width := 0
  for index, flag := range flags {
    label := formatFlagLabel(flag)
    labels[index] = label
    if length := len(label); length > width {
      width = length
    }
  }
  for index, flag := range flags {
    fmt.Fprintf(builder, "  %-*s  %s\n", width, labels[index], flag.Usage)
  }
}

func formatFlagLabel(flag Flag) string {
  names := make([]string, 0, len(flag.Names))
  for _, name := range flag.Names {
    if strings.HasPrefix(name, "-") {
      names = append(names, name)
    }
  }
  label := strings.Join(names, ", ")
  if suffix := flagValueType(flag); suffix != "" {
    return label + " " + suffix
  }
  return label
}

func flagValueType(flag Flag) string {
  if flag.ValueKind == ValueNone {
    return ""
  }
  if flag.Repeatable {
    return "stringArray"
  }
  return "string"
}

func commandSynopsis(command Command) string {
  if command.Name == "completion" {
    return "pipebox completion <zsh|bash|fish>"
  }
  if len(command.Flags) == 0 {
    return "pipebox " + command.Name
  }
  return "pipebox " + command.Name + " [flags]"
}
