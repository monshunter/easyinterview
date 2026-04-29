#!/usr/bin/env python3
"""Contract tests for scripts/lint/conventions_yaml.py."""

from __future__ import annotations

import copy
import importlib.util
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("conventions_yaml.py")


def load_linter():
    spec = importlib.util.spec_from_file_location("conventions_yaml_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


REQUIRED_AI_FIELDS = [
    "model_profile_name",
    "model_profile_version",
    "model_family",
    "model_id",
    "fallback_chain",
    "route",
    "validation_status",
    "output_schema_version",
    "prompt_version",
    "rubric_version",
    "language",
    "feature_flag",
    "data_source_version",
]


def valid_data() -> dict:
    return {
        "version": 1.0,
        "schemaVersion": 1,
        "sampleUuidV7": "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
        "uuidV7Regex": "^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
        "tmpIdPrefix": "tmp_",
        "pagination": {"defaultPageSize": 20, "maxPageSize": 100},
        "idempotency": {"ttlSeconds": 86400},
        "errors": [
            {
                "code": "AUTH_UNAUTHORIZED",
                "message": "authentication required or invalid",
                "retryable": False,
            },
            {
                "code": "TARGET_IMPORT_FAILED",
                "message": "failed to import target job",
                "retryable": True,
            },
            {
                "code": "PRACTICE_SESSION_CONFLICT",
                "message": "practice session conflict",
                "retryable": False,
            },
            {
                "code": "REPORT_NOT_READY",
                "message": "report is not ready",
                "retryable": True,
            },
            {
                "code": "VALIDATION_FAILED",
                "message": "validation failed",
                "retryable": False,
            },
            {
                "code": "RATE_LIMITED",
                "message": "rate limited",
                "retryable": True,
            },
        ],
        "jobStatuses": ["queued", "running", "succeeded", "failed", "cancelled", "dead"],
        "enums": [
            {
                "name": f"Enum{i}",
                "sourceSection": f"5.{i}",
                "jsonField": f"field{i}",
                "values": ["ready"],
            }
            for i in range(1, 14)
        ],
        "structures": {
            "PageInfo": {
                "fields": [
                    {"name": "nextCursor", "type": "string", "nullable": True},
                    {"name": "pageSize", "type": "int"},
                    {"name": "hasMore", "type": "bool"},
                ]
            },
            "ApiError": {
                "fields": [
                    {"name": "code", "type": "string"},
                    {"name": "message", "type": "string"},
                    {"name": "requestId", "type": "string"},
                    {"name": "retryable", "type": "bool"},
                ]
            },
        },
        "aiVocabulary": {"fields": [{"name": name} for name in REQUIRED_AI_FIELDS]},
    }


class ConventionsYAMLTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()

    def test_ai_vocabulary_complete_shape_passes(self) -> None:
        self.assertEqual([], self.linter.validate(valid_data()))

    def test_requires_ai_vocabulary_namespace(self) -> None:
        data = valid_data()
        data.pop("aiVocabulary")

        errs = self.linter.validate(data)

        self.assertTrue(any("missing top-level keys" in err and "aiVocabulary" in err for err in errs), errs)

    def test_rejects_missing_ai_vocabulary_field(self) -> None:
        data = valid_data()
        data["aiVocabulary"]["fields"] = [
            field
            for field in data["aiVocabulary"]["fields"]
            if field["name"] != "validation_status"
        ]

        errs = self.linter.validate(data)

        self.assertTrue(any("validation_status" in err for err in errs), errs)

    def test_rejects_non_snake_case_ai_vocabulary_field(self) -> None:
        data = copy.deepcopy(valid_data())
        data["aiVocabulary"]["fields"][0]["name"] = "modelProfileName"

        errs = self.linter.validate(data)

        self.assertTrue(any("modelProfileName" in err and "lower_snake_case" in err for err in errs), errs)


if __name__ == "__main__":
    unittest.main()
