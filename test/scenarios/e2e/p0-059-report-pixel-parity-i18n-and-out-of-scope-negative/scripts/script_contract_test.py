#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
SCENARIO_DIR = SCRIPT_DIR.parent
REPO_ROOT = SCRIPT_DIR.parents[4]

FRONTEND_FILES = (
    "src/api/targetJobReports.test.ts",
    "src/app/screens/reports/__tests__/ReportsScreen.test.tsx",
    "src/app/screens/report/__tests__/reportContract.test.ts",
    "src/app/screens/report/__tests__/ConversationReport.test.tsx",
    "src/app/screens/generating/__tests__/GeneratingBackNavigation.test.tsx",
    "src/app/screens/generating/__tests__/GeneratingScreen.test.tsx",
    "src/app/screens/report/__tests__/preflight.test.ts",
    "src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts",
    "src/app/screens/report/__tests__/outOfScopeNegative.test.ts",
    "src/app/screens/generating/__tests__/outOfScopeNegative.test.ts",
)

PLAYWRIGHT_SPECS = (
    "tests/pixel-parity/reports.spec.ts",
    "tests/pixel-parity/report.spec.ts",
    "tests/pixel-parity/generating.spec.ts",
)


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


class ScenarioScriptContractTest(unittest.TestCase):
    def test_docs_own_current_plan_reports_states_and_parity(self) -> None:
        combined = "\n".join(
            read(path)
            for path in (
                SCENARIO_DIR / "README.md",
                SCENARIO_DIR / "data/seed-input.md",
                SCENARIO_DIR / "data/expected-outcome.md",
            )
        )
        for marker in (
            "ReportsScreen",
            "current plan",
            "currentReport",
            "latestAttempt",
            "loading",
            "empty",
            "error",
            "latest-ready",
            "Back",
            "1440x900",
            "390x844",
            "current/latest-only",
            "only production screen consumer",
        ):
            self.assertIn(marker, combined)

    def test_trigger_runs_source_unit_build_and_all_browser_parity_gates(self) -> None:
        trigger = read(SCRIPT_DIR / "trigger.sh")
        for test_file in FRONTEND_FILES:
            self.assertIn(test_file, trigger)
        for spec in PLAYWRIGHT_SPECS:
            self.assertIn(spec, trigger)
        for marker in (
            "script_contract_test.py",
            "frontend_report_dashboard_out_of_scope.py",
            "frontend_report_dashboard_out_of_scope_test.py",
            "frontend build",
            "CI=1",
            "--project=desktop",
            "--project=mobile",
            "--workers=1",
            "--retries=0",
            "2>&1",
        ):
            self.assertIn(marker, trigger)

    def test_verifier_binds_reports_states_unique_consumer_and_parity(self) -> None:
        verify = read(SCRIPT_DIR / "verify.sh")
        for test_file in FRONTEND_FILES:
            self.assertIn(pathlib.Path(test_file).name, verify)
        for spec in PLAYWRIGHT_SPECS:
            self.assertIn(pathlib.Path(spec).name, verify)
        for marker in (
            "current-plan reports ready state matches the UI truth",
            "reports loading empty error latest-ready and mismatch states match the UI truth",
            "currentPlanIsolation=true",
            "currentLatestOnly=true",
            "backTarget=parse",
            "changedRatio=",
            "only consumer=frontend/src/app/screens/reports/ReportsScreen.tsx",
            "no tests to run",
            "--- FAIL:",
        ):
            self.assertIn(marker, verify)

    def test_success_cleanup_removes_only_scenario_outputs(self) -> None:
        cleanup = read(SCRIPT_DIR / "cleanup.sh")
        self.assertIn("trigger.log", cleanup)
        self.assertIn("setup.env", cleanup)
        self.assertNotIn("frontend/.playwright-output", cleanup)

    def test_report_and_generating_parity_return_to_the_independent_reports_page(self) -> None:
        for relative_path in (
            "frontend/tests/pixel-parity/report.spec.ts",
            "frontend/tests/pixel-parity/generating.spec.ts",
        ):
            source = read(REPO_ROOT / relative_path)
            self.assertNotIn("/parse?section=reports", source, relative_path)
            self.assertIn("/reports?targetJobId=", source, relative_path)


if __name__ == "__main__":
    unittest.main()
