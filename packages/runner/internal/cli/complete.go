package cli

import "strings"

const PathSentinel = "__PIPEBOX_PATH__"

type DynamicResolver interface {
  Workflows(repo string) []string
  Jobs(repo, workflow, provider string) []string
  Providers(repo string) []string
}

type CompleteOptions struct {
  Catalog  Catalog
  Resolver DynamicResolver
}

func Complete(args []string, options CompleteOptions) []string {
  catalog := options.Catalog
  if len(catalog.Commands) == 0 {
    catalog = DefaultCatalog()
  }

  partial := ""
  words := args
  if len(words) > 0 {
    partial = words[len(words)-1]
    words = words[:len(words)-1]
  }

  // --flag=partial keeps the flag token in words and uses the inline value as partial.
  if name, value, ok := splitInlineFlag(partial); ok {
    words = append(words, name+"=")
    partial = value
  }

  if len(words) == 0 {
    return filterPrefix(rootCandidates(catalog), partial)
  }

  first := words[0]
  if command, ok := catalog.Command(first); ok {
    return filterPrefix(commandCandidates(command, words[1:], options.Resolver), partial)
  }

  if strings.HasPrefix(first, "-") || first == "help" {
    return filterPrefix(rootFlagCandidates(catalog, words), partial)
  }

  return filterPrefix(rootCandidates(catalog), partial)
}
