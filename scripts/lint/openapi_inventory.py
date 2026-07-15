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
    ("Reports", "get", "/reports/{reportId}/conversation", "getReportConversation"),
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

RESUME_SUMMARY_FIELDS = [
    "id",
    "title",
    "displayName",
    "language",
    "sourceType",
    "parseStatus",
    "summaryHeadline",
    "hasReadableContent",
    "updatedAt",
]


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


def validate_resume_summary_contract(data: dict[str, Any], errors: list[str]) -> None:
    """Lock OPENAPI-005's summary-only list and full-detail read boundary."""
    schemas = ((data.get("components") or {}).get("schemas") or {})
    summary = schemas.get("ResumeSummary")
    if not isinstance(summary, dict):
        errors.append("missing components.schemas.ResumeSummary")
        return

    if set(summary) != {"type", "additionalProperties", "required", "properties"}:
        errors.append(
            "ResumeSummary must contain only type/additionalProperties/required/properties"
        )
    if summary.get("type") != "object":
        errors.append("ResumeSummary.type must be object")
    if summary.get("additionalProperties") is not False:
        errors.append("ResumeSummary.additionalProperties must be false")
    if summary.get("required") != RESUME_SUMMARY_FIELDS:
        errors.append(
            "ResumeSummary.required must exactly equal "
            f"{RESUME_SUMMARY_FIELDS}; got {summary.get('required')!r}"
        )

    properties = summary.get("properties") or {}
    if list(properties) != RESUME_SUMMARY_FIELDS:
        errors.append(
            "ResumeSummary.properties must exactly equal "
            f"{RESUME_SUMMARY_FIELDS}; got {list(properties)!r}"
        )
    expected_properties = {
        "id": {"type": "string", "format": "uuid"},
        "title": {"type": "string"},
        "displayName": {"type": "string"},
        "language": {"type": "string"},
        "sourceType": {"type": "string", "enum": ["upload", "paste"]},
        "parseStatus": {"$ref": "#/components/schemas/TargetJobParseStatus"},
        "summaryHeadline": {
            "oneOf": [{"type": "string"}, {"type": "null"}]
        },
        "hasReadableContent": {"type": "boolean"},
        "updatedAt": {"type": "string", "format": "date-time"},
    }
    if properties != expected_properties:
        errors.append("ResumeSummary property schemas must match OPENAPI-005 exactly")

    paginated = schemas.get("PaginatedResume") or {}
    all_of = paginated.get("allOf") or []
    expected_envelope = {"$ref": "#/components/schemas/PaginatedEnvelope"}
    if len(all_of) != 2 or all_of[0] != expected_envelope:
        errors.append("PaginatedResume must preserve the PaginatedEnvelope allOf")
    item_ref = None
    if len(all_of) == 2 and isinstance(all_of[1], dict):
        item_ref = (
            ((((all_of[1].get("properties") or {}).get("items") or {}).get("items")) or {}).get("$ref")
        )
    if item_ref != "#/components/schemas/ResumeSummary":
        errors.append("PaginatedResume.items must reference ResumeSummary")

    paths = data.get("paths") or {}
    list_operation = ((paths.get("/resumes") or {}).get("get") or {})
    if list_operation.get("operationId") != "listResumes":
        errors.append("GET /resumes operationId must remain listResumes")
    list_response_ref = (
        (((((list_operation.get("responses") or {}).get("200") or {}).get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
    )
    if list_response_ref != "#/components/schemas/PaginatedResume":
        errors.append("listResumes 200 response must remain PaginatedResume")

    full_detail_operations = {
        ("/resumes/{resumeId}", "get", "200", "getResume"),
        ("/resumes/{resumeId}", "patch", "200", "updateResume"),
        ("/resumes/{resumeId}/duplicate", "post", "201", "duplicateResume"),
        ("/resumes/{resumeId}/archive", "post", "202", "archiveResume"),
    }
    for path, method, status, operation_id in full_detail_operations:
        operation = ((paths.get(path) or {}).get(method) or {})
        if operation.get("operationId") != operation_id:
            errors.append(f"{method.upper()} {path} must remain {operation_id}")
            continue
        response_ref = (
            (((((operation.get("responses") or {}).get(status) or {}).get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
        )
        if response_ref != "#/components/schemas/Resume":
            errors.append(f"{operation_id} {status} response must remain full Resume")

    resume_properties = _schema_properties(schemas, "Resume")
    for field in (
        "fileObjectId",
        "originalText",
        "parsedTextSnapshot",
        "parsedSummary",
        "structuredProfile",
        "createdAt",
        "deletedAt",
    ):
        if field not in resume_properties:
            errors.append(f"full Resume detail must retain {field}")
    provenance_ref = (
        (((resume_properties.get("structuredProfile") or {}).get("properties") or {}).get("provenance") or {}).get("$ref")
    )
    if provenance_ref != "#/components/schemas/GenerationProvenance":
        errors.append("full Resume structuredProfile must retain GenerationProvenance")


def validate_targetjob_paste_only_contract(data: dict[str, Any], errors: list[str]) -> None:
    """Lock OPENAPI-002's paste-only TargetJob wire surface.

    The pre-launch correction intentionally has no source-wrapper or read-side
    provenance compatibility path. Resume and privacy uploads remain valid.
    """
    schemas = ((data.get("components") or {}).get("schemas") or {})
    request = schemas.get("ImportTargetJobRequest") or {}
    request_properties = request.get("properties") or {}
    expected_request_properties = ["rawText", "targetLanguage", "resumeId"]

    if request.get("type") != "object":
        errors.append("ImportTargetJobRequest.type must be object")
    if request.get("additionalProperties") is not False:
        errors.append("ImportTargetJobRequest must set additionalProperties=false")
    if request.get("required") != expected_request_properties:
        errors.append(
            "ImportTargetJobRequest.required must exactly equal "
            f"{expected_request_properties}; got {request.get('required')!r}"
        )
    if list(request_properties) != expected_request_properties:
        errors.append(
            "ImportTargetJobRequest.properties must exactly equal "
            f"{expected_request_properties}; got {list(request_properties)!r}"
        )

    raw_text = request_properties.get("rawText") or {}
    if raw_text.get("type") != "string":
        errors.append("ImportTargetJobRequest.rawText.type must be string")
    if raw_text.get("minLength") != 1:
        errors.append("ImportTargetJobRequest.rawText.minLength must be 1")
    if raw_text.get("pattern") != r"\S":
        errors.append(r"ImportTargetJobRequest.rawText.pattern must be \\S")

    for removed_schema in (
        "TargetJobImportSourceURL",
        "TargetJobImportSourceManualText",
        "TargetJobImportSourceFile",
        "TargetJobImportSourceManualForm",
        "TargetJobImportSource",
    ):
        if removed_schema in schemas:
            errors.append(f"removed TargetJob source schema still present: {removed_schema}")

    target_job = schemas.get("TargetJob") or {}
    target_properties = target_job.get("properties") or {}
    target_required = target_job.get("required") or []
    for removed_property in ("sourceType", "sourceUrl"):
        if removed_property in target_properties:
            errors.append(f"TargetJob must not expose {removed_property}")
        if removed_property in target_required:
            errors.append(f"TargetJob.required must not contain {removed_property}")

    purpose_enum = (
        (((schemas.get("UploadPresignRequest") or {}).get("properties") or {}).get("purpose") or {}).get("enum")
    )
    if purpose_enum != ["resume", "privacy_export"]:
        errors.append(
            "UploadPresignRequest.purpose must exactly preserve resume/privacy_export; "
            f"got {purpose_enum!r}"
        )

    upload_tag = next(
        (tag for tag in (data.get("tags") or []) if isinstance(tag, dict) and tag.get("name") == "Uploads"),
        {},
    )
    expected_upload_description = (
        "Pre-signed object-storage URLs for resume uploads and privacy-export artifacts."
    )
    if upload_tag.get("description") != expected_upload_description:
        errors.append(
            "Uploads tag description must describe only resume/privacy-export consumers; "
            f"got {upload_tag.get('description')!r}"
        )

    operation_invariants = (
        (
            "/targets/import",
            "post",
            "importTargetJob",
            "202",
            "#/components/schemas/TargetJobWithJob",
        ),
        (
            "/uploads/presign",
            "post",
            "createUploadPresign",
            "201",
            "#/components/schemas/UploadPresign",
        ),
    )
    paths = data.get("paths") or {}
    for path, method, operation_id, status, response_ref in operation_invariants:
        operation = ((paths.get(path) or {}).get(method) or {})
        if operation.get("operationId") != operation_id:
            errors.append(
                f"{method.upper()} {path} operationId must remain {operation_id!r}; "
                f"got {operation.get('operationId')!r}"
            )
        response = ((operation.get("responses") or {}).get(status) or {})
        actual_ref = (
            (((response.get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
        )
        if actual_ref != response_ref:
            errors.append(
                f"{method.upper()} {path} {status} response must remain {response_ref}; got {actual_ref!r}"
            )


def validate_targetjob_report_overview_contract(
    data: dict[str, Any], errors: list[str]
) -> None:
    """Lock OPENAPI-004's no-pagination canonical-round overview wire."""
    schemas = ((data.get("components") or {}).get("schemas") or {})
    operation = (
        (((data.get("paths") or {}).get("/targets/{targetJobId}/reports") or {}).get("get"))
        or {}
    )
    if operation.get("operationId") != "listTargetJobReports":
        errors.append(
            "GET /targets/{targetJobId}/reports operationId must remain listTargetJobReports"
        )

    named_parameters = [
        parameter.get("name")
        for parameter in (operation.get("parameters") or [])
        if isinstance(parameter, dict) and "name" in parameter
    ]
    if named_parameters != ["targetJobId"]:
        errors.append(
            "listTargetJobReports named parameters must exactly equal ['targetJobId']; "
            f"got {named_parameters!r}"
        )
    parameter_refs = [
        parameter.get("$ref")
        for parameter in (operation.get("parameters") or [])
        if isinstance(parameter, dict) and "$ref" in parameter
    ]
    expected_parameter_refs = [
        "#/components/parameters/XRequestID",
        "#/components/parameters/Traceparent",
        "#/components/parameters/AcceptLanguage",
        "#/components/parameters/XClientVersion",
    ]
    if parameter_refs != expected_parameter_refs:
        errors.append(
            "listTargetJobReports shared header refs must remain unchanged; "
            f"got {parameter_refs!r}"
        )

    response_ref = (
        (((((operation.get("responses") or {}).get("200") or {}).get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
    )
    expected_response_ref = "#/components/schemas/TargetJobReportsOverview"
    if response_ref != expected_response_ref:
        errors.append(
            "listTargetJobReports 200 response must reference "
            f"{expected_response_ref}; got {response_ref!r}"
        )

    if "PaginatedFeedbackReport" in schemas:
        errors.append("removed PaginatedFeedbackReport schema must not be present")
    target_job_properties = ((schemas.get("TargetJob") or {}).get("properties") or {})
    if "latestReportId" in target_job_properties:
        errors.append("TargetJob must not expose latestReportId")
    if (schemas.get("PracticeRoundRef") or {}).get("additionalProperties") is not False:
        errors.append(
            "PracticeRoundRef must set additionalProperties=false for the closed overview round"
        )

    expected_shapes: dict[str, tuple[list[str], list[str]]] = {
        "TargetJobReportsOverview": (
            ["targetJobId", "rounds"],
            ["targetJobId", "rounds"],
        ),
        "TargetJobReportRoundOverview": (
            ["round", "currentReport", "latestAttempt"],
            ["round", "currentReport", "latestAttempt"],
        ),
        "TargetJobCurrentReportSummary": (
            ["id", "generatedAt"],
            ["id", "generatedAt"],
        ),
        "TargetJobReportAttemptSummary": (
            ["id", "status", "errorCode", "createdAt"],
            ["id", "status", "errorCode", "createdAt"],
        ),
    }
    for schema_name, (required, properties) in expected_shapes.items():
        schema = schemas.get(schema_name)
        if not isinstance(schema, dict):
            errors.append(f"missing components.schemas.{schema_name}")
            continue
        if schema.get("type") != "object":
            errors.append(f"{schema_name}.type must be object")
        if schema.get("additionalProperties") is not False:
            errors.append(f"{schema_name} must set additionalProperties=false")
        if schema.get("required") != required:
            errors.append(
                f"{schema_name}.required must exactly equal {required}; got {schema.get('required')!r}"
            )
        actual_properties = list((schema.get("properties") or {}).keys())
        if actual_properties != properties:
            errors.append(
                f"{schema_name}.properties must exactly equal {properties}; got {actual_properties!r}"
            )

    round_overview = schemas.get("TargetJobReportRoundOverview") or {}
    round_properties = round_overview.get("properties") or {}
    if (round_properties.get("round") or {}).get("$ref") != "#/components/schemas/PracticeRoundRef":
        errors.append("TargetJobReportRoundOverview.round must reference PracticeRoundRef")
    overview_rounds = (
        ((schemas.get("TargetJobReportsOverview") or {}).get("properties") or {}).get("rounds")
        or {}
    )
    if overview_rounds.get("minItems") != 2 or overview_rounds.get("maxItems") != 5:
        errors.append(
            "TargetJobReportsOverview.rounds must preserve the canonical 2..5 catalog bounds"
        )
    expected_nullable_refs = {
        "currentReport": "#/components/schemas/TargetJobCurrentReportSummary",
        "latestAttempt": "#/components/schemas/TargetJobReportAttemptSummary",
    }
    for field_name, expected_ref in expected_nullable_refs.items():
        actual = (round_properties.get(field_name) or {}).get("oneOf")
        expected = [{"$ref": expected_ref}, {"type": "null"}]
        if actual != expected:
            errors.append(
                f"TargetJobReportRoundOverview.{field_name} must be required explicit nullable {expected_ref}"
            )

    attempt = schemas.get("TargetJobReportAttemptSummary") or {}
    attempt_properties = attempt.get("properties") or {}
    if (attempt_properties.get("status") or {}).get("$ref") != "#/components/schemas/ReportStatus":
        errors.append("TargetJobReportAttemptSummary.status must reference ReportStatus")
    expected_error_code = [
        {"$ref": "#/components/schemas/ApiErrorCode"},
        {"type": "null"},
    ]
    if (attempt_properties.get("errorCode") or {}).get("oneOf") != expected_error_code:
        errors.append(
            "TargetJobReportAttemptSummary.errorCode must be required explicit nullable ApiErrorCode"
        )
    conditionals = attempt.get("allOf") or []
    if len(conditionals) != 1:
        errors.append(
            "TargetJobReportAttemptSummary must define exactly one failed-only errorCode conditional"
        )
    else:
        conditional = conditionals[0]
        failed_const = (
            ((((conditional.get("if") or {}).get("properties") or {}).get("status") or {}).get("const"))
        )
        then_ref = (
            ((((conditional.get("then") or {}).get("properties") or {}).get("errorCode") or {}).get("$ref"))
        )
        else_type = (
            ((((conditional.get("else") or {}).get("properties") or {}).get("errorCode") or {}).get("type"))
        )
        if (
            failed_const != "failed"
            or then_ref != "#/components/schemas/ApiErrorCode"
            or else_type != "null"
        ):
            errors.append(
                "TargetJobReportAttemptSummary.errorCode must be non-null only when status=failed"
            )


def validate_report_conversation_contract(data: dict[str, Any], errors: list[str]) -> None:
    """Lock OPENAPI-001's report-owned, locator-free conversation projection."""
    paths = data.get("paths") or {}
    practice_sessions = paths.get("/practice/sessions") or {}
    if "get" in practice_sessions:
        errors.append("public GET /practice/sessions must be removed")

    operation = ((paths.get("/reports/{reportId}/conversation") or {}).get("get")) or {}
    if operation.get("operationId") != "getReportConversation":
        errors.append(
            "GET /reports/{reportId}/conversation operationId must be getReportConversation"
        )
    if operation.get("tags") != ["Reports"]:
        errors.append("getReportConversation must be owned by the Reports tag")

    named_parameters = [
        parameter.get("name")
        for parameter in (operation.get("parameters") or [])
        if isinstance(parameter, dict) and "name" in parameter
    ]
    if named_parameters != ["reportId"]:
        errors.append(
            "getReportConversation named parameters must exactly equal ['reportId']; "
            f"got {named_parameters!r}"
        )
    parameter_refs = [
        parameter.get("$ref")
        for parameter in (operation.get("parameters") or [])
        if isinstance(parameter, dict) and "$ref" in parameter
    ]
    expected_parameter_refs = [
        "#/components/parameters/XRequestID",
        "#/components/parameters/Traceparent",
        "#/components/parameters/AcceptLanguage",
        "#/components/parameters/XClientVersion",
    ]
    if parameter_refs != expected_parameter_refs:
        errors.append(
            "getReportConversation shared header refs must remain unchanged; "
            f"got {parameter_refs!r}"
        )
    response_ref = (
        (((((operation.get("responses") or {}).get("200") or {}).get("content") or {}).get("application/json") or {}).get("schema") or {}).get("$ref")
    )
    if response_ref != "#/components/schemas/ReportConversation":
        errors.append(
            "getReportConversation 200 response must reference ReportConversation; "
            f"got {response_ref!r}"
        )

    schemas = ((data.get("components") or {}).get("schemas") or {})
    if "PaginatedPracticeSession" in schemas:
        errors.append("removed PaginatedPracticeSession schema must not be present")
    expected_shapes: dict[str, tuple[list[str], list[str]]] = {
        "ReportConversation": (
            ["reportId", "reportStatus", "context", "messages"],
            ["reportId", "reportStatus", "context", "messages"],
        ),
        "ReportConversationMessage": (
            ["sequence", "role", "content", "createdAt"],
            ["sequence", "role", "content", "createdAt"],
        ),
    }
    for schema_name, (required, properties) in expected_shapes.items():
        schema = schemas.get(schema_name)
        if not isinstance(schema, dict):
            errors.append(f"missing components.schemas.{schema_name}")
            continue
        if schema.get("type") != "object":
            errors.append(f"{schema_name}.type must be object")
        if schema.get("additionalProperties") is not False:
            errors.append(f"{schema_name} must set additionalProperties=false")
        if schema.get("required") != required:
            errors.append(
                f"{schema_name}.required must exactly equal {required}; got {schema.get('required')!r}"
            )
        actual_properties = list((schema.get("properties") or {}).keys())
        if actual_properties != properties:
            errors.append(
                f"{schema_name}.properties must exactly equal {properties}; got {actual_properties!r}"
            )

    conversation_properties = _schema_properties(schemas, "ReportConversation")
    if (conversation_properties.get("reportId") or {}).get("format") != "uuid":
        errors.append("ReportConversation.reportId must be a UUID")
    if (conversation_properties.get("reportStatus") or {}).get("$ref") != "#/components/schemas/ReportStatus":
        errors.append("ReportConversation.reportStatus must reference ReportStatus")
    if (conversation_properties.get("context") or {}).get("$ref") != "#/components/schemas/ReportContextSnapshot":
        errors.append("ReportConversation.context must reference ReportContextSnapshot")
    if (
        ((conversation_properties.get("messages") or {}).get("items") or {}).get("$ref")
        != "#/components/schemas/ReportConversationMessage"
    ):
        errors.append("ReportConversation.messages must reference ReportConversationMessage")

    message_properties = _schema_properties(schemas, "ReportConversationMessage")
    sequence = message_properties.get("sequence") or {}
    if (
        sequence.get("type") != "integer"
        or sequence.get("format") != "int32"
        or sequence.get("minimum") != 1
    ):
        errors.append("ReportConversationMessage.sequence must be a positive int32")
    role = message_properties.get("role") or {}
    if role.get("type") != "string" or role.get("enum") != ["user", "assistant"]:
        errors.append("ReportConversationMessage.role must exactly be user|assistant")
    content = message_properties.get("content") or {}
    if content != {"type": "string", "minLength": 1, "pattern": r"\S"}:
        errors.append(
            "ReportConversationMessage.content must be a nonblank string with exact minLength/pattern"
        )
    created_at = message_properties.get("createdAt") or {}
    if created_at.get("type") != "string" or created_at.get("format") != "date-time":
        errors.append("ReportConversationMessage.createdAt must be a date-time string")
    for forbidden_property in (
        "sessionId",
        "id",
        "clientMessageId",
        "replyStatus",
        "replyGeneration",
        "anchor",
    ):
        if forbidden_property in message_properties:
            errors.append(
                "ReportConversationMessage must not expose internal locator "
                f"{forbidden_property!r}"
            )


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
    for prop in (
        "summary",
        "context",
        "dimensionAssessments",
        "retryFocusDimensionCodes",
        "provenance",
    ):
        if prop not in feedback_props:
            errors.append(f"FeedbackReport missing current product property {prop!r}")
    for prop in (
        "questionAssessments",
        "retryFocusTurnIds",
        "retryFocusCompetencyCodes",
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
    validate_targetjob_paste_only_contract(data, errors)
    validate_targetjob_report_overview_contract(data, errors)
    validate_report_conversation_contract(data, errors)
    validate_resume_summary_contract(data, errors)

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
