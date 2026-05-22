#!/usr/bin/env python3
"""Legacy-negative gate for backend-async-runner (spec D-12).

After plan 001 consolidates every in-process drainer / runner into the single
runner.Runtime kernel, the retired entry points below must not reappear in
production Go sources. Tests, history docs, and lint self keep the old terms as
evidence, so they are excluded from the scan.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

# Forbidden production references. Each entry is a compiled regex matched line
# by line against non-test Go sources under backend/.
FORBIDDEN_PATTERNS = (
    (re.compile(r"\b(?:review|reviewdomain|domainreview)\.NewRunner\b"), "legacy review.NewRunner (use runner.Runtime + review.GenerateHandler)"),
    (re.compile(r"\b(?:review|reviewdomain|domainreview)\.NewReaper\b"), "legacy review.NewReaper (use runner.Runtime reaper)"),
    (re.compile(r"\bComputeReportFailureBackoff\b"), "legacy ComputeReportFailureBackoff (use runner.BackoffPolicy)"),
    (re.compile(r"\bDefaultReportFailureBackoff\b"), "legacy DefaultReportFailureBackoff (use runner.BackoffPolicy)"),
    (re.compile(r"\bBackgroundMailDispatcher\b"), "legacy auth.BackgroundMailDispatcher (use EmailDispatchEnqueuer + EmailDispatchHandler)"),
    (re.compile(r"\bNewBackgroundMailDispatcher\b"), "legacy auth.NewBackgroundMailDispatcher (use EmailDispatchEnqueuer)"),
    (re.compile(r"\btargetjob\.NewDrainer\("), "legacy targetjob.NewDrainer instantiation (register handlers on runner.Runtime)"),
    (re.compile(r"\b(?:jdmatchRuntime|resumeRuntime|targetJobRuntime)\.Drainer\b"), "legacy per-domain Drainer field (use runner.Runtime)"),
    (re.compile(r"\breportRuntime\.(?:Runner|Reaper)\b"), "legacy reportRuntime Runner/Reaper field (use runner.Runtime)"),
)

# Production Go roots that must stay clean. Tests are excluded by suffix.
SCAN_ROOTS = (
    ("backend", "cmd"),
    ("backend", "internal"),
)

EXCLUDED_PARTS = {".git", "node_modules", "dist", "coverage", ".test-output", "__pycache__"}


def is_production_go(path: Path) -> bool:
    if path.suffix != ".go":
        return False
    if path.name.endswith("_test.go"):
        return False
    return True


def iter_scan_files(repo_root: Path) -> list[Path]:
    files: list[Path] = []
    for prefix in SCAN_ROOTS:
        root = repo_root.joinpath(*prefix)
        if not root.exists():
            continue
        for path in root.rglob("*.go"):
            if path.is_symlink() or not path.is_file():
                continue
            if any(part in EXCLUDED_PARTS for part in path.parts):
                continue
            if is_production_go(path):
                files.append(path)
    return sorted(set(files))


def scan_paths(paths: list[Path]) -> list[str]:
    problems: list[str] = []
    for path in paths:
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            for pattern, message in FORBIDDEN_PATTERNS:
                if pattern.search(line):
                    problems.append(f"{path}:{lineno}: {message}")
    return problems


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(ROOT), help="repository root")
    parser.add_argument("--phase", default="all", choices=["all"], help="reserved for future phased gates")
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    problems = scan_paths(iter_scan_files(repo_root))
    if problems:
        for problem in problems:
            print(f"FAIL: {problem}", file=sys.stderr)
        return 1
    print("OK: backend-async-runner legacy runner/drainer/dispatcher entry points absent from production sources")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
