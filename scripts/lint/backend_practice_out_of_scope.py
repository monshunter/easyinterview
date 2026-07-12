#!/usr/bin/env python3
"""Reject retired structured-question practice contracts from active surfaces."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


EXCLUDED_PARTS = {".git", ".test-output", "node_modules", "dist", "coverage", "generated"}
ACTIVE_PREFIXES = (
    ("backend", "internal", "api", "practice"),
    ("backend", "internal", "practice"),
    ("backend", "internal", "store", "practice"),
    ("frontend", "src", "app", "screens", "practice"),
    ("frontend", "src", "app", "screens", "report"),
    ("openapi", "fixtures", "PracticePlans"),
    ("openapi", "fixtures", "PracticeSessions"),
    ("shared", "events", "schemas"),
)
STALE_PATTERNS = {
    "appendSessionEvent": re.compile(r"\bappendSessionEvent\b"),
    "PracticeTurn": re.compile(r"\bPracticeTurn\b"),
    "QuestionAssessment": re.compile(r"\bQuestionAssessment\b"),
    "questionBudget": re.compile(r"\bquestionBudget\b"),
    "hintsEnabled": re.compile(r"\bhintsEnabled\b"),
    "practiceMode": re.compile(r"\bpracticeMode\b"),
    "first-question feature key": re.compile(r"practice\.session\.first_question"),
    "follow-up feature key": re.compile(r"practice\.session\.follow_up|practice\.followup"),
    "turn feature key": re.compile(r"practice\.turn\."),
}
REQUIRED_FILES = (
    ("backend", "internal", "practice", "message_service.go"),
    ("backend", "internal", "store", "practice", "messages.go"),
    ("openapi", "fixtures", "PracticeSessions", "sendPracticeMessage.json"),
)
FORBIDDEN_FILES = (
    ("backend", "internal", "practice", "question_generation.go"),
    ("backend", "internal", "practice", "hint_ai.go"),
    ("backend", "internal", "practice", "session_event.go"),
    ("backend", "cmd", "api", "practice_http_scenario_test.go"),
    ("shared", "events", "schemas", "practice.turn.completed.v1.json"),
)


def is_active_file(path: Path, repo_root: Path) -> bool:
    try:
        parts = path.relative_to(repo_root).parts
    except ValueError:
        return False
    if any(part in EXCLUDED_PARTS for part in parts):
        return False
    if path.name.endswith("_test.go") or "/__tests__/" in path.as_posix() or path.name.endswith(".test.ts") or path.name.endswith(".test.tsx"):
        return False
    return any(parts[: len(prefix)] == prefix for prefix in ACTIVE_PREFIXES)


def scan_active_surfaces(repo_root: Path) -> list[str]:
    problems: list[str] = []
    for path in repo_root.rglob("*"):
        if not path.is_file() or not is_active_file(path, repo_root):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            for label, pattern in STALE_PATTERNS.items():
                if pattern.search(line):
                    problems.append(f"{path}:{lineno}: stale structured-practice contract {label}")
    return problems


def scan_file_inventory(repo_root: Path) -> list[str]:
    problems: list[str] = []
    for parts in REQUIRED_FILES:
        path = repo_root.joinpath(*parts)
        if not path.is_file():
            problems.append(f"{path}: missing current conversation artifact")
    for parts in FORBIDDEN_FILES:
        path = repo_root.joinpath(*parts)
        if path.exists():
            problems.append(f"{path}: retired structured-practice artifact still exists")
    return problems


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--phase", choices=("phase0", "phase3", "all"), default="phase0")
    args = parser.parse_args(argv)
    repo_root = Path(args.repo_root).resolve()
    problems = scan_file_inventory(repo_root)
    if args.phase in {"phase3", "all"}:
        problems.extend(scan_active_surfaces(repo_root))
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print(f"backend_practice_out_of_scope {args.phase}: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
