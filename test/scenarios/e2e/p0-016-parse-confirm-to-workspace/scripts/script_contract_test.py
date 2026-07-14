#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
SCENARIO_DIR = SCRIPT_DIR.parent

FRONTEND_FILES = (
    "src/api/targetJob.realApiMode.test.ts",
    "src/app/screens/parse/ParseReports.test.tsx",
    "src/app/screens/parse/ParseScreen.test.tsx",
    "src/app/screens/parse/ParseEdit.test.tsx",
    "src/app/screens/parse/ParseAuthGate.test.tsx",
    "src/app/screens/parse/ParseResumeBinding.test.tsx",
    "src/app/screens/home/MockInterviewCard.test.tsx",
    "src/app/screens/home/HomeRecentMocks.test.tsx",
    "src/app/navigation/interviewContext.test.ts",
    "src/app/routeUrl.test.ts",
    "src/app/topbar/TopBar.test.tsx",
)

PLAYWRIGHT_TITLES = (
    "readonly plan detail exposes only direct start with bound resume context",
    "plan-detail reports entry matches the UI truth and keeps Parse report-free",
    "start interview hands off directly to practice with bound resume",
)


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


class ScenarioScriptContractTest(unittest.TestCase):
    def test_docs_own_page_entry_without_embedded_reports(self) -> None:
        combined = "\n".join(
            read(path)
            for path in (
                SCENARIO_DIR / "README.md",
                SCENARIO_DIR / "data/seed-input.md",
                SCENARIO_DIR / "data/expected-outcome.md",
            )
        )
        for marker in (
            "readonly",
            "Start",
            "parse-reports-entry",
            "/reports?targetJobId=",
            "TopBar",
            "listTargetJobReports",
            "section=reports",
            "1440x900",
            "390x844",
        ):
            self.assertIn(marker, combined)
        self.assertIn("does not call `listTargetJobReports`", combined)
        self.assertIn("does not embed", combined)

    def test_trigger_runs_current_source_unit_build_and_browser_gates(self) -> None:
        trigger = read(SCRIPT_DIR / "trigger.sh")
        for test_file in FRONTEND_FILES:
            self.assertIn(test_file, trigger)
        for title in PLAYWRIGHT_TITLES:
            self.assertIn(title, trigger)
        for marker in (
            "source_contract_test.py",
            "frontend build",
            "tests/pixel-parity/parse.spec.ts",
            "CI=1",
            "--project=desktop",
            "--project=mobile",
            "--workers=1",
            "--retries=0",
            "2>&1",
        ):
            self.assertIn(marker, trigger)
        for stale in (
            "./internal/review",
            "./internal/store/review",
            "./internal/api/reports",
            "parse_reports_evidence_test.mjs",
            "section=reports",
        ):
            self.assertNotIn(stale, trigger)

    def test_verifier_binds_runner_passes_and_entry_evidence(self) -> None:
        verify = read(SCRIPT_DIR / "verify.sh")
        for test_file in FRONTEND_FILES:
            self.assertIn(pathlib.Path(test_file).name, verify)
        for title in PLAYWRIGHT_TITLES:
            self.assertIn(title, verify)
        for marker in (
            "6 passed",
            "parseListRequestsBeforeClick=0",
            "topbarReportsEntry=0",
            "embeddedReports=0",
            "sectionReportsAccepted=false",
            "changedRatio=",
            "no tests to run",
            "--- FAIL:",
        ):
            self.assertIn(marker, verify)

    def test_source_gate_owns_entry_and_reports_screen_only_consumption(self) -> None:
        source_gate = read(SCRIPT_DIR / "source_contract_test.py")
        for marker in (
            "listTargetJobReports",
            "ReportsScreen.tsx",
            "ParseScreen.tsx",
            "parse-reports-entry",
            "topbar-nav-reports",
            "section=reports",
            "parse-reports",
        ):
            self.assertIn(marker, source_gate)

    def test_old_embedded_reports_evidence_contract_has_zero_references(self) -> None:
        combined = "\n".join(
            read(path)
            for path in SCENARIO_DIR.rglob("*")
            if path.is_file() and path.name != "script_contract_test.py"
        )
        self.assertNotIn("parse_reports_evidence", combined)
        self.assertNotIn("canonical-round reports match", combined)

    def test_setup_and_cleanup_keep_scenario_output_bounded(self) -> None:
        setup = read(SCRIPT_DIR / "setup.sh")
        cleanup = read(SCRIPT_DIR / "cleanup.sh")
        self.assertIn("setup.env", setup)
        self.assertIn("trigger.log", setup)
        self.assertIn("setup.env", cleanup)
        self.assertNotIn("evidence-server.pid", setup + cleanup)
        self.assertNotIn("playwright", cleanup)


if __name__ == "__main__":
    unittest.main()
