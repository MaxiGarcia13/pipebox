#!/usr/bin/env python3
import argparse
import sys
from pathlib import Path

SKIP_DIRS = {"vendor"}


def iter_go_files(root: Path):
  for path in root.rglob("*.go"):
    if any(part in SKIP_DIRS for part in path.parts):
      continue
    yield path


def check(root: Path) -> int:
  offenders = []
  for path in iter_go_files(root):
    text = path.read_text(encoding="utf-8")
    if "\t" in text:
      for i, line in enumerate(text.splitlines(), start=1):
        if "\t" in line:
          offenders.append(f"{path}:{i}")
  if not offenders:
    return 0
  for item in offenders:
    print(item, file=sys.stderr)
  print(
    "error: Go files must use 2-space indent (no tabs; see .editorconfig)",
    file=sys.stderr,
  )
  return 1


def fix(root: Path) -> int:
  for path in iter_go_files(root):
    text = path.read_text(encoding="utf-8")
    if "\t" not in text:
      continue
    path.write_text(text.replace("\t", "  "), encoding="utf-8")
  return 0


def main() -> int:
  parser = argparse.ArgumentParser(
    description="Enforce 2-space indent in Go sources (no tabs).",
  )
  mode = parser.add_mutually_exclusive_group(required=True)
  mode.add_argument(
    "--check",
    action="store_true",
    help="fail if any *.go file contains tabs",
  )
  mode.add_argument(
    "--fix",
    action="store_true",
    help="expand tabs to 2 spaces in *.go files",
  )
  parser.add_argument(
    "root",
    nargs="?",
    default=".",
    help="directory to scan (default: .)",
  )
  args = parser.parse_args()
  root = Path(args.root)
  if args.check:
    return check(root)
  return fix(root)


if __name__ == "__main__":
  sys.exit(main())
