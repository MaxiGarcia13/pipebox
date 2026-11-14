package install

import (
  "archive/zip"
  "os"
  "path/filepath"
  "testing"
)

func TestUnzipRestoresSymlinks(t *testing.T) {
  tmp := t.TempDir()
  zipPath := filepath.Join(tmp, "app.zip")
  extractDir := filepath.Join(tmp, "extract")

  if err := writeZipWithSymlink(zipPath); err != nil {
    t.Fatalf("write zip: %v", err)
  }
  if err := os.MkdirAll(extractDir, 0o755); err != nil {
    t.Fatalf("mkdir extract: %v", err)
  }
  if err := unzip(zipPath, extractDir); err != nil {
    t.Fatalf("unzip: %v", err)
  }

  linkPath := filepath.Join(extractDir, "Framework", "Current")
  info, err := os.Lstat(linkPath)
  if err != nil {
    t.Fatalf("lstat symlink: %v", err)
  }
  if info.Mode()&os.ModeSymlink == 0 {
    t.Fatalf("expected symlink at %s, got mode %v", linkPath, info.Mode())
  }
  target, err := os.Readlink(linkPath)
  if err != nil {
    t.Fatalf("readlink: %v", err)
  }
  if target != "A" {
    t.Fatalf("symlink target = %q, want %q", target, "A")
  }

  body, err := os.ReadFile(filepath.Join(extractDir, "Framework", "A", "payload"))
  if err != nil {
    t.Fatalf("read payload via layout: %v", err)
  }
  if string(body) != "ok" {
    t.Fatalf("payload = %q, want %q", body, "ok")
  }
}

func writeZipWithSymlink(zipPath string) error {
  out, err := os.Create(zipPath)
  if err != nil {
    return err
  }
  defer func() { _ = out.Close() }()

  writer := zip.NewWriter(out)
  defer func() { _ = writer.Close() }()

  dirHeader := &zip.FileHeader{Name: "Framework/A/"}
  dirHeader.SetMode(os.ModeDir | 0o755)
  if _, err := writer.CreateHeader(dirHeader); err != nil {
    return err
  }

  fileWriter, err := writer.Create("Framework/A/payload")
  if err != nil {
    return err
  }
  if _, err := fileWriter.Write([]byte("ok")); err != nil {
    return err
  }

  linkHeader := &zip.FileHeader{Name: "Framework/Current"}
  linkHeader.SetMode(os.ModeSymlink | 0o777)
  linkWriter, err := writer.CreateHeader(linkHeader)
  if err != nil {
    return err
  }
  if _, err := linkWriter.Write([]byte("A")); err != nil {
    return err
  }
  return nil
}
