package install

import "testing"

func TestDesktopAssetNameFor(t *testing.T) {
  cases := []struct {
    goos, goarch, want string
  }{
    {goos: "darwin", goarch: "arm64", want: "PipeBox-darwin-arm64.zip"},
    {goos: "darwin", goarch: "amd64", want: "PipeBox-darwin-x64.zip"},
    {goos: "linux", goarch: "amd64", want: "PipeBox-linux-amd64.zip"},
    {goos: "windows", goarch: "amd64", want: "PipeBox-windows-x64.zip"},
    {goos: "linux", goarch: "arm64"},
    {goos: "windows", goarch: "arm64"},
    {goos: "darwin", goarch: "386"},
    {goos: "freebsd", goarch: "amd64"},
  }

  for _, tc := range cases {
    t.Run(tc.goos+"/"+tc.goarch, func(t *testing.T) {
      got, err := desktopAssetNameFor(tc.goos, tc.goarch)
      if tc.want == "" {
        if err == nil {
          t.Fatalf("expected error, got %q", got)
        }
        return
      }
      if err != nil {
        t.Fatalf("desktopAssetNameFor: %v", err)
      }
      if got != tc.want {
        t.Fatalf("got %q want %q", got, tc.want)
      }
    })
  }
}
