package cli

import "fmt"

type Kind int

const (
  KindHelp Kind = iota
  KindVersion
  KindSubcommand
  KindOpenUI
  KindError
)

type Args struct {
  Kind        Kind
  Command     string
  CommandArgs []string
  Repo        string
  NewWindow   bool
  Err         error
}

func Parse(args []string) Args {
  return ParseCatalog(DefaultCatalog(), args)
}

func ParseCatalog(catalog Catalog, args []string) Args {
  if len(args) == 0 {
    return Args{Kind: KindOpenUI}
  }

  first := args[0]
  switch first {
  case "help", "-h", "--help":
    return Args{Kind: KindHelp}
  case "-v", "--version":
    return Args{Kind: KindVersion}
  case "version":
    return Args{Kind: KindVersion, Command: "version"}
  }

  if catalog.IsSubcommand(first) {
    return Args{
      Kind:        KindSubcommand,
      Command:     first,
      CommandArgs: args[1:],
    }
  }

  repo := ""
  newWindow := false
  for index := 0; index < len(args); index++ {
    arg := args[index]
    switch {
    case arg == "-h" || arg == "--help" || arg == "help":
      return Args{Kind: KindHelp}
    case arg == "-v" || arg == "--version" || arg == "version":
      return Args{Kind: KindVersion}
    case arg == "-n" || arg == "--new":
      newWindow = true
    case arg == "--repo":
      if index+1 >= len(args) {
        return Args{
          Kind: KindError,
          Err:  fmt.Errorf("usage: pipebox --repo <path>"),
        }
      }
      repo = args[index+1]
      index++
    case len(arg) > 7 && arg[:7] == "--repo=":
      repo = arg[7:]
    default:
      return Args{
        Kind: KindError,
        Err:  fmt.Errorf("unknown argument %q", arg),
      }
    }
  }

  return Args{Kind: KindOpenUI, Repo: repo, NewWindow: newWindow}
}
