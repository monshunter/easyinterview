"""Self-test for frontend_debrief_legacy.py.

Verifies the lint enumerates every forbidden term required by Phase 8.7 and
passes on a clean repository.
"""

from __future__ import annotations

import subprocess
import sys
import tempfile
from pathlib import Path

import pytest


SCRIPT = (Path(__file__).resolve().parent / "frontend_debrief_legacy.py").resolve()
REPO_ROOT = SCRIPT.parents[2]


REQUIRED_TERMS = (
    "experience_library",
    "experienceLibrary",
    "star_editor",
    "starEditor",
    "drill_builder",
    "drillBuilder",
    "mistakes_book",
    "mistakesBook",
    "growth_center",
    "growthCenter",
    "report_timeline",
    "reportTimeline",
    "debrief_full",
)


def test_frontend_debrief_legacy_includes_terms() -> None:
    text = SCRIPT.read_text(encoding="utf-8")
    for term in REQUIRED_TERMS:
        assert term in text, f"lint script missing forbidden term: {term}"


def test_frontend_debrief_legacy_passes_on_clean_repo() -> None:
    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(REPO_ROOT), "--phase", "all"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, (
        f"lint failed unexpectedly on a clean repo:\nstdout={result.stdout}\nstderr={result.stderr}"
    )


def test_frontend_debrief_legacy_flags_offender() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        target = root / "frontend" / "src" / "app" / "screens" / "debrief"
        target.mkdir(parents=True, exist_ok=True)
        offender = target / "OffenderModule.tsx"
        offender.write_text(
            "export const x = 'experienceLibrary'; // retired vocabulary\n",
            encoding="utf-8",
        )
        scripts_dir = root / "scripts" / "lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        copy_target = scripts_dir / "frontend_debrief_legacy.py"
        copy_target.write_bytes(SCRIPT.read_bytes())
        result = subprocess.run(
            [sys.executable, str(copy_target), "--repo-root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        assert result.returncode == 1, (
            f"lint should have caught offender:\nstdout={result.stdout}\nstderr={result.stderr}"
        )


def test_frontend_debrief_legacy_skips_test_files() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        target = root / "frontend" / "src" / "app" / "screens" / "debrief"
        target.mkdir(parents=True, exist_ok=True)
        legacy_test = target / "Something.test.tsx"
        legacy_test.write_text(
            "// negative: experienceLibrary / drillBuilder must not appear\n",
            encoding="utf-8",
        )
        scripts_dir = root / "scripts" / "lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        copy_target = scripts_dir / "frontend_debrief_legacy.py"
        copy_target.write_bytes(SCRIPT.read_bytes())
        result = subprocess.run(
            [sys.executable, str(copy_target), "--repo-root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        assert result.returncode == 0, (
            f"lint should skip *.test.tsx files:\nstdout={result.stdout}\nstderr={result.stderr}"
        )


if __name__ == "__main__":
    pytest.main([__file__, "-q"])
