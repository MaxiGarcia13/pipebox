package install

import (
  "archive/zip"
  "fmt"
  "io"
  "os"
  "path/filepath"
  "strings"
)

func unzip(zipPath, destDir string) (err error) {
  reader, err := zip.OpenReader(zipPath)
  if err != nil {
    return fmt.Errorf("open zip: %w", err)
  }
  defer func() {
    if closeErr := reader.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close zip: %w", closeErr)
    }
  }()

  destDir = filepath.Clean(destDir)
  for _, file := range reader.File {
    if extractErr := extractZipFile(file, destDir); extractErr != nil {
      return extractErr
    }
  }
  return nil
}

func extractZipFile(file *zip.File, destDir string) (err error) {
  name := filepath.Clean(file.Name)
  if name == "." || strings.HasPrefix(name, "..") {
    return fmt.Errorf("invalid zip entry %q", file.Name)
  }

  target := filepath.Join(destDir, name)
  if !strings.HasPrefix(target, destDir+string(os.PathSeparator)) && target != destDir {
    return fmt.Errorf("zip entry escapes dest: %s", file.Name)
  }

  mode := file.Mode()
  if file.FileInfo().IsDir() {
    return os.MkdirAll(target, 0o755)
  }

  if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
    return fmt.Errorf("mkdir for %s: %w", target, err)
  }

  source, err := file.Open()
  if err != nil {
    return fmt.Errorf("open zip entry: %w", err)
  }
  defer func() {
    if closeErr := source.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close zip entry: %w", closeErr)
    }
  }()

  // ZIP stores symlink targets as the entry body; restore them or Electron
  // frameworks become tiny text files and dyld fails with "not valid mach-o".
  if mode&os.ModeSymlink != 0 {
    linkTarget, readErr := io.ReadAll(source)
    if readErr != nil {
      return fmt.Errorf("read symlink %s: %w", target, readErr)
    }
    _ = os.Remove(target)
    if err := os.Symlink(string(linkTarget), target); err != nil {
      return fmt.Errorf("symlink %s: %w", target, err)
    }
    return nil
  }

  perm := mode.Perm()
  if perm == 0 {
    perm = 0o644
  }
  out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
  if err != nil {
    return fmt.Errorf("create %s: %w", target, err)
  }
  defer func() {
    if closeErr := out.Close(); err == nil && closeErr != nil {
      err = fmt.Errorf("close %s: %w", target, closeErr)
    }
  }()

  if _, err := io.Copy(out, source); err != nil {
    return fmt.Errorf("extract %s: %w", target, err)
  }
  return nil
}
