package cli

import (
  "fmt"
  "strings"
)

func Usage() string {
  return UsageCatalog(DefaultCatalog())
}

func UsageCatalog(catalog Catalog) string {
  var builder strings.Builder
  builder.WriteString("pipebox — local CI runner CLI\n\n")
  builder.WriteString("Usage:\n")
  builder.WriteString("  pipebox\n")

  for _, flag := range catalog.RootFlags {
    line := formatRootFlagUsage(flag)
    if line != "" {
      builder.WriteString("  pipebox ")
      builder.WriteString(line)
      builder.WriteByte('\n')
    }
  }

  for _, command := range catalog.VisibleCommands() {
    builder.WriteString("  pipebox ")
    builder.WriteString(formatCommandUsage(command))
    builder.WriteByte('\n')
  }

  if len(catalog.Notes) > 0 {
    builder.WriteByte('\n')
    for _, note := range catalog.Notes {
      builder.WriteString(note)
      builder.WriteByte('\n')
    }
  }

  return strings.TrimRight(builder.String(), "\n")
}

func formatRootFlagUsage(flag Flag) string {
  switch {
  case flagHasName(flag, "--repo"):
    return "--repo <path>"
  case flagHasName(flag, "--new"):
    return "-n | --new [--repo <path>]"
  case flagHasName(flag, "--help"):
    return "--help | -h"
  case flagHasName(flag, "--version"):
    return "--version | -v"
  default:
    return strings.Join(flag.Names, " | ")
  }
}

func formatCommandUsage(command Command) string {
  if command.Name == "completion" {
    return "completion <zsh|bash|fish>"
  }
  if len(command.Flags) == 0 {
    return command.Name
  }

  parts := []string{command.Name}
  for _, flag := range command.Flags {
    parts = append(parts, formatFlagUsage(flag))
  }
  return strings.Join(parts, " ")
}

func formatFlagUsage(flag Flag) string {
  primary := flag.Names[0]
  switch flag.ValueKind {
  case ValueNone:
    return fmt.Sprintf("[%s]", primary)
  case ValueOpaque:
    if flag.Repeatable {
      return fmt.Sprintf("[%s KEY=VALUE]", primary)
    }
    return fmt.Sprintf("[%s <value>]", primary)
  case ValuePath:
    if primary == "--repo" {
      return fmt.Sprintf("%s <path>", primary)
    }
    if primary == "--secret-file" {
      return fmt.Sprintf("[%s <path>]", primary)
    }
    return fmt.Sprintf("[%s <path>]", primary)
  case ValueWorkflow:
    return fmt.Sprintf("%s <file>", primary)
  case ValueJob:
    return fmt.Sprintf("[%s <id>]", primary)
  case ValueProvider:
    return fmt.Sprintf("[%s <id>]", primary)
  case ValueVersion:
    return fmt.Sprintf("[%s <v>]", primary)
  default:
    return fmt.Sprintf("[%s]", primary)
  }
}
