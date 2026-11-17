package cli

import (
	"sort"
	"strings"
)

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

func awaitingFlagValue(flags []Flag, words []string) (awaiting bool, flag Flag, ok bool) {
	for index := 0; index < len(words); index++ {
		word := words[index]
		if name, value, hasValue := splitInlineFlag(word); hasValue {
			if matched, found := findFlag(flags, name); found && matched.ValueKind != ValueNone {
				flag = matched
				ok = true
				awaiting = value == ""
			}
			continue
		}
		if matched, found := findFlag(flags, word); found {
			flag = matched
			ok = true
			if matched.ValueKind == ValueNone {
				awaiting = false
				continue
			}
			if index+1 >= len(words) {
				awaiting = true
				return awaiting, flag, ok
			}
			index++
			awaiting = false
			continue
		}
	}
	return awaiting, flag, ok
}

func splitInlineFlag(word string) (name, value string, ok bool) {
	if !strings.HasPrefix(word, "--") {
		return "", "", false
	}
	name, value, found := strings.Cut(word, "=")
	if !found {
		return "", "", false
	}
	return name, value, true
}

func unusedFlagNames(flags []Flag, words []string) []string {
	used := usedFlagNames(flags, words)
	names := make([]string, 0)
	for _, flag := range flags {
		primary := flag.Names[0]
		if !flag.Repeatable {
			if _, seen := used[primary]; seen {
				continue
			}
		}
		names = append(names, flag.Names...)
	}
	return uniqueSorted(names)
}

func usedFlagNames(flags []Flag, words []string) map[string]struct{} {
	used := make(map[string]struct{})
	for index := 0; index < len(words); index++ {
		word := words[index]
		name := word
		if inlineName, _, ok := splitInlineFlag(word); ok {
			name = inlineName
		}
		flag, found := findFlag(flags, name)
		if !found {
			continue
		}
		primary := flag.Names[0]
		used[primary] = struct{}{}
		if flag.ValueKind != ValueNone && !strings.Contains(word, "=") && index+1 < len(words) {
			index++
		}
	}
	return used
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
		return resolver.Workflows(flagValueInWords(words, "--repo"))
	case ValueJob:
		if resolver == nil {
			return nil
		}
		return resolver.Jobs(
			flagValueInWords(words, "--repo"),
			flagValueInWords(words, "--workflow"),
			flagValueInWords(words, "--provider"),
		)
	case ValueProvider:
		if resolver == nil {
			return []string{"github", "jenkins"}
		}
		providers := resolver.Providers(flagValueInWords(words, "--repo"))
		if len(providers) == 0 {
			return []string{"github", "jenkins"}
		}
		return providers
	default:
		return nil
	}
}

func flagValueInWords(words []string, name string) string {
	prefix := name + "="
	for index := 0; index < len(words); index++ {
		word := words[index]
		if word == name {
			if index+1 < len(words) {
				return words[index+1]
			}
			return ""
		}
		if strings.HasPrefix(word, prefix) {
			return word[len(prefix):]
		}
	}
	return ""
}

func filterPrefix(candidates []string, partial string) []string {
	if partial == "" {
		return candidates
	}
	filtered := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == PathSentinel || strings.HasPrefix(candidate, partial) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
