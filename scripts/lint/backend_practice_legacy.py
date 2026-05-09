#!/usr/bin/env python3
"""Legacy-negative gate for backend-practice Phase 0.

The removed practice-mode literal must not remain in active code, specs,
tests, scenario assets, generated artifacts, or contract files. Historical
work journals and completed reports are excluded because they are immutable
delivery evidence rather than executable truth sources.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


REMOVED_MODE = "debrief" + "_replay"
EXCLUDED_PARTS = {
    ".git",
    ".test-output",
    "node_modules",
    "dist",
    "coverage",
}
EXCLUDED_PREFIXES = (
    ("docs", "work-journal"),
    ("docs", "reports"),
)


def is_excluded(path: Path, repo_root: Path) -> bool:
    rel = path.resolve().relative_to(repo_root.resolve())
    parts = rel.parts
    if any(part in EXCLUDED_PARTS for part in parts):
        return True
    return any(parts[: len(prefix)] == prefix for prefix in EXCLUDED_PREFIXES)


def iter_repo_files(repo_root: Path) -> list[Path]:
    out: list[Path] = []
    for path in repo_root.rglob("*"):
        if not path.is_file():
            continue
        if is_excluded(path, repo_root):
            continue
        out.append(path)
    return out


def is_allowed_line(line: str) -> bool:
    if "PracticeGoal" not in line:
        return False
    active_mode_contexts = ("PracticeMode", "practiceMode", "practice_plans.mode", "session.mode")
    return not any(context in line for context in active_mode_contexts)


def scan_paths(paths: list[Path], repo_root: Path) -> list[str]:
    problems: list[str] = []
    for path in paths:
        if is_excluded(path, repo_root):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            if REMOVED_MODE in line and not is_allowed_line(line):
                problems.append(f"{path}:{lineno}: removed practice mode literal in active context")
    return problems


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    problems = scan_paths(iter_repo_files(repo_root), repo_root)
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print("backend_practice_legacy: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
