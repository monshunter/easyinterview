#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
SCENARIO_DIR = SCRIPT_DIR.parent
ROOT = SCRIPT_DIR.parents[4]


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


class ReportsCanonicalRoutingScriptContractTest(unittest.TestCase):
    def test_assets_and_index_define_target_scoped_reports_history(self) -> None:
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
            "section=reports",
            "reload",
            "back / forward",
            "topbar-nav-reports",
        ):
            self.assertIn(marker, assets)
        for superseded_claim in (
            "report-failure-state",
            "并保留 errorCode",
            "generating/report 保留各自的 `reportId` / `reportStatus`",
        ):
            self.assertNotIn(superseded_claim, assets)

        index = read(ROOT / "test/scenarios/e2e/INDEX.md")
        row = next(line for line in index.splitlines() if "E2E.P0.088" in line)
        self.assertIn("/reports?targetJobId=", row)
        self.assertIn("targetJobId", row)

    def test_wrapper_runs_source_contract_and_verbose_scenario(self) -> None:
        trigger = read(SCRIPT_DIR / "trigger.sh")
        verify = read(SCRIPT_DIR / "verify.sh")
        for marker in (
            "source_contract_test.py",
            "p0-088-url-addressable-routing-canonical.test.tsx",
            "--reporter=verbose",
            "2>&1",
        ):
            self.assertIn(marker, trigger)
        for marker in (
            "source_contract_test.py",
            "direct-open /reports with hostile legacy params",
            "reload preserves the canonical Reports target context",
            "back/forward restores Reports with targetJobId only",
            "Test Files",
            "Tests",
        ):
            self.assertIn(marker, verify)


if __name__ == "__main__":
    unittest.main()
