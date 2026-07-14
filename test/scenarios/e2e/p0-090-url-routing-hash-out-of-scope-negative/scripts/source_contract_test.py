#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]


def read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


class ReportsHashFallbackSourceContractTest(unittest.TestCase):
    def test_scenario_covers_reports_hash_fallback_legacy_strip_and_topbar_negative(self) -> None:
        source = read(
            "frontend/src/app/scenarios/"
            "p0-090-url-routing-hash-out-of-scope-negative.test.tsx"
        )
        for marker in (
            "`#route=workspace&targetJobId=...` bootstrap rewrites to target-scoped detail",
            'expect(window.location.search).toBe("?targetJobId=tj-1")',
            'screen.getByTestId("workspace-detail-loading")',
            "#route=reports&targetJobId=",
            "/reports?targetJobId=",
            "section=reports",
            "reportId=",
            "status=",
            "roundId=",
            'screen.queryByTestId("topbar-nav-reports")',
            "isCanonicalFrontendPath(",
        ):
            self.assertIn(marker, source)

    def test_formal_hash_codec_and_fallback_sources_cover_reports(self) -> None:
        bootstrap = read("frontend/src/app/bootstrapRoute.test.ts")
        route_url = read("frontend/src/app/routeUrl.test.ts")
        fallback = read("frontend/src/app/spaFallback.test.ts")
        topbar = read("frontend/src/app/topbar/TopBar.test.tsx")
        app = read("frontend/src/app/App.test.tsx")
        self.assertIn(
            "canonicalizes a Reports hash to targetJobId-only /reports", bootstrap
        )
        self.assertIn(
            "drops the retired Parse reports section and all report business authority",
            route_url,
        )
        for marker in (
            "retains targetJobId as the sole workspace detail locator",
            "parses targetJobId as the sole canonical workspace detail locator",
            "autoStartPractice",
            "unknownKey",
            "rawText",
            "token",
        ):
            self.assertIn(marker, route_url)
        self.assertIn(
            "hash adapter and canonical codec retain only the workspace target locator",
            bootstrap,
        )
        for marker in (
            "renders a target-scoped workspace as read-only detail with one getTargetJob",
            'queryByTestId("parse-loading-step-0")',
        ):
            self.assertIn(marker, app)
        self.assertIn('/reports?targetJobId=tj-1", "/tmp/dist"', fallback)
        self.assertIn('queryByTestId("topbar-nav-reports")', topbar)


if __name__ == "__main__":
    unittest.main()
