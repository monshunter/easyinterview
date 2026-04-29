#!/usr/bin/env python3
"""Structural validator for shared/conventions.yaml.

Acts as the local lint gate that enforces the cross-language truth source defined by
[shared-conventions-codified spec](../../docs/spec/shared-conventions-codified/spec.md)
and [00-shared-conventions.md](../../easyinterview-tech-docs/00-shared-conventions.md).

Run: `python3 scripts/lint/conventions_yaml.py [path]`
"""
from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Any

import yaml

ENUM_VALUE_RE = re.compile(r"^[a-z][a-z0-9_]*$")
ERROR_CODE_RE = re.compile(r"^[A-Z][A-Z0-9_]*$")
UUIDV7_RE_FALLBACK = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)

EXPECTED_TOP_LEVEL = {
    "version",
    "schemaVersion",
    "sampleUuidV7",
    "uuidV7Regex",
    "tmpIdPrefix",
    "pagination",
    "idempotency",
    "errors",
    "jobStatuses",
    "enums",
    "structures",
    "aiVocabulary",
}

EXPECTED_STRUCTURES = {"PageInfo", "ApiError"}
EXPECTED_ENUM_SECTIONS = {f"5.{i}" for i in range(1, 14)}  # §5.1 .. §5.13
EXPECTED_JOB_STATUSES = {"queued", "running", "succeeded", "failed", "cancelled", "dead"}
REQUIRED_ERROR_CODES = {
    "AUTH_UNAUTHORIZED",
    "TARGET_IMPORT_FAILED",
    "PRACTICE_SESSION_CONFLICT",
    "REPORT_NOT_READY",
    "VALIDATION_FAILED",
    "RATE_LIMITED",
}
REQUIRED_AI_VOCABULARY_FIELDS = {
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
}


class ValidationError(Exception):
    pass


def _require(cond: bool, msg: str) -> None:
    if not cond:
        raise ValidationError(msg)


def validate(data: dict[str, Any]) -> list[str]:
    errors: list[str] = []

    missing = EXPECTED_TOP_LEVEL - set(data)
    if missing:
        errors.append(f"missing top-level keys: {sorted(missing)}")

    if "tmpIdPrefix" in data and data["tmpIdPrefix"] != "tmp_":
        errors.append(f"tmpIdPrefix must be 'tmp_', got {data['tmpIdPrefix']!r}")

    pagination = data.get("pagination") or {}
    if pagination.get("defaultPageSize") != 20:
        errors.append("pagination.defaultPageSize must be 20")
    if pagination.get("maxPageSize") != 100:
        errors.append("pagination.maxPageSize must be 100")

    idempotency = data.get("idempotency") or {}
    if idempotency.get("ttlSeconds") != 86400:
        errors.append("idempotency.ttlSeconds must be 86400 (24h)")

    sample = data.get("sampleUuidV7", "")
    if not UUIDV7_RE_FALLBACK.match(sample):
        errors.append(f"sampleUuidV7 must be a UUIDv7 string, got {sample!r}")

    error_list = data.get("errors") or []
    seen_error_codes: set[str] = set()
    for entry in error_list:
        code = entry.get("code", "")
        if not ERROR_CODE_RE.match(code):
            errors.append(f"error code must be UPPER_SNAKE_CASE, got {code!r}")
        if code in seen_error_codes:
            errors.append(f"duplicate error code: {code!r}")
        seen_error_codes.add(code)
        if "message" not in entry:
            errors.append(f"error {code!r} missing 'message'")
        if "retryable" not in entry:
            errors.append(f"error {code!r} missing 'retryable' bool")
        elif not isinstance(entry["retryable"], bool):
            errors.append(f"error {code!r} retryable must be bool")
    missing_error_codes = REQUIRED_ERROR_CODES - seen_error_codes
    if missing_error_codes:
        errors.append(
            "errors must include all 6 upstream examples from 00-shared-conventions §3.2; "
            f"missing {sorted(missing_error_codes)}"
        )

    job_statuses = set(data.get("jobStatuses") or [])
    if job_statuses != EXPECTED_JOB_STATUSES:
        errors.append(
            "jobStatuses must equal "
            f"{sorted(EXPECTED_JOB_STATUSES)}, got {sorted(job_statuses)}"
        )

    enums = data.get("enums") or []
    seen_sections: set[str] = set()
    seen_names: set[str] = set()
    for enum in enums:
        name = enum.get("name", "")
        if not name or not name[0].isupper():
            errors.append(f"enum name must be PascalCase, got {name!r}")
        if name in seen_names:
            errors.append(f"duplicate enum name: {name!r}")
        seen_names.add(name)

        section = enum.get("sourceSection", "")
        if section not in EXPECTED_ENUM_SECTIONS:
            errors.append(f"enum {name!r} sourceSection must be one of §5.1..§5.13, got {section!r}")
        seen_sections.add(section)

        json_field = enum.get("jsonField", "")
        if not json_field or json_field[0].isupper() or "_" in json_field:
            errors.append(f"enum {name!r} jsonField must be camelCase, got {json_field!r}")

        values = enum.get("values") or []
        if not values:
            errors.append(f"enum {name!r} must have at least one value")
        for value in values:
            if not ENUM_VALUE_RE.match(value):
                errors.append(
                    f"enum {name!r} value must be lower_snake_case, got {value!r}"
                )

    missing_sections = EXPECTED_ENUM_SECTIONS - seen_sections
    if missing_sections:
        errors.append(
            f"enums must cover all 13 §5 sections; missing {sorted(missing_sections)}"
        )

    structures = data.get("structures") or {}
    missing_structs = EXPECTED_STRUCTURES - set(structures)
    if missing_structs:
        errors.append(f"missing structures: {sorted(missing_structs)}")
    for struct_name in EXPECTED_STRUCTURES & set(structures):
        fields = structures[struct_name].get("fields") or []
        if not fields:
            errors.append(f"structure {struct_name!r} must declare fields")
        for field in fields:
            field_name = field.get("name", "")
            if not field_name or field_name[0].isupper() or "_" in field_name:
                errors.append(
                    f"structure {struct_name!r} field name must be camelCase, got {field_name!r}"
                )

    ai_vocabulary = data.get("aiVocabulary") or {}
    ai_fields = ai_vocabulary.get("fields") or []
    if not isinstance(ai_fields, list) or not ai_fields:
        errors.append("aiVocabulary.fields must declare fields")
    seen_ai_fields: set[str] = set()
    for field in ai_fields:
        if isinstance(field, dict):
            field_name = field.get("name", "")
        else:
            field_name = ""
        if not ENUM_VALUE_RE.match(field_name):
            errors.append(
                f"aiVocabulary field name must be lower_snake_case, got {field_name!r}"
            )
            continue
        if field_name in seen_ai_fields:
            errors.append(f"duplicate aiVocabulary field: {field_name!r}")
        seen_ai_fields.add(field_name)

    missing_ai_fields = REQUIRED_AI_VOCABULARY_FIELDS - seen_ai_fields
    if missing_ai_fields:
        errors.append(
            "aiVocabulary.fields must include the required AI meta fields; "
            f"missing {sorted(missing_ai_fields)}"
        )

    return errors


def main() -> int:
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("shared/conventions.yaml")
    if not path.exists():
        print(f"FAIL: {path} does not exist", file=sys.stderr)
        return 2

    try:
        data = yaml.safe_load(path.read_text(encoding="utf-8"))
    except yaml.YAMLError as exc:
        print(f"FAIL: {path} is not valid YAML: {exc}", file=sys.stderr)
        return 2

    if not isinstance(data, dict):
        print(f"FAIL: {path} root must be a mapping", file=sys.stderr)
        return 2

    errors = validate(data)
    if errors:
        for err in errors:
            print(f"FAIL: {err}", file=sys.stderr)
        return 1

    print(f"OK: {path} ({len(data['enums'])} enum types, {len(data['errors'])} error codes)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
