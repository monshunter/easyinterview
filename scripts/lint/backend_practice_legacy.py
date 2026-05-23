#!/usr/bin/env python3
"""Legacy-negative gates for backend-practice plans.

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
    "hint_disabled_globally",
    "legacy_hint_policy",
    "legacy_mode_assisted_value",
    "legacy debrief replay value",
    "warmup",
    "single_drill",
    "drill_builder",
    "mistake_queue",
    "growth_center",
    "practiceModeCard",
)
PHASE3_VOICE_ROUTE = re.compile(
    r"(/voice(?:[/?#\"'\s]|$)|\bvoice_route\b|\b(?:standalone|independent) voice route\b|\bvoice route alias\b)",
    re.IGNORECASE,
)
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
    ("test", "scenarios", "e2e", "p0-048-practice-hint-assisted-across-goals"),
    ("test", "scenarios", "e2e", "p0-049-practice-hint-strict-refusal"),
    ("test", "scenarios", "e2e", "p0-050-practice-hint-provenance-task-runs"),
    ("test", "scenarios", "e2e", "p0-051-practice-hint-degrade-privacy"),
)
PHASE3_EXCLUDED_SUFFIXES = (
    ("scripts", "verify.sh"),
)
BACKEND_PRACTICE_002_BDD_PLAN = (
    "docs",
    "spec",
    "backend-practice",
    "plans",
    "002-event-loop-and-completion",
    "bdd-plan.md",
)
BACKEND_PRACTICE_002_HTTP_SCENARIOS = (
    "backend",
    "cmd",
    "api",
    "practice_http_scenario_test.go",
)
E2E_INDEX = ("test", "scenarios", "e2e", "INDEX.md")
E2E_ID_RE = re.compile(r"E2E\.P0\.(\d{3})")
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


def repo_relative_path(path: Path, repo_root: Path) -> Path | None:
    try:
        return path.absolute().relative_to(repo_root.absolute())
    except ValueError:
        return None


def is_excluded(path: Path, repo_root: Path) -> bool:
    rel = repo_relative_path(path, repo_root)
    if rel is None:
        return True
    parts = rel.parts
    if any(part in EXCLUDED_PARTS for part in parts):
        return True
    return any(parts[: len(prefix)] == prefix for prefix in EXCLUDED_PREFIXES)


def iter_repo_files(repo_root: Path) -> list[Path]:
    out: list[Path] = []
    for path in repo_root.rglob("*"):
        if path.is_symlink():
            continue
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
    rel = repo_relative_path(path, repo_root)
    if rel is None:
        return False
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
                    if term == "legacy debrief replay value" and path.name.endswith("_test.go"):
                        continue
                    problems.append(f"{path}:{lineno}: retired backend-practice term {term!r}")
            if PHASE3_VOICE_ROUTE.search(line) and "practice-voice-mvp" not in line:
                problems.append(f"{path}:{lineno}: retired standalone voice route")
    return problems


def parse_e2e_index(index_path: Path) -> tuple[dict[str, str], list[str]]:
    entries: dict[str, str] = {}
    problems: list[str] = []
    if not index_path.exists():
        return entries, [f"{index_path}: missing e2e scenario index"]
    for lineno, line in enumerate(index_path.read_text(encoding="utf-8").splitlines(), start=1):
        match = E2E_ID_RE.search(line)
        if not match:
            continue
        scenario_id = f"E2E.P0.{match.group(1)}"
        if scenario_id in entries:
            problems.append(f"{index_path}:{lineno}: duplicate scenario id {scenario_id}")
            continue
        entries[scenario_id] = line
    return entries, problems


def backend_practice_002_assigned_bdd_ids(bdd_plan_path: Path) -> tuple[list[str], list[str]]:
    if not bdd_plan_path.exists():
        return [], [f"{bdd_plan_path}: missing backend-practice 002 bdd-plan.md"]
    assigned: list[str] = []
    for line in bdd_plan_path.read_text(encoding="utf-8").splitlines():
        if line.startswith("- 编号分配:"):
            assigned = [f"E2E.P0.{match}" for match in E2E_ID_RE.findall(line)]
            break
    if not assigned:
        return [], [f"{bdd_plan_path}: missing backend-practice 002 BDD 编号分配 line"]
    duplicates = sorted({scenario_id for scenario_id in assigned if assigned.count(scenario_id) > 1})
    if duplicates:
        return assigned, [f"{bdd_plan_path}: duplicate backend-practice 002 BDD ids {', '.join(duplicates)}"]
    if len(assigned) != 6:
        return assigned, [f"{bdd_plan_path}: expected 6 backend-practice 002 BDD ids, got {len(assigned)}"]
    return assigned, []


def scan_backend_practice_002_bdd_ids(repo_root: Path) -> list[str]:
    problems: list[str] = []
    assigned, assigned_problems = backend_practice_002_assigned_bdd_ids(repo_root.joinpath(*BACKEND_PRACTICE_002_BDD_PLAN))
    problems.extend(assigned_problems)
    index_entries, index_problems = parse_e2e_index(repo_root.joinpath(*E2E_INDEX))
    problems.extend(index_problems)
    for scenario_id in assigned:
        index_line = index_entries.get(scenario_id)
        if index_line is None:
            continue
        if "practice" not in index_line.lower():
            problems.append(f"{repo_root.joinpath(*E2E_INDEX)}: backend-practice 002 id {scenario_id} collides with indexed scenario: {index_line}")

    scenario_test_path = repo_root.joinpath(*BACKEND_PRACTICE_002_HTTP_SCENARIOS)
    if not scenario_test_path.exists():
        problems.append(f"{scenario_test_path}: missing backend-practice 002 HTTP scenario test file")
        return problems
    scenario_test = scenario_test_path.read_text(encoding="utf-8")
    for scenario_id in assigned:
        digits = scenario_id.rsplit(".", maxsplit=1)[1]
        test_name_fragment = f"TestE2EP0{digits}"
        if test_name_fragment not in scenario_test:
            problems.append(f"{scenario_test_path}: missing Go HTTP scenario test for {scenario_id} ({test_name_fragment}*)")
    if "TestE2EP00Practice" in scenario_test:
        problems.append(f"{scenario_test_path}: malformed backend-practice E2E test name without numeric id")
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
        problems.extend(scan_backend_practice_002_bdd_ids(repo_root))
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
