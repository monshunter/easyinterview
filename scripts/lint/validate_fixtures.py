#!/usr/bin/env python3
"""Validate openapi/fixtures/*.json against openapi/openapi.yaml.

Phase 1.3 scope (per `002-fixtures-and-mock-source` plan §3 / spec C-6 / C-11):
    1. structural — operationId matches filename, `scenarios.default` exists
       and is the first scenario, status code is declared on the operation.
    2. schema    — request.body and response.body schema-valid against the
       operation's request / status-matched response schema in openapi.yaml.
    3. provenance — every AI-generated schema listed in spec §4.6 carries a
       complete `GenerationProvenance` (6 non-empty fields) and uses a
       provider-neutral model id in fixture data.
    4. privacy   — emails restricted to example.{com,org,net} or `.example`,
       phones to `+1-555-01xx`, and the employer-brand blacklist below.
    5. ids       — `format: uuid` values must match UUIDv7 layout, and any
       string with `tmp_` prefix is rejected.
    6. coverage  — every operationId currently exposed by openapi.yaml must
       have a fixture file.
    7. D-20 out-of-scope resume contract keys — flat resume fixtures must not
       reintroduce resumeAssetId / resumeVersionId request or response fields.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from collections import OrderedDict
from pathlib import Path
from typing import Any, Iterable, Iterator, List, Tuple

import yaml


# ---------- shared regexes / constants ---------------------------------------

UUID_SHAPE_RE = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
)
UUID_V7_RE = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)
ISO_DATETIME_RE = re.compile(
    r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$"
)
EMAIL_RE = re.compile(r"\b[A-Za-z0-9._%+-]+@([A-Za-z0-9.-]+\.[A-Za-z]{2,})\b")
PHONE_RE = re.compile(r"\+\d[\d\-\s()]{7,}\d")
TEMP_ID_RE = re.compile(r"\btmp_[A-Za-z0-9_-]+\b")

ALLOWED_EMAIL_DOMAINS = {"example.com", "example.org", "example.net"}
ALLOWED_PHONE_PREFIX = "+1-555-01"

# Real employer-style brands.
COMPANY_BLACKLIST = (
    "alibaba", "tencent", "bytedance", "baidu", "meituan", "didi", "huawei",
    "字节", "腾讯", "阿里巴巴", "百度", "美团", "滴滴", "华为", "星环",
)
COMPANY_BLACKLIST_RE = re.compile(
    r"(?:^|[^A-Za-z0-9_])(" + "|".join(re.escape(b) for b in COMPANY_BLACKLIST) +
    r")(?:[^A-Za-z0-9_]|$)",
    re.IGNORECASE,
)
PROVIDER_NEUTRAL_MODEL_ID_RE = re.compile(r"^(?:model-profile|fixture-model):[a-z][a-z0-9_.-]*$")
VENDOR_MODEL_TOKEN_RE = re.compile(
    r"(?:openrouter|anthropic|claude|openai|gpt-|mistral|gemini|cohere)",
    re.IGNORECASE,
)
PRACTICE_VOICE_PLAYABLE_AUDIO_REF_RE = re.compile(
    r"^data:audio/(?:mpeg|wav|ogg);base64,[A-Za-z0-9+/=]+$",
    re.IGNORECASE,
)
PRACTICE_VOICE_RESOLVER_AUDIO_REF_RE = re.compile(
    r"^/api/v1/practice/voice-turns/[^/]+/chunks/[^/]+/audio$"
)
SHA256_DIGEST_RE = re.compile(r"^(?:sha256:)?[0-9a-f]{64}$")

# AI-generated schemas listed in spec §4.6. Provenance must resolve from each
# value path, where `[*]` expands a list and yields its members.
AI_PROVENANCE_PATHS: dict[str, Tuple[str, ...]] = {
    "getTargetJob": ("summary.provenance", "fitSummary.provenance"),
    "listTargetJobs": ("items[*].summary.provenance", "items[*].fitSummary.provenance"),
    "updateTargetJob": ("summary.provenance", "fitSummary.provenance"),
    "getFeedbackReport": ("provenance",),
    "getResumeTailorRun": ("provenance",),
    "getResume": ("structuredProfile.provenance",),
    "updateResume": ("structuredProfile.provenance",),
    "duplicateResume": ("structuredProfile.provenance",),
}
PROVENANCE_REQUIRED_FIELDS = (
    "promptVersion",
    "rubricVersion",
    "modelId",
    "language",
    "featureFlag",
    "dataSourceVersion",
)
P0_EXPORT_ERROR_CODES: dict[str, str] = {
    "requestPrivacyExport": "PRIVACY_EXPORT_NOT_AVAILABLE",
    "exportResume": "RESUME_EXPORT_NOT_AVAILABLE",
}
REQUIRED_NAMED_SCENARIOS: dict[str, frozenset[str]] = {
    "importTargetJob": frozenset(
        {"default", "paste-primary", "validation-blank-raw-text"}
    ),
    "createUploadPresign": frozenset({"default", "privacy-export"}),
    "createPracticePlan": frozenset(
        {"default", "retry-derived", "next-derived", "round-mismatch"}
    ),
    "getFeedbackReport": frozenset(
        {
            "default",
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
    ),
    "getPracticePlan": frozenset({"default", "legacy-null-round-identity"}),
    "getTargetJob": frozenset({"default", "not-started-progress", "all-completed-progress"}),
    "listTargetJobs": frozenset({"default", "not-started-progress", "all-completed-progress"}),
    "completePracticeSession": frozenset(
        {
            "default",
            "replay",
            "mismatch",
            "session-already-completed",
            "cross-user-not-found",
        }
    ),
    "createPracticeVoiceTurn": frozenset({"default"}),
    "getPracticeSession": frozenset(
        {
            "default",
            "reply-pending",
            "reply-retryable-failed",
            "reply-terminal-failed",
            "reply-complete",
        }
    ),
    "sendPracticeMessage": frozenset(
        {
            "default",
            "validation-empty-text",
            "auth-unauthorized",
            "session-not-found",
            "reply-pending-conflict",
            "client-message-mismatch",
            "ai-timeout-retryable",
            "retry-success-same-client-message",
        }
    ),
    "listTargetJobReports": frozenset(
        {
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
        }
    ),
}
OUT_OF_SCOPE_D20_FIXTURE_KEYS = frozenset({"resumeAssetId", "resumeVersionId"})
CANONICAL_NEGATIVE_REQUEST_SCHEMA_FAILURES = {
    ("importTargetJob", "validation-blank-raw-text"): "rawText",
}
TARGET_JOB_READ_OPERATION_IDS = frozenset(
    {"listTargetJobs", "getTargetJob", "updateTargetJob", "archiveTargetJob"}
)
REMOVED_TARGET_JOB_FIELDS = frozenset(
    {"sourceType", "sourceUrl", "latestReportId"}
)
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
REPORT_OVERVIEW_ERROR_MATRIX = {
    "cross-user-not-found": (404, "TARGET_JOB_NOT_FOUND"),
    "target-not-found": (404, "TARGET_JOB_NOT_FOUND"),
    "invalid-frozen-context": (500, "AI_OUTPUT_INVALID"),
    "missing-frozen-context": (500, "AI_OUTPUT_INVALID"),
}
REPORT_OVERVIEW_FORBIDDEN_BODY_KEYS = frozenset(
    {
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
)
PRACTICE_REPLY_STATUSES = frozenset(
    {"pending", "retryable_failed", "terminal_failed", "complete"}
)
PRACTICE_SEND_SCENARIO_ORDER = (
    "default",
    "validation-empty-text",
    "auth-unauthorized",
    "session-not-found",
    "reply-pending-conflict",
    "client-message-mismatch",
    "ai-timeout-retryable",
    "retry-success-same-client-message",
)
PRACTICE_SEND_FAILURE_MATRIX = {
    "validation-empty-text": (
        422,
        "VALIDATION_FAILED",
        False,
        {"field": "text", "reservation": "not_created"},
    ),
    "auth-unauthorized": (
        401,
        "AUTH_UNAUTHORIZED",
        False,
        {"reservation": "not_created"},
    ),
    "session-not-found": (
        404,
        "PRACTICE_SESSION_NOT_FOUND",
        False,
        {"reservation": "not_created"},
    ),
    "reply-pending-conflict": (
        409,
        "PRACTICE_SESSION_CONFLICT",
        False,
        {"reservation": "unchanged", "replyStatus": "pending"},
    ),
    "client-message-mismatch": (
        409,
        "IDEMPOTENCY_KEY_MISMATCH",
        False,
        {"field": "clientMessageId", "reservation": "unchanged"},
    ),
    "ai-timeout-retryable": (
        502,
        "AI_PROVIDER_TIMEOUT",
        True,
        {
            "clientMessageId": "01918fa0-0000-7000-8000-000000007010",
            "reservation": "persisted",
            "replyStatus": "retryable_failed",
        },
    ),
}


# ---------- helpers -----------------------------------------------------------

def walk_fixtures(fixtures_root: Path) -> List[Tuple[str, str, Path, dict]]:
    """Walk `<fixtures_root>/<tag>/<operationId>.json` entries."""
    if not fixtures_root.is_dir():
        return []
    entries: List[Tuple[str, str, Path, dict]] = []
    for tag_dir in sorted(p for p in fixtures_root.iterdir() if p.is_dir()):
        for fixture_path in sorted(p for p in tag_dir.iterdir() if p.suffix == ".json"):
            with fixture_path.open("r", encoding="utf-8") as f:
                data = json.load(f, object_pairs_hook=OrderedDict)
            entries.append((tag_dir.name, fixture_path.stem, fixture_path, data))
    return entries


def _walk_strings(data: Any, prefix: str = "") -> Iterator[Tuple[str, str]]:
    if isinstance(data, str):
        yield prefix, data
    elif isinstance(data, dict):
        for k, v in data.items():
            yield from _walk_strings(v, f"{prefix}.{k}" if prefix else k)
    elif isinstance(data, list):
        for i, v in enumerate(data):
            yield from _walk_strings(v, f"{prefix}[{i}]")


def _walk_keys(data: Any, prefix: str = "") -> Iterator[Tuple[str, str]]:
    if isinstance(data, dict):
        for k, v in data.items():
            path = f"{prefix}.{k}" if prefix else k
            yield path, k
            yield from _walk_keys(v, path)
    elif isinstance(data, list):
        for i, v in enumerate(data):
            yield from _walk_keys(v, f"{prefix}[{i}]")


def _resolve_path(data: Any, path: str) -> List[Any]:
    parts = re.findall(r"[^.\[\]*]+|\[\*\]", path)
    cursor: List[Any] = [data]
    for part in parts:
        next_cursor: List[Any] = []
        for c in cursor:
            if part == "[*]":
                if isinstance(c, list):
                    next_cursor.extend(c)
            elif isinstance(c, dict) and part in c:
                next_cursor.append(c[part])
        cursor = next_cursor
    return cursor


# ---------- OpenAPI loading + minimal schema validator -----------------------

def load_openapi(openapi_path: Path) -> dict:
    with openapi_path.open("r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def build_operation_index(spec: dict) -> dict[str, dict]:
    """opId → {tag, method, path, operation}."""
    index: dict[str, dict] = {}
    for path, methods in (spec.get("paths") or {}).items():
        if not isinstance(methods, dict):
            continue
        for method, op in methods.items():
            if not isinstance(op, dict) or "operationId" not in op:
                continue
            opid = op["operationId"]
            tag = (op.get("tags") or [None])[0]
            index[opid] = {
                "tag": tag,
                "method": method.lower(),
                "path": path,
                "operation": op,
            }
    return index


def expected_fixture_operations(spec: dict) -> Tuple[Tuple[str, str], ...]:
    """Return `(tag, operationId)` rows derived from the live OpenAPI document.

    `make lint-openapi` owns the frozen/additive operation inventory. The fixture
    validator only checks that whatever operation inventory OpenAPI currently
    exposes has exactly one matching fixture, so this gate cannot become a
    second hand-maintained operation list.
    """
    rows: list[Tuple[str, str]] = []
    for operation_id, meta in build_operation_index(spec).items():
        tag = meta.get("tag")
        if isinstance(tag, str):
            rows.append((tag, operation_id))
    return tuple(sorted(rows, key=lambda row: (row[0], row[1])))


def _resolve_ref(ref: str, root: dict) -> dict:
    parts = ref.lstrip("#/").split("/")
    cursor: Any = root
    for p in parts:
        cursor = cursor[p]
    return cursor


def schema_validate(value: Any, schema: dict | None, *, root: dict, path: str,
                    errors: List[str]) -> None:
    """Lightweight OpenAPI 3.1 schema validator covering the constructs used
    by openapi.yaml. Reports human-readable error strings into `errors`."""
    if schema is None:
        return
    if "$ref" in schema:
        schema_validate(value, _resolve_ref(schema["$ref"], root),
                        root=root, path=path, errors=errors)
        return

    # JSON Schema permits a bare `const` without an adjacent `type`. This is
    # used by conditional predicates and discriminated request branches, so it
    # must be evaluated before type-specific handling.
    if "const" in schema and value != schema["const"]:
        errors.append(f"{path}: value {value!r} != const {schema['const']!r}")
        return

    # 3.1 supports type as list. Normalize.
    types_field = schema.get("type")
    if isinstance(types_field, list):
        for t in types_field:
            tmp: List[str] = []
            sub = dict(schema)
            sub["type"] = t
            schema_validate(value, sub, root=root, path=path, errors=tmp)
            if not tmp:
                return
        errors.append(f"{path}: matched none of type {types_field}")
        return

    if "if" in schema:
        predicate_errors: List[str] = []
        schema_validate(
            value,
            schema["if"],
            root=root,
            path=path,
            errors=predicate_errors,
        )
        branch = schema.get("then") if not predicate_errors else schema.get("else")
        if branch is not None:
            schema_validate(value, branch, root=root, path=path, errors=errors)

    if "not" in schema:
        forbidden_errors: List[str] = []
        schema_validate(
            value,
            schema["not"],
            root=root,
            path=path,
            errors=forbidden_errors,
        )
        if not forbidden_errors:
            errors.append(f"{path}: matched forbidden schema")

    if "oneOf" in schema:
        matched = 0
        last_branch_errors: List[List[str]] = []
        for sub in schema["oneOf"]:
            tmp: List[str] = []
            schema_validate(value, sub, root=root, path=path, errors=tmp)
            if not tmp:
                matched += 1
            else:
                last_branch_errors.append(tmp)
        if matched != 1:
            errors.append(
                f"{path}: oneOf expected exactly one match, matched {matched} "
                f"(branch errors: {last_branch_errors[:1]})"
            )
        return
    if "allOf" in schema:
        for sub in schema["allOf"]:
            schema_validate(value, sub, root=root, path=path, errors=errors)
        # fall through to inspect inline fields
    if "anyOf" in schema:
        for sub in schema["anyOf"]:
            tmp: List[str] = []
            schema_validate(value, sub, root=root, path=path, errors=tmp)
            if not tmp:
                return
        errors.append(f"{path}: anyOf matched no branch")
        return

    t = types_field

    # Nullable handling.
    if value is None:
        if t == "null" or schema.get("nullable") is True:
            return
        if t in (None, "object") and not schema.get("required"):
            return
        errors.append(f"{path}: null not allowed (type={t!r})")
        return

    # Object.
    if t == "object" or "properties" in schema or "required" in schema:
        if not isinstance(value, dict):
            errors.append(f"{path}: expected object, got {type(value).__name__}")
            return
        for req in schema.get("required", []):
            if req not in value:
                errors.append(f"{path}: missing required field {req!r}")
        for trigger, dependencies in (schema.get("dependentRequired") or {}).items():
            if trigger not in value:
                continue
            for dependency in dependencies:
                if dependency not in value:
                    errors.append(
                        f"{path}: dependentRequired {trigger!r} requires {dependency!r}"
                    )
        props = schema.get("properties", {}) or {}
        addl = schema.get("additionalProperties", True)
        for k, v in value.items():
            sub_schema = props.get(k)
            if sub_schema is not None:
                schema_validate(v, sub_schema, root=root,
                                path=f"{path}.{k}" if path else k, errors=errors)
            elif addl is False:
                errors.append(f"{path}: unexpected property {k!r}")
            elif isinstance(addl, dict):
                schema_validate(v, addl, root=root,
                                path=f"{path}.{k}" if path else k, errors=errors)
        return

    if t == "array":
        if not isinstance(value, list):
            errors.append(f"{path}: expected array, got {type(value).__name__}")
            return
        if "minItems" in schema and len(value) < schema["minItems"]:
            errors.append(f"{path}: {len(value)} items < minItems {schema['minItems']}")
        if "maxItems" in schema and len(value) > schema["maxItems"]:
            errors.append(f"{path}: {len(value)} items > maxItems {schema['maxItems']}")
        items_schema = schema.get("items")
        for i, item in enumerate(value):
            schema_validate(item, items_schema, root=root,
                            path=f"{path}[{i}]", errors=errors)
        if schema.get("uniqueItems") is True:
            seen: set[str] = set()
            for item in value:
                encoded = json.dumps(item, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
                if encoded in seen:
                    errors.append(f"{path}: uniqueItems requires distinct array entries")
                    break
                seen.add(encoded)
        return

    if t == "string":
        if not isinstance(value, str):
            errors.append(f"{path}: expected string, got {type(value).__name__}")
            return
        if "minLength" in schema and len(value) < schema["minLength"]:
            errors.append(f"{path}: length {len(value)} < minLength {schema['minLength']}")
        if "maxLength" in schema and len(value) > schema["maxLength"]:
            errors.append(f"{path}: length {len(value)} > maxLength {schema['maxLength']}")
        if "enum" in schema and value not in schema["enum"]:
            errors.append(f"{path}: value {value!r} not in enum {schema['enum']}")
        pattern = schema.get("pattern")
        if isinstance(pattern, str) and re.search(pattern, value) is None:
            errors.append(f"{path}: value {value!r} does not match pattern {pattern!r}")
        fmt = schema.get("format")
        if fmt == "uuid" and not UUID_SHAPE_RE.match(value):
            errors.append(f"{path}: invalid uuid format ({value!r})")
        elif fmt == "date-time" and not ISO_DATETIME_RE.match(value):
            errors.append(f"{path}: invalid date-time ({value!r})")
        elif fmt == "date" and not re.match(r"^\d{4}-\d{2}-\d{2}$", value):
            errors.append(f"{path}: invalid date ({value!r})")
        elif fmt == "uri" and "://" not in value:
            errors.append(f"{path}: invalid uri ({value!r})")
        elif fmt == "email" and "@" not in value:
            errors.append(f"{path}: invalid email ({value!r})")
        return

    if t == "integer":
        if isinstance(value, bool) or not isinstance(value, int):
            errors.append(f"{path}: expected integer, got {type(value).__name__}")
            return
        if "minimum" in schema and value < schema["minimum"]:
            errors.append(f"{path}: {value} < minimum {schema['minimum']}")
        if "maximum" in schema and value > schema["maximum"]:
            errors.append(f"{path}: {value} > maximum {schema['maximum']}")
        if "enum" in schema and value not in schema["enum"]:
            errors.append(f"{path}: integer {value!r} not in enum {schema['enum']}")
        return

    if t == "number":
        if isinstance(value, bool) or not isinstance(value, (int, float)):
            errors.append(f"{path}: expected number, got {type(value).__name__}")
        return

    if t == "boolean":
        if not isinstance(value, bool):
            errors.append(f"{path}: expected boolean, got {type(value).__name__}")
        return

    if t == "null":
        if value is not None:
            errors.append(f"{path}: expected null, got {type(value).__name__}")
        return

    # Unknown type or pure constraint schema — skip silently.


# ---------- per-fixture checks -----------------------------------------------

def _select_response_schema(op: dict, status: int) -> Tuple[dict | None, str | None]:
    """Pick the response schema for a given status code.

    Returns (schema, status-key-used). Falls back to `default` for declared
    error responses. If no match is possible, schema is None.
    """
    responses = op.get("responses", {}) or {}
    # Exact match (str key in OpenAPI).
    if str(status) in responses:
        body_schema = _content_schema(responses[str(status)])
        return body_schema, str(status)
    if "default" in responses:
        body_schema = _content_schema(responses["default"])
        return body_schema, "default"
    return None, None


def _content_schema(response: Any) -> dict | None:
    if not isinstance(response, dict):
        return None
    if "$ref" in response:
        return None  # tests can't easily resolve component-level response refs
    content = response.get("content") or {}
    json_block = content.get("application/json") or {}
    return json_block.get("schema")


def _request_schema(op: dict) -> dict | None:
    rb = op.get("requestBody") or {}
    return _content_schema(rb)


def check_structural(
    opid: str, data: dict, fixture_path: Path, op: dict, errors: List[str]
) -> None:
    if data.get("operationId") != opid:
        errors.append(
            f"{opid}: operationId field {data.get('operationId')!r} != filename {fixture_path.name!r}"
        )
    scenarios = data.get("scenarios")
    if not isinstance(scenarios, dict) or not scenarios:
        errors.append(f"{opid}: scenarios object missing or empty")
        return
    if next(iter(scenarios)) != "default":
        errors.append(f"{opid}: first scenario must be 'default'")
    if "default" not in scenarios:
        errors.append(f"{opid}: scenarios.default required")


def check_status_declared(
    opid: str, op: dict, scenario_name: str, scenario: dict, errors: List[str]
) -> None:
    response = scenario.get("response") or {}
    status = response.get("status")
    if not isinstance(status, int):
        errors.append(f"{opid}.{scenario_name}: response.status must be integer")
        return
    declared = set(op.get("responses", {}).keys())
    if str(status) not in declared and "default" not in declared:
        errors.append(
            f"{opid}.{scenario_name}: status {status} not declared on operation "
            f"(declared: {sorted(declared)})"
        )
    elif str(status) not in declared and not (400 <= status < 600):
        # default response only covers error space.
        errors.append(
            f"{opid}.{scenario_name}: status {status} not declared, "
            f"and `default` only covers 4xx/5xx — saw 2xx/3xx"
        )


def check_schema(
    opid: str, op: dict, scenario_name: str, scenario: dict, root: dict,
    errors: List[str],
) -> None:
    response = scenario.get("response") or {}
    status = response.get("status")
    if isinstance(status, int):
        body_schema, _ = _select_response_schema(op, status)
        if body_schema is not None and "body" in response:
            schema_validate(
                response.get("body"), body_schema, root=root,
                path=f"{opid}.{scenario_name}.response.body", errors=errors,
            )
    if "request" in scenario:
        body_schema = _request_schema(op)
        if body_schema is not None and "body" in scenario["request"]:
            request_errors: List[str] = []
            schema_validate(
                scenario["request"]["body"], body_schema, root=root,
                path=f"{opid}.{scenario_name}.request.body", errors=request_errors,
            )
            expected_field = CANONICAL_NEGATIVE_REQUEST_SCHEMA_FAILURES.get(
                (opid, scenario_name)
            )
            if expected_field is None:
                errors.extend(request_errors)
                return
            expected_path = (
                f"{opid}.{scenario_name}.request.body.{expected_field}:"
            )
            if not (
                len(request_errors) == 1
                and request_errors[0].startswith(expected_path)
                and "does not match pattern" in request_errors[0]
            ):
                errors.append(
                    f"{opid}.{scenario_name}.request.body: canonical negative request "
                    f"must fail exactly at /{expected_field}; saw {request_errors!r}"
                )


def check_provenance(opid: str, scenario: dict, errors: List[str]) -> None:
    paths = AI_PROVENANCE_PATHS.get(opid)
    if not paths:
        return
    body = (scenario.get("response") or {}).get("body")
    if body is None:
        errors.append(f"{opid}: missing response.body for provenance check")
        return
    if (
        opid == "getFeedbackReport"
        and body.get("status") is not None
        and body.get("status") != "ready"
    ):
        if body.get("provenance") is not None:
            errors.append(
                "getFeedbackReport.provenance: non-ready report provenance must be null"
            )
        return
    for path in paths:
        resolved = _resolve_path(body, path)
        if not resolved:
            errors.append(f"{opid}: provenance path {path} did not resolve")
            continue
        for prov in resolved:
            if not isinstance(prov, dict):
                errors.append(f"{opid}.{path}: provenance must be object")
                continue
            for field in PROVENANCE_REQUIRED_FIELDS:
                v = prov.get(field)
                if not isinstance(v, str) or not v.strip():
                    errors.append(
                        f"{opid}.{path}.{field}: provenance field missing or blank"
                    )
                    continue
                if field == "modelId":
                    if not PROVIDER_NEUTRAL_MODEL_ID_RE.match(v):
                        errors.append(
                            f"{opid}.{path}.modelId: fixture provenance must use "
                            f"a provider-neutral model id (`model-profile:<name>` or "
                            f"`fixture-model:<name>`), got {v!r}"
                        )
                    if VENDOR_MODEL_TOKEN_RE.search(v):
                        errors.append(
                            f"{opid}.{path}.modelId: fixture provenance must not hardcode "
                            f"vendor/model tokens, got {v!r}"
                        )


def check_p0_export_error_code(opid: str, scenarios: dict, errors: List[str]) -> None:
    """Spec D-12 / D-18: P0 export exceptions return their locked 501 error codes.

    Hand-written examples live in fixtures as the single source of truth.
    """
    expected_code = P0_EXPORT_ERROR_CODES.get(opid)
    if expected_code is None:
        return
    default = scenarios.get("default") or {}
    response = default.get("response") or {}
    if response.get("status") != 501:
        errors.append(f"{opid}: default.response.status must be 501 (spec D-12 / D-18)")
        return
    body = response.get("body") or {}
    code = (body.get("error") or {}).get("code")
    if code != expected_code:
        errors.append(
            f"{opid}: default.response.body.error.code must be {expected_code!r} (spec D-12 / D-18); got {code!r}"
        )


def check_required_named_scenarios(opid: str, scenarios: dict, errors: List[str]) -> None:
    required = REQUIRED_NAMED_SCENARIOS.get(opid)
    if required is None:
        return
    missing = required - set(scenarios)
    if missing:
        errors.append(f"{opid}: missing required scenarios {sorted(missing)}")


def check_target_job_paste_only_semantics(
    opid: str, scenarios: dict, errors: List[str]
) -> None:
    if opid == "importTargetJob":
        expected_scenarios = [
            "default",
            "paste-primary",
            "validation-blank-raw-text",
        ]
        if list(scenarios) != expected_scenarios:
            errors.append(
                "importTargetJob: scenarios must be exactly "
                f"{expected_scenarios!r}; got {list(scenarios)!r}"
            )
        for scenario_name in ("default", "paste-primary"):
            scenario = scenarios.get(scenario_name) or {}
            body = ((scenario.get("request") or {}).get("body") or {})
            path = f"{opid}.{scenario_name}.request.body"
            if set(body) != {"rawText", "targetLanguage", "resumeId"}:
                errors.append(
                    f"{path}: paste-only request must contain exactly "
                    "rawText, targetLanguage and resumeId"
                )
            raw_text = body.get("rawText")
            if not isinstance(raw_text, str) or not raw_text.strip():
                errors.append(f"{path}.rawText: positive fixture must be non-whitespace")

        negative = scenarios.get("validation-blank-raw-text") or {}
        request_body = ((negative.get("request") or {}).get("body") or {})
        response = negative.get("response") or {}
        error = ((response.get("body") or {}).get("error") or {})
        if set(request_body) != {"rawText", "targetLanguage", "resumeId"}:
            errors.append(
                f"{opid}.validation-blank-raw-text.request.body: must retain the exact "
                "flattened request shape"
            )
        raw_text = request_body.get("rawText")
        if not isinstance(raw_text, str) or not raw_text or raw_text.strip():
            errors.append(
                f"{opid}.validation-blank-raw-text.request.body.rawText: "
                "must be non-empty whitespace-only text"
            )
        if response.get("status") != 422:
            errors.append(
                f"{opid}.validation-blank-raw-text.response.status: expected 422"
            )
        if error.get("code") != "VALIDATION_FAILED":
            errors.append(
                f"{opid}.validation-blank-raw-text.response.body.error.code: "
                "expected 'VALIDATION_FAILED'"
            )
        if error.get("retryable") is not False:
            errors.append(
                f"{opid}.validation-blank-raw-text.response.body.error.retryable: "
                "expected false"
            )
        if (error.get("details") or {}).get("field") != "rawText":
            errors.append(
                f"{opid}.validation-blank-raw-text.response.body.error.details.field: "
                "expected 'rawText'"
            )
        return

    if opid == "createUploadPresign":
        expected_scenarios = ["default", "privacy-export"]
        if list(scenarios) != expected_scenarios:
            errors.append(
                "createUploadPresign: scenarios must be exactly "
                f"{expected_scenarios!r}; got {list(scenarios)!r}"
            )
        purposes = {
            ((scenario.get("request") or {}).get("body") or {}).get("purpose")
            for scenario in scenarios.values()
        }
        if purposes != {"resume", "privacy_export"}:
            errors.append(
                "createUploadPresign: positive fixture purposes must be exactly "
                f"resume/privacy_export; got {sorted(repr(value) for value in purposes)}"
            )
        return

    if opid not in TARGET_JOB_READ_OPERATION_IDS:
        return
    for scenario_name, scenario in scenarios.items():
        body = ((scenario.get("response") or {}).get("body"))
        for key_path, key in _walk_keys(body):
            if key in REMOVED_TARGET_JOB_FIELDS:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.{key_path}: "
                    f"removed TargetJob field {key!r} is forbidden"
                )


def check_target_job_reports_overview_semantics(
    scenarios: dict, errors: List[str]
) -> None:
    opid = "listTargetJobReports"
    if tuple(scenarios) != REPORT_OVERVIEW_SCENARIO_ORDER:
        errors.append(
            f"{opid}: scenarios must be exactly "
            f"{list(REPORT_OVERVIEW_SCENARIO_ORDER)!r}; got {list(scenarios)!r}"
        )

    for scenario_name, scenario in scenarios.items():
        response = scenario.get("response") or {}
        expected_error = REPORT_OVERVIEW_ERROR_MATRIX.get(scenario_name)
        if expected_error is not None:
            expected_status, expected_code = expected_error
            body = response.get("body") or {}
            error = body.get("error") or {}
            if response.get("status") != expected_status:
                errors.append(
                    f"{opid}.{scenario_name}.response.status: expected {expected_status}"
                )
            if set(body) != {"error"}:
                errors.append(
                    f"{opid}.{scenario_name}.response.body: typed ApiErrorResponse required"
                )
            if error.get("code") != expected_code:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.error.code: "
                    f"expected {expected_code!r}"
                )
            if error.get("retryable") is not False:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.error.retryable: expected false"
                )
            continue

        if response.get("status") != 200:
            errors.append(
                f"{opid}.{scenario_name}.response.status: overview scenario must return 200"
            )
            continue
        body = response.get("body")
        if not isinstance(body, dict):
            errors.append(f"{opid}.{scenario_name}.response.body: object required")
            continue
        body_path = f"{opid}.{scenario_name}.response.body"
        if set(body) != {"targetJobId", "rounds"}:
            errors.append(
                f"{body_path}: closed overview requires exactly targetJobId and rounds"
            )
        for key_path, key in _walk_keys(body):
            if key in REPORT_OVERVIEW_FORBIDDEN_BODY_KEYS:
                errors.append(f"{body_path}.{key_path}: forbidden overview field {key!r}")

        rounds = body.get("rounds")
        if not isinstance(rounds, list) or not 2 <= len(rounds) <= 5:
            errors.append(f"{body_path}.rounds: requires 2 to 5 canonical rounds")
            continue
        seen_rounds: set[tuple[Any, Any]] = set()
        previous_sequence = 0
        for index, item in enumerate(rounds):
            item_path = f"{body_path}.rounds[{index}]"
            if not isinstance(item, dict):
                errors.append(f"{item_path}: round overview object required")
                continue
            if set(item) != {"round", "currentReport", "latestAttempt"}:
                errors.append(
                    f"{item_path}: requires exactly round/currentReport/latestAttempt; "
                    "nullable fields must be explicit"
                )
            round_ref = item.get("round")
            if not isinstance(round_ref, dict) or set(round_ref) != {
                "roundId",
                "roundSequence",
            }:
                errors.append(f"{item_path}.round: closed PracticeRoundRef required")
                continue
            round_id = round_ref.get("roundId")
            sequence = round_ref.get("roundSequence")
            identity = (round_id, sequence)
            if identity in seen_rounds:
                errors.append(f"{item_path}.round: duplicate canonical round identity")
            seen_rounds.add(identity)
            if not isinstance(sequence, int) or sequence <= previous_sequence:
                errors.append(f"{item_path}.round.roundSequence: rounds must be ordered")
            elif not isinstance(round_id, str) or not round_id.startswith(
                f"round-{sequence}-"
            ):
                errors.append(
                    f"{item_path}.round: roundId sequence must equal roundSequence"
                )
            previous_sequence = sequence if isinstance(sequence, int) else previous_sequence

            current = item.get("currentReport")
            if current is not None and (
                not isinstance(current, dict) or set(current) != {"id", "generatedAt"}
            ):
                errors.append(f"{item_path}.currentReport: closed id/generatedAt summary required")
            latest = item.get("latestAttempt")
            if latest is None:
                continue
            if not isinstance(latest, dict) or set(latest) != {
                "id",
                "status",
                "errorCode",
                "createdAt",
            }:
                errors.append(
                    f"{item_path}.latestAttempt: closed id/status/errorCode/createdAt summary required"
                )
                continue
            status = latest.get("status")
            error_code = latest.get("errorCode")
            if status == "failed":
                if not isinstance(error_code, str) or not error_code:
                    errors.append(
                        f"{item_path}.latestAttempt.errorCode: failed attempt requires a code"
                    )
            elif error_code is not None:
                errors.append(
                    f"{item_path}.latestAttempt.errorCode: non-failed attempt requires null"
                )

    positive = {
        name: ((scenarios.get(name) or {}).get("response") or {}).get("body") or {}
        for name in REPORT_OVERVIEW_SCENARIO_ORDER
        if name not in REPORT_OVERVIEW_ERROR_MATRIX
    }
    if positive.get("default", {}).get("rounds") and any(
        item.get("currentReport") is not None or item.get("latestAttempt") is not None
        for item in positive["default"]["rounds"]
        if isinstance(item, dict)
    ):
        errors.append(f"{opid}.default: all canonical rounds must be explicitly empty")

    expected_latest_status = {
        "prior-ready-newer-queued": "queued",
        "prior-ready-newer-generating": "generating",
        "prior-ready-newer-failed": "failed",
    }
    for scenario_name, expected_status in expected_latest_status.items():
        rounds = positive.get(scenario_name, {}).get("rounds") or []
        matches = [
            item for item in rounds
            if isinstance(item, dict)
            and item.get("currentReport") is not None
            and isinstance(item.get("latestAttempt"), dict)
            and item["latestAttempt"].get("status") == expected_status
            and item["latestAttempt"].get("id") != item["currentReport"].get("id")
        ]
        if not matches:
            errors.append(
                f"{opid}.{scenario_name}: prior ready plus newer {expected_status} attempt required"
            )

    current_ready_rounds = positive.get("current-ready", {}).get("rounds") or []
    if not any(
        isinstance(item, dict)
        and item.get("currentReport") is not None
        and item.get("latestAttempt") is None
        for item in current_ready_rounds
    ):
        errors.append(f"{opid}.current-ready: current report with null latest attempt required")

    latest_ready_rounds = positive.get("latest-ready-is-current", {}).get("rounds") or []
    if not any(
        isinstance(item, dict)
        and isinstance(item.get("currentReport"), dict)
        and isinstance(item.get("latestAttempt"), dict)
        and item["latestAttempt"].get("status") == "ready"
        and item["latestAttempt"].get("id") == item["currentReport"].get("id")
        for item in latest_ready_rounds
    ):
        errors.append(
            f"{opid}.latest-ready-is-current: latest ready must share current report id"
        )


def _practice_round_ref(sequence: int, round_type: str) -> dict[str, Any]:
    return {
        "roundId": f"round-{sequence}-{round_type}",
        "roundSequence": sequence,
    }


def _check_practice_plan_round_pair(
    path: str, body: dict, errors: List[str], *, allow_null: bool = False
) -> None:
    has_round_id = "roundId" in body
    has_round_sequence = "roundSequence" in body
    if has_round_id != has_round_sequence:
        errors.append(f"{path}: roundId and roundSequence must be present as a pair")
        return
    if not has_round_id:
        errors.append(f"{path}: current practice plan must include roundId and roundSequence")
        return

    round_id = body.get("roundId")
    round_sequence = body.get("roundSequence")
    if allow_null and round_id is None and round_sequence is None:
        return
    if not isinstance(round_id, str) or not isinstance(round_sequence, int):
        errors.append(f"{path}: roundId and roundSequence must both be non-null current identity values")
        return
    match = re.fullmatch(
        r"round-([1-9][0-9]*)-(hr|technical|manager|cross_functional|culture|final|other)",
        round_id,
    )
    if match is None or int(match.group(1)) != round_sequence:
        errors.append(f"{path}: roundId sequence must equal roundSequence")


def check_feedback_report_semantics(scenarios: dict, errors: List[str]) -> None:
    expected_status = {
        "default": "ready",
        "ready-needs-practice": "ready",
        "queued": "queued",
        "generating": "generating",
        "ready-well-prepared": "ready",
        "ready-empty-focus": "ready",
        "failed": "failed",
        "failed-context-too-large": "failed",
        "invalid-contract": "failed",
        "long-content": "ready",
    }
    expected_error = {
        "queued": None,
        "generating": None,
        "failed": "AI_PROVIDER_TIMEOUT",
        "failed-context-too-large": "REPORT_CONTEXT_TOO_LARGE",
        "invalid-contract": "AI_OUTPUT_INVALID",
    }
    null_when_not_ready = ("summary", "preparednessLevel", "provenance")
    empty_when_not_ready = (
        "dimensionAssessments",
        "highlights",
        "issues",
        "nextActions",
        "retryFocusDimensionCodes",
    )
    forbidden_keys = {
        "dimension",
        "retryFocusCompetencyCodes",
        "questionAssessments",
        "retryFocusTurnIds",
    }

    for scenario_name, status in expected_status.items():
        scenario = scenarios.get(scenario_name)
        if not isinstance(scenario, dict):
            continue
        body = ((scenario.get("response") or {}).get("body") or {})
        path = f"getFeedbackReport.{scenario_name}.response.body"
        if body.get("status") != status:
            errors.append(f"{path}.status: expected {status!r}")
        if not isinstance(body.get("context"), dict):
            errors.append(f"{path}.context: frozen report context is required")
        for key_path, key in _walk_keys(body):
            if key in forbidden_keys:
                errors.append(f"{path}.{key_path}: legacy report field {key!r} is forbidden")

        if status == "ready":
            if body.get("errorCode") is not None:
                errors.append(f"{path}.errorCode: ready report must use null")
            if not isinstance(body.get("summary"), str) or not body["summary"].strip():
                errors.append(f"{path}.summary: ready report must be non-empty")
            if not isinstance(body.get("preparednessLevel"), str) or not body["preparednessLevel"].strip():
                errors.append(f"{path}.preparednessLevel: ready report must be non-null")
            if not isinstance(body.get("provenance"), dict):
                errors.append(f"{path}.provenance: ready report must be non-null object")
            continue

        if body.get("errorCode") != expected_error.get(scenario_name):
            errors.append(
                f"{path}.errorCode: expected {expected_error.get(scenario_name)!r}"
            )
        for field in null_when_not_ready:
            if body.get(field) is not None:
                errors.append(f"{path}.{field}: non-ready report must use null")
        for field in empty_when_not_ready:
            if body.get(field) != []:
                errors.append(f"{path}.{field}: non-ready report must use []")

    ready_empty_focus = scenarios.get("ready-empty-focus") or {}
    ready_empty_body = ((ready_empty_focus.get("response") or {}).get("body") or {})
    if ready_empty_body.get("retryFocusDimensionCodes") != []:
        errors.append(
            "getFeedbackReport.ready-empty-focus.response.body.retryFocusDimensionCodes: "
            "must be [] to prove generic retry fallback"
        )


def check_target_job_practice_progress(path: str, target: dict, errors: List[str]) -> None:
    summary = target.get("summary")
    if not isinstance(summary, dict):
        return
    rounds = summary.get("interviewRounds")
    if not isinstance(rounds, list) or not rounds:
        return

    canonical: list[dict[str, Any]] = []
    for index, round_item in enumerate(rounds):
        if not isinstance(round_item, dict):
            errors.append(f"{path}.summary.interviewRounds[{index}]: must be object")
            return
        sequence = round_item.get("sequence")
        round_type = round_item.get("type")
        if not isinstance(sequence, int) or not isinstance(round_type, str):
            errors.append(f"{path}.summary.interviewRounds[{index}]: sequence/type required")
            return
        canonical.append(_practice_round_ref(sequence, round_type))

    progress = target.get("practiceProgress")
    if not isinstance(progress, dict):
        errors.append(f"{path}.practiceProgress: structured TargetJob requires backend progress projection")
        return
    completed = progress.get("completedRounds")
    if not isinstance(completed, list):
        errors.append(f"{path}.practiceProgress.completedRounds: must be array")
        return
    if completed != canonical[:len(completed)]:
        errors.append(
            f"{path}.practiceProgress.completedRounds: must be an ordered, deduplicated canonical prefix"
        )
        return

    if len(completed) == 0:
        expected_status = "not_started"
        expected_current: dict[str, Any] | None = canonical[0]
    elif len(completed) < len(canonical):
        expected_status = "in_progress"
        expected_current = canonical[len(completed)]
    else:
        expected_status = "completed"
        expected_current = None

    if progress.get("status") != expected_status:
        errors.append(
            f"{path}.practiceProgress.status: expected {expected_status!r} from completed session facts"
        )
    if progress.get("currentRound") != expected_current:
        errors.append(
            f"{path}.practiceProgress.currentRound: expected first incomplete canonical round {expected_current!r}"
        )
    if expected_status == "completed" and target.get("currentPracticePlanId") is not None:
        errors.append(f"{path}.currentPracticePlanId: completed progress must not expose a current plan")


def check_practice_round_semantics(opid: str, scenarios: dict, errors: List[str]) -> None:
    if opid == "createPracticePlan":
        if "report-derived" in scenarios:
            errors.append(
                "createPracticePlan.report-derived: compatibility scenario is forbidden; "
                "use retry-derived and next-derived"
            )
        for scenario_name in ("default", "retry-derived", "next-derived"):
            scenario = scenarios.get(scenario_name) or {}
            response_body = ((scenario.get("response") or {}).get("body") or {})
            _check_practice_plan_round_pair(
                f"{opid}.{scenario_name}.response.body", response_body, errors
            )
            request_body = ((scenario.get("request") or {}).get("body") or {})
            if scenario_name == "default" and request_body.get("roundId") != response_body.get("roundId"):
                errors.append(
                    f"{opid}.{scenario_name}: request roundId must equal persisted response roundId"
                )
            if scenario_name == "default":
                continue
            expected_goal = {
                "retry-derived": "retry_current_round",
                "next-derived": "next_round",
            }[scenario_name]
            if set(request_body) != {"goal", "sourceReportId"}:
                errors.append(
                    f"{opid}.{scenario_name}.request.body: derived request must contain "
                    "exactly goal and sourceReportId"
                )
            if request_body.get("goal") != expected_goal:
                errors.append(
                    f"{opid}.{scenario_name}.request.body.goal: expected {expected_goal!r}"
                )
            if response_body.get("goal") != expected_goal:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.goal: expected {expected_goal!r}"
                )
            if request_body.get("sourceReportId") != response_body.get("sourceReportId"):
                errors.append(
                    f"{opid}.{scenario_name}: sourceReportId must match persisted response"
                )
        return

    if opid == "getPracticePlan":
        default_body = (((scenarios.get("default") or {}).get("response") or {}).get("body") or {})
        _check_practice_plan_round_pair(f"{opid}.default.response.body", default_body, errors)
        legacy_body = (
            ((scenarios.get("legacy-null-round-identity") or {}).get("response") or {}).get("body") or {}
        )
        _check_practice_plan_round_pair(
            f"{opid}.legacy-null-round-identity.response.body",
            legacy_body,
            errors,
            allow_null=True,
        )
        return

    if opid not in {"getTargetJob", "listTargetJobs"}:
        return

    expected_variant_status = {
        "default": "in_progress",
        "not-started-progress": "not_started",
        "all-completed-progress": "completed",
    }
    for scenario_name, scenario in scenarios.items():
        body = ((scenario.get("response") or {}).get("body") or {})
        targets = body.get("items") if opid == "listTargetJobs" else [body]
        if not isinstance(targets, list):
            continue
        projected_statuses: list[str] = []
        for index, target in enumerate(targets):
            if not isinstance(target, dict):
                continue
            target_path = f"{opid}.{scenario_name}.response.body"
            if opid == "listTargetJobs":
                target_path += f".items[{index}]"
            check_target_job_practice_progress(target_path, target, errors)
            summary = target.get("summary")
            if isinstance(summary, dict) and summary.get("interviewRounds"):
                progress = target.get("practiceProgress")
                if isinstance(progress, dict) and isinstance(progress.get("status"), str):
                    projected_statuses.append(progress["status"])
        expected = expected_variant_status.get(scenario_name)
        if expected is not None and expected not in projected_statuses:
            errors.append(
                f"{opid}.{scenario_name}: requires a structured TargetJob with practiceProgress.status={expected!r}"
            )


def check_practice_voice_playable_refs(opid: str, scenarios: dict, errors: List[str]) -> None:
    if opid != "createPracticeVoiceTurn":
        return
    for scenario_name, scenario in scenarios.items():
        body = ((scenario.get("response") or {}).get("body") or {})
        for index, chunk in enumerate(body.get("ttsChunks") or []):
            audio_ref = chunk.get("audioRef")
            if not isinstance(audio_ref, str):
                errors.append(
                    f"{opid}.{scenario_name}.response.body.ttsChunks[{index}].audioRef: must be string"
                )
                continue
            if PRACTICE_VOICE_PLAYABLE_AUDIO_REF_RE.match(audio_ref):
                continue
            if PRACTICE_VOICE_RESOLVER_AUDIO_REF_RE.match(audio_ref):
                continue
            errors.append(
                f"{opid}.{scenario_name}.response.body.ttsChunks[{index}].audioRef: "
                "must be browser-playable data:audio/...;base64,... or a checked-in resolver URL; "
                f"got {audio_ref!r}"
            )


def check_practice_conversation_semantics(
    opid: str, scenarios: dict, errors: List[str]
) -> None:
    if opid == "appendSessionEvent":
        for scenario_name, scenario in scenarios.items():
            request_payload = (((scenario.get("request") or {}).get("body") or {}).get("payload") or {})
            for field in ("playedTextHash", "committedTextHash"):
                digest = request_payload.get(field)
                if digest is not None and (
                    not isinstance(digest, str) or not SHA256_DIGEST_RE.fullmatch(digest)
                ):
                    errors.append(
                        f"{opid}.{scenario_name}.request.body.payload.{field}: "
                        f"must be a bare or sha256:-prefixed 64-hex SHA-256 digest"
                    )

            body = ((scenario.get("response") or {}).get("body") or {})
            action = body.get("assistantAction") or {}
            if action.get("type") != "ask_follow_up":
                continue
            current_turn = ((body.get("session") or {}).get("currentTurn") or {})
            if current_turn.get("id") != action.get("turnId"):
                errors.append(
                    f"{opid}.{scenario_name}.response.body.session.currentTurn.id: "
                    "must match assistantAction.turnId for ask_follow_up"
                )
            if current_turn.get("questionText") != action.get("questionText"):
                errors.append(
                    f"{opid}.{scenario_name}.response.body.session.currentTurn.questionText: "
                    "must match assistantAction.questionText for ask_follow_up"
                )
            if current_turn.get("status") != "follow_up_requested":
                errors.append(
                    f"{opid}.{scenario_name}.response.body.session.currentTurn.status: "
                    "must be follow_up_requested for ask_follow_up"
                )

        default = scenarios.get("default") or {}
        replay = scenarios.get("replay") or {}
        default_body = ((default.get("response") or {}).get("body") or {})
        replay_body = ((replay.get("response") or {}).get("body") or {})
        if replay_body != default_body:
            errors.append(
                f"{opid}.replay.response.body: must equal the default original snapshot"
            )

        for scenario_name, scenario in scenarios.items():
            request = ((scenario.get("request") or {}).get("body") or {})
            response = scenario.get("response") or {}
            body = response.get("body") or {}
            action = body.get("assistantAction") or {}
            if request.get("kind") == "session_resumed" and (
                response.get("status") != 200
                or action.get("type") != "session_wait"
                or action.get("questionText") not in (None, "")
            ):
                errors.append(
                    f"{opid}.{scenario_name}.response: session_resumed must return "
                    "200 session_wait without a new question"
                )

        timeout = scenarios.get("ai-timeout") or {}
        timeout_request = ((timeout.get("request") or {}).get("body") or {})
        timeout_response = timeout.get("response") or {}
        timeout_body = timeout_response.get("body") or {}
        timeout_action = timeout_body.get("assistantAction") or {}
        timeout_turn = ((timeout_body.get("session") or {}).get("currentTurn") or {})
        timeout_turn_id = ((timeout_request.get("payload") or {}).get("turnId"))
        if (
            timeout_response.get("status") != 200
            or timeout_action.get("type") != "session_wait"
            or timeout_action.get("questionText") not in (None, "")
            or timeout_turn.get("id") != timeout_turn_id
            or timeout_turn.get("status") != "asked"
        ):
            errors.append(
                f"{opid}.ai-timeout.response: provider timeout must return 200 "
                "session_wait and preserve the original asked turn"
            )
        return

    if opid != "createPracticeVoiceTurn":
        return
    default_response = ((scenarios.get("default") or {}).get("response") or {})
    default_code = (((default_response.get("body") or {}).get("error") or {}).get("code"))
    if default_response.get("status") != 422 or default_code != "AI_UNSUPPORTED_CAPABILITY":
        errors.append(
            "createPracticeVoiceTurn.default.response: disabled phone contract must return "
            "422 AI_UNSUPPORTED_CAPABILITY"
        )
    if set(scenarios) != {"default"}:
        errors.append("createPracticeVoiceTurn: disabled contract permits only the default scenario")
    return


def _check_practice_message_sequence(
    path: str, messages: Any, errors: List[str]
) -> None:
    if not isinstance(messages, list):
        errors.append(f"{path}: messages must be an array")
        return
    seen_message_ids: set[str] = set()
    seen_client_ids: set[str] = set()
    for index, message in enumerate(messages):
        message_path = f"{path}[{index}]"
        if not isinstance(message, dict):
            errors.append(f"{message_path}: message must be an object")
            continue
        message_id = message.get("id")
        if not isinstance(message_id, str) or not message_id:
            errors.append(f"{message_path}.id: non-empty message id required")
        elif message_id in seen_message_ids:
            errors.append(f"{message_path}.id: duplicate message id {message_id!r}")
        else:
            seen_message_ids.add(message_id)

        role = message.get("role")
        if role == "assistant":
            for forbidden in ("clientMessageId", "replyStatus"):
                if forbidden in message:
                    errors.append(
                        f"{message_path}.{forbidden}: assistant recovery field forbidden"
                    )
            continue
        if role != "user":
            errors.append(f"{message_path}.role: expected user or assistant")
            continue

        client_message_id = message.get("clientMessageId")
        if not isinstance(client_message_id, str) or not client_message_id:
            errors.append(f"{message_path}.clientMessageId: user field required")
        elif client_message_id in seen_client_ids:
            errors.append(
                f"{message_path}.clientMessageId: duplicate user replay identity"
            )
        else:
            seen_client_ids.add(client_message_id)
        reply_status = message.get("replyStatus")
        if reply_status not in PRACTICE_REPLY_STATUSES:
            errors.append(
                f"{message_path}.replyStatus: expected one of "
                f"{sorted(PRACTICE_REPLY_STATUSES)!r}"
            )
            continue
        next_message = messages[index + 1] if index + 1 < len(messages) else None
        has_immediate_assistant = (
            isinstance(next_message, dict) and next_message.get("role") == "assistant"
        )
        if reply_status == "complete" and not has_immediate_assistant:
            errors.append(
                f"{message_path}.replyStatus: complete requires exactly one following assistant"
            )
        if reply_status != "complete" and index != len(messages) - 1:
            errors.append(
                f"{message_path}.replyStatus: unresolved user message must be the final projection"
            )


def check_practice_reply_recovery_semantics(
    opid: str, scenarios: dict, errors: List[str]
) -> None:
    if opid == "getPracticeSession":
        for scenario_name, scenario in scenarios.items():
            response = scenario.get("response") or {}
            if response.get("status") != 200:
                continue
            body = response.get("body") or {}
            _check_practice_message_sequence(
                f"{opid}.{scenario_name}.response.body.messages",
                body.get("messages"),
                errors,
            )

        expected_status = {
            "reply-pending": "pending",
            "reply-retryable-failed": "retryable_failed",
            "reply-terminal-failed": "terminal_failed",
            "reply-complete": "complete",
        }
        shared_identity: tuple[Any, ...] | None = None
        for scenario_name, reply_status in expected_status.items():
            scenario = scenarios.get(scenario_name) or {}
            response = scenario.get("response") or {}
            body = response.get("body") or {}
            messages = body.get("messages") or []
            users = [
                message
                for message in messages
                if isinstance(message, dict) and message.get("role") == "user"
            ]
            if not users:
                errors.append(f"{opid}.{scenario_name}: recovery user message required")
                continue
            user = users[-1]
            if user.get("replyStatus") != reply_status:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.messages: expected "
                    f"replyStatus={reply_status!r}"
                )
            identity = (
                body.get("id"),
                user.get("id"),
                user.get("clientMessageId"),
                user.get("content"),
            )
            if shared_identity is None:
                shared_identity = identity
            elif identity != shared_identity:
                errors.append(
                    f"{opid}.{scenario_name}: recovery scenarios must share session/user identity"
                )
        retryable = scenarios.get("reply-retryable-failed") or {}
        marker = ((retryable.get("response") or {}).get("headers") or {}).get(
            "X-Recovery-Marker"
        )
        if marker != "reload-same-client-message-id":
            errors.append(
                f"{opid}.reply-retryable-failed.response.headers.X-Recovery-Marker: "
                "expected reload-same-client-message-id"
            )
        return

    if opid != "sendPracticeMessage":
        return
    if tuple(scenarios) != PRACTICE_SEND_SCENARIO_ORDER:
        errors.append(
            f"{opid}: scenarios must be exactly {list(PRACTICE_SEND_SCENARIO_ORDER)!r}; "
            f"got {list(scenarios)!r}"
        )

    for scenario_name, expected in PRACTICE_SEND_FAILURE_MATRIX.items():
        scenario = scenarios.get(scenario_name) or {}
        response = scenario.get("response") or {}
        body = response.get("body")
        status, code, retryable, details = expected
        if not isinstance(body, dict) or set(body) != {"error"}:
            errors.append(
                f"{opid}.{scenario_name}.response.body: typed ApiErrorResponse required"
            )
            continue
        error = body.get("error")
        if not isinstance(error, dict):
            errors.append(f"{opid}.{scenario_name}.response.body.error: object required")
            continue
        if response.get("status") != status:
            errors.append(f"{opid}.{scenario_name}.response.status: expected {status}")
        if error.get("code") != code:
            errors.append(
                f"{opid}.{scenario_name}.response.body.error.code: expected {code!r}"
            )
        if error.get("retryable") is not retryable:
            errors.append(
                f"{opid}.{scenario_name}.response.body.error.retryable: expected {retryable}"
            )
        if error.get("details") != details:
            errors.append(
                f"{opid}.{scenario_name}.response.body.error.details: expected {details!r}"
            )
        for required in ("message", "requestId"):
            if not isinstance(error.get(required), str) or not error[required]:
                errors.append(
                    f"{opid}.{scenario_name}.response.body.error.{required}: non-empty string required"
                )

    for scenario_name in ("default", "retry-success-same-client-message"):
        scenario = scenarios.get(scenario_name) or {}
        body = ((scenario.get("response") or {}).get("body") or {})
        session = body.get("session") or {}
        messages = session.get("messages")
        _check_practice_message_sequence(
            f"{opid}.{scenario_name}.response.body.session.messages",
            messages,
            errors,
        )
        user = body.get("userMessage") or {}
        assistant = body.get("assistantMessage") or {}
        request = ((scenario.get("request") or {}).get("body") or {})
        if user.get("role") != "user" or user.get("replyStatus") != "complete":
            errors.append(
                f"{opid}.{scenario_name}.response.body.userMessage: complete user required"
            )
        if user.get("clientMessageId") != request.get("clientMessageId"):
            errors.append(
                f"{opid}.{scenario_name}: response user must preserve request clientMessageId"
            )
        if assistant.get("role") != "assistant" or any(
            field in assistant for field in ("clientMessageId", "replyStatus")
        ):
            errors.append(
                f"{opid}.{scenario_name}.response.body.assistantMessage: assistant recovery fields forbidden"
            )
        if isinstance(messages, list):
            if sum(
                isinstance(message, dict) and message.get("id") == user.get("id")
                for message in messages
            ) != 1:
                errors.append(
                    f"{opid}.{scenario_name}: exactly one projected user message required"
                )
            if sum(
                isinstance(message, dict) and message.get("id") == assistant.get("id")
                for message in messages
            ) != 1:
                errors.append(
                    f"{opid}.{scenario_name}: exactly one projected assistant message required"
                )

    timeout_request = (
        ((scenarios.get("ai-timeout-retryable") or {}).get("request") or {}).get("body")
        or {}
    )
    retry_request = (
        ((scenarios.get("retry-success-same-client-message") or {}).get("request") or {}).get("body")
        or {}
    )
    if retry_request != timeout_request:
        errors.append(
            f"{opid}.retry-success-same-client-message.request.body: retry must reuse "
            "the exact clientMessageId and text from ai-timeout-retryable"
        )


def check_privacy_and_ids(opid: str, data: dict, errors: List[str]) -> None:
    for path, value in _walk_strings(data):
        if TEMP_ID_RE.search(value):
            errors.append(f"{opid}.{path}: tmp_ id forbidden ({value!r})")
        if UUID_SHAPE_RE.match(value) and not UUID_V7_RE.match(value):
            errors.append(f"{opid}.{path}: id {value!r} matches UUID shape but not UUIDv7 layout")
        for match in EMAIL_RE.findall(value):
            domain = match.lower()
            if not (domain in ALLOWED_EMAIL_DOMAINS or domain.endswith(".example")):
                errors.append(f"{opid}.{path}: email domain {match!r} not on allowlist")
        for phone in PHONE_RE.findall(value):
            if not phone.startswith(ALLOWED_PHONE_PREFIX):
                errors.append(f"{opid}.{path}: phone {phone!r} not on +1-555-01xx allowlist")
        m = COMPANY_BLACKLIST_RE.search(value)
        if m:
            errors.append(
                f"{opid}.{path}: blacklisted employer brand {m.group(1)!r} present"
            )


def check_d20_out_of_scope_fixture_keys(opid: str, data: dict, errors: List[str]) -> None:
    for path, key in _walk_keys(data):
        if key in OUT_OF_SCOPE_D20_FIXTURE_KEYS:
            errors.append(
                f"{opid}.{path}: D-20 flat resume fixtures must use resumeId, "
                f"not out-of-scope key {key!r}"
            )


# ---------- top-level entry --------------------------------------------------

def validate(repo_root: Path) -> List[str]:
    fixtures_root = repo_root / "openapi" / "fixtures"
    openapi_path = repo_root / "openapi" / "openapi.yaml"
    spec = load_openapi(openapi_path)
    op_index = build_operation_index(spec)

    errors: List[str] = []
    seen: set[str] = set()

    for tag, opid, fixture_path, data in walk_fixtures(fixtures_root):
        seen.add(opid)
        op_meta = op_index.get(opid)
        if op_meta is None:
            errors.append(f"{opid}: not present in openapi.yaml")
            continue
        expected_tag = op_meta.get("tag")
        if expected_tag != tag:
            errors.append(f"{opid}: fixture in tag {tag!r} does not match openapi tag {expected_tag!r}")
        op = op_meta.get("operation")
        check_structural(opid, data, fixture_path, op, errors)
        scenarios = data.get("scenarios", {}) or {}
        for scenario_name, scenario in scenarios.items():
            check_status_declared(opid, op, scenario_name, scenario, errors)
            check_schema(opid, op, scenario_name, scenario, spec, errors)
        check_required_named_scenarios(opid, scenarios, errors)
        check_target_job_paste_only_semantics(opid, scenarios, errors)
        if opid == "listTargetJobReports":
            check_target_job_reports_overview_semantics(scenarios, errors)
        if opid == "getFeedbackReport":
            check_feedback_report_semantics(scenarios, errors)
        check_practice_round_semantics(opid, scenarios, errors)
        check_practice_voice_playable_refs(opid, scenarios, errors)
        check_practice_conversation_semantics(opid, scenarios, errors)
        check_practice_reply_recovery_semantics(opid, scenarios, errors)
        check_provenance(opid, scenarios.get("default") or {}, errors)
        check_p0_export_error_code(opid, scenarios, errors)
        check_d20_out_of_scope_fixture_keys(opid, data, errors)
        check_privacy_and_ids(opid, data, errors)

    expected = {opid for _tag, opid in expected_fixture_operations(spec)}
    missing = expected - seen
    for opid in sorted(missing):
        errors.append(f"missing fixture for operationId {opid}")

    return errors


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path(__file__).resolve().parents[2],
        help="Repository root containing openapi/ (default: walk up from script).",
    )
    args = parser.parse_args(list(argv))

    errors = validate(args.repo_root)
    if errors:
        for err in errors:
            print(f"validate-fixtures: {err}", file=sys.stderr)
        print(
            f"validate-fixtures: FAILED with {len(errors)} error(s) under {args.repo_root}",
            file=sys.stderr,
        )
        return 1
    spec = load_openapi(args.repo_root / "openapi" / "openapi.yaml")
    expected_count = len(expected_fixture_operations(spec))
    print(
        f"validate-fixtures: OK — {expected_count} fixtures under {args.repo_root / 'openapi' / 'fixtures'}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
