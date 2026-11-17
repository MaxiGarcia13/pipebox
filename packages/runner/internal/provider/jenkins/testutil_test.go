package jenkins

import (
  "os"
  "path/filepath"
)

func writeFile(dir, name, content string) error {
  return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}
