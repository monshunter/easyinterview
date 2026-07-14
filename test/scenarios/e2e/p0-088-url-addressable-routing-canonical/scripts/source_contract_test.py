#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]


def read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


class ReportsCanonicalRoutingSourceContractTest(unittest.TestCase):
    def test_scenario_covers_reports_direct_reload_navigation_and_history(self) -> None:
        source = read(
            "frontend/src/app/scenarios/"
            "p0-088-url-addressable-routing-canonical.test.tsx"
        )
        for marker in (
            'name: "reports"',
            "/reports?targetJobId=",
            "section=reports",
            'screen.getByTestId("reports-screen")',
            'screen.queryByTestId("topbar-nav-reports")',
            "window.history.back()",
            "window.history.forward()",
            "unmount()",
        ):
            self.assertIn(marker, source)

    def test_route_sources_keep_reports_target_scoped_and_non_primary(self) -> None:
        history = read("frontend/src/app/AppRoutingHistory.test.tsx")
        route_url = read("frontend/src/app/routeUrl.test.ts")
        topbar = read("frontend/src/app/topbar/TopBar.test.tsx")

        for marker in (
            "direct-opens Reports with targetJobId only and keeps chrome visible",
            "replaces an untrusted Reports deep link with workspace without adding a back-loop",
            "navigate(next) pushes a target-scoped Reports URL without legacy params",
            "popstate restores Reports and scrubs incompatible query state",
        ):
            self.assertIn(marker, history)
        for marker in (
            "serializes Reports with targetJobId as its only safe context",
            "parses Reports with targetJobId only",
            "drops the retired Parse reports section and all report business authority",
        ):
            self.assertIn(marker, route_url)
        self.assertIn('not.toContain("reports")', topbar)
        self.assertIn('queryByTestId("topbar-nav-reports")', topbar)


if __name__ == "__main__":
    unittest.main()
