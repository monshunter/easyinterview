#!/usr/bin/env python3
"""Out-of-scope gate for backend-review report generation.

The owner plan intentionally names out-of-scope terms in markdown gates. This script
therefore scans only runtime, fixtures, and scenario assets that must not carry
old report-dashboard or worker-schema vocabulary.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

OUT_OF_SCOPE_TERMS = (
    "reportLayout",
    "readiness_score",
    "mistakes_queue",
    "mistake_queue",
    "drill_builder",
    "growth_center",
    "report_timeline",
    "report_form",
    "review_method_version",
    "leased_at",
    "attempt_count",
    "retry_mistakes",
    "build_drill",
    "open_growth_center",
)

AI_TASK_SUCCEEDED_RE = re.compile(r"\bai_task_runs\b.*(['\"]succeeded['\"]|status\s*=\s*succeeded)", re.IGNORECASE)

SCAN_PREFIXES = (
    ("backend", "internal", "review"),
    ("backend", "internal", "api", "reports"),
    ("backend", "internal", "store", "review"),
    ("openapi", "fixtures", "Reports"),
    ("test", "scenarios", "e2e", "p0-052-report-generation-happy-path"),
    ("test", "scenarios", "e2e", "p0-053-report-read-and-listing"),
    ("test", "scenarios", "e2e", "p0-054-report-ai-failure-and-retry"),
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
            for term in OUT_OF_SCOPE_TERMS:
                if term in line:
                    problems.append(f"{path}:{lineno}: out-of-scope backend-review term {term!r}")
            if "worker_id" in line and "slog.String(\"worker_id\"" not in line:
                problems.append(f"{path}:{lineno}: out-of-scope async worker_id persistence vocabulary")
            if AI_TASK_SUCCEEDED_RE.search(line):
                problems.append(f"{path}:{lineno}: ai_task_runs status must use 'success', not 'succeeded'")
    return problems


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(ROOT), help="repository root")
    parser.add_argument("--phase", default="all", choices=["all"], help="reserved for future phased gates")
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    problems = scan_paths(iter_scan_files(repo_root), repo_root)
    if problems:
        for problem in problems:
            print(f"FAIL: {problem}", file=sys.stderr)
        return 1
    print("OK: backend-review out-of-scope terms absent from runtime report surfaces")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
