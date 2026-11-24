package cli

import (
  "os"
  "path/filepath"
  "reflect"
  "runtime"
  "testing"
)

func TestProviderResolverUsesWorkingDirectory(t *testing.T) {
  repo := filepath.Join(moduleRoot(t), "testdata", "repos", "mixed")
  previous, err := os.Getwd()
  if err != nil {
    t.Fatal(err)
  }
  if err := os.Chdir(repo); err != nil {
    t.Fatal(err)
  }
  t.Cleanup(func() {
    _ = os.Chdir(previous)
  })

  resolver := ProviderResolver{}
  workflows := resolver.Workflows("")
  wantWorkflows := []string{
    filepath.Join(".github", "workflows", "ci.yml"),
    "Jenkinsfile",
  }
  if !reflect.DeepEqual(workflows, wantWorkflows) {
    t.Fatalf("workflows: got %#v", workflows)
  }

  jobs := resolver.Jobs("", filepath.Join(".github", "workflows", "ci.yml"), "")
  if !reflect.DeepEqual(jobs, []string{"build"}) {
    t.Fatalf("jobs: got %#v", jobs)
  }

  providers := resolver.Providers("")
  if !reflect.DeepEqual(providers, []string{"github", "jenkins"}) {
    t.Fatalf("providers: got %#v", providers)
  }
}

func moduleRoot(t *testing.T) string {
  t.Helper()
  _, file, _, ok := runtime.Caller(0)
  if !ok {
    t.Fatal("runtime.Caller failed")
  }
  return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}
