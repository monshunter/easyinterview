#!/usr/bin/env python3
"""scripts/lint/prompt_hardcode_lint.py - F3 hardcoded-prompt boundary lint.

Thin Python wrapper around `scripts/lint/prompt_hardcode_lint.go`. The Go
helper does the actual AST scan via `go/parser`; this wrapper provides a
consistent CLI shape, allows tests to redirect the scan roots, and keeps the
exit-code contract aligned with the rest of the lint suite.

Run: `python3 scripts/lint/prompt_hardcode_lint.py [--roots p1,p2,...]`
Exit: 0 on success, 1 on any violation.
"""
from __future__ import annotations

import argparse
import pathlib
import subprocess
import sys

DEFAULT_ROOTS = [
    "backend/internal/practice",
    "backend/internal/report",
    "backend/internal/resume",
    "backend/internal/targetjob",
]

GO_HELPER = pathlib.Path(__file__).with_suffix(".go")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--roots",
        default=",".join(DEFAULT_ROOTS),
        help="comma-separated roots passed to the Go helper",
    )
    parser.add_argument(
        "--repo-root",
        default=".",
        help="working directory the Go helper resolves relative paths against",
    )
    args = parser.parse_args(argv)

    cmd = ["go", "run", str(GO_HELPER), "-roots", args.roots]
    completed = subprocess.run(cmd, cwd=args.repo_root, check=False)
    return completed.returncode


if __name__ == "__main__":
    sys.exit(main())
