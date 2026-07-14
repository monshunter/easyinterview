#!/usr/bin/env python3
"""Scoped gate for frontend-report-dashboard/001 production boundaries.

Asserts that out-of-scope identifiers are absent from the report surfaces and
that ReportsScreen is the only production screen allowed to consume
``listTargetJobReports``. The plan / BDD / test docs / spec / preflight test
and this script itself may mention the terms as negative assertions.

Run from repo root: `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py
--repo-root . --phase all`.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path
from typing import Iterable

FORBIDDEN_PATTERNS: tuple[str, ...] = (
    r"reportLayout",
    r"report_layout",
    r"fully_prepared",  # out-of-scope 5-tier readiness ordinal
    r"readinessScore",
    r"readiness_score",
    r"mistakes_queue",
    r"mistakesQueue",
    r"drill_builder",
    r"drillBuilder",
    r"growth_center",
    r"growthCenter",
    r"report_timeline",
    r"reportTimeline",
    r"report_form",
    r"reportForm",
    r"createPracticeVoiceTurn",
    r"getCompanyIntel",
    r"getDebrief",
    r"VoiceSessionSurface",
    r"PracticeWaveformBars",
    r"window\.EI_DATA",
    r"ui-design/src/data",
    r"reportHistory",
    r"report_history",
    r"reportVersions",
    r"report_versions",
    r"Report Center",
    r"报告中心",
    # Prototype-only short CSS variables are not defined by the formal D2
    # token system. Report/generating implementation must use --ei-color-* and
    # --ei-font-* tokens so theme / dark / customAccent changes are real gates.
    r"--ei-bg(?:-[a-z]+)?\b",
    r"--ei-ink[23]?\b",
    r"--ei-accent(?:-soft)?\b",
    r"--ei-rule\b",
    r"--ei-danger(?:-soft)?\b",
    r"--ei-ok(?:-soft)?\b",
    r"--ei-warn(?:-soft)?\b",
    r"--ei-cool(?:-soft)?\b",
    r"--ei-amber(?:-soft)?\b",
    r"--ei-sans\b",
    r"--ei-mono\b",
)

LIST_OPERATION = "listTargetJobReports"
EXPECTED_LIST_CONSUMER = "frontend/src/app/screens/reports/ReportsScreen.tsx"

# Files (relative to repo root) that legitimately mention out-of-scope terms because
# they are negative assertions / documentation / this lint script itself.
ALLOWED_FILES: tuple[str, ...] = (
    "scripts/lint/frontend_report_dashboard_out_of_scope.py",
    "scripts/lint/frontend_report_dashboard_out_of_scope_test.py",
    "frontend/src/app/screens/report/__tests__/preflight.test.ts",
    "frontend/src/app/screens/report/__tests__/outOfScopeNegative.test.ts",
    "frontend/src/app/screens/generating/__tests__/outOfScopeNegative.test.ts",
)


def production_typescript_files(base: Path) -> Iterable[Path]:
    if not base.exists():
        return
    for path in base.rglob("*"):
        if path.is_file() and path.suffix in {".ts", ".tsx"}:
            if "__tests__" in path.parts or ".test." in path.name or ".spec." in path.name:
                continue
            yield path


def walk_report_surfaces(repo_root: Path) -> Iterable[Path]:
    target_dirs = [
        repo_root / "frontend" / "src" / "app" / "screens" / "reports",
        repo_root / "frontend" / "src" / "app" / "screens" / "report",
        repo_root / "frontend" / "src" / "app" / "screens" / "generating",
    ]
    for base in target_dirs:
        yield from production_typescript_files(base)


def list_operation_consumers(repo_root: Path) -> list[str]:
    screens = repo_root / "frontend" / "src" / "app" / "screens"
    return sorted(
        path.relative_to(repo_root).as_posix()
        for path in production_typescript_files(screens)
        if LIST_OPERATION in path.read_text(encoding="utf-8")
    )


def is_allowed(repo_root: Path, file_path: Path) -> bool:
    rel = file_path.relative_to(repo_root).as_posix()
    return rel in ALLOWED_FILES


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".", help="Repo root path")
    parser.add_argument(
        "--phase",
        default="all",
        help="Phase tag (informational only)",
    )
    args = parser.parse_args()
    repo_root = Path(args.repo_root).resolve()

    failures: list[str] = []
    for file_path in walk_report_surfaces(repo_root):
        if is_allowed(repo_root, file_path):
            continue
        text = file_path.read_text(encoding="utf-8")
        for pattern in FORBIDDEN_PATTERNS:
            if re.search(pattern, text):
                failures.append(
                    f"{file_path.relative_to(repo_root)} contains forbidden literal: {pattern}"
                )

    consumers = list_operation_consumers(repo_root)
    if consumers != [EXPECTED_LIST_CONSUMER]:
        failures.append(
            f"{LIST_OPERATION} production screen consumers must equal "
            f"[{EXPECTED_LIST_CONSUMER}], got {consumers}"
        )

    if failures:
        print("frontend-report-dashboard out-of-scope lint FAILED:")
        for failure in failures:
            print(f"  - {failure}")
        return 1
    print(
        "frontend-report-dashboard out-of-scope lint OK "
        f"(phase={args.phase}, files scanned in "
        "frontend/src/app/screens/{reports,report,generating}; "
        f"only consumer={EXPECTED_LIST_CONSUMER})"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
