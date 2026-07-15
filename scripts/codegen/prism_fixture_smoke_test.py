#!/usr/bin/env python3
"""Contract tests for the fixture-backed Prism smoke matrix."""

from __future__ import annotations

import unittest
from unittest.mock import patch

from scripts.codegen import prism_fixture_smoke as smoke


class PrismFixtureSmokeTest(unittest.TestCase):
    def test_resume_matrix_covers_summary_list_and_full_detail_defaults(self) -> None:
        rows = {row[0]: row for row in smoke.SMOKE_MATRIX}

        self.assertEqual(
            (
                "listResumes",
                "GET",
                "/resumes",
                200,
                "openapi/fixtures/Resumes/listResumes.json",
            ),
            rows["listResumes"],
        )
        self.assertEqual(
            (
                "getResume",
                "GET",
                "/resumes/01918fa0-0000-7000-8000-000000001000",
                200,
                "openapi/fixtures/Resumes/getResume.json",
            ),
            rows["getResume"],
        )

    def test_report_handoff_matrix_includes_the_report_owned_conversation(self) -> None:
        rows = {row[0]: row for row in smoke.SMOKE_MATRIX}

        self.assertTrue(
            {"getFeedbackReport", "getReportConversation", "listTargetJobReports", "createPracticePlan"}
            <= set(rows)
        )
        self.assertEqual(
            (
                "getReportConversation",
                "GET",
                "/reports/01918fa0-0070-7000-8000-000000000070/conversation",
                200,
                "openapi/fixtures/Reports/getReportConversation.json",
            ),
            rows["getReportConversation"],
        )

    def test_target_job_report_pointer_removal_matrix_covers_all_affected_defaults(self) -> None:
        operation_ids = {row[0] for row in smoke.SMOKE_MATRIX}

        self.assertTrue(
            {
                "listTargetJobs",
                "getTargetJob",
                "importTargetJob",
                "updateTargetJob",
                "archiveTargetJob",
                "listTargetJobReports",
            }
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

    @patch("scripts.codegen.prism_fixture_smoke.subprocess.run")
    def test_patch_request_uses_patch_method_and_default_fixture_body(self, run) -> None:
        run.return_value.stdout = '{"ok":true}\nHTTP=200\n'

        status, _ = smoke._curl(
            "PATCH",
            "http://127.0.0.1:4010/targets/target-1",
            "example=default",
            {"status": "interviewing"},
        )

        self.assertEqual(200, status)
        command = run.call_args.args[0]
        self.assertIn("PATCH", command)
        self.assertIn('{"status":"interviewing"}', command)


if __name__ == "__main__":
    unittest.main()
