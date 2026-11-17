package cli

func rootCandidates(catalog Catalog) []string {
  names := make([]string, 0, len(catalog.Commands)+len(catalog.RootFlags))
  for _, command := range catalog.VisibleCommands() {
    names = append(names, command.Name)
  }
  for _, flag := range catalog.RootFlags {
    for _, name := range flag.Names {
      if name == "help" {
        continue
      }
      names = append(names, name)
    }
  }
  return uniqueSorted(names)
}

func rootFlagCandidates(catalog Catalog, words []string) []string {
  awaiting, flag, ok := awaitingFlagValue(catalog.RootFlags, words)
  if ok && awaiting {
    return valueCandidates(flag, words, nil)
  }
  return unusedFlagNames(catalog.RootFlags, words)
}

func commandCandidates(
  command Command,
  words []string,
  resolver DynamicResolver,
) []string {
  if command.Name == "completion" {
    if len(words) == 0 {
      return uniqueSorted([]string{"zsh", "bash", "fish"})
    }
    return nil
  }

  awaiting, flag, ok := awaitingFlagValue(command.Flags, words)
  if ok && awaiting {
    return valueCandidates(flag, words, resolver)
  }
  return unusedFlagNames(command.Flags, words)
}

func valueCandidates(flag Flag, words []string, resolver DynamicResolver) []string {
  switch flag.ValueKind {
  case ValuePath:
    return []string{PathSentinel}
  case ValueOpaque, ValueNone, ValueVersion:
    return nil
  case ValueWorkflow:
    if resolver == nil {
      return nil
    }
    return resolver.Workflows(FlagValue(words, "--repo"))
  case ValueJob:
    if resolver == nil {
      return nil
    }
    return resolver.Jobs(
      FlagValue(words, "--repo"),
      FlagValue(words, "--workflow"),
      FlagValue(words, "--provider"),
    )
  case ValueProvider:
    if resolver == nil {
      return []string{"github", "jenkins"}
    }
    providers := resolver.Providers(FlagValue(words, "--repo"))
    if len(providers) == 0 {
      return []string{"github", "jenkins"}
    }
    return providers
  default:
    return nil
  }
}
