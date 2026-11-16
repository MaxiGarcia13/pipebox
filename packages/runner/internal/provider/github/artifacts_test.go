package github

import (
  "os"
  "path/filepath"
  "testing"
)

func TestListArtifactsV4Zip(t *testing.T) {
  t.Parallel()

  base := t.TempDir()
  zipPath := filepath.Join(base, "1", "cv-pdf", "cv-pdf.zip")
  if err := os.MkdirAll(filepath.Dir(zipPath), 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(zipPath, []byte("pdf-bytes"), 0o600); err != nil {
    t.Fatal(err)
  }

  artifacts, err := listArtifacts(base)
  if err != nil {
    t.Fatal(err)
  }
  if len(artifacts) != 1 {
    t.Fatalf("got %#v", artifacts)
  }
  if artifacts[0].Name != "cv-pdf" {
    t.Fatalf("name: %#v", artifacts[0])
  }
  if artifacts[0].Path != zipPath {
    t.Fatalf("path: %#v", artifacts[0])
  }
  if artifacts[0].Size != int64(len("pdf-bytes")) {
    t.Fatalf("size: %#v", artifacts[0])
  }
}

func TestListArtifactsLooseDirectory(t *testing.T) {
  t.Parallel()

  base := t.TempDir()
  filePath := filepath.Join(base, "1", "logs", "out.txt")
  if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
    t.Fatal(err)
  }
  if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
    t.Fatal(err)
  }

  artifacts, err := listArtifacts(base)
  if err != nil {
    t.Fatal(err)
  }
  if len(artifacts) != 1 {
    t.Fatalf("got %#v", artifacts)
  }
  if artifacts[0].Name != "logs" {
    t.Fatalf("name: %#v", artifacts[0])
  }
  if artifacts[0].Path != filepath.Join(base, "1", "logs") {
    t.Fatalf("path: %#v", artifacts[0])
  }
  if artifacts[0].Size != int64(len("hello")) {
    t.Fatalf("size: %#v", artifacts[0])
  }
}

func TestListArtifactsEmpty(t *testing.T) {
  t.Parallel()

  artifacts, err := listArtifacts(t.TempDir())
  if err != nil {
    t.Fatal(err)
  }
  if len(artifacts) != 0 {
    t.Fatalf("got %#v", artifacts)
  }
}

func TestAllocatePort(t *testing.T) {
  t.Parallel()

  port, err := allocatePort()
  if err != nil {
    t.Fatal(err)
  }
  if port <= 0 {
    t.Fatalf("port: %d", port)
  }
}
