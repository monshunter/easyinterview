#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
SCENARIO_DIR = SCRIPT_DIR.parent
ROOT = SCRIPT_DIR.parents[4]


def read(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


class ReportsHashFallbackScriptContractTest(unittest.TestCase):
    def test_assets_and_index_define_reports_hash_and_negative_boundaries(self) -> None:
        assets = "\n".join(
            read(path)
            for path in (
                SCENARIO_DIR / "README.md",
                SCENARIO_DIR / "data/seed-input.md",
                SCENARIO_DIR / "data/expected-outcome.md",
            )
        )
        for marker in (
            "#route=reports&targetJobId=",
            "/reports?targetJobId=",
            "section=reports",
            "reportId",
            "status",
            "roundId",
            "topbar-nav-reports",
        ):
            self.assertIn(marker, assets)

        index = read(ROOT / "test/scenarios/e2e/INDEX.md")
        row = next(line for line in index.splitlines() if "E2E.P0.090" in line)
        self.assertIn("#route=reports", row)
        self.assertIn("/reports", row)

    def test_wrapper_runs_source_scenario_codec_fallback_and_topbar_gates(self) -> None:
        trigger = read(SCRIPT_DIR / "trigger.sh")
        verify = read(SCRIPT_DIR / "verify.sh")
        for marker in (
            "source_contract_test.py",
            "p0-090-url-routing-hash-out-of-scope-negative.test.tsx",
            "bootstrapRoute.test.ts",
            "routeUrl.test.ts",
            "spaFallback.test.ts",
            "topbar/TopBar.test.tsx",
            "--reporter=verbose",
            "2>&1",
        ):
            self.assertIn(marker, trigger)
        for marker in (
            "source_contract_test.py",
            "Reports hash bootstrap keeps targetJobId only",
            "legacy Parse report params are stripped",
            "SPA fallback explicitly serves the known /reports path",
            "canonicalizes a Reports hash to targetJobId-only /reports",
            "Test Files",
            "Tests",
        ):
            self.assertIn(marker, verify)


if __name__ == "__main__":
    unittest.main()
