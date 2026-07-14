"""Self-test for frontend_report_dashboard_out_of_scope.py.

Verifies the lint includes every forbidden term required by plan §5.6 and
allows negative-assertion docs to mention them without tripping the gate.
"""

from __future__ import annotations

import subprocess
import sys
import tempfile
from pathlib import Path

import pytest


SCRIPT = (Path(__file__).resolve().parent / "frontend_report_dashboard_out_of_scope.py").resolve()
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
    "reportHistory",
    "reportVersions",
    "Report Center",
    "报告中心",
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


def test_frontend_report_dashboard_out_of_scope_includes_terms() -> None:
    text = SCRIPT.read_text(encoding="utf-8")
    for term in REQUIRED_TERMS:
        assert term in text, f"lint script missing forbidden term: {term}"


def test_frontend_report_dashboard_out_of_scope_passes_on_clean_repo() -> None:
    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(REPO_ROOT), "--phase", "all"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, (
        f"lint failed unexpectedly on a clean repo:\nstdout={result.stdout}\nstderr={result.stderr}"
    )


def test_frontend_report_dashboard_out_of_scope_allows_negative_docs() -> None:
    # Stand up a throwaway frontend skeleton and confirm allowed files are not
    # flagged even when they enumerate out-of-scope vocabulary.
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        target = root / "frontend" / "src" / "app" / "screens" / "report" / "__tests__"
        target.mkdir(parents=True, exist_ok=True)
        out_of_scope_negative = target / "outOfScopeNegative.test.ts"
        out_of_scope_negative.write_text(
            "// negative: reportLayout / mistakesQueue / VoiceSessionSurface\n",
            encoding="utf-8",
        )
        reports = root / "frontend" / "src" / "app" / "screens" / "reports" / "ReportsScreen.tsx"
        reports.parent.mkdir(parents=True, exist_ok=True)
        reports.write_text("client.listTargetJobReports(targetJobId);\n", encoding="utf-8")
        scripts_dir = root / "scripts" / "lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        # Copy the lint script into the temp tree so its self-path skip still works.
        copy_target = scripts_dir / "frontend_report_dashboard_out_of_scope.py"
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


def test_frontend_report_dashboard_out_of_scope_allows_only_reports_screen_consumer() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        reports = root / "frontend/src/app/screens/reports/ReportsScreen.tsx"
        reports.parent.mkdir(parents=True, exist_ok=True)
        reports.write_text("client.listTargetJobReports(targetJobId);\n", encoding="utf-8")
        scripts_dir = root / "scripts/lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        copy_target = scripts_dir / "frontend_report_dashboard_out_of_scope.py"
        copy_target.write_bytes(SCRIPT.read_bytes())

        result = subprocess.run(
            [sys.executable, str(copy_target), "--repo-root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        assert result.returncode == 0, result.stdout + result.stderr
        assert "only consumer=frontend/src/app/screens/reports/ReportsScreen.tsx" in result.stdout


def test_frontend_report_dashboard_out_of_scope_rejects_any_second_list_consumer() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        root = Path(tmpdir)
        reports = root / "frontend/src/app/screens/reports/ReportsScreen.tsx"
        parse = root / "frontend/src/app/screens/parse/ParseScreen.tsx"
        reports.parent.mkdir(parents=True, exist_ok=True)
        parse.parent.mkdir(parents=True, exist_ok=True)
        reports.write_text("client.listTargetJobReports(targetJobId);\n", encoding="utf-8")
        parse.write_text("client.listTargetJobReports(targetJobId);\n", encoding="utf-8")
        scripts_dir = root / "scripts/lint"
        scripts_dir.mkdir(parents=True, exist_ok=True)
        copy_target = scripts_dir / "frontend_report_dashboard_out_of_scope.py"
        copy_target.write_bytes(SCRIPT.read_bytes())

        result = subprocess.run(
            [sys.executable, str(copy_target), "--repo-root", str(root)],
            capture_output=True,
            text=True,
            check=False,
        )
        assert result.returncode == 1
        assert "listTargetJobReports production screen consumers" in result.stdout


if __name__ == "__main__":
    pytest.main([__file__, "-q"])
