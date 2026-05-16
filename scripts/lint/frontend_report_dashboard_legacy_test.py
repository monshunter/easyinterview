"""Self-test for frontend_report_dashboard_legacy.py.

Verifies the lint includes every forbidden term required by plan §5.6 and
allows negative-assertion docs to mention them without tripping the gate.
"""

from __future__ import annotations

import subprocess
import sys
import tempfile
from pathlib import Path

import pytest


SCRIPT = (Path(__file__).resolve().parent / "frontend_report_dashboard_legacy.py").resolve()
REPO_ROOT = SCRIPT.parents[2]


REQUIRED_TERMS = (
    "reportLayout",
    "report_layout",
    "fully_prepared",
    "readinessScore",
    "readiness_score",
    "mistakes_queue",
    "mistakesQueue",
    "drill_builder",
    "drillBuilder",
    "growth_center",
    "growthCenter",
    "report_timeline",
    "reportTimeline",
    "report_form",
    "reportForm",
    "createPracticeVoiceTurn",
    "getCompanyIntel",
    "getDebrief",
    "VoiceSessionSurface",
    "PracticeWaveformBars",
    "listTargetJobReports",
    "--ei-bg",
    "--ei-ink",
    "--ei-accent",
    "--ei-rule",
    "--ei-danger",
    "--ei-ok",
    "--ei-warn",
    "--ei-cool",
    "--ei-amber",
    "--ei-sans",
    "--ei-mono",
)


def test_frontend_report_dashboard_legacy_includes_terms() -> None:
    text = SCRIPT.read_text(encoding="utf-8")
    for term in REQUIRED_TERMS:
        assert term in text, f"lint script missing forbidden term: {term}"


def test_frontend_report_dashboard_legacy_passes_on_clean_repo() -> None:
    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(REPO_ROOT), "--phase", "all"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, (
        f"lint failed unexpectedly on a clean repo:\nstdout={result.stdout}\nstderr={result.stderr}"
    )


def test_frontend_report_dashboard_legacy_allows_negative_docs() -> None:
    # Stand up a throwaway frontend skeleton and confirm allowed files are not
    # flagged even when they enumerate retired vocabulary.
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        target = root / "frontend" / "src" / "app" / "screens" / "report" / "__tests__"
        target.mkdir(parents=True, exist_ok=True)
        legacy_negative = target / "legacyNegative.test.ts"
        legacy_negative.write_text(
            "// negative: reportLayout / mistakesQueue / VoiceSessionSurface\n",
            encoding="utf-8",
        )
        scripts_dir = root / "scripts" / "lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        # Copy the lint script into the temp tree so its self-path skip still works.
        copy_target = scripts_dir / "frontend_report_dashboard_legacy.py"
        copy_target.write_bytes(SCRIPT.read_bytes())
        result = subprocess.run(
            [sys.executable, str(copy_target), "--repo-root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        assert result.returncode == 0, (
            f"lint flagged an allowed negative-assertion file:\nstdout={result.stdout}\nstderr={result.stderr}"
        )


if __name__ == "__main__":
    pytest.main([__file__, "-q"])
