#!/usr/bin/env python3
"""Scoped grep gate for frontend-report-dashboard/001 retired vocabulary.

Asserts that the following retired identifiers are NOT present in the
implementation code under frontend/src/app/screens/{report,generating}/. The
plan / BDD / test docs / spec / preflight test and this script itself are
allowed to mention them as negative assertions.

Run from repo root: `python3 scripts/lint/frontend_report_dashboard_legacy.py
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
    r"fully_prepared",  # legacy 5-tier readiness ordinal
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
    r"listTargetJobReports",
)

# Files (relative to repo root) that legitimately mention retired terms because
# they are negative assertions / documentation / this lint script itself.
ALLOWED_FILES: tuple[str, ...] = (
    "scripts/lint/frontend_report_dashboard_legacy.py",
    "scripts/lint/frontend_report_dashboard_legacy_test.py",
    "frontend/src/app/screens/report/__tests__/preflight.test.ts",
    "frontend/src/app/screens/report/__tests__/legacyNegative.test.ts",
    "frontend/src/app/screens/generating/__tests__/legacyNegative.test.ts",
)


def walk_screens(repo_root: Path) -> Iterable[Path]:
    target_dirs = [
        repo_root / "frontend" / "src" / "app" / "screens" / "report",
        repo_root / "frontend" / "src" / "app" / "screens" / "generating",
    ]
    for base in target_dirs:
        if not base.exists():
            continue
        for path in base.rglob("*"):
            if path.is_file() and path.suffix in {".ts", ".tsx"}:
                # Skip __tests__ entirely — by design tests carry negative
                # assertions that name the forbidden terms.
                if "__tests__" in path.parts:
                    continue
                yield path


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
    for file_path in walk_screens(repo_root):
        if is_allowed(repo_root, file_path):
            continue
        text = file_path.read_text(encoding="utf-8")
        for pattern in FORBIDDEN_PATTERNS:
            if re.search(pattern, text):
                failures.append(
                    f"{file_path.relative_to(repo_root)} contains forbidden literal: {pattern}"
                )

    if failures:
        print("frontend-report-dashboard legacy lint FAILED:")
        for failure in failures:
            print(f"  - {failure}")
        return 1
    print(
        f"frontend-report-dashboard legacy lint OK (phase={args.phase}, files scanned in frontend/src/app/screens/{{report,generating}})"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
