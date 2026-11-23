package jenkins

import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)

func TestWriteSecretsCasc(t *testing.T) {
  dir := t.TempDir()
  if err := writeSecretsCasc(dir, []string{`TOKEN=a"b`, "EMPTY="}); err != nil {
    t.Fatal(err)
  }
  data, err := os.ReadFile(filepath.Join(dir, "casc_configs", "secrets.yaml"))
  if err != nil {
    t.Fatal(err)
  }
  text := string(data)
  for _, want := range []string{
    "globalNodeProperties:",
    `key: "TOKEN"`,
    `value: "a\"b"`,
    `key: "EMPTY"`,
    `value: ""`,
  } {
    if !strings.Contains(text, want) {
      t.Fatalf("missing %q in\n%s", want, text)
    }
  }
  if strings.Contains(text, "numExecutors") || strings.Contains(text, "securityRealm") {
    t.Fatalf("secrets file repeated base casc keys:\n%s", text)
  }
}

func TestWriteSecretsCascSkipsEmpty(t *testing.T) {
  dir := t.TempDir()
  if err := writeSecretsCasc(dir, nil); err != nil {
    t.Fatal(err)
  }
  if _, err := os.Stat(filepath.Join(dir, "casc_configs", "secrets.yaml")); !os.IsNotExist(err) {
    t.Fatalf("expected no file, err=%v", err)
  }
}
