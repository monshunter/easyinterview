#!/usr/bin/env python3
"""Tests for the fixture-backed mock operation registry."""

from __future__ import annotations

import unittest
from pathlib import Path

from scripts.mock_contract.fixture_registry import (
    FixtureRegistryError,
    build_fixture_registry,
)


REPO_ROOT = Path(__file__).resolve().parents[2]


class FixtureRegistryTest(unittest.TestCase):
    def test_lookup_exposes_operation_metadata_for_p0_domains(self) -> None:
        registry = build_fixture_registry(REPO_ROOT)

        cases = {
            "getMe": ("Auth", 200, "#/components/schemas/UserContext"),
            "listTargetJobs": ("TargetJobs", 200, "#/components/schemas/PaginatedTargetJob"),
            "getPracticeSession": ("PracticeSessions", 200, "#/components/schemas/PracticeSession"),
            "getFeedbackReport": ("Reports", 200, "#/components/schemas/FeedbackReport"),
            "getResume": ("Resumes", 200, "#/components/schemas/Resume"),
        }

        for operation_id, (tag, status, schema_ref) in cases.items():
            with self.subTest(operation_id=operation_id):
                entry = registry.lookup(operation_id)
                self.assertEqual(tag, entry.tag)
                self.assertEqual(operation_id, entry.operation_id)
                self.assertEqual(status, entry.default_status)
                self.assertEqual(schema_ref, entry.response_schema_ref)
                self.assertEqual(
                    REPO_ROOT / "openapi" / "fixtures" / tag / f"{operation_id}.json",
                    entry.fixture_path,
                )
                self.assertIn("default", entry.scenarios)

    def test_unknown_operation_fails_loudly(self) -> None:
        registry = build_fixture_registry(REPO_ROOT)

        with self.assertRaisesRegex(FixtureRegistryError, "unknown operationId: noSuchOperation"):
            registry.lookup("noSuchOperation")


if __name__ == "__main__":
    unittest.main()
