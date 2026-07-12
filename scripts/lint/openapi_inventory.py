#!/usr/bin/env python3
"""Structural inventory linter for openapi/openapi.yaml.

Enforces the v1.0.0 freeze locked by
[openapi-v1-contract spec](../../docs/spec/openapi-v1-contract/spec.md) §3.1.1
and the §4.1 status-code / Idempotency / privacy-export rules. Runs locally as
part of `make codegen-check`; the spec C-1 acceptance test is a `swagger-cli
validate` plus this script returning exit 0.

Run: `python3 scripts/lint/openapi_inventory.py [openapi.yaml]`
"""
from __future__ import annotations

import sys
from pathlib import Path
from typing import Any

import yaml

# Ordered tag list per spec §2.1.
EXPECTED_TAGS: list[str] = [
    "Auth",
    "Uploads",
    "Resumes",
    "TargetJobs",
    "PracticePlans",
    "PracticeSessions",
    "Reports",
    "ResumeTailor",
    "Jobs",
    "Privacy",
]

# (tag, method, path, operationId) tuples per spec §3.1.1 plus additive owner plans.
EXPECTED_OPERATIONS: list[tuple[str, str, str, str]] = [
    ("Auth", "get", "/me", "getMe"),
    ("Auth", "patch", "/me", "completeMyProfile"),
    ("Auth", "delete", "/me", "deleteMe"),
    ("Auth", "post", "/auth/email/start", "startAuthEmailChallenge"),
    ("Auth", "get", "/auth/email/verify", "verifyAuthEmailChallenge"),
    ("Auth", "post", "/auth/logout", "logout"),
    ("Uploads", "post", "/uploads/presign", "createUploadPresign"),
    ("Resumes", "post", "/resumes", "registerResume"),
    ("Resumes", "get", "/resumes", "listResumes"),
    ("Resumes", "get", "/resumes/{resumeId}", "getResume"),
    ("Resumes", "get", "/resumes/{resumeId}/source", "getResumeSource"),
    ("Resumes", "patch", "/resumes/{resumeId}", "updateResume"),
    ("Resumes", "post", "/resumes/{resumeId}/duplicate", "duplicateResume"),
    ("Resumes", "post", "/resumes/{resumeId}/archive", "archiveResume"),
    ("Resumes", "post", "/resumes/{resumeId}/exports", "exportResume"),
    ("TargetJobs", "post", "/targets/import", "importTargetJob"),
    ("TargetJobs", "get", "/targets", "listTargetJobs"),
    ("TargetJobs", "get", "/targets/{targetJobId}", "getTargetJob"),
    ("TargetJobs", "patch", "/targets/{targetJobId}", "updateTargetJob"),
    ("TargetJobs", "post", "/targets/{targetJobId}/archive", "archiveTargetJob"),
    ("PracticePlans", "post", "/practice/plans", "createPracticePlan"),
    ("PracticePlans", "get", "/practice/plans/{planId}", "getPracticePlan"),
    ("PracticeSessions", "get", "/practice/sessions", "listPracticeSessions"),
    ("PracticeSessions", "post", "/practice/sessions", "startPracticeSession"),
    ("PracticeSessions", "get", "/practice/sessions/{sessionId}", "getPracticeSession"),
    ("PracticeSessions", "post", "/practice/sessions/{sessionId}/voice-turns", "createPracticeVoiceTurn"),
    ("PracticeSessions", "post", "/practice/sessions/{sessionId}/messages", "sendPracticeMessage"),
    ("PracticeSessions", "post", "/practice/sessions/{sessionId}/complete", "completePracticeSession"),
    ("Reports", "get", "/reports/{reportId}", "getFeedbackReport"),
    ("Reports", "get", "/targets/{targetJobId}/reports", "listTargetJobReports"),
    ("ResumeTailor", "post", "/resume/tailor", "requestResumeTailor"),
    ("ResumeTailor", "get", "/resume/tailor-runs/{tailorRunId}", "getResumeTailorRun"),
    ("Jobs", "get", "/jobs/{jobId}", "getJob"),
    ("Privacy", "post", "/privacy/exports", "requestPrivacyExport"),
    ("Privacy", "post", "/privacy/deletions", "requestPrivacyDelete"),
    ("Privacy", "get", "/privacy/requests/{privacyRequestId}", "getPrivacyRequest"),
    ("Auth", "get", "/runtime-config", "getRuntimeConfig"),
]

# Side-effect endpoints that must reference `Idempotency-Key` per plan §1.3 / spec D-6.
IK_REQUIRED: set[tuple[str, str]] = {
    ("delete", "/me"),
    ("post", "/uploads/presign"),
    ("post", "/resumes"),
    ("patch", "/resumes/{resumeId}"),
    ("post", "/resumes/{resumeId}/duplicate"),
    ("post", "/resumes/{resumeId}/archive"),
    ("post", "/resumes/{resumeId}/exports"),
    ("post", "/targets/import"),
    ("patch", "/targets/{targetJobId}"),
    ("post", "/targets/{targetJobId}/archive"),
    ("post", "/practice/plans"),
    ("post", "/practice/sessions"),
    ("post", "/practice/sessions/{sessionId}/voice-turns"),
    ("post", "/practice/sessions/{sessionId}/complete"),
    ("post", "/resume/tailor"),
    ("post", "/privacy/exports"),
    ("post", "/privacy/deletions"),
}

# Endpoints that must NOT carry `Idempotency-Key` per plan §1.3 (ADR-Q1 + clientEventId).
IK_FORBIDDEN: set[tuple[str, str]] = {
    ("post", "/auth/email/start"),
    ("post", "/practice/sessions/{sessionId}/messages"),
}

# Public endpoints per spec §4.1 — must declare `security: []` to override doc-level cookie auth.
PUBLIC_ENDPOINTS: set[tuple[str, str]] = {
    ("post", "/auth/email/start"),
    ("get", "/auth/email/verify"),
    ("get", "/runtime-config"),
}

# AI-generation schemas listed in spec §4.6 — each must reach `GenerationProvenance` via $ref topology.
AI_PROVENANCE_SCHEMAS: list[str] = [
    "TargetJobSummary",
    "TargetJobFitSummary",
    "FeedbackReport",
    "ResumeTailorRun",
    "Resume",
]

# P0 export 501 exceptions locked by the current contract.
P0_501_ENDPOINTS: dict[tuple[str, str], str] = {
    ("post", "/privacy/exports"): "PRIVACY_EXPORT_NOT_AVAILABLE",
    ("post", "/resumes/{resumeId}/exports"): "RESUME_EXPORT_NOT_AVAILABLE",
}

FORBIDDEN_PRODUCT_SCOPE_TOKENS: tuple[str, ...] = (
    "Mistakes",
    "Growth",
    "MistakeEntry",
    "GrowthOverview",
    "MistakeStatus",
    "listMistakes",
    "retestMistake",
    "getGrowthOverview",
    "openMistakeCount",
    "writtenToMistakeBook",
    "mistakeIds",
    "single_drill",
    "counter_questions",
    "warmup",
    "core_interview",
    "fix_mistake",
    "getMyProfile",
    "updateMyProfile",
    "listExperienceCards",
    "createExperienceCard",
    "updateExperienceCard",
    "createDebrief",
    "suggestDebriefQuestions",
    "getDebrief",
    "CandidateProfile",
    "ExperienceCard",
    "DebriefWithJob",
    "sourceDebriefId",
)

EXPECTED_PRODUCT_ENUMS: dict[str, list[str]] = {
    "PracticeGoal": ["baseline", "retry_current_round", "next_round"],
    "ReportNextAction.type": ["retry_current_round", "next_round", "review_evidence"],
    "JobType": [
        "target_import",
        "resume_parse",
        "report_generate",
        "resume_tailor",
        "privacy_export",
        "privacy_delete",
    ],
    "ResourceType": [
        "target_job",
        "feedback_report",
        "resume_asset",
        "resume_tailor_run",
        "privacy_request",
    ],
}

DEFAULT_OPENAPI_PATH = Path("openapi/openapi.yaml")
APIERROR_REF = "#/components/responses/ApiErrorResponse"
IDEMPOTENCY_REF = "#/components/parameters/IdempotencyKey"
HTTP_METHODS = ("get", "post", "put", "patch", "delete")


def fail(errors: list[str]) -> None:
    """Print errors to stderr and exit 1."""
    sys.stderr.write("ERROR: openapi/openapi.yaml inventory check failed:\n")
    for line in errors:
        sys.stderr.write(f"  - {line}\n")
    sys.exit(1)


def collect_refs(node: Any, found: set[str]) -> None:
    """Recursively collect every `$ref` value reachable from node."""
    if isinstance(node, dict):
        for key, value in node.items():
            if key == "$ref" and isinstance(value, str):
                found.add(value)
            else:
                collect_refs(value, found)
    elif isinstance(node, list):
        for item in node:
            collect_refs(item, found)


def reachable_schemas(schemas: dict[str, Any], roots: list[str]) -> set[str]:
    """Walk the components.schemas $ref graph from roots, return reachable names."""
    seen: set[str] = set()
    stack = list(roots)
    while stack:
        name = stack.pop()
        if name in seen or name not in schemas:
            continue
        seen.add(name)
        refs: set[str] = set()
        collect_refs(schemas[name], refs)
        for ref in refs:
            if ref.startswith("#/components/schemas/"):
                stack.append(ref.rsplit("/", 1)[-1])
    return seen


def _schema_properties(schemas: dict[str, Any], schema_name: str) -> dict[str, Any]:
    schema = schemas.get(schema_name) or {}
    props = schema.get("properties") or {}
    return props if isinstance(props, dict) else {}


def validate_product_scope_contract(data: dict[str, Any], errors: list[str]) -> None:
    """Enforce product-scope v1.2 / current UI semantic invariants that are
    stronger than structural OpenAPI validity."""
    text = yaml.safe_dump(data, sort_keys=False, allow_unicode=True)
    for token in FORBIDDEN_PRODUCT_SCOPE_TOKENS:
        if token in text:
            errors.append(f"product-scope v1.2 forbidden token still present: {token!r}")

    paths = data.get("paths") or {}
    for path_str in paths:
        for forbidden_prefix in ("/mistakes", "/growth", "/voice", "/drill", "/profiles", "/debriefs"):
            if path_str.startswith(forbidden_prefix):
                errors.append(f"forbidden product-scope path {path_str!r} (matches {forbidden_prefix!r})")

    tags = {tag.get("name") for tag in (data.get("tags") or []) if isinstance(tag, dict)}
    for tag in ("Mistakes", "Growth", "Voice", "Drill", "Profile", "Debriefs"):
        if tag in tags:
            errors.append(f"forbidden product-scope tag {tag!r}")

    schemas = ((data.get("components") or {}).get("schemas") or {})
    for stale_schema in ("PracticeMode", "PracticeTurn", "QuestionReviewStatus", "QuestionAssessment"):
        if stale_schema in schemas:
            errors.append(f"stale practice/report schema must be removed: {stale_schema}")
    for name, expected in EXPECTED_PRODUCT_ENUMS.items():
        if name == "ReportNextAction.type":
            report_next_action_type = ((_schema_properties(schemas, "ReportNextAction").get("type") or {}).get("enum") or [])
            if report_next_action_type != expected:
                errors.append(f"ReportNextAction.type enum mismatch: expected {expected}, got {report_next_action_type}")
            continue
        actual = (schemas.get(name) or {}).get("enum")
        if actual != expected:
            errors.append(f"{name} enum mismatch: expected {expected}, got {actual}")

    feedback_report = schemas.get("FeedbackReport") or {}
    feedback_props = _schema_properties(schemas, "FeedbackReport")
    feedback_required = set(feedback_report.get("required") or [])
    for required in ("sessionId", "targetJobId"):
        if required not in feedback_required:
            errors.append(f"FeedbackReport must be session-scoped and require {required!r}")
    for prop in ("dimensionAssessments", "retryFocusCompetencyCodes", "provenance"):
        if prop not in feedback_props:
            errors.append(f"FeedbackReport missing current product property {prop!r}")
    for prop in (
        "questionAssessments",
        "retryFocusTurnIds",
        "mistakes",
        "mistakeEntries",
        "mistakeIds",
        "openMistakeCount",
    ):
        if prop in feedback_props:
            errors.append(f"FeedbackReport must not expose stale property {prop!r}")
    if "QuestionAssessment" in schemas:
        errors.append("QuestionAssessment schema must be removed from conversation-level reports")

    target_props = _schema_properties(schemas, "TargetJob")
    if "openQuestionIssueCount" not in target_props:
        errors.append("TargetJob must expose openQuestionIssueCount")
    if "openMistakeCount" in target_props:
        errors.append("TargetJob must not expose old openMistakeCount")


def has_parameter_ref(operation: dict[str, Any], target_ref: str) -> bool:
    for param in operation.get("parameters") or []:
        if isinstance(param, dict) and param.get("$ref") == target_ref:
            return True
    return False


def main(argv: list[str]) -> int:
    path = Path(argv[1]) if len(argv) > 1 else DEFAULT_OPENAPI_PATH
    if not path.exists():
        fail([f"openapi file not found: {path}"])
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        fail(["openapi document is not a YAML mapping"])

    errors: list[str] = []

    # Doc head invariants.
    if data.get("openapi") != "3.1.0":
        errors.append(f"`openapi` must be 3.1.0; got {data.get('openapi')!r}")
    info = data.get("info") or {}
    if info.get("version") != "1.0.0":
        errors.append(f"`info.version` must be 1.0.0; got {info.get('version')!r}")
    servers = data.get("servers") or []
    if not servers or servers[0].get("url") != "/api/v1":
        errors.append(f"`servers[0].url` must be /api/v1; got {servers!r}")

    # Tag presence and order (spec D-11).
    tags = [tag.get("name") for tag in (data.get("tags") or [])]
    if tags != EXPECTED_TAGS:
        errors.append(f"tag list mismatch:\n      expected: {EXPECTED_TAGS}\n      got     : {tags}")

    # Document-level security must require sessionCookie.
    doc_security = data.get("security")
    if doc_security != [{"sessionCookie": []}]:
        errors.append(f"document-level `security` must be [{{sessionCookie: []}}]; got {doc_security!r}")

    # ADR-Q1: no `http`/`bearer` security scheme allowed.
    schemes = (data.get("components") or {}).get("securitySchemes") or {}
    for name, scheme in schemes.items():
        if not isinstance(scheme, dict):
            continue
        if scheme.get("type") == "http" and (scheme.get("scheme") or "").lower() == "bearer":
            errors.append(f"forbidden `http/bearer` security scheme `{name}` (ADR-Q1 locks session cookie)")
    if "sessionCookie" not in schemes:
        errors.append("missing `sessionCookie` security scheme (ADR-Q1)")
    else:
        sc = schemes["sessionCookie"]
        if sc.get("type") != "apiKey" or sc.get("in") != "cookie" or not sc.get("name"):
            errors.append(f"`sessionCookie` must be apiKey/cookie with `name`; got {sc!r}")

    # Operation enumeration (spec §3.1.1 D-9 plus additive owner plans).
    paths = data.get("paths") or {}
    seen_ops: set[tuple[str, str, str, str]] = set()
    operation_count = 0
    for path_str, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, operation in item.items():
            if method not in HTTP_METHODS or not isinstance(operation, dict):
                continue
            operation_count += 1
            tags_op = operation.get("tags") or []
            tag = tags_op[0] if tags_op else None
            seen_ops.add((tag, method, path_str, operation.get("operationId")))

    expected_set = set(EXPECTED_OPERATIONS)
    missing = expected_set - seen_ops
    extra = seen_ops - expected_set
    if missing:
        errors.append("missing operations: " + ", ".join(sorted(f"{m.upper()} {p} ({o})" for _, m, p, o in missing)))
    if extra:
        errors.append("unexpected operations: " + ", ".join(sorted(f"{m.upper()} {p} ({o})" for _, m, p, o in extra)))
    expected_operation_count = len(EXPECTED_OPERATIONS)
    if operation_count != expected_operation_count:
        errors.append(
            f"operation count must be {expected_operation_count} (spec §3.1.1 plus additive owner plans); got {operation_count}"
        )

    # operationId uniqueness.
    op_ids = [op for _, _, _, op in seen_ops]
    duplicates = sorted({oid for oid in op_ids if op_ids.count(oid) > 1})
    if duplicates:
        errors.append(f"duplicate operationIds: {duplicates}")

    # Each operation needs `default: $ref ApiErrorResponse`.
    for path_str, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, operation in item.items():
            if method not in HTTP_METHODS or not isinstance(operation, dict):
                continue
            default_response = (operation.get("responses") or {}).get("default")
            if not isinstance(default_response, dict) or default_response.get("$ref") != APIERROR_REF:
                errors.append(f"{method.upper()} {path_str}: response.default must be $ref {APIERROR_REF}")

    # Idempotency-Key required / forbidden sets (spec D-6 + ADR-Q1).
    for method, path_str in IK_REQUIRED:
        operation = (paths.get(path_str) or {}).get(method)
        if not isinstance(operation, dict):
            errors.append(f"IK_REQUIRED endpoint missing: {method.upper()} {path_str}")
            continue
        if not has_parameter_ref(operation, IDEMPOTENCY_REF):
            errors.append(f"{method.upper()} {path_str}: must reference $ref {IDEMPOTENCY_REF}")
    for method, path_str in IK_FORBIDDEN:
        operation = (paths.get(path_str) or {}).get(method)
        if not isinstance(operation, dict):
            errors.append(f"IK_FORBIDDEN endpoint missing: {method.upper()} {path_str}")
            continue
        if has_parameter_ref(operation, IDEMPOTENCY_REF):
            errors.append(f"{method.upper()} {path_str}: must NOT reference {IDEMPOTENCY_REF} (uses clientEventId / rate-limit)")

    # Public endpoints must declare `security: []`.
    for method, path_str in PUBLIC_ENDPOINTS:
        operation = (paths.get(path_str) or {}).get(method)
        if not isinstance(operation, dict):
            continue
        if operation.get("security") != []:
            errors.append(f"{method.upper()} {path_str}: must declare `security: []` (public per spec §4.1)")

    # 501 uniqueness — only explicit P0 export exceptions may declare it (spec D-12 / D-18).
    five_oh_one_ops: list[tuple[str, str]] = []
    for path_str, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, operation in item.items():
            if method not in HTTP_METHODS or not isinstance(operation, dict):
                continue
            if "501" in (operation.get("responses") or {}):
                five_oh_one_ops.append((method, path_str))
    expected_501 = sorted(P0_501_ENDPOINTS.keys())
    if sorted(five_oh_one_ops) != expected_501:
        errors.append(f"501 must appear only on P0 export exceptions {expected_501}; got {five_oh_one_ops}")

    # P0 export 501 responses must declare ApiErrorResponse on JSON content.
    # The operation-specific error.code examples are owned by fixtures and
    # verified by `scripts/lint/validate_fixtures.py`.
    for method, path_str in expected_501:
        operation = ((paths.get(path_str) or {}).get(method) or {})
        response_501 = ((operation.get("responses") or {}).get("501") or {})
        content = (response_501.get("content") or {}).get("application/json") or {}
        schema_ref = (content.get("schema") or {}).get("$ref")
        if schema_ref != "#/components/schemas/ApiErrorResponse":
            errors.append(
                f"{method.upper()} {path_str} 501 content.application/json.schema must `$ref` ApiErrorResponse"
            )

    # GenerationProvenance contract (spec §4.6).
    schemas = ((data.get("components") or {}).get("schemas") or {})
    api_error = schemas.get("ApiError") or {}
    api_error_response = schemas.get("ApiErrorResponse") or {}
    api_error_required = sorted(api_error.get("required") or [])
    if api_error_required != ["code", "message", "requestId", "retryable"]:
        errors.append(f"ApiError must be the inner error object; required mismatch: {api_error_required}")
    if "error" in (api_error.get("properties") or {}):
        errors.append("ApiError must not contain the outer `error` envelope property")
    response_error_ref = (((api_error_response.get("properties") or {}).get("error") or {}).get("$ref"))
    if response_error_ref != "#/components/schemas/ApiError":
        errors.append(f"ApiErrorResponse.error must $ref ApiError; got {response_error_ref!r}")
    response_ref = (((data.get("components") or {}).get("responses") or {}).get("ApiErrorResponse") or {})
    response_schema_ref = (((response_ref.get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
    if response_schema_ref != "#/components/schemas/ApiErrorResponse":
        errors.append(f"components.responses.ApiErrorResponse must $ref schema ApiErrorResponse; got {response_schema_ref!r}")

    provenance = schemas.get("GenerationProvenance")
    if not isinstance(provenance, dict):
        errors.append("missing components.schemas.GenerationProvenance")
    else:
        expected_required = sorted(["promptVersion", "rubricVersion", "modelId", "language", "featureFlag", "dataSourceVersion"])
        actual_required = sorted(provenance.get("required") or [])
        if actual_required != expected_required:
            errors.append(f"GenerationProvenance.required mismatch: expected {expected_required}, got {actual_required}")
        actual_props = sorted((provenance.get("properties") or {}).keys())
        if actual_props != expected_required:
            errors.append(f"GenerationProvenance.properties mismatch: expected {expected_required}, got {actual_props}")
        rubric_desc = ((provenance.get("properties") or {}).get("rubricVersion") or {}).get("description") or ""
        if "not_applicable" not in rubric_desc:
            errors.append("GenerationProvenance.rubricVersion.description must mention literal `not_applicable` (spec §4.6)")

    # Each AI-generation schema must reach GenerationProvenance via $ref topology.
    for schema_name in AI_PROVENANCE_SCHEMAS:
        if schema_name not in schemas:
            errors.append(f"missing AI schema components.schemas.{schema_name}")
            continue
        reachable = reachable_schemas(schemas, [schema_name])
        if "GenerationProvenance" not in reachable:
            errors.append(f"{schema_name} cannot reach GenerationProvenance via $ref topology (spec §4.6)")

    validate_product_scope_contract(data, errors)

    # Sync against B1 truth source (spec C-8 partial).
    conventions_path = Path("shared/conventions.yaml")
    if conventions_path.exists():
        b1 = yaml.safe_load(conventions_path.read_text(encoding="utf-8"))
        b1_enum_map = {entry["name"]: list(entry["values"]) for entry in (b1.get("enums") or [])}
        for name, values in b1_enum_map.items():
            schema = schemas.get(name)
            if schema is None:
                errors.append(f"missing B1 enum mirror components.schemas.{name}")
                continue
            actual = schema.get("enum")
            if actual != values:
                errors.append(f"enum {name} drift vs shared/conventions.yaml: openapi={actual}, b1={values}")
        b1_codes = sorted({entry["code"] for entry in (b1.get("errors") or [])} | {"PRIVACY_EXPORT_NOT_AVAILABLE"})
        actual_codes = sorted((schemas.get("ApiErrorCode") or {}).get("enum") or [])
        if actual_codes != b1_codes:
            errors.append(f"ApiErrorCode drift vs shared/conventions.yaml#errors: openapi={actual_codes}, b1={b1_codes}")
        b1_job_statuses = list(b1.get("jobStatuses") or [])
        actual_job_statuses = (schemas.get("JobStatus") or {}).get("enum")
        if actual_job_statuses != b1_job_statuses:
            errors.append(f"JobStatus drift vs shared/conventions.yaml#jobStatuses: openapi={actual_job_statuses}, b1={b1_job_statuses}")

    if errors:
        fail(errors)
    sys.stdout.write(
        f"openapi inventory OK: {len(EXPECTED_TAGS)} tags, {len(EXPECTED_OPERATIONS)} operations, "
        "ApiErrorResponse/IK/501/Provenance invariants enforced; B1 enums in sync.\n"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
