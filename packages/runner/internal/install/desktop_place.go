package install

import (
  "fmt"
  "io"
  "os"
  "path/filepath"
  "runtime"
  "strings"
)

func userDesktopPath() (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("resolve home: %w", err)
  }

  switch runtime.GOOS {
  case "darwin":
    return filepath.Join(homeDir, "Applications", "PipeBox.app"), nil
  case "linux":
    return filepath.Join(homeDir, ".local", "share", "pipebox-app"), nil
  case "windows":
    localAppData := os.Getenv("LOCALAPPDATA")
    if localAppData == "" {
      return "", fmt.Errorf("LOCALAPPDATA is not set")
    }
    return filepath.Join(localAppData, "PipeBox"), nil
  default:
    return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
  }
}

func placeDesktop(extractDir string) (string, error) {
  switch runtime.GOOS {
  case "darwin":
    return placeDarwin(extractDir)
  case "linux":
    return placeLinux(extractDir)
  case "windows":
    return placeWindows(extractDir)
  default:
    return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
  }
}

func placeDarwin(extractDir string) (string, error) {
  appPath, err := findMacApp(extractDir)
  if err != nil {
    return "", err
  }

  dest, err := userDesktopPath()
  if err != nil {
    return "", err
  }
  if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
    return "", fmt.Errorf("create Applications: %w", err)
  }
  if err := os.RemoveAll(dest); err != nil {
    return "", fmt.Errorf("replace existing app: %w", err)
  }
  if err := copyTree(appPath, dest); err != nil {
    return "", err
  }
  return dest, nil
}

func placeLinux(extractDir string) (string, error) {
  binary, err := findNamedExecutable(extractDir, "PipeBox")
  if err != nil {
    return "", err
  }
  sourceRoot := filepath.Dir(binary)

  dest, err := userDesktopPath()
  if err != nil {
    return "", err
  }
  if err := os.RemoveAll(dest); err != nil {
    return "", fmt.Errorf("replace existing app: %w", err)
  }
  if err := copyTree(sourceRoot, dest); err != nil {
    return "", err
  }
  return filepath.Join(dest, "PipeBox"), nil
}

func placeWindows(extractDir string) (string, error) {
  binary, err := findNamedExecutable(extractDir, "PipeBox.exe")
  if err != nil {
    return "", err
  }
  sourceRoot := filepath.Dir(binary)

  dest, err := userDesktopPath()
  if err != nil {
    return "", err
  }
  if err := os.RemoveAll(dest); err != nil {
    return "", fmt.Errorf("replace existing app: %w", err)
  }
  if err := copyTree(sourceRoot, dest); err != nil {
    return "", err
  }
  return filepath.Join(dest, "PipeBox.exe"), nil
}

func findMacApp(root string) (string, error) {
  var found string
  err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
      found = path
      return io.EOF
    }
    // MakerZIP may zip the .app contents at the archive root (Contents/...).
    if entry.IsDir() && entry.Name() == "Contents" && filepath.Dir(path) == root {
      synthetic := filepath.Join(root, "PipeBox.app")
      if mkErr := os.MkdirAll(synthetic, 0o755); mkErr != nil {
        return mkErr
      }
      if renameErr := os.Rename(path, filepath.Join(synthetic, "Contents")); renameErr != nil {
        return renameErr
      }
      found = synthetic
      return io.EOF
    }
    return nil
  })
  if err != nil && err != io.EOF {
    return "", fmt.Errorf("find .app: %w", err)
  }
  if found == "" {
    return "", fmt.Errorf("PipeBox.app not found in archive")
  }
  return found, nil
}

func findNamedExecutable(root, name string) (string, error) {
  var found string
  err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    if entry.IsDir() {
      return nil
    }
    if entry.Name() == name {
      found = path
      return io.EOF
    }
    return nil
  })
  if err != nil && err != io.EOF {
    return "", fmt.Errorf("find %s: %w", name, err)
  }
  if found == "" {
    return "", fmt.Errorf("%s not found in archive", name)
  }
  return found, nil
}

func copyTree(source, dest string) error {
  return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }
    rel, err := filepath.Rel(source, path)
    if err != nil {
      return err
    }
    target := filepath.Join(dest, rel)
    info, err := entry.Info()
    if err != nil {
      return err
    }
    if entry.IsDir() {
      return os.MkdirAll(target, 0o755)
    }
    if info.Mode()&os.ModeSymlink != 0 {
      link, err := os.Readlink(path)
      if err != nil {
        return err
      }
      _ = os.Remove(target)
      return os.Symlink(link, target)
    }
    if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
      return err
    }
    if err := copyFile(path, target); err != nil {
      return err
    }
    return os.Chmod(target, info.Mode().Perm())
  })
}
