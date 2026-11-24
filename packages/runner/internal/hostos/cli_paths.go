package hostos

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

func cliGoBinPath(binaryName string) (string, error) {
  if gobin := strings.TrimSpace(os.Getenv("GOBIN")); gobin != "" {
    return filepath.Join(gobin, binaryName), nil
  }

  if gopath := strings.TrimSpace(os.Getenv("GOPATH")); gopath != "" {
    first := strings.Split(gopath, string(os.PathListSeparator))[0]
    first = strings.TrimSpace(first)
    if first != "" {
      return filepath.Join(first, "bin", binaryName), nil
    }
  }

  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }
  return filepath.Join(homeDir, "go", "bin", binaryName), nil
}
