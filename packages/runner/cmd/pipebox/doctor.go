package main

import (
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/doctor"
  "github.com/MaxiGarcia13/pipebox/packages/runner/internal/event"
)

func runDoctor() {
  if err := event.EmitStdout(doctor.Run(version)); err != nil {
    fail(err)
  }
}
