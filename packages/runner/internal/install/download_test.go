package install

import (
  "io"
  "net/http"
  "net/http/httptest"
  "os"
  "path/filepath"
  "testing"
)

type recordingProgress struct {
  steps   []string
  label   string
  total   int64
  tracked bool
}

func (r *recordingProgress) Step(message string) {
  r.steps = append(r.steps, message)
}

func (r *recordingProgress) Track(label string, total int64, body io.Reader) io.Reader {
  r.tracked = true
  r.label = label
  r.total = total
  return body
}

func TestDownloadFileTracksProgress(t *testing.T) {
  payload := []byte("pipebox-bytes")
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Length", "13")
    _, _ = w.Write(payload)
  }))
  defer server.Close()

  dest := filepath.Join(t.TempDir(), "asset.bin")
  progress := &recordingProgress{}
  if err := downloadFile(server.URL, dest, "cli", progress); err != nil {
    t.Fatalf("downloadFile: %v", err)
  }

  if !progress.tracked {
    t.Fatal("expected Track to be called")
  }
  if progress.label != "cli" {
    t.Fatalf("label = %q", progress.label)
  }
  if progress.total != 13 {
    t.Fatalf("total = %d", progress.total)
  }

  got, err := os.ReadFile(dest)
  if err != nil {
    t.Fatalf("ReadFile: %v", err)
  }
  if string(got) != string(payload) {
    t.Fatalf("got %q", got)
  }
}

func TestDownloadFileNilProgress(t *testing.T) {
  payload := []byte("ok")
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
    _, _ = w.Write(payload)
  }))
  defer server.Close()

  dest := filepath.Join(t.TempDir(), "asset.bin")
  if err := downloadFile(server.URL, dest, "desktop", nil); err != nil {
    t.Fatalf("downloadFile: %v", err)
  }
  got, err := os.ReadFile(dest)
  if err != nil {
    t.Fatalf("ReadFile: %v", err)
  }
  if string(got) != string(payload) {
    t.Fatalf("got %q", got)
  }
}
