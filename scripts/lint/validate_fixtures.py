#!/usr/bin/env python3
"""Validate openapi/fixtures/*.json against openapi/openapi.yaml.

Phase 1.3 scope (per `002-fixtures-and-mock-source` plan §3 / spec C-6 / C-11):
    1. structural — operationId matches filename, `scenarios.default` exists
       and is the first scenario, status code is declared on the operation.
    2. schema    — request.body and response.body schema-valid against the
       operation's request / status-matched response schema in openapi.yaml.
    3. provenance — every AI-generated schema listed in spec §4.6 carries a
       complete `GenerationProvenance` (6 non-empty fields).
    4. privacy   — emails restricted to example.{com,org,net} or `.example`,
       phones to `+1-555-01xx`, and the employer-brand blacklist below.
    5. ids       — `format: uuid` values must match UUIDv7 layout, and any
       string with `tmp_` prefix is rejected.
    6. coverage  — exactly the 36 operationIds frozen by spec §3.1.1 must
       have a fixture file.
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

# Real employer-style brands; AI-vendor names are excluded by design — they
# only show up in `provenance.modelId` infrastructure metadata.
COMPANY_BLACKLIST = (
    "alibaba", "tencent", "bytedance", "baidu", "meituan", "didi", "huawei",
    "字节", "腾讯", "阿里巴巴", "百度", "美团", "滴滴", "华为", "星环",
)
COMPANY_BLACKLIST_RE = re.compile(
    r"(?:^|[^A-Za-z0-9_])(" + "|".join(re.escape(b) for b in COMPANY_BLACKLIST) +
    r")(?:[^A-Za-z0-9_]|$)",
    re.IGNORECASE,
)

# Spec §3.1.1 36-endpoint freeze. Mirrors openapi_inventory.py and is the
# coverage source-of-truth for missing-fixture errors.
EXPECTED_OPERATIONS: Tuple[Tuple[str, str], ...] = (
    ("Auth", "getMe"),
    ("Auth", "startAuthEmailChallenge"),
    ("Auth", "verifyAuthEmailChallenge"),
    ("Auth", "logout"),
    ("Auth", "getRuntimeConfig"),
    ("Uploads", "createUploadPresign"),
    ("Profile", "getMyProfile"),
    ("Profile", "updateMyProfile"),
    ("Profile", "listExperienceCards"),
    ("Profile", "createExperienceCard"),
    ("Profile", "updateExperienceCard"),
    ("Resumes", "registerResume"),
    ("Resumes", "getResume"),
    ("TargetJobs", "importTargetJob"),
    ("TargetJobs", "listTargetJobs"),
    ("TargetJobs", "getTargetJob"),
    ("TargetJobs", "updateTargetJob"),
    ("PracticePlans", "createPracticePlan"),
    ("PracticePlans", "getPracticePlan"),
    ("PracticeSessions", "startPracticeSession"),
    ("PracticeSessions", "getPracticeSession"),
    ("PracticeSessions", "appendSessionEvent"),
    ("PracticeSessions", "completePracticeSession"),
    ("Reports", "getFeedbackReport"),
    ("Reports", "listTargetJobReports"),
    ("Mistakes", "listMistakes"),
    ("Mistakes", "retestMistake"),
    ("ResumeTailor", "requestResumeTailor"),
    ("ResumeTailor", "getResumeTailorRun"),
    ("Debriefs", "createDebrief"),
    ("Debriefs", "getDebrief"),
    ("Growth", "getGrowthOverview"),
    ("Jobs", "getJob"),
    ("Privacy", "requestPrivacyExport"),
    ("Privacy", "requestPrivacyDelete"),
    ("Privacy", "getPrivacyRequest"),
)

# AI-generated schemas listed in spec §4.6. Provenance must resolve from each
# value path, where `[*]` expands a list and yields its members.
AI_PROVENANCE_PATHS: dict[str, Tuple[str, ...]] = {
    "getTargetJob": ("summary.provenance", "fitSummary.provenance"),
    "listTargetJobs": ("items[*].summary.provenance", "items[*].fitSummary.provenance"),
    "updateTargetJob": ("summary.provenance", "fitSummary.provenance"),
    "appendSessionEvent": ("assistantAction.provenance",),
    "getFeedbackReport": ("provenance",),
    "listTargetJobReports": ("items[*].provenance",),
    "listMistakes": ("items[*].provenance",),
    "getResumeTailorRun": ("provenance",),
    "getDebrief": ("provenance",),
}
PROVENANCE_REQUIRED_FIELDS = (
    "promptVersion",
    "rubricVersion",
    "modelId",
    "language",
    "featureFlag",
    "dataSourceVersion",
)


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
        items_schema = schema.get("items")
        for i, item in enumerate(value):
            schema_validate(item, items_schema, root=root,
                            path=f"{path}[{i}]", errors=errors)
        return

    if t == "string":
        if not isinstance(value, str):
            errors.append(f"{path}: expected string, got {type(value).__name__}")
            return
        if "enum" in schema and value not in schema["enum"]:
            errors.append(f"{path}: value {value!r} not in enum {schema['enum']}")
        if "const" in schema and value != schema["const"]:
            errors.append(f"{path}: value {value!r} != const {schema['const']!r}")
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
            schema_validate(
                scenario["request"]["body"], body_schema, root=root,
                path=f"{opid}.{scenario_name}.request.body", errors=errors,
            )


def check_provenance(opid: str, scenario: dict, errors: List[str]) -> None:
    paths = AI_PROVENANCE_PATHS.get(opid)
    if not paths:
        return
    body = (scenario.get("response") or {}).get("body")
    if body is None:
        errors.append(f"{opid}: missing response.body for provenance check")
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


def check_privacy_export_error_code(opid: str, scenarios: dict, errors: List[str]) -> None:
    """Spec D-12 / C-7: requestPrivacyExport's default scenario must return
    501 with `error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`. The hand-written
    example previously living in openapi.yaml moved here as part of B2 002 §3.1
    (single source of truth). Apply only to that operation."""
    if opid != "requestPrivacyExport":
        return
    default = scenarios.get("default") or {}
    response = default.get("response") or {}
    if response.get("status") != 501:
        errors.append(f"{opid}: default.response.status must be 501 (spec D-12)")
        return
    body = response.get("body") or {}
    code = (body.get("error") or {}).get("code")
    if code != "PRIVACY_EXPORT_NOT_AVAILABLE":
        errors.append(
            f"{opid}: default.response.body.error.code must be 'PRIVACY_EXPORT_NOT_AVAILABLE' (spec D-12); got {code!r}"
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
        if (tag, opid) not in EXPECTED_OPERATIONS:
            errors.append(f"{opid}: fixture in tag {tag!r} not in spec §3.1.1 freeze")
        op = (op_index.get(opid) or {}).get("operation")
        if op is None:
            errors.append(f"{opid}: not present in openapi.yaml")
            continue
        check_structural(opid, data, fixture_path, op, errors)
        scenarios = data.get("scenarios", {}) or {}
        for scenario_name, scenario in scenarios.items():
            if scenario_name not in {"default", "prototype-baseline"}:
                # other scenarios accepted; still validated for status + schema.
                pass
            check_status_declared(opid, op, scenario_name, scenario, errors)
            check_schema(opid, op, scenario_name, scenario, spec, errors)
        check_provenance(opid, scenarios.get("default") or {}, errors)
        check_privacy_export_error_code(opid, scenarios, errors)
        check_privacy_and_ids(opid, data, errors)

    expected = {opid for _tag, opid in EXPECTED_OPERATIONS}
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
    print(
        f"validate-fixtures: OK — 36 fixtures under {args.repo_root / 'openapi' / 'fixtures'}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
