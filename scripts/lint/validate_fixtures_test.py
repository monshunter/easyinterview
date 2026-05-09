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


# (tag, operationId, expected default status, has_request_body)
EXPECTED_OPERATIONS = [
    ("Auth", "getMe", 200, False),
    ("Auth", "deleteMe", 202, False),
    ("Auth", "startAuthEmailChallenge", 202, True),
    ("Auth", "verifyAuthEmailChallenge", 200, False),
    ("Auth", "logout", 204, False),
    ("Auth", "getRuntimeConfig", 200, False),
    ("Uploads", "createUploadPresign", 201, True),
    ("Profile", "getMyProfile", 200, False),
    ("Profile", "updateMyProfile", 200, True),
    ("Profile", "listExperienceCards", 200, False),
    ("Profile", "createExperienceCard", 201, True),
    ("Profile", "updateExperienceCard", 200, True),
    ("Resumes", "registerResume", 202, True),
    ("Resumes", "getResume", 200, False),
    ("TargetJobs", "importTargetJob", 202, True),
    ("TargetJobs", "listTargetJobs", 200, False),
    ("TargetJobs", "getTargetJob", 200, False),
    ("TargetJobs", "updateTargetJob", 200, True),
    ("PracticePlans", "createPracticePlan", 201, True),
    ("PracticePlans", "getPracticePlan", 200, False),
    ("PracticeSessions", "startPracticeSession", 201, True),
    ("PracticeSessions", "getPracticeSession", 200, False),
    ("PracticeSessions", "appendSessionEvent", 200, True),
    ("PracticeSessions", "completePracticeSession", 202, True),
    ("Reports", "getFeedbackReport", 200, False),
    ("Reports", "listTargetJobReports", 200, False),
    ("ResumeTailor", "requestResumeTailor", 202, True),
    ("ResumeTailor", "getResumeTailorRun", 200, False),
    ("Debriefs", "createDebrief", 202, True),
    ("Debriefs", "getDebrief", 200, False),
    ("Jobs", "getJob", 200, False),
    ("Privacy", "requestPrivacyExport", 501, False),
    ("Privacy", "requestPrivacyDelete", 202, False),
    ("Privacy", "getPrivacyRequest", 200, False),
]


class FixtureSkeletonTest(unittest.TestCase):
    """Phase 1.1 structural contract."""

    def test_thirty_four_operations_expected(self) -> None:
        self.assertEqual(len(EXPECTED_OPERATIONS), 34)

    def test_twelve_unique_tags(self) -> None:
        tags = {tag for tag, *_ in EXPECTED_OPERATIONS}
        self.assertEqual(len(tags), 12)

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
                if opid == "deleteMe":
                    self.assertIn("request", default, f"{path}: request headers must be present")
                    self.assertIn("headers", default["request"], f"{path}: request.headers must be present")
                    self.assertIn("Idempotency-Key", default["request"]["headers"])
                    self.assertNotIn("body", default["request"], f"{path}: DELETE /me has no request body")
                    continue
                if has_req:
                    self.assertIn("request", default, f"{path}: request must be present")
                    self.assertIn("body", default["request"], f"{path}: request.body must be present")
                else:
                    self.assertNotIn(
                        "request",
                        default,
                        f"{path}: request must be omitted when operation has no requestBody",
                    )


class FixtureValidatorWalkerTest(unittest.TestCase):
    """Validator helper exposes a structural walk over openapi/fixtures/."""

    def test_walker_returns_34_entries(self) -> None:
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
    "appendSessionEvent": ["assistantAction.provenance"],
    "getFeedbackReport": ["provenance"],
    "getResumeTailorRun": ["provenance"],
    "getDebrief": ["provenance"],
}

LIST_OPERATIONS = [
    "listExperienceCards",
    "listTargetJobs",
    "listTargetJobReports",
]

# *WithJob async operations and the JobType they must emit.
WITH_JOB_OPERATIONS = {
    "startAuthEmailChallenge": None,  # 202 but no Job envelope (auth challenge)
    "registerResume": "resume_parse",
    "importTargetJob": "target_import",
    "completePracticeSession": "report_generate",
    "requestResumeTailor": "resume_tailor",
    "createDebrief": "debrief_generate",
    "deleteMe": "privacy_delete",
    "requestPrivacyDelete": "privacy_delete",
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

    def test_debrief_default_fixture_excludes_p1_followup_fields(self) -> None:
        body = _load_fixture("getDebrief", "Debriefs")["scenarios"]["default"]["response"]["body"]
        self.assertNotIn("thankYouDraft", body)
        self.assertNotIn("nextRoundChecklist", body)

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
