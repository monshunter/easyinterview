#!/usr/bin/env python3
"""Contract tests for scripts/lint/validate_fixtures.py + openapi/fixtures/.

Phase 1.1: structural shape only (skeleton). Schema and content invariants are
covered as they are added in Phase 1.2 / 1.3.
"""

from __future__ import annotations

import importlib.util
import json
import unittest
from pathlib import Path

import yaml

import scripts.lint.openapi_inventory as inventory

ROOT = Path(__file__).resolve().parents[2]
FIXTURES_ROOT = ROOT / "openapi" / "fixtures"
OPENAPI_PATH = ROOT / "openapi" / "openapi.yaml"
SCRIPT = Path(__file__).with_name("validate_fixtures.py")


def _load_validator():
    spec = importlib.util.spec_from_file_location("validate_fixtures_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _load_openapi() -> dict:
    return yaml.safe_load(OPENAPI_PATH.read_text(encoding="utf-8"))


def _preferred_default_status(operation: dict) -> int:
    for code in ("200", "201", "202", "204", "422", "501"):
        if code in (operation.get("responses") or {}):
            return int(code)
    raise AssertionError(f"operation has no preferred default status: {operation}")


def _expected_operations() -> list[tuple[str, str, int, bool]]:
    spec = _load_openapi()
    rows: list[tuple[str, str, int, bool]] = []
    for tag, method, path, opid in inventory.EXPECTED_OPERATIONS:
        operation = spec["paths"][path][method]
        rows.append((tag, opid, _preferred_default_status(operation), "requestBody" in operation))
    return rows


# (tag, operationId, expected default status, has_request_body)
EXPECTED_OPERATIONS = _expected_operations()
IK_REQUIRED_OPERATION_IDS = {
    opid
    for _tag, method, path, opid in inventory.EXPECTED_OPERATIONS
    if (method, path) in inventory.IK_REQUIRED
}


class FixtureSkeletonTest(unittest.TestCase):
    """Phase 1.1 structural contract."""

    def test_thirty_seven_operations_expected(self) -> None:
        self.assertEqual(len(EXPECTED_OPERATIONS), 37)

    def test_ten_unique_tags(self) -> None:
        tags = {tag for tag, *_ in EXPECTED_OPERATIONS}
        self.assertEqual(len(tags), 10)

    def test_each_fixture_file_exists(self) -> None:
        missing = []
        for tag, opid, _status, _has_req in EXPECTED_OPERATIONS:
            path = FIXTURES_ROOT / tag / f"{opid}.json"
            if not path.is_file():
                missing.append(str(path.relative_to(ROOT)))
        self.assertEqual([], missing, f"missing fixture files: {missing}")

    def test_each_fixture_has_correct_operationId(self) -> None:
        for tag, opid, _status, _has_req in EXPECTED_OPERATIONS:
            path = FIXTURES_ROOT / tag / f"{opid}.json"
            with self.subTest(operationId=opid):
                with path.open("r", encoding="utf-8") as f:
                    data = json.load(f)
                self.assertEqual(
                    data.get("operationId"),
                    opid,
                    f"{path}: operationId field must match filename",
                )

    def test_scenarios_default_is_first_key(self) -> None:
        for tag, opid, _status, _has_req in EXPECTED_OPERATIONS:
            path = FIXTURES_ROOT / tag / f"{opid}.json"
            with self.subTest(operationId=opid):
                with path.open("r", encoding="utf-8") as f:
                    data = json.load(f)
                scenarios = data.get("scenarios")
                self.assertIsInstance(scenarios, dict, f"{path}: scenarios must be object")
                first = next(iter(scenarios), None)
                self.assertEqual(first, "default", f"{path}: first scenario must be 'default'")

    def test_default_response_status_matches_spec(self) -> None:
        for tag, opid, status, _has_req in EXPECTED_OPERATIONS:
            path = FIXTURES_ROOT / tag / f"{opid}.json"
            with self.subTest(operationId=opid):
                with path.open("r", encoding="utf-8") as f:
                    data = json.load(f)
                resp = data["scenarios"]["default"]["response"]
                self.assertEqual(
                    resp.get("status"),
                    status,
                    f"{path}: default.response.status must equal {status}",
                )

    def test_request_field_present_when_operation_has_requestBody(self) -> None:
        for tag, opid, _status, has_req in EXPECTED_OPERATIONS:
            path = FIXTURES_ROOT / tag / f"{opid}.json"
            with self.subTest(operationId=opid):
                with path.open("r", encoding="utf-8") as f:
                    data = json.load(f)
                default = data["scenarios"]["default"]
                if has_req:
                    self.assertIn("request", default, f"{path}: request must be present")
                    self.assertIn("body", default["request"], f"{path}: request.body must be present")
                elif opid in IK_REQUIRED_OPERATION_IDS and "request" in default:
                    self.assertIn("headers", default["request"], f"{path}: request.headers must be present")
                    self.assertIn("Idempotency-Key", default["request"]["headers"])
                    self.assertNotIn("body", default["request"], f"{path}: operation has no request body")
                else:
                    self.assertNotIn(
                        "request",
                        default,
                        f"{path}: request must be omitted when operation has no requestBody",
                    )


class FixtureValidatorWalkerTest(unittest.TestCase):
    """Validator helper exposes a structural walk over openapi/fixtures/."""

    def test_walker_returns_55_entries(self) -> None:
        validator = _load_validator()
        entries = validator.walk_fixtures(FIXTURES_ROOT)
        self.assertEqual(
            sorted(opid for _tag, opid, _path, _data in entries),
            sorted(opid for _tag, opid, *_ in EXPECTED_OPERATIONS),
        )


# Phase 1.2: content invariants ------------------------------------------------

import re

UUID_SHAPE_RE = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
)
UUID_V7_RE = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)
ISO_DATETIME_RE = re.compile(
    r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$"
)

# AI-generated schemas listed in spec §4.6 — these must carry `provenance`.
PROVENANCE_OPERATIONS = {
    # operationId -> json-pointer-style path inside scenarios.default.response.body
    "getTargetJob": [
        "summary.provenance",
        "fitSummary.provenance",
    ],
    "getFeedbackReport": ["provenance"],
    "getResumeTailorRun": ["provenance"],
    "getResume": ["structuredProfile.provenance"],
    "listResumes": ["items[*].structuredProfile.provenance"],
    "updateResume": ["structuredProfile.provenance"],
    "duplicateResume": ["structuredProfile.provenance"],
}

LIST_OPERATIONS = [
    "listTargetJobs",
    "listResumes",
]

REPORT_OVERVIEW_SCENARIO_ORDER = (
    "default",
    "current-ready",
    "prior-ready-newer-queued",
    "prior-ready-newer-generating",
    "prior-ready-newer-failed",
    "latest-ready-is-current",
    "ready-tie-break",
    "cross-user-not-found",
    "target-not-found",
    "invalid-frozen-context",
    "missing-frozen-context",
)

# *WithJob async operations and the JobType they must emit.
WITH_JOB_OPERATIONS = {
    "startAuthEmailChallenge": None,  # 202 but no Job envelope (auth challenge)
    "registerResume": "resume_parse",
    "importTargetJob": "target_import",
    "completePracticeSession": "report_generate",
    "requestResumeTailor": "resume_tailor",
    "deleteMe": "privacy_delete",
    "requestPrivacyDelete": "privacy_delete",
}

REQUIRED_PRACTICE_SESSION_SCENARIOS = {
    "completePracticeSession": {
        "default",
        "replay",
        "mismatch",
        "session-already-completed",
        "cross-user-not-found",
    },
    "createPracticeVoiceTurn": {"default"},
    "getPracticeSession": {
        "default",
        "reply-pending",
        "reply-retryable-failed",
        "reply-terminal-failed",
        "reply-complete",
    },
    "sendPracticeMessage": {
        "default",
        "validation-empty-text",
        "auth-unauthorized",
        "session-not-found",
        "reply-pending-conflict",
        "client-message-mismatch",
        "ai-timeout-retryable",
        "retry-success-same-client-message",
    },
}

PROVENANCE_REQUIRED_FIELDS = [
    "promptVersion",
    "rubricVersion",
    "modelId",
    "language",
    "featureFlag",
    "dataSourceVersion",
]

# Privacy allowlist (spec §4.7).
ALLOWED_EMAIL_DOMAINS = {"example.com", "example.org", "example.net"}
ALLOWED_PHONE_PREFIX = "+1-555-01"  # +1-555-0100..0199
EMAIL_RE = re.compile(r"\b[A-Za-z0-9._%+-]+@([A-Za-z0-9.-]+\.[A-Za-z]{2,})\b")
PHONE_RE = re.compile(r"\+\d[\d\-\s()]{7,}\d")
# Real employer-style brands that must never leak into fixtures.
COMPANY_BLACKLIST = {
    "alibaba", "tencent", "bytedance", "baidu", "meituan", "didi", "huawei",
    "字节", "腾讯", "阿里巴巴", "百度", "美团", "滴滴", "华为", "星环",
}
COMPANY_BLACKLIST_RE = re.compile(
    r"(?:^|[^A-Za-z0-9_])(" + "|".join(re.escape(b) for b in COMPANY_BLACKLIST) +
    r")(?:[^A-Za-z0-9_]|$)",
    re.IGNORECASE,
)
TEMP_ID_RE = re.compile(r"\btmp_[A-Za-z0-9_-]+\b")
PROVIDER_NEUTRAL_MODEL_ID_RE = re.compile(r"^(?:model-profile|fixture-model):[a-z][a-z0-9_.-]*$")
VENDOR_MODEL_TOKEN_RE = re.compile(
    r"(?:openrouter|anthropic|claude|openai|gpt-|mistral|gemini|cohere)",
    re.IGNORECASE,
)


def _walk_strings(data, prefix=""):
    """Yield (path, str) tuples for every string value reachable from data."""
    if isinstance(data, str):
        yield prefix, data
    elif isinstance(data, dict):
        for k, v in data.items():
            yield from _walk_strings(v, f"{prefix}.{k}" if prefix else k)
    elif isinstance(data, list):
        for i, v in enumerate(data):
            yield from _walk_strings(v, f"{prefix}[{i}]")


def _resolve(data, path):
    """Walk a `summary.provenance` / `items[*].provenance` style path. `[*]`
    expands a list and yields each element's resolved value."""
    parts = re.findall(r"[^.\[\]*]+|\[\*\]", path)
    cursor = [data]
    for part in parts:
        next_cursor = []
        for c in cursor:
            if part == "[*]":
                if isinstance(c, list):
                    next_cursor.extend(c)
            else:
                if isinstance(c, dict) and part in c:
                    next_cursor.append(c[part])
        cursor = next_cursor
    return cursor


def _load_fixture(opid: str, tag: str) -> dict:
    path = FIXTURES_ROOT / tag / f"{opid}.json"
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


class FixtureContentTest(unittest.TestCase):
    """Phase 1.2 content invariants."""

    def test_privacy_export_returns_501_with_correct_error_code(self) -> None:
        data = _load_fixture("requestPrivacyExport", "Privacy")
        resp = data["scenarios"]["default"]["response"]
        self.assertEqual(resp["status"], 501)
        self.assertEqual(
            resp["body"]["error"]["code"],
            "PRIVACY_EXPORT_NOT_AVAILABLE",
            "spec D-12 requires this exact error code on P0",
        )

    def test_resume_export_returns_501_with_correct_error_code(self) -> None:
        data = _load_fixture("exportResume", "Resumes")
        resp = data["scenarios"]["default"]["response"]
        self.assertEqual(resp["status"], 501)
        self.assertEqual(
            resp["body"]["error"]["code"],
            "RESUME_EXPORT_NOT_AVAILABLE",
            "spec D-18 requires this exact error code on P0",
        )

    def test_privacy_delete_returns_202_with_job(self) -> None:
        data = _load_fixture("requestPrivacyDelete", "Privacy")
        body = data["scenarios"]["default"]["response"]["body"]
        self.assertEqual(data["scenarios"]["default"]["response"]["status"], 202)
        self.assertIn("privacyRequestId", body)
        self.assertIn("job", body)
        self.assertEqual(body["job"]["jobType"], "privacy_delete")

    def test_delete_me_returns_idempotent_privacy_delete_job(self) -> None:
        data = _load_fixture("deleteMe", "Auth")
        default = data["scenarios"]["default"]
        self.assertEqual(default["request"]["headers"]["Idempotency-Key"], "idem-delete-me-2026-04-29")
        body = default["response"]["body"]
        self.assertEqual(default["response"]["status"], 202)
        self.assertIn("privacyRequestId", body)
        self.assertIn("job", body)
        self.assertEqual(body["job"]["jobType"], "privacy_delete")
        self.assertEqual(body["job"]["resourceType"], "privacy_request")

    def test_list_endpoints_have_pageInfo(self) -> None:
        for opid in LIST_OPERATIONS:
            tag = next(t for t, o, *_ in EXPECTED_OPERATIONS if o == opid)
            with self.subTest(operationId=opid):
                body = _load_fixture(opid, tag)["scenarios"]["default"]["response"]["body"]
                self.assertIn("items", body)
                self.assertIn("pageInfo", body)
                page = body["pageInfo"]
                self.assertIsNone(page.get("nextCursor"))
                self.assertEqual(page.get("hasMore"), False)
                self.assertIsInstance(page.get("pageSize"), int)
                self.assertGreaterEqual(len(body["items"]), 1)
                self.assertLessEqual(len(body["items"]), 3)

    def test_ai_schemas_carry_complete_provenance(self) -> None:
        for opid, paths in PROVENANCE_OPERATIONS.items():
            tag = next(t for t, o, *_ in EXPECTED_OPERATIONS if o == opid)
            body = _load_fixture(opid, tag)["scenarios"]["default"]["response"]["body"]
            for path in paths:
                with self.subTest(operationId=opid, path=path):
                    found = _resolve(body, path)
                    self.assertTrue(
                        found,
                        f"{opid}: provenance path {path} did not resolve to any value",
                    )
                    for prov in found:
                        self.assertIsInstance(prov, dict, f"{opid}.{path} must be object")
                        for field in PROVENANCE_REQUIRED_FIELDS:
                            value = prov.get(field)
                            self.assertIsInstance(
                                value, str,
                                f"{opid}.{path}.{field} must be a string",
                            )
                        self.assertTrue(
                            value.strip(),
                            f"{opid}.{path}.{field} must be non-empty",
                        )
                        if field == "modelId":
                            self.assertRegex(
                                value,
                                PROVIDER_NEUTRAL_MODEL_ID_RE,
                                f"{opid}.{path}.{field} must be a provider-neutral model id",
                            )
                            self.assertNotRegex(
                                value,
                                VENDOR_MODEL_TOKEN_RE,
                                f"{opid}.{path}.{field} must not hardcode vendor/model tokens",
                            )

    def test_validator_rejects_vendor_specific_model_id(self) -> None:
        validator = _load_validator()
        errors = []
        validator.check_provenance(
            "getFeedbackReport",
            {
                "response": {
                    "body": {
                        "provenance": {
                            "promptVersion": "feedback_report.v3",
                            "rubricVersion": "feedback_report.rubric.v2",
                            "modelId": "openrouter:anthropic/claude-sonnet-4.6",
                            "language": "zh-CN",
                            "featureFlag": "none",
                            "dataSourceVersion": "practice_session.v9",
                        }
                    }
                }
            },
            errors,
        )

        self.assertTrue(any("modelId" in err and "provider-neutral" in err for err in errors), errors)
        self.assertTrue(any("modelId" in err and "vendor/model tokens" in err for err in errors), errors)

    def test_validator_accepts_provider_neutral_fixture_model_id(self) -> None:
        validator = _load_validator()
        errors = []
        validator.check_provenance(
            "getTargetJob",
            {
                "response": {
                    "body": {
                        "summary": {
                            "provenance": {
                                "promptVersion": "v0.1.0",
                                "rubricVersion": "v0.1.0",
                                "modelId": "fixture-model:target-import-parse",
                                "language": "zh-CN",
                                "featureFlag": "none",
                                "dataSourceVersion": "registry.v1",
                            }
                        },
                        "fitSummary": {
                            "provenance": {
                                "promptVersion": "v0.1.0",
                                "rubricVersion": "not_applicable",
                                "modelId": "fixture-model:target-import-parse",
                                "language": "zh-CN",
                                "featureFlag": "none",
                                "dataSourceVersion": "registry.v1",
                            }
                        },
                    }
                }
            },
            errors,
        )

        self.assertFalse([err for err in errors if "modelId" in err], errors)

    def test_non_ready_feedback_report_requires_null_provenance(self) -> None:
        validator = _load_validator()
        errors: list[str] = []
        validator.check_provenance(
            "getFeedbackReport",
            {
                "response": {
                    "body": {
                        "status": "generating",
                        "provenance": None,
                    }
                }
            },
            errors,
        )

        self.assertEqual([], errors)

    def test_with_job_operations_carry_correct_jobType(self) -> None:
        for opid, expected_job_type in WITH_JOB_OPERATIONS.items():
            if expected_job_type is None:
                continue
            tag = next(t for t, o, *_ in EXPECTED_OPERATIONS if o == opid)
            body = _load_fixture(opid, tag)["scenarios"]["default"]["response"]["body"]
            with self.subTest(operationId=opid):
                self.assertIn("job", body, f"{opid}: response must include job envelope")
                self.assertEqual(body["job"]["jobType"], expected_job_type)
                self.assertIn(body["job"]["status"], {"queued", "running"})

    def test_target_job_import_fixture_uses_exact_paste_only_matrix(self) -> None:
        scenarios = _load_fixture("importTargetJob", "TargetJobs")["scenarios"]

        self.assertEqual(
            ["default", "paste-primary", "validation-blank-raw-text"],
            list(scenarios),
        )
        for scenario_name in ("default", "paste-primary"):
            with self.subTest(scenario=scenario_name):
                body = scenarios[scenario_name]["request"]["body"]
                self.assertEqual({"rawText", "targetLanguage", "resumeId"}, set(body))
                self.assertIsInstance(body["rawText"], str)
                self.assertTrue(body["rawText"].strip())

        negative = scenarios["validation-blank-raw-text"]
        self.assertTrue(negative["request"]["body"]["rawText"])
        self.assertFalse(negative["request"]["body"]["rawText"].strip())
        self.assertEqual(422, negative["response"]["status"])
        error = negative["response"]["body"]["error"]
        self.assertEqual("VALIDATION_FAILED", error["code"])
        self.assertIs(False, error["retryable"])
        self.assertEqual("rawText", error["details"]["field"])

    def test_target_job_import_schema_rejects_retired_and_blank_requests(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schema = spec["components"]["schemas"]["ImportTargetJobRequest"]
        resume_id = "01918fa0-0000-7000-8000-000000001000"
        valid = {
            "rawText": "Senior frontend engineer with React and SSR experience.",
            "targetLanguage": "zh-CN",
            "resumeId": resume_id,
        }

        valid_errors: list[str] = []
        validator.schema_validate(
            valid, schema, root=spec, path="request", errors=valid_errors
        )
        self.assertEqual([], valid_errors)

        invalid = [
            {**valid, "rawText": ""},
            {**valid, "rawText": "   "},
            {**valid, "rawText": "\t"},
            {**valid, "rawText": "\n"},
            {
                "source": {"type": "url", "url": "https://acme.example/jobs/1"},
                "targetLanguage": "zh-CN",
                "resumeId": resume_id,
            },
            {**valid, "source": {"type": "manual_text", "rawText": valid["rawText"]}},
            {**valid, "source": {"type": "file", "fileObjectId": resume_id}},
            {
                **valid,
                "source": {
                    "type": "manual_form",
                    "title": "Senior frontend engineer",
                    "companyName": "Acme",
                    "rawDescription": valid["rawText"],
                },
            },
            {**valid, "fileObjectId": resume_id},
            {**valid, "titleHint": "Senior frontend engineer"},
            {**valid, "companyNameHint": "Acme"},
            {**valid, "unexpected": True},
        ]
        for index, body in enumerate(invalid):
            with self.subTest(case=index):
                errors: list[str] = []
                validator.schema_validate(
                    body, schema, root=spec, path="request", errors=errors
                )
                self.assertTrue(errors, body)

    def test_only_canonical_blank_raw_text_scenario_may_fail_request_schema(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        op = validator.build_operation_index(spec)["importTargetJob"]["operation"]
        scenario = {
            "request": {
                "headers": {},
                "body": {
                    "rawText": "   ",
                    "targetLanguage": "zh-CN",
                    "resumeId": "01918fa0-0000-7000-8000-000000001000",
                },
            },
            "response": {
                "status": 422,
                "headers": {"X-Request-ID": "req_2026-07-13-blank-raw-text"},
                "body": {
                    "error": {
                        "code": "VALIDATION_FAILED",
                        "message": "rawText must contain non-whitespace content",
                        "requestId": "req_2026-07-13-blank-raw-text",
                        "retryable": False,
                        "details": {"field": "rawText"},
                    }
                },
            },
        }

        errors: list[str] = []
        validator.check_schema(
            "importTargetJob",
            op,
            "validation-blank-raw-text",
            scenario,
            spec,
            errors,
        )
        self.assertEqual([], errors)

        missing_resume = json.loads(json.dumps(scenario))
        del missing_resume["request"]["body"]["resumeId"]
        errors = []
        validator.check_schema(
            "importTargetJob",
            op,
            "validation-blank-raw-text",
            missing_resume,
            spec,
            errors,
        )
        self.assertTrue(errors, "the canonical exception must require exactly /rawText")

        errors = []
        validator.check_schema(
            "importTargetJob",
            op,
            "another-blank-request",
            scenario,
            spec,
            errors,
        )
        self.assertTrue(errors, "scenario-name wildcards must not bypass request validation")

    def test_target_job_read_fixture_semantics_reject_removed_source_fields(self) -> None:
        validator = _load_validator()
        for operation_id in (
            "listTargetJobs",
            "getTargetJob",
            "updateTargetJob",
            "archiveTargetJob",
        ):
            scenarios = _load_fixture(operation_id, "TargetJobs")["scenarios"]
            errors: list[str] = []
            validator.check_target_job_paste_only_semantics(
                operation_id, scenarios, errors
            )
            with self.subTest(operationId=operation_id):
                self.assertEqual([], errors)

        clean = {
            "default": {
                "response": {"status": 200, "body": {"id": "target"}}
            }
        }
        for removed in ("sourceType", "sourceUrl"):
            mutated = json.loads(json.dumps(clean))
            mutated["default"]["response"]["body"][removed] = (
                "manual_text" if removed == "sourceType" else None
            )
            errors = []
            validator.check_target_job_paste_only_semantics(
                "getTargetJob", mutated, errors
            )
            with self.subTest(removed=removed):
                self.assertTrue(any(removed in error for error in errors), errors)

    def test_target_job_read_fixtures_forbid_latest_report_pointer(self) -> None:
        validator = _load_validator()
        for operation_id in (
            "listTargetJobs",
            "getTargetJob",
            "updateTargetJob",
            "archiveTargetJob",
        ):
            scenarios = _load_fixture(operation_id, "TargetJobs")["scenarios"]
            errors: list[str] = []
            validator.check_target_job_paste_only_semantics(
                operation_id, scenarios, errors
            )
            encoded = json.dumps(scenarios, ensure_ascii=False)
            with self.subTest(operationId=operation_id):
                self.assertNotIn("latestReportId", encoded)
                self.assertEqual([], errors)

        mutated = {
            "default": {
                "response": {
                    "status": 200,
                    "body": {"id": "target", "latestReportId": None},
                }
            }
        }
        errors = []
        validator.check_target_job_paste_only_semantics(
            "getTargetJob", mutated, errors
        )
        self.assertTrue(
            any("latestReportId" in error and "forbidden" in error for error in errors),
            errors,
        )

    def test_target_job_reports_overview_fixture_is_closed_and_canonical(self) -> None:
        validator = _load_validator()
        scenarios = _load_fixture("listTargetJobReports", "Reports")["scenarios"]
        errors: list[str] = []
        validator.check_target_job_reports_overview_semantics(scenarios, errors)

        self.assertEqual(REPORT_OVERVIEW_SCENARIO_ORDER, tuple(scenarios))
        self.assertEqual([], errors)
        for scenario_name, scenario in scenarios.items():
            response = scenario["response"]
            if response["status"] != 200:
                continue
            body = response["body"]
            with self.subTest(scenario=scenario_name):
                self.assertEqual({"targetJobId", "rounds"}, set(body))
                self.assertGreaterEqual(len(body["rounds"]), 2)
                self.assertLessEqual(len(body["rounds"]), 5)
                for item in body["rounds"]:
                    self.assertEqual(
                        {"round", "currentReport", "latestAttempt"}, set(item)
                    )
                    self.assertEqual(
                        {"roundId", "roundSequence"}, set(item["round"])
                    )

        forbidden = {
            "items",
            "pageInfo",
            "cursor",
            "sessionId",
            "sourcePlanId",
            "provenance",
            "modelId",
            "rubricVersion",
            "summary",
            "dimensionAssessments",
            "practicePlanId",
            "latestReportId",
        }
        for key_path, key in validator._walk_keys(scenarios):
            with self.subTest(path=key_path):
                self.assertNotIn(key, forbidden)

    def test_target_job_reports_overview_semantics_reject_invalid_variants(self) -> None:
        validator = _load_validator()
        scenarios = _load_fixture("listTargetJobReports", "Reports")["scenarios"]

        mutations: list[tuple[str, dict]] = []
        flat = json.loads(json.dumps(scenarios))
        flat["default"]["response"]["body"] = {"items": [], "pageInfo": {}}
        mutations.append(("flat", flat))

        missing_rounds = json.loads(json.dumps(scenarios))
        del missing_rounds["default"]["response"]["body"]["rounds"]
        mutations.append(("missing-rounds", missing_rounds))

        missing_nullable = json.loads(json.dumps(scenarios))
        del missing_nullable["default"]["response"]["body"]["rounds"][0][
            "latestAttempt"
        ]
        mutations.append(("missing-nullable", missing_nullable))

        invalid_error = json.loads(json.dumps(scenarios))
        invalid_error["prior-ready-newer-failed"]["response"]["body"]["rounds"][
            0
        ]["latestAttempt"]["errorCode"] = None
        mutations.append(("failed-without-error", invalid_error))

        error_on_generating = json.loads(json.dumps(scenarios))
        error_on_generating["prior-ready-newer-generating"]["response"]["body"][
            "rounds"
        ][0]["latestAttempt"]["errorCode"] = "AI_PROVIDER_TIMEOUT"
        mutations.append(("non-failed-with-error", error_on_generating))

        invalid_status = json.loads(json.dumps(scenarios))
        invalid_status["prior-ready-newer-generating"]["response"]["body"][
            "rounds"
        ][0]["latestAttempt"]["status"] = "retryable_failed"
        mutations.append(("invalid-status", invalid_status))

        for name, mutation in mutations:
            errors: list[str] = []
            validator.check_target_job_reports_overview_semantics(mutation, errors)
            with self.subTest(mutation=name):
                self.assertTrue(errors)

    def test_upload_presign_keeps_resume_and_privacy_only(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schema = spec["components"]["schemas"]["UploadPresignRequest"]
        scenarios = _load_fixture("createUploadPresign", "Uploads")["scenarios"]

        self.assertEqual(["default", "privacy-export"], list(scenarios))
        self.assertEqual(
            {"resume", "privacy_export"},
            {
                scenario["request"]["body"]["purpose"]
                for scenario in scenarios.values()
            },
        )
        for purpose in ("resume", "privacy_export"):
            body = json.loads(json.dumps(scenarios["default"]["request"]["body"]))
            body["purpose"] = purpose
            errors: list[str] = []
            validator.schema_validate(
                body, schema, root=spec, path="request", errors=errors
            )
            self.assertEqual([], errors, purpose)

        rejected = json.loads(json.dumps(scenarios["default"]["request"]["body"]))
        rejected["purpose"] = "target_job_attachment"
        errors = []
        validator.schema_validate(
            rejected, schema, root=spec, path="request", errors=errors
        )
        self.assertTrue(any("target_job_attachment" in error for error in errors), errors)

    def test_practice_session_fixtures_declare_required_named_scenarios(self) -> None:
        for opid, expected in REQUIRED_PRACTICE_SESSION_SCENARIOS.items():
            scenarios = _load_fixture(opid, "PracticeSessions")["scenarios"]
            with self.subTest(operationId=opid):
                self.assertTrue(
                    expected.issubset(scenarios),
                    f"{opid}: missing required scenarios {sorted(expected - set(scenarios))}",
                )

    def test_practice_reply_recovery_fixtures_are_role_typed_and_replay_safe(self) -> None:
        validator = _load_validator()
        get_scenarios = _load_fixture("getPracticeSession", "PracticeSessions")["scenarios"]
        send_scenarios = _load_fixture("sendPracticeMessage", "PracticeSessions")["scenarios"]

        errors: list[str] = []
        validator.check_practice_reply_recovery_semantics(
            "getPracticeSession", get_scenarios, errors
        )
        validator.check_practice_reply_recovery_semantics(
            "sendPracticeMessage", send_scenarios, errors
        )
        self.assertEqual([], errors)

        invalid_get_mutations = []
        missing_client_id = json.loads(json.dumps(get_scenarios))
        del missing_client_id["reply-pending"]["response"]["body"]["messages"][-1]["clientMessageId"]
        invalid_get_mutations.append(missing_client_id)

        assistant_recovery = json.loads(json.dumps(get_scenarios))
        assistant = assistant_recovery["reply-complete"]["response"]["body"]["messages"][-1]
        assistant["clientMessageId"] = "01918fa0-0000-7000-8000-000000007010"
        assistant["replyStatus"] = "complete"
        invalid_get_mutations.append(assistant_recovery)

        invalid_status = json.loads(json.dumps(get_scenarios))
        invalid_status["reply-retryable-failed"]["response"]["body"]["messages"][-1]["replyStatus"] = "unknown"
        invalid_get_mutations.append(invalid_status)

        duplicate_retry = json.loads(json.dumps(get_scenarios))
        duplicate_retry["reply-complete"]["response"]["body"]["messages"].append(
            json.loads(json.dumps(duplicate_retry["reply-complete"]["response"]["body"]["messages"][-1]))
        )
        invalid_get_mutations.append(duplicate_retry)

        for index, scenarios in enumerate(invalid_get_mutations):
            with self.subTest(get_mutation=index):
                mutation_errors: list[str] = []
                validator.check_practice_reply_recovery_semantics(
                    "getPracticeSession", scenarios, mutation_errors
                )
                self.assertTrue(mutation_errors)

        untyped_error = json.loads(json.dumps(send_scenarios))
        untyped_error["ai-timeout-retryable"]["response"]["body"] = {
            "message": "raw provider timeout"
        }
        mutation_errors = []
        validator.check_practice_reply_recovery_semantics(
            "sendPracticeMessage", untyped_error, mutation_errors
        )
        self.assertTrue(mutation_errors)

        changed_retry_id = json.loads(json.dumps(send_scenarios))
        changed_retry_id["retry-success-same-client-message"]["request"]["body"]["clientMessageId"] = (
            "01918fa0-0000-7000-8000-000000007011"
        )
        mutation_errors = []
        validator.check_practice_reply_recovery_semantics(
            "sendPracticeMessage", changed_retry_id, mutation_errors
        )
        self.assertTrue(mutation_errors)

    def test_schema_validator_enforces_round_pair_pattern_and_unique_progress(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schemas = spec["components"]["schemas"]

        missing_pair_errors: list[str] = []
        validator.schema_validate(
            {"roundId": "round-1-technical"},
            schemas["PracticePlan"],
            root=spec,
            path="plan",
            errors=missing_pair_errors,
        )
        self.assertTrue(any("dependent" in error and "roundSequence" in error for error in missing_pair_errors))

        pattern_errors: list[str] = []
        validator.schema_validate(
            "01918fa0-0000-7000-8000-000000004000",
            schemas["CreatePracticePlanRequest"]["properties"]["roundId"],
            root=spec,
            path="request.roundId",
            errors=pattern_errors,
        )
        self.assertTrue(any("pattern" in error for error in pattern_errors))

        duplicate_errors: list[str] = []
        duplicate = {"roundId": "round-1-technical", "roundSequence": 1}
        validator.schema_validate(
            [duplicate, duplicate],
            schemas["PracticeProgress"]["properties"]["completedRounds"],
            root=spec,
            path="progress.completedRounds",
            errors=duplicate_errors,
        )
        self.assertTrue(any("uniqueItems" in error for error in duplicate_errors))

    def test_create_practice_plan_conditional_request_matrix(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schema = spec["components"]["schemas"]["CreatePracticePlanRequest"]
        baseline = {
            "targetJobId": "01918fa0-0001-7000-8000-000000000001",
            "goal": "baseline",
            "interviewerPersona": "technical_manager",
            "difficulty": "standard",
            "language": "zh-CN",
            "timeBudgetMinutes": 45,
            "resumeId": "01918fa0-0002-7000-8000-000000000002",
        }
        source_report_id = "01918fa0-0003-7000-8000-000000000003"
        valid = [
            baseline,
            {"goal": "retry_current_round", "sourceReportId": source_report_id},
            {"goal": "next_round", "sourceReportId": source_report_id},
        ]
        for index, body in enumerate(valid):
            with self.subTest(valid=index):
                errors: list[str] = []
                validator.schema_validate(body, schema, root=spec, path="request", errors=errors)
                self.assertEqual([], errors)

        invalid = [
            {**baseline, "sourceReportId": source_report_id},
            {"goal": "retry_current_round"},
            {"goal": "retry_current_round", "sourceReportId": None},
            {"goal": "retry_current_round", "sourceReportId": ""},
            {"goal": "retry_current_round", "sourceReportId": "not-a-uuid"},
            {"goal": "retry_current_round", "sourceReportId": source_report_id, "targetJobId": baseline["targetJobId"]},
            {"goal": "next_round"},
            {"goal": "next_round", "sourceReportId": None},
            {"goal": "next_round", "sourceReportId": ""},
            {"goal": "next_round", "sourceReportId": "not-a-uuid"},
            {"goal": "next_round", "sourceReportId": source_report_id, "difficulty": "standard"},
            {"goal": "next_round", "sourceReportId": source_report_id, "focusCompetencyCodes": ["delivery"]},
        ]
        for index, body in enumerate(invalid):
            with self.subTest(invalid=index):
                errors = []
                validator.schema_validate(body, schema, root=spec, path="request", errors=errors)
                self.assertTrue(errors, body)

    def test_schema_validator_enforces_report_bounds_and_closed_objects(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schemas = spec["components"]["schemas"]

        cases = [
            (schemas["DimensionAssessment"], {"code": "x", "label": "X", "status": "strong", "confidence": "high", "dimension": "legacy"}),
            (schemas["ReportHighlight"], {"dimensionCode": "x", "evidence": "ok", "confidence": "high", "questionId": "legacy"}),
            (schemas["ReportIssue"], {"dimensionCode": "x", "evidence": "ok", "confidence": "high", "unknown": True}),
            (schemas["ReportNextAction"], {"type": "review_evidence", "label": "x" * 201}),
        ]
        for index, (schema, value) in enumerate(cases):
            with self.subTest(case=index):
                errors: list[str] = []
                validator.schema_validate(value, schema, root=spec, path="report", errors=errors)
                self.assertTrue(errors, value)

    def test_schema_validator_selects_feedback_status_conditional_branch(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schema = spec["components"]["schemas"]["FeedbackReport"]
        queued = _load_fixture("getFeedbackReport", "Reports")["scenarios"]["queued"]
        body = queued["response"]["body"]

        errors: list[str] = []
        validator.schema_validate(body, schema, root=spec, path="report", errors=errors)
        self.assertEqual([], errors)

        leaked = json.loads(json.dumps(body))
        leaked["summary"] = "generation is not ready"
        errors = []
        validator.schema_validate(leaked, schema, root=spec, path="report", errors=errors)
        self.assertTrue(any("expected null" in error for error in errors), errors)

    def test_schema_validator_rejects_ready_null_state_fields(self) -> None:
        validator = _load_validator()
        spec = _load_openapi()
        schema = spec["components"]["schemas"]["FeedbackReport"]
        ready = _load_fixture("getFeedbackReport", "Reports")["scenarios"]["ready-needs-practice"]
        for field in ("preparednessLevel", "provenance"):
            with self.subTest(field=field):
                body = json.loads(json.dumps(ready["response"]["body"]))
                body[field] = None
                errors: list[str] = []
                validator.schema_validate(body, schema, root=spec, path="report", errors=errors)
                self.assertTrue(errors, field)

        body = json.loads(json.dumps(ready["response"]["body"]))
        body["errorCode"] = "AI_OUTPUT_INVALID"
        errors = []
        validator.schema_validate(body, schema, root=spec, path="report", errors=errors)
        self.assertTrue(errors, "ready errorCode must be null")

        queued = _load_fixture("getFeedbackReport", "Reports")["scenarios"]["queued"]
        body = json.loads(json.dumps(queued["response"]["body"]))
        body["errorCode"] = "AI_OUTPUT_INVALID"
        errors = []
        validator.schema_validate(body, schema, root=spec, path="report", errors=errors)
        self.assertTrue(errors, "queued errorCode must be null")

        failed = _load_fixture("getFeedbackReport", "Reports")["scenarios"]["failed"]
        body = json.loads(json.dumps(failed["response"]["body"]))
        body["errorCode"] = None
        errors = []
        validator.schema_validate(body, schema, root=spec, path="report", errors=errors)
        self.assertTrue(errors, "failed errorCode must be non-null")

    def test_practice_round_fixtures_cover_current_legacy_and_progress_states(self) -> None:
        validator = _load_validator()
        expected_scenarios = {
            ("PracticePlans", "createPracticePlan"): {
                "default",
                "retry-derived",
                "next-derived",
                "round-mismatch",
            },
            ("PracticePlans", "getPracticePlan"): {"default", "legacy-null-round-identity"},
            ("TargetJobs", "getTargetJob"): {"default", "not-started-progress", "all-completed-progress"},
            ("TargetJobs", "listTargetJobs"): {"default", "not-started-progress", "all-completed-progress"},
        }

        for (tag, opid), expected in expected_scenarios.items():
            scenarios = _load_fixture(opid, tag)["scenarios"]
            with self.subTest(operationId=opid):
                self.assertTrue(expected.issubset(scenarios), sorted(expected - set(scenarios)))
                errors: list[str] = []
                validator.check_practice_round_semantics(opid, scenarios, errors)
                self.assertEqual([], errors)

    def test_feedback_report_fixtures_cover_current_status_matrix(self) -> None:
        validator = _load_validator()
        scenarios = _load_fixture("getFeedbackReport", "Reports")["scenarios"]
        expected = {
            "default",
            "prototype-baseline",
            "ready-needs-practice",
            "queued",
            "generating",
            "ready-well-prepared",
            "ready-empty-focus",
            "failed",
            "failed-context-too-large",
            "invalid-contract",
            "long-content",
        }
        self.assertTrue(expected.issubset(scenarios), sorted(expected - set(scenarios)))

        errors: list[str] = []
        validator.check_feedback_report_semantics(scenarios, errors)
        self.assertEqual([], errors)

        self.assertEqual(
            "REPORT_CONTEXT_TOO_LARGE",
            scenarios["failed-context-too-large"]["response"]["body"]["errorCode"],
        )

    def test_feedback_report_semantics_rejects_oversize_error_alias(self) -> None:
        validator = _load_validator()
        scenarios = _load_fixture("getFeedbackReport", "Reports")["scenarios"]
        mutated = json.loads(json.dumps(scenarios))
        mutated["failed-context-too-large"]["response"]["body"]["errorCode"] = (
            "REPORT_INPUT_TOO_LARGE"
        )
        errors: list[str] = []

        validator.check_feedback_report_semantics(mutated, errors)

        self.assertTrue(any("REPORT_CONTEXT_TOO_LARGE" in error for error in errors), errors)

    def test_feedback_report_semantics_rejects_ready_null_state_fields(self) -> None:
        validator = _load_validator()
        scenarios = _load_fixture("getFeedbackReport", "Reports")["scenarios"]
        for field in ("preparednessLevel", "provenance"):
            with self.subTest(field=field):
                mutated = json.loads(json.dumps(scenarios))
                mutated["ready-needs-practice"]["response"]["body"][field] = None
                errors: list[str] = []
                validator.check_feedback_report_semantics(mutated, errors)
                self.assertTrue(any(field in error for error in errors), errors)

    def test_practice_progress_validator_rejects_non_prefix_completion(self) -> None:
        validator = _load_validator()
        target = {
            "summary": {
                "interviewRounds": [
                    {"sequence": 1, "type": "technical"},
                    {"sequence": 2, "type": "manager"},
                ]
            },
            "practiceProgress": {
                "status": "in_progress",
                "completedRounds": [{"roundId": "round-2-manager", "roundSequence": 2}],
                "currentRound": {"roundId": "round-1-technical", "roundSequence": 1},
            },
        }
        errors: list[str] = []

        validator.check_target_job_practice_progress("fixture.target", target, errors)

        self.assertTrue(any("completedRounds" in error and "prefix" in error for error in errors), errors)

    def test_register_resume_fileless_source_variants_omit_file_object_id(self) -> None:
        scenarios = _load_fixture("registerResume", "Resumes")["scenarios"]

        paste_body = scenarios["paste-text"]["request"]["body"]
        self.assertEqual("paste", paste_body["sourceType"])
        self.assertIn("rawText", paste_body)
        self.assertNotIn("fileObjectId", paste_body)

        for scenario in scenarios.values():
            body = scenario["request"]["body"]
            self.assertIn(body["sourceType"], {"upload", "paste"})
            self.assertNotIn("guidedAnswers", body)

    def test_list_resumes_represents_fileless_assets_without_file_object_id(self) -> None:
        scenarios = _load_fixture("listResumes", "Resumes")["scenarios"]
        items = []
        for scenario in scenarios.values():
            items.extend(scenario["response"]["body"]["items"])
        by_source = {item["sourceType"]: item for item in items}

        self.assertIsInstance(by_source["upload"]["fileObjectId"], str)
        self.assertIsNone(by_source["paste"]["fileObjectId"])
        self.assertEqual({"upload", "paste"}, set(by_source))

    def test_uuid_format_ids_are_uuidv7_no_tmp_prefix(self) -> None:
        validator = _load_validator()
        for tag, opid, _path, data in validator.walk_fixtures(FIXTURES_ROOT):
            for path, value in _walk_strings(data):
                with self.subTest(operationId=opid, field=path):
                    self.assertNotRegex(
                        value, TEMP_ID_RE,
                        f"{opid}.{path}: tmp_ id forbidden in fixtures",
                    )
                    # Only enforce v7 shape on values that already look like a
                    # UUID: this avoids false positives on string identifiers
                    # like operationId / modelId / clientEventId-as-token, which
                    # share the *Id suffix but are not UUID-typed in the schema.
                    if UUID_SHAPE_RE.match(value):
                        self.assertRegex(
                            value, UUID_V7_RE,
                            f"{opid}.{path}={value!r} must be UUIDv7 (matched UUID shape but not v7 layout)",
                        )

    def test_datetime_strings_are_iso8601_utc(self) -> None:
        validator = _load_validator()
        datetime_fields = {"createdAt", "updatedAt", "askedAt", "occurredAt",
                           "expiresAt", "sessionExpiresAt", "requestedAt",
                           "completedAt", "clientCompletedAt"}
        for tag, opid, _path, data in validator.walk_fixtures(FIXTURES_ROOT):
            for path, value in _walk_strings(data):
                leaf = path.rsplit(".", 1)[-1].rstrip("0123456789")
                leaf = leaf.split("[", 1)[0]
                if leaf in datetime_fields and value:
                    with self.subTest(operationId=opid, field=path):
                        self.assertRegex(
                            value, ISO_DATETIME_RE,
                            f"{opid}.{path}={value!r} must be RFC3339 UTC",
                        )

    def test_privacy_allowlist_emails_phones_companies(self) -> None:
        validator = _load_validator()
        for tag, opid, _path, data in validator.walk_fixtures(FIXTURES_ROOT):
            for path, value in _walk_strings(data):
                with self.subTest(operationId=opid, field=path):
                    for match in EMAIL_RE.findall(value):
                        domain = match.lower()
                        # `.example` reserved suffix is also allowed
                        if not (domain in ALLOWED_EMAIL_DOMAINS or domain.endswith(".example")):
                            self.fail(
                                f"{opid}.{path}: email domain {match!r} is not on allowlist"
                            )
                    for phone in PHONE_RE.findall(value):
                        if not phone.startswith(ALLOWED_PHONE_PREFIX):
                            self.fail(
                                f"{opid}.{path}: phone {phone!r} not on +1-555-01xx allowlist"
                            )
                    match = COMPANY_BLACKLIST_RE.search(value)
                    if match:
                        self.fail(
                            f"{opid}.{path}: blacklisted employer brand "
                            f"{match.group(1)!r} present in {value!r}"
                        )


if __name__ == "__main__":
    unittest.main()
