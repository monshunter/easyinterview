#!/usr/bin/env python3
"""Scoped grep gate for frontend-debrief/001 retired vocabulary.

Asserts that the following retired identifiers are NOT present in the
implementation code under frontend/src/app/screens/debrief/ and the
frontend/src/app/i18n/locales/ catalog. The plan / BDD / test docs /
spec / scenario directories are allowed to mention them as negative
assertions.

Run from repo root:
    python3 scripts/lint/frontend_debrief_legacy.py --repo-root . --phase all
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path
from typing import Iterable

FORBIDDEN_PATTERNS: tuple[str, ...] = (
    r"experience_library",
    r"experienceLibrary",
    r"star_editor",
    r"starEditor",
    r"drill_builder",
    r"drillBuilder",
    r"mistakes_book",
    r"mistakesBook",
    r"growth_center",
    r"growthCenter",
    r"report_timeline",
    r"reportTimeline",
    # Historical standalone-voice route alias; current spec collapses voice
    # back into practice params (see normalizeRoute.ts ALIAS table).
    r'route\s*:\s*["\']voice["\']',
    # Retired full-screen route alias must NEVER appear in the formal
    # debrief screen — it is normalized once in normalizeRoute.ts only.
    r'name\s*:\s*["\']debrief_full["\']',
)

ALLOWED_FILES: tuple[str, ...] = (
    "scripts/lint/frontend_debrief_legacy.py",
    "scripts/lint/frontend_debrief_legacy_test.py",
)


def walk_targets(repo_root: Path) -> Iterable[Path]:
    target_dirs = [
        repo_root / "frontend" / "src" / "app" / "screens" / "debrief",
        repo_root / "frontend" / "src" / "app" / "i18n" / "locales",
    ]
    for base in target_dirs:
        if not base.exists():
            continue
        for path in base.rglob("*"):
            if not path.is_file():
                continue
            if path.suffix not in {".ts", ".tsx"}:
                continue
            # Skip co-located tests / __tests__ — by design they may carry
            # negative assertions naming the retired terms.
            if "__tests__" in path.parts:
                continue
            if path.name.endswith(".test.ts") or path.name.endswith(".test.tsx"):
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
    for file_path in walk_targets(repo_root):
        if is_allowed(repo_root, file_path):
            continue
        text = file_path.read_text(encoding="utf-8")
        for pattern in FORBIDDEN_PATTERNS:
            if re.search(pattern, text):
                failures.append(
                    f"{file_path.relative_to(repo_root)} contains forbidden literal: {pattern}"
                )

    if failures:
        print("frontend-debrief legacy lint FAILED:")
        for failure in failures:
            print(f"  - {failure}")
        return 1
    print(
        f"frontend-debrief legacy lint OK (phase={args.phase}, "
        f"files scanned in frontend/src/app/screens/debrief + "
        f"frontend/src/app/i18n/locales)"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
