package jenkins

import (
  "os"
  "path/filepath"
  "testing"
)

func TestListRunHomeArtifacts(t *testing.T) {
  runHome := t.TempDir()
  archive := filepath.Join(runHome, "jobs", "job", "builds", "1", "archive")
  if err := os.MkdirAll(archive, 0o755); err != nil {
    t.Fatal(err)
  }
  filePath := filepath.Join(archive, "report.txt")
  if err := os.WriteFile(filePath, []byte("ok"), 0o644); err != nil {
    t.Fatal(err)
  }

  artifacts, err := listRunHomeArtifacts(runHome)
  if err != nil {
    t.Fatal(err)
  }
  if len(artifacts) != 1 || artifacts[0].Name != "report.txt" {
    t.Fatalf("artifacts: %#v", artifacts)
  }
}

func TestListRunHomeArtifactsEmpty(t *testing.T) {
  artifacts, err := listRunHomeArtifacts(t.TempDir())
  if err != nil {
    t.Fatal(err)
  }
  if len(artifacts) != 0 {
    t.Fatalf("expected empty, got %#v", artifacts)
  }
}
