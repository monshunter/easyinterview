#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
SCENARIO_DIR = SCRIPT_DIR.parent
ROOT = SCRIPT_DIR.parents[4]


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


class ReportsAuthPrivacyScriptContractTest(unittest.TestCase):
    def test_assets_and_index_define_reports_auth_restore(self) -> None:
        assets = "\n".join(
            read(path)
            for path in (
                SCENARIO_DIR / "README.md",
                SCENARIO_DIR / "data/seed-input.md",
                SCENARIO_DIR / "data/expected-outcome.md",
            )
        )
        for marker in (
            "/reports?targetJobId=",
            "pendingRoute=reports",
            "/workspace?targetJobId=",
            "autoStartPractice",
            "unknown",
            "section",
            "reportId",
            "status",
            "roundId",
            "ZERO",
        ):
            self.assertIn(marker, assets)

        index = read(ROOT / "test/scenarios/e2e/INDEX.md")
        row = next(line for line in index.splitlines() if "E2E.P0.089" in line)
        self.assertIn("Reports", row)
        self.assertIn("targetJobId", row)

    def test_wrapper_runs_source_auth_privacy_and_scenario_gates(self) -> None:
        trigger = read(SCRIPT_DIR / "trigger.sh")
        verify = read(SCRIPT_DIR / "verify.sh")
        for marker in (
            "source_contract_test.py",
            "p0-089-url-routing-auth-privacy.test.tsx",
            "auth/AppPendingAction.test.tsx",
            "AppRoutingPrivacy.test.tsx",
            "routeUrl.test.ts",
            "App.test.tsx",
            "--reporter=verbose",
            "2>&1",
        ):
            self.assertIn(marker, trigger)
        for marker in (
            "source_contract_test.py",
            "restores an unauthenticated Reports deep link with targetJobId only",
            "navigate(reports) keeps only targetJobId",
            "navigate(workspace) with raw markers drops every marker",
            "renders a target-scoped workspace as read-only detail with one getTargetJob",
            "Test Files",
            "Tests",
        ):
            self.assertIn(marker, verify)


if __name__ == "__main__":
    unittest.main()
