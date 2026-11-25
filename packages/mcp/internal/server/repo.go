package server

import "os"

func resolveRepo(repo string) (string, error) {
  if repo != "" {
    return repo, nil
  }
  cwd, err := os.Getwd()
  if err != nil {
    return "", err
  }
  return cwd, nil
}
