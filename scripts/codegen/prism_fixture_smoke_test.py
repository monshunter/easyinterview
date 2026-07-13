#!/usr/bin/env python3
"""Contract tests for the fixture-backed Prism smoke matrix."""

from __future__ import annotations

import unittest
from unittest.mock import patch

from scripts.codegen import prism_fixture_smoke as smoke


class PrismFixtureSmokeTest(unittest.TestCase):
    def test_report_handoff_matrix_includes_both_reports_and_create_plan(self) -> None:
        operation_ids = {row[0] for row in smoke.SMOKE_MATRIX}

        self.assertTrue(
            {"getFeedbackReport", "listTargetJobReports", "createPracticePlan"}
            <= operation_ids
        )

    @patch("scripts.codegen.prism_fixture_smoke.subprocess.run")
    def test_post_request_sends_default_fixture_body(self, run) -> None:
        run.return_value.stdout = '{"ok":true}\nHTTP=201\n'
        request_body = {"goal": "baseline"}

        status, _ = smoke._curl(
            "POST",
            "http://127.0.0.1:4010/practice/plans",
            "code=201, example=default",
            request_body,
        )

        self.assertEqual(201, status)
        command = run.call_args.args[0]
        self.assertIn("Content-Type: application/json", command)
        self.assertIn('{"goal":"baseline"}', command)


if __name__ == "__main__":
    unittest.main()
