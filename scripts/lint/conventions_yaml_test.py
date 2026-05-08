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
    "provider",
    "capability",
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
    "from_provider",
    "from_model_family",
    "to_provider",
    "to_model_family",
]

REQUIRED_AI_CAPABILITIES = ["chat", "stt", "tts", "realtime", "judge"]

REQUIRED_PROVIDER_REGISTRY_FIELDS = [
    "name",
    "protocol",
    "base_url_env",
    "api_key_env",
    "capabilities",
    "version",
]

REQUIRED_MODEL_PROFILE_FIELDS = [
    "name",
    "capability",
    "status",
    "unsupported_reason",
    "default",
    "provider_ref",
    "model",
    "params",
    "fallback",
    "when",
    "timeout_ms",
    "max_tokens",
    "rate_limit",
    "route",
    "version",
    "privacy_policy",
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
                "name": "TargetJobStatus",
                "sourceSection": "5.1",
                "jsonField": "status",
                "values": ["draft"],
            },
            {
                "name": "TargetJobParseStatus",
                "sourceSection": "5.2",
                "jsonField": "parseStatus",
                "values": ["ready"],
            },
            {
                "name": "PracticeMode",
                "sourceSection": "5.3",
                "jsonField": "mode",
                "values": ["assisted", "strict", "debrief_replay"],
            },
            {
                "name": "PracticeGoal",
                "sourceSection": "5.4",
                "jsonField": "goal",
                "values": ["baseline", "retry_current_round", "next_round", "debrief"],
            },
            {
                "name": "InterviewerRole",
                "sourceSection": "5.5",
                "jsonField": "interviewerRole",
                "values": ["generalist"],
            },
            {
                "name": "SessionStatus",
                "sourceSection": "5.6",
                "jsonField": "sessionStatus",
                "values": ["running"],
            },
            {
                "name": "ReportStatus",
                "sourceSection": "5.7",
                "jsonField": "reportStatus",
                "values": ["ready"],
            },
            {
                "name": "ReadinessTier",
                "sourceSection": "5.8",
                "jsonField": "readinessTier",
                "values": ["needs_practice"],
            },
            {
                "name": "DimensionStatus",
                "sourceSection": "5.9",
                "jsonField": "dimensionStatus",
                "values": ["meets_bar"],
            },
            {
                "name": "Confidence",
                "sourceSection": "5.10",
                "jsonField": "confidence",
                "values": ["medium"],
            },
            {
                "name": "QuestionReviewStatus",
                "sourceSection": "5.11",
                "jsonField": "questionReviewStatus",
                "values": ["open", "queued_for_retry", "resolved"],
            },
            {
                "name": "DebriefStatus",
                "sourceSection": "5.12",
                "jsonField": "debriefStatus",
                "values": ["draft"],
            },
            {
                "name": "PrivacyRequestType",
                "sourceSection": "5.13",
                "jsonField": "privacyRequestType",
                "values": ["export"],
            },
            {
                "name": "PrivacyRequestStatus",
                "sourceSection": "5.13",
                "jsonField": "privacyRequestStatus",
                "values": ["queued"],
            },
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
        "aiVocabulary": {
            "capabilities": REQUIRED_AI_CAPABILITIES,
            "providerRegistryFields": [
                {"name": name} for name in REQUIRED_PROVIDER_REGISTRY_FIELDS
            ],
            "modelProfileFields": [
                {"name": name} for name in REQUIRED_MODEL_PROFILE_FIELDS
            ],
            "fields": [{"name": name} for name in REQUIRED_AI_FIELDS],
        },
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

    def test_rejects_missing_ai_capability(self) -> None:
        for capability in REQUIRED_AI_CAPABILITIES:
            with self.subTest(capability=capability):
                data = valid_data()
                data["aiVocabulary"]["capabilities"] = [
                    value
                    for value in data["aiVocabulary"]["capabilities"]
                    if value != capability
                ]

                errs = self.linter.validate(data)

                self.assertTrue(any(capability in err for err in errs), errs)

    def test_rejects_missing_ai_provider_registry_field(self) -> None:
        data = valid_data()
        data["aiVocabulary"]["providerRegistryFields"] = [
            field
            for field in data["aiVocabulary"]["providerRegistryFields"]
            if field["name"] != "api_key_env"
        ]

        errs = self.linter.validate(data)

        self.assertTrue(any("api_key_env" in err for err in errs), errs)

    def test_rejects_missing_ai_model_profile_field(self) -> None:
        data = valid_data()
        data["aiVocabulary"]["modelProfileFields"] = [
            field
            for field in data["aiVocabulary"]["modelProfileFields"]
            if field["name"] != "provider_ref"
        ]

        errs = self.linter.validate(data)

        self.assertTrue(any("provider_ref" in err for err in errs), errs)

    def test_rejects_non_snake_case_ai_vocabulary_field(self) -> None:
        data = copy.deepcopy(valid_data())
        data["aiVocabulary"]["fields"][0]["name"] = "modelProfileName"

        errs = self.linter.validate(data)

        self.assertTrue(any("modelProfileName" in err and "lower_snake_case" in err for err in errs), errs)

    def test_rejects_removed_mistake_status(self) -> None:
        data = copy.deepcopy(valid_data())
        data["enums"].append(
            {
                "name": "MistakeStatus",
                "sourceSection": "5.11",
                "jsonField": "mistakeStatus",
                "values": ["open"],
            }
        )

        errs = self.linter.validate(data)

        self.assertTrue(
            any("MistakeStatus" in err and "removed by product-scope v1.2" in err for err in errs),
            errs,
        )

    def test_rejects_legacy_practice_mode_values(self) -> None:
        data = copy.deepcopy(valid_data())
        for enum in data["enums"]:
            if enum["name"] == "PracticeMode":
                enum["values"] = ["warmup", "core_interview", "single_drill"]
                break

        errs = self.linter.validate(data)

        self.assertTrue(
            any("PracticeMode" in err and "product-scope v1.2 values" in err for err in errs),
            errs,
        )


if __name__ == "__main__":
    unittest.main()
