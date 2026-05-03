#!/usr/bin/env python3
"""Local lint and breaking-change gate for event/outbox contract assets."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

import yaml


BREAKING = "breaking change requires eventVersion + 1"
SOURCE_SUFFIXES = {".go", ".js", ".jsx", ".ts", ".tsx"}
IGNORED_DIRS = {".git", ".next", "coverage", "dist", "node_modules", "vendor"}
CONTRACT_GENERATED_DIRS = (
    Path("backend/internal/shared/events"),
    Path("backend/internal/shared/jobs"),
    Path("frontend/src/lib/events"),
    Path("frontend/src/lib/jobs"),
)
EVENT_GENERATOR_DIR = Path("backend/cmd/codegen/events")
FIXTURE_PARTS = {"__fixtures__", "fixtures", "testdata", "generated"}
TEST_FILE_SUFFIXES = ("_test.go", ".test.ts", ".test.tsx", ".spec.ts", ".spec.tsx")
GO_EVENT_CONST_RE = re.compile(r"\b(EventName[A-Za-z0-9_]*)\s+(?:EventName\s*)?=")
TS_EVENT_CONST_RE = re.compile(r"\b(EVENT_NAME_[A-Z0-9_]*)\s*=")
REMOVED_EVENT_NAMES = {"mistake.created", "mistake.status.changed"}
REMOVED_PAYLOAD_FIELDS = {
    ("report.generated", "mistakeCount"),
    ("debrief.completed", "generatedMistakeCount"),
}


def _by_event_name(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {
        event["name"]: event
        for event in data.get("events", [])
        if isinstance(event, dict) and isinstance(event.get("name"), str)
    }


def _by_job_type(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {
        job["canonical"]: job
        for job in data.get("jobs", [])
        if isinstance(job, dict) and isinstance(job.get("canonical"), str)
    }


def _by_enum_name(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {
        enum["name"]: enum
        for enum in data.get("eventLocalEnums", [])
        if isinstance(enum, dict) and isinstance(enum.get("name"), str)
    }


def _event_names(data: dict[str, Any]) -> list[str]:
    return sorted(_by_event_name(data))


def _canonical_job_types(data: dict[str, Any]) -> list[str]:
    return sorted(_by_job_type(data))


def _asynq_task_names(data: dict[str, Any]) -> list[str]:
    return sorted(
        job["asynqTask"]
        for job in data.get("jobs", [])
        if isinstance(job, dict) and isinstance(job.get("asynqTask"), str)
    )


def compare_events_baseline(current: dict[str, Any], baseline: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    current_events = _by_event_name(current)
    baseline_events = _by_event_name(baseline)
    current_enums = _by_enum_name(current)
    baseline_enums = _by_enum_name(baseline)

    for name in sorted(set(baseline_enums) - set(current_enums)):
        errors.append(f"{name}: deleted event-local enum is a breaking change; {BREAKING}")
    for name, baseline_enum in baseline_enums.items():
        current_enum = current_enums.get(name)
        if not current_enum:
            continue
        baseline_values = set(baseline_enum.get("values") or [])
        current_values = set(current_enum.get("values") or [])
        for value in sorted(baseline_values - current_values):
            errors.append(f"{name}.{value}: deleted enum member; {BREAKING}")

    for name in sorted(set(baseline_events) - set(current_events)):
        errors.append(f"{name}: deleted eventName is a breaking change; {BREAKING}")
    for name, baseline_event in baseline_events.items():
        current_event = current_events.get(name)
        if not current_event:
            continue
        baseline_payload = baseline_event.get("requiredPayload") or {}
        current_payload = current_event.get("requiredPayload") or {}
        for field in sorted(set(baseline_payload) - set(current_payload)):
            errors.append(f"{name}.{field}: deleted required payload field; {BREAKING}")
        for field, baseline_spec in baseline_payload.items():
            current_spec = current_payload.get(field)
            if not isinstance(current_spec, dict) or not isinstance(baseline_spec, dict):
                continue
            if current_spec.get("type") != baseline_spec.get("type"):
                errors.append(
                    f"{name}.{field}: required payload type changed "
                    f"{baseline_spec.get('type')!r} -> {current_spec.get('type')!r}; {BREAKING}"
                )
    return errors


def compare_jobs_baseline(current: dict[str, Any], baseline: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    current_jobs = _by_job_type(current)
    baseline_jobs = _by_job_type(baseline)

    for canonical in sorted(set(baseline_jobs) - set(current_jobs)):
        errors.append(f"{canonical}: deleted jobType is a breaking change; {BREAKING}")
    for canonical, baseline_job in baseline_jobs.items():
        current_job = current_jobs.get(canonical)
        if not current_job:
            continue
        if current_job.get("apiFacing") != baseline_job.get("apiFacing"):
            errors.append(
                f"{canonical}.apiFacing changed "
                f"{baseline_job.get('apiFacing')!r} -> {current_job.get('apiFacing')!r}; {BREAKING}"
            )
        if current_job.get("asynqTask") != baseline_job.get("asynqTask"):
            errors.append(
                f"{canonical}.asynqTask changed "
                f"{baseline_job.get('asynqTask')!r} -> {current_job.get('asynqTask')!r}; {BREAKING}"
            )
    baseline_subset = baseline.get("apiFacingSubset") or []
    current_subset = current.get("apiFacingSubset") or []
    if current_subset != baseline_subset:
        errors.append(f"apiFacingSubset changed {baseline_subset!r} -> {current_subset!r}; {BREAKING}")
    return errors


def validate_jobs_contract_shape(jobs: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    job_by_type = _by_job_type(jobs)
    subset = jobs.get("apiFacingSubset") or []
    if not isinstance(subset, list):
        return ["apiFacingSubset must be a list"]
    if len(subset) != 7:
        errors.append(f"apiFacingSubset must contain exactly 7 job types, got {len(subset)}: {subset!r}")
    for canonical in subset:
        job = job_by_type.get(canonical)
        if job is None:
            errors.append(f"apiFacingSubset contains unknown jobType {canonical!r}")
            continue
        if job.get("apiFacing") is not True:
            errors.append(f"{canonical}: apiFacingSubset entry must have apiFacing=true")
    email_dispatch = job_by_type.get("email_dispatch")
    if email_dispatch:
        payload_fields = set((email_dispatch.get("payloadSchema") or {}).keys())
        redacted_fields = set(email_dispatch.get("redactedFields") or [])
        for field in sorted(payload_fields & redacted_fields):
            errors.append(f"email_dispatch.payloadSchema contains redacted field {field!r}")
    return errors


def validate_product_scope_removals(events: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    for event in events.get("events") or []:
        if not isinstance(event, dict):
            continue
        name = event.get("name")
        if name in REMOVED_EVENT_NAMES:
            errors.append(f"{name}: removed by product-scope v1.2; do not restore independent mistake events")
        payload = event.get("requiredPayload") or {}
        for removed_event, removed_field in REMOVED_PAYLOAD_FIELDS:
            if name == removed_event and removed_field in payload:
                errors.append(
                    f"{name}.{removed_field}: removed by product-scope v1.2; use "
                    f"{'questionIssueCount' if name == 'report.generated' else 'practiceFocusCount'}"
                )
    return errors


def scan_source_literals(root: Path, events: dict[str, Any], jobs: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    forbidden = _forbidden_literals(events, jobs)
    for path in _source_files(root):
        rel = path.relative_to(root)
        if _is_literal_allowed_path(rel):
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            text = path.read_text(encoding="utf-8", errors="ignore")
        for value in sorted(forbidden):
            if _contains_string_literal(text, value):
                errors.append(f"{rel}: naked event/job literal {value!r}; use generated constants")
        for match in GO_EVENT_CONST_RE.finditer(text):
            errors.append(f"{rel}: handwritten {match.group(1)} constant is forbidden outside generated events package")
        for match in TS_EVENT_CONST_RE.finditer(text):
            errors.append(f"{rel}: handwritten {match.group(1)} constant is forbidden outside generated events package")
    return errors


def validate_generated_contracts(root: Path, events: dict[str, Any], jobs: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    expected_events = _event_names(events)
    expected_subset = jobs.get("apiFacingSubset") or []

    event_paths = [
        root / "backend/internal/shared/events/events.go",
        root / "frontend/src/lib/events/events.ts",
    ]
    for path in event_paths:
        if not path.exists():
            errors.append(f"{path.relative_to(root)}: generated contract file is missing; run make codegen-events")
            continue
        generated = _parse_generated_event_names(path)
        if generated != expected_events:
            missing = sorted(set(expected_events) - set(generated))
            extra = sorted(set(generated) - set(expected_events))
            errors.append(
                f"{path.relative_to(root)}: generated event names must match 16 shared/events.yaml entries; "
                f"missing={missing!r} extra={extra!r}"
            )

    go_jobs = root / "backend/internal/shared/jobs/jobs.go"
    if not go_jobs.exists():
        errors.append(f"{go_jobs.relative_to(root)}: generated contract file is missing; run make codegen-events")
    else:
        generated_subset = _parse_go_api_facing_jobs(go_jobs)
        if generated_subset != expected_subset:
            errors.append(
                f"{go_jobs.relative_to(root)}: APIFacingJobTypes must match shared/jobs.yaml.apiFacingSubset; "
                f"expected={expected_subset!r} actual={generated_subset!r}"
            )

    ts_jobs = root / "frontend/src/lib/jobs/jobs.ts"
    if not ts_jobs.exists():
        errors.append(f"{ts_jobs.relative_to(root)}: generated contract file is missing; run make codegen-events")
    else:
        generated_subset = _parse_ts_api_facing_jobs(ts_jobs)
        if generated_subset != expected_subset:
            errors.append(
                f"{ts_jobs.relative_to(root)}: API_FACING_JOB_TYPES must match shared/jobs.yaml.apiFacingSubset; "
                f"expected={expected_subset!r} actual={generated_subset!r}"
            )
    return errors


def _forbidden_literals(events: dict[str, Any], jobs: dict[str, Any]) -> set[str]:
    return set(_event_names(events)) | set(_canonical_job_types(jobs)) | set(_asynq_task_names(jobs))


def _source_files(root: Path) -> list[Path]:
    files: list[Path] = []
    for top in ("backend", "frontend"):
        base = root / top
        if not base.exists():
            continue
        for path in base.rglob("*"):
            if not path.is_file() or path.suffix not in SOURCE_SUFFIXES:
                continue
            rel_parts = path.relative_to(root).parts
            if any(part in IGNORED_DIRS for part in rel_parts):
                continue
            files.append(path)
    return files


def _is_literal_allowed_path(rel: Path) -> bool:
    if any(_is_relative_to(rel, generated_dir) for generated_dir in CONTRACT_GENERATED_DIRS):
        return True
    if _is_relative_to(rel, EVENT_GENERATOR_DIR):
        return True
    if any(part in FIXTURE_PARTS for part in rel.parts):
        return True
    return rel.name.endswith(TEST_FILE_SUFFIXES)


def _is_relative_to(path: Path, base: Path) -> bool:
    try:
        path.relative_to(base)
    except ValueError:
        return False
    return True


def _contains_string_literal(text: str, value: str) -> bool:
    return any(f"{quote}{value}{quote}" in text for quote in ('"', "'", "`"))


def _parse_generated_event_names(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    if path.suffix == ".go":
        values = re.findall(r"\bEventName[A-Za-z0-9_]*\s+EventName\s*=\s*\"([^\"]+)\"", text)
    else:
        values = re.findall(r"\bEVENT_NAME_[A-Z0-9_]+\s*=\s*\"([^\"]+)\"", text)
    return sorted(values)


def _parse_go_api_facing_jobs(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    constants = {
        name: value
        for name, value in re.findall(r"\b(JobType[A-Za-z0-9_]*)\s+JobType\s*=\s*\"([^\"]+)\"", text)
    }
    match = re.search(r"\bAPIFacingJobTypes\s*=\s*\[\]JobType\s*\{(?P<body>.*?)\n\}", text, re.S)
    if not match:
        return []
    return [constants.get(name, name) for name in re.findall(r"\bJobType[A-Za-z0-9_]*\b", match.group("body"))]


def _parse_ts_api_facing_jobs(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    match = re.search(r"\bAPI_FACING_JOB_TYPES\s*=\s*\[(?P<body>.*?)\]\s+as\s+const", text, re.S)
    if not match:
        return []
    return re.findall(r"\"([^\"]+)\"", match.group("body"))


def _load_yaml(path: Path) -> dict[str, Any]:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def _load_json(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(Path(__file__).resolve().parents[2]))
    args = parser.parse_args()

    root = Path(args.repo_root).resolve()
    try:
        current_events = _load_yaml(root / "shared/events.yaml")
        current_jobs = _load_yaml(root / "shared/jobs.yaml")
        baseline_events = _load_json(root / "shared/events/baseline/events.v1.json")
        baseline_jobs = _load_json(root / "shared/jobs/baseline/jobs.v1.json")
    except (OSError, ValueError, yaml.YAMLError, json.JSONDecodeError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        return 2

    errors = compare_events_baseline(current_events, baseline_events)
    errors.extend(compare_jobs_baseline(current_jobs, baseline_jobs))
    errors.extend(validate_product_scope_removals(current_events))
    errors.extend(validate_jobs_contract_shape(current_jobs))
    errors.extend(validate_generated_contracts(root, current_events, current_jobs))
    errors.extend(scan_source_literals(root, current_events, current_jobs))
    if errors:
        for error in errors:
            print(f"FAIL: {error}", file=sys.stderr)
        return 1
    print("OK: event/job baselines match current truth sources")
    return 0


if __name__ == "__main__":
    sys.exit(main())
