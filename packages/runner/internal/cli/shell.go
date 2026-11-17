package cli

import (
  "fmt"
  "strings"
)

func CompletionScript(shell string) (string, error) {
  switch strings.ToLower(strings.TrimSpace(shell)) {
  case "zsh":
    return zshCompletionScript(), nil
  case "bash":
    return bashCompletionScript(), nil
  case "fish":
    return fishCompletionScript(), nil
  default:
    return "", fmt.Errorf("unsupported shell %q (want zsh, bash, or fish)", shell)
  }
}

func zshCompletionScript() string {
  return `#compdef pipebox
_pipebox() {
  local out
  out="$(pipebox __complete -- "${words[2,CURRENT]}" 2>/dev/null)"
  if [[ "$out" == *__PIPEBOX_PATH__* ]]; then
    _files
    return
  fi
  local -a candidates
  candidates=("${(@f)out}")
  (( $#candidates )) && _describe 'pipebox' candidates
}

compdef _pipebox pipebox
`
}

func bashCompletionScript() string {
  return `_pipebox() {
  local cur out line
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  local words=("${COMP_WORDS[@]:1:COMP_CWORD}")
  out="$(pipebox __complete -- "${words[@]}" 2>/dev/null)"
  if [[ "$out" == *__PIPEBOX_PATH__* ]]; then
    COMPREPLY=($(compgen -f -- "$cur"))
    return
  fi
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    COMPREPLY+=("$line")
  done < <(compgen -W "$(printf '%s\n' "$out")" -- "$cur")
}

complete -F _pipebox pipebox
`
}

func fishCompletionScript() string {
  return `function __pipebox_complete
  set -l tokens (commandline -opc)
  set -e tokens[1]
  set -l current (commandline -ct)
  if test (count $tokens) -eq 0
    set tokens $current
  else
    set tokens $tokens $current
  end
  set -l out (pipebox __complete -- $tokens 2>/dev/null)
  if string match -q '*__PIPEBOX_PATH__*' -- $out
    __fish_complete_path
    return
  end
  for line in $out
    echo $line
  end
end

complete -c pipebox -f -a '(__pipebox_complete)'
`
}
