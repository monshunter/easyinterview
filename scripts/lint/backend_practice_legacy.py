#!/usr/bin/env python3
"""Legacy-negative gate for backend-practice plan 001.

The removed practice-mode literal must not remain in active code, specs,
tests, scenario assets, generated artifacts, or contract files. Historical
work journals and completed reports are excluded because they are immutable
delivery evidence rather than executable truth sources.

Phase 3 also pins the retired module vocabulary on implementation/runtime
surfaces. The owner plan docs intentionally name the retired terms as gate
inputs, so Phase 3 scans a constrained output set instead of recursively
scanning every markdown line in the repository.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


REMOVED_MODE = "debrief" + "_replay"
PHASE3_RETIRED_TERMS = (
    "warmup",
    "single_drill",
    "drill_builder",
    "mistake_queue",
    "growth_center",
    "practiceModeCard",
)
PHASE3_VOICE_ROUTE = re.compile(r"(/voice\b|practice\.voice|voice_route|voice route)", re.IGNORECASE)
PHASE3_SCAN_PREFIXES = (
    ("backend", "cmd", "api"),
    ("backend", "internal", "api", "practice"),
    ("backend", "internal", "practice"),
    ("backend", "internal", "store", "practice"),
    ("openapi", "fixtures", "PracticePlans"),
    ("openapi", "fixtures", "PracticeSessions"),
    ("test", "scenarios", "e2e", "p0-022-practice-plan-baseline-create-and-read"),
    ("test", "scenarios", "e2e", "p0-023-practice-session-start-and-first-question"),
    ("test", "scenarios", "e2e", "p0-024-practice-session-ai-failure-retry"),
    ("test", "scenarios", "e2e", "p0-025-practice-idempotency-and-isolation-matrix"),
    ("test", "scenarios", "e2e", "p0-026-practice-observability-and-privacy-redlines"),
)
PHASE3_EXCLUDED_SUFFIXES = (
    ("scripts", "verify.sh"),
)
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


def is_phase3_scan_path(path: Path, repo_root: Path) -> bool:
    rel = path.resolve().relative_to(repo_root.resolve())
    parts = rel.parts
    if any(parts[: len(prefix)] == prefix for prefix in PHASE3_SCAN_PREFIXES):
        return not any(parts[-len(suffix) :] == suffix for suffix in PHASE3_EXCLUDED_SUFFIXES)
    return False


def scan_phase3_paths(paths: list[Path], repo_root: Path) -> list[str]:
    problems: list[str] = []
    for path in paths:
        if is_excluded(path, repo_root) or not is_phase3_scan_path(path, repo_root):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            for term in PHASE3_RETIRED_TERMS:
                if term in line:
                    problems.append(f"{path}:{lineno}: retired backend-practice term {term!r}")
            if PHASE3_VOICE_ROUTE.search(line) and "practice-voice-mvp" not in line:
                problems.append(f"{path}:{lineno}: retired standalone voice route")
    return problems


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--phase", choices=("phase0", "phase3", "all"), default="phase0")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    files = iter_repo_files(repo_root)
    problems: list[str] = []
    if args.phase in {"phase0", "all"}:
        problems.extend(scan_paths(files, repo_root))
    if args.phase in {"phase3", "all"}:
        problems.extend(scan_phase3_paths(files, repo_root))
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print(f"backend_practice_legacy {args.phase}: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
