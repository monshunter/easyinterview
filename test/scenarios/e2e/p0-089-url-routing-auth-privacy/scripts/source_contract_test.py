#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]


def read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


class ReportsAuthPrivacySourceContractTest(unittest.TestCase):
    def test_scenario_restores_reports_and_strips_hostile_authority(self) -> None:
        source = read(
            "frontend/src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx"
        )
        for marker in (
            "restores an unauthenticated Reports deep link with targetJobId only",
            "/reports?targetJobId=",
            'pendingRoute")).toBe("reports")',
            'screen.getByTestId("reports-screen")',
            '"section"',
            '"reportId"',
            '"status"',
            '"roundId"',
            "expectNoRawMarkerLeak",
        ):
            self.assertIn(marker, source)

    def test_formal_auth_and_privacy_tests_cover_reports(self) -> None:
        pending = read("frontend/src/app/auth/AppPendingAction.test.tsx")
        privacy = read("frontend/src/app/AppRoutingPrivacy.test.tsx")
        self.assertIn(
            "restores an unauthenticated Reports deep link with targetJobId only",
            pending,
        )
        self.assertIn("pendingRoute\")).toBe(\"reports\")", pending)
        self.assertIn(
            "navigate(reports) keeps only targetJobId and drops report state plus raw markers",
            privacy,
        )
        for marker in ("section", "reportId", "status", "roundId", "rawText"):
            self.assertIn(marker, pending)
            self.assertIn(marker, privacy)

    def test_workspace_detail_keeps_only_target_locator_and_stays_read_only(self) -> None:
        privacy = read("frontend/src/app/AppRoutingPrivacy.test.tsx")
        route_url = read("frontend/src/app/routeUrl.test.ts")
        app = read("frontend/src/app/App.test.tsx")
        for marker in (
            "navigate(workspace) with raw markers drops every marker",
            'expect(window.location.search).toBe("?targetJobId=tj-redline")',
            'toBe("/workspace?targetJobId=tj-popstate")',
        ):
            self.assertIn(marker, privacy)
        for marker in (
            "retains targetJobId as the sole workspace detail locator",
            "parses targetJobId as the sole canonical workspace detail locator",
            "autoStartPractice",
            "unknownKey",
            "rawText",
            "token",
        ):
            self.assertIn(marker, route_url)
        for marker in (
            "renders a target-scoped workspace as read-only detail with one getTargetJob",
            'queryByTestId("parse-loading-step-0")',
        ):
            self.assertIn(marker, app)


if __name__ == "__main__":
    unittest.main()
