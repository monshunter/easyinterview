#!/usr/bin/env python3
"""Legacy-negative gate for backend-debrief runtime surfaces.

The owner docs intentionally name retired terms in negative-gate prose. This
script scans runtime code, event/job contracts, fixtures, and scenario runtime
assets where those terms must not reappear as active behavior.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

RETIRED_TERMS = (
    "mistakes_count",
    "generatedMistakeCount",
    "experience_library",
    "drill_builder",
    "growth_center",
    "star_editor",
    "debrief_voice",
)

SCAN_PREFIXES = (
    ("backend", "internal", "debrief"),
    ("backend", "internal", "api", "debriefs"),
    ("backend", "internal", "store", "debrief"),
    ("shared", "events.yaml"),
    ("shared", "jobs.yaml"),
    ("openapi", "fixtures", "Debriefs"),
    ("test", "scenarios", "e2e", "p0-060-debrief-create-worker-happy"),
    ("test", "scenarios", "e2e", "p0-061-debrief-get-isolation"),
    ("test", "scenarios", "e2e", "p0-062-debrief-worker-retry-failure"),
    ("test", "scenarios", "e2e", "p0-063-debrief-suggest-questions"),
    ("test", "scenarios", "e2e", "p0-064-debrief-privacy-legacy"),
)

EXCLUDED_PARTS = {".git", "node_modules", "dist", "coverage", ".test-output", "__pycache__"}


def repo_relative_path(path: Path, repo_root: Path) -> Path | None:
    try:
        return path.absolute().relative_to(repo_root.absolute())
    except ValueError:
        return None


def is_scan_path(path: Path, repo_root: Path) -> bool:
    rel = repo_relative_path(path, repo_root)
    if rel is None:
        return False
    parts = rel.parts
    if any(part in EXCLUDED_PARTS for part in parts):
        return False
    return any(parts[: len(prefix)] == prefix for prefix in SCAN_PREFIXES)


def iter_scan_files(repo_root: Path) -> list[Path]:
    files: list[Path] = []
    for prefix in SCAN_PREFIXES:
        root = repo_root.joinpath(*prefix)
        if root.is_file():
            files.append(root)
            continue
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.is_symlink() or not path.is_file():
                continue
            if is_scan_path(path, repo_root):
                files.append(path)
    return sorted(set(files))


def scan_paths(paths: list[Path], repo_root: Path) -> list[str]:
    problems: list[str] = []
    for path in paths:
        if not is_scan_path(path, repo_root):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            for term in RETIRED_TERMS:
                if term in line:
                    problems.append(f"{path}:{lineno}: retired backend-debrief term {term!r}")
    return problems


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(ROOT), help="repository root")
    parser.add_argument("--phase", default="all", choices=["all"], help="reserved for phased gates")
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    problems = scan_paths(iter_scan_files(repo_root), repo_root)
    if problems:
        for problem in problems:
            print(f"FAIL: {problem}", file=sys.stderr)
        return 1
    print("OK: backend-debrief legacy terms absent from runtime surfaces")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
