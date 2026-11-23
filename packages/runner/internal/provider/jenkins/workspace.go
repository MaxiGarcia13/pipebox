package jenkins

import "strings"

func applyWorkspace(source, workspace string) string {
  if _, ok := extractFirstBlock(stripJenkinsNoise(source), "pipeline"); !ok {
    return source
  }

  var builder strings.Builder
  builder.Grow(len(source) + len(workspace))
  index := 0
  for index < len(source) {
    found := indexOfIdent(source, "agent", index)
    if found < 0 {
      builder.WriteString(source[index:])
      break
    }
    clause, end, ok := readAgentClause(source, found)
    if !ok {
      builder.WriteString(source[index : found+len("agent")])
      index = found + len("agent")
      continue
    }
    builder.WriteString(source[index:found])
    builder.WriteString(rewriteAgentClause(clause, workspace))
    index = end
  }
  return builder.String()
}

func readAgentClause(source string, from int) (clause string, end int, ok bool) {
  cursor := skipSpace(source, from+len("agent"))
  if cursor >= len(source) {
    return "", from, false
  }
  if source[cursor] == '{' {
    _, end, ok = balancedBody(source, cursor)
    if !ok {
      return "", from, false
    }
    return source[from:end], end, true
  }
  end = cursor
  for end < len(source) && !isAgentKeywordStop(source[end]) {
    end++
  }
  if end == cursor {
    return "", from, false
  }
  return source[from:end], end, true
}

func isAgentKeywordStop(char byte) bool {
  return isSpaceOrBrace(char) || char == ';'
}

func rewriteAgentClause(clause, workspace string) string {
  cleaned := stripJenkinsNoise(clause)
  if indexOfIdent(cleaned, "customWorkspace", 0) >= 0 || reuseNodeEnabled(cleaned) {
    return clause
  }

  kind := agentKind(cleaned)
  switch kind {
  case "any":
    return builtInWorkspaceAgent("'built-in'", workspace)
  case "none":
    return clause
  case "node", "docker", "dockerfile":
    rewritten, ok := insertWorkspaceOption(clause, kind, workspace)
    if !ok {
      return clause
    }
    return rewritten
  case "label":
    label, ok := labelExpression(cleaned)
    if !ok {
      return clause
    }
    return builtInWorkspaceAgent(label, workspace)
  default:
    return clause
  }
}

func agentKind(clause string) string {
  cursor := skipSpace(clause, len("agent"))
  if cursor >= len(clause) {
    return ""
  }
  if clause[cursor] != '{' {
    end := cursor
    for end < len(clause) && !isAgentKeywordStop(clause[end]) {
      end++
    }
    return clause[cursor:end]
  }
  body, _, ok := balancedBody(clause, cursor)
  if !ok {
    return ""
  }
  body = strings.TrimLeft(body, " \t\r\n")
  end := 0
  for end < len(body) && isIdentChar(rune(body[end])) {
    end++
  }
  return body[:end]
}

func reuseNodeEnabled(clause string) bool {
  found := indexOfIdent(clause, "reuseNode", 0)
  if found < 0 {
    return false
  }
  cursor := skipSpace(clause, found+len("reuseNode"))
  return strings.HasPrefix(clause[cursor:], "true")
}

func labelExpression(clause string) (string, bool) {
  found := indexOfIdent(clause, "label", 0)
  if found < 0 {
    return "", false
  }
  cursor := skipSpace(clause, found+len("label"))
  token, _, ok := readGroovyString(clause, cursor)
  return token, ok
}

func readGroovyString(source string, from int) (token string, end int, ok bool) {
  if from >= len(source) {
    return "", from, false
  }
  quote := source[from]
  if quote != '\'' && quote != '"' {
    return "", from, false
  }
  index := from + 1
  for index < len(source) {
    if source[index] == '\\' && index+1 < len(source) {
      index += 2
      continue
    }
    if source[index] == quote {
      return source[from : index+1], index + 1, true
    }
    index++
  }
  return "", from, false
}

func insertWorkspaceOption(clause, kind, workspace string) (string, bool) {
  found := indexOfIdent(clause, kind, 0)
  if found < 0 {
    return "", false
  }
  cursor := skipSpace(clause, found+len(kind))
  if cursor >= len(clause) || clause[cursor] != '{' {
    return "", false
  }
  _, end, ok := balancedBody(clause, cursor)
  if !ok {
    return "", false
  }
  option := "\n    customWorkspace '" + groovySingleQuote(workspace) + "'\n  "
  return clause[:end-1] + option + clause[end-1:], true
}

func builtInWorkspaceAgent(labelExpr, workspace string) string {
  return "agent {\n  node {\n    label " + labelExpr + "\n    customWorkspace '" + groovySingleQuote(workspace) + "'\n  }\n}"
}

func groovySingleQuote(value string) string {
  var builder strings.Builder
  for _, char := range value {
    if char == '\\' || char == '\'' {
      builder.WriteByte('\\')
    }
    builder.WriteRune(char)
  }
  return builder.String()
}
