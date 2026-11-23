package jenkins

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

func writeSecretsCasc(runHome string, pairs []string) error {
  if len(pairs) == 0 {
    return nil
  }
  dir := filepath.Join(runHome, "casc_configs")
  if err := os.MkdirAll(dir, 0o755); err != nil {
    return err
  }

  var builder strings.Builder
  builder.WriteString("jenkins:\n")
  builder.WriteString("  globalNodeProperties:\n")
  builder.WriteString("    - envVars:\n")
  builder.WriteString("        env:\n")
  for _, pair := range pairs {
    key, value, ok := strings.Cut(pair, "=")
    if !ok || strings.TrimSpace(key) == "" {
      return fmt.Errorf("invalid secret %q (want KEY=VALUE)", pair)
    }
    builder.WriteString("          - key: \"")
    builder.WriteString(yamlDoubleQuoted(strings.TrimSpace(key)))
    builder.WriteString("\"\n            value: \"")
    builder.WriteString(yamlDoubleQuoted(value))
    builder.WriteString("\"\n")
  }
  return os.WriteFile(filepath.Join(dir, "secrets.yaml"), []byte(builder.String()), 0o644)
}

func yamlDoubleQuoted(value string) string {
  var builder strings.Builder
  for _, char := range value {
    switch char {
    case '\\', '"':
      builder.WriteByte('\\')
      builder.WriteRune(char)
    case '\n':
      builder.WriteString(`\n`)
    case '\r':
      builder.WriteString(`\r`)
    case '\t':
      builder.WriteString(`\t`)
    default:
      builder.WriteRune(char)
    }
  }
  return builder.String()
}
