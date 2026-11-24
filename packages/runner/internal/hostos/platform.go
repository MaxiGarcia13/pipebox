package hostos

import "runtime"

type Platform interface {
  ID() string
  CLIBinaryName() string
  CLILocalBinPath() (string, error)
  CLIGoBinPath() (string, error)
  DesktopDir() (string, error)
  DesktopLaunchCandidates() []string
  IsLaunchable(path string) bool
  ReplaceFile(source, target string) error
  RemoveFile(path string) error
  RemoveAll(path string) error
  PathEqualKey(path string) string
}

func Current() Platform {
  switch runtime.GOOS {
  case "darwin":
    return darwinPlatform{}
  case "windows":
    return windowsPlatform{}
  default:
    return linuxPlatform{}
  }
}
