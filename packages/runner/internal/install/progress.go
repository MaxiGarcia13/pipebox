package install

import "io"

type Progress interface {
  Step(message string)
  Track(label string, total int64, r io.Reader) io.Reader
}

type NopProgress struct{}

func (NopProgress) Step(string) {}

func (NopProgress) Track(_ string, _ int64, r io.Reader) io.Reader {
  return r
}

func resolveProgress(progress Progress) Progress {
  if progress == nil {
    return NopProgress{}
  }
  return progress
}
