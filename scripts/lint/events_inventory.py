#!/usr/bin/env python3
"""Validate the B3 event/job contract truth sources.

Run: `python3 scripts/lint/events_inventory.py shared/events.yaml [shared/jobs.yaml]`
"""
from __future__ import annotations

import sys
import re
from pathlib import Path
from typing import Any

import yaml


EVENT_NAME_RE = re.compile(r"^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*){1,2}$")
ALLOWED_EVENT_DOMAINS = {
    "target",
    "practice",
    "report",
    "resume",
    "debrief",
    "source",
    "privacy",
}
ALLOWED_PAST_TENSE_VERBS = {
    "requested",
    "parsed",
    "started",
    "completed",
    "generated",
    "failed",
    "created",
    "changed",
    "refreshed",
}
EXPECTED_ENVELOPE_FIELDS = {
    "eventId": {"type": "$ref:b1.UUIDv7", "required": True},
    "eventName": {"type": "dot_case_event_name", "required": True},
    "eventVersion": {"type": "int", "required": True},
    "aggregateType": {"type": "snake_case", "required": True},
    "aggregateId": {"type": "$ref:b1.UUIDv7", "required": True},
    "occurredAt": {"type": "rfc3339", "required": True},
    "producer": {"type": "enum", "required": True},
    "traceId": {"type": "string", "required": False, "softRequired": True},
    "payload": {"type": "polymorphic", "required": True},
}
EXPECTED_PRODUCERS = ["api", "backend_async", "dispatcher", "review"]
EXPECTED_EVENT_LOCAL_ENUMS = {
    "TargetImportSourceType": ["url", "text", "file"],
    "ResumeTailorMode": ["inline", "rewrite", "mirror"],
    "SourceFreshnessStatus": ["fresh", "stale", "failed"],
}
EXPECTED_EVENTS = {
    "target.import.requested": {
        "producer": "api",
        "aggregateType": "target_job",
        "requiredPayload": {
            "targetJobId": "uuidv7",
            "userId": "uuidv7",
            "sourceType": "$ref:event.TargetImportSourceType",
            "targetLanguage": "string",
        },
    },
    "target.parsed": {
        "producer": "backend_async",
        "aggregateType": "target_job",
        "requiredPayload": {
            "targetJobId": "uuidv7",
            "userId": "uuidv7",
            "analysisStatus": "$ref:b1.TargetJobParseStatus",
            "requirementCount": "int",
            "coreThemes": "string[]",
        },
    },
    "target.analysis.failed": {
        "producer": "backend_async",
        "aggregateType": "target_job",
        "requiredPayload": {
            "targetJobId": "uuidv7",
            "errorCode": "string",
            "retryable": "bool",
        },
    },
    "practice.session.started": {
        "producer": "api",
        "aggregateType": "practice_session",
        "requiredPayload": {
            "sessionId": "uuidv7",
            "planId": "uuidv7",
            "targetJobId": "uuidv7",
            "goal": "$ref:b1.PracticeGoal",
            "mode": "$ref:b1.PracticeMode",
            "language": "string",
        },
    },
    "practice.turn.completed": {
        "producer": "api",
        "aggregateType": "practice_turn",
        "requiredPayload": {
            "sessionId": "uuidv7",
            "turnId": "uuidv7",
            "turnIndex": "int",
            "questionIntent": "string",
            "followUpCount": "int",
            "answerCharLength": "int",
        },
    },
    "practice.session.completed": {
        "producer": "api",
        "aggregateType": "practice_session",
        "requiredPayload": {
            "sessionId": "uuidv7",
            "planId": "uuidv7",
            "targetJobId": "uuidv7",
            "turnCount": "int",
            "language": "string",
        },
    },
    "report.generation.requested": {
        "producer": ["api", "dispatcher"],
        "aggregateType": "feedback_report",
        "requiredPayload": {
            "reportId": "uuidv7",
            "sessionId": "uuidv7",
            "targetJobId": "uuidv7",
        },
    },
    "report.generated": {
        "producer": "backend_async",
        "aggregateType": "feedback_report",
        "requiredPayload": {
            "reportId": "uuidv7",
            "sessionId": "uuidv7",
            "targetJobId": "uuidv7",
            "preparednessLevel": "$ref:b1.ReadinessTier",
            "questionIssueCount": "int",
            "promptVersion": "string",
            "rubricVersion": "string",
            "modelId": "string",
        },
    },
    "report.generation.failed": {
        "producer": "backend_async",
        "aggregateType": "feedback_report",
        "requiredPayload": {
            "reportId": "uuidv7",
            "sessionId": "uuidv7",
            "errorCode": "string",
            "retryable": "bool",
        },
    },
    "resume.parse.completed": {
        "producer": "backend_async",
        "aggregateType": "resume_asset",
        "requiredPayload": {
            "resumeAssetId": "uuidv7",
            "userId": "uuidv7",
            "parseStatus": "$ref:b1.TargetJobParseStatus",
        },
    },
    "resume.tailor.completed": {
        "producer": "backend_async",
        "aggregateType": "resume_tailor_run",
        "requiredPayload": {
            "tailorRunId": "uuidv7",
            "resumeAssetId": "uuidv7",
            "targetJobId": "uuidv7",
            "mode": "$ref:event.ResumeTailorMode",
            "status": "$ref:b1.ReportStatus",
        },
    },
    "debrief.created": {
        "producer": "api",
        "aggregateType": "debrief",
        "requiredPayload": {
            "debriefId": "uuidv7",
            "targetJobId": "uuidv7",
            "roundType": "$ref:b1.InterviewerRole",
            "questionCount": "int",
        },
    },
    "debrief.completed": {
        "producer": "backend_async",
        "aggregateType": "debrief",
        "requiredPayload": {
            "debriefId": "uuidv7",
            "targetJobId": "uuidv7",
            "riskItemCount": "int",
            "practiceFocusCount": "int",
        },
    },
    "source.refreshed": {
        "producer": "backend_async",
        "aggregateType": "source_record",
        "requiredPayload": {
            "sourceRecordId": "uuidv7",
            "ownerType": "string",
            "ownerId": "uuidv7",
            "freshnessStatus": "$ref:event.SourceFreshnessStatus",
        },
    },
    "privacy.request.created": {
        "producer": "api",
        "aggregateType": "privacy_request",
        "requiredPayload": {
            "privacyRequestId": "uuidv7",
            "userId": "uuidv7",
            "requestType": "$ref:b1.PrivacyRequestType",
        },
    },
    "privacy.request.completed": {
        "producer": "backend_async",
        "aggregateType": "privacy_request",
        "requiredPayload": {
            "privacyRequestId": "uuidv7",
            "userId": "uuidv7",
            "requestType": "$ref:b1.PrivacyRequestType",
            "status": "$ref:b1.PrivacyRequestStatus",
        },
    },
}
EXPECTED_JOBS = {
    "target_import": {
        "asynqTask": "target.import",
        "apiFacing": True,
        "triggerEvent": "target.import.requested",
        "ownerDomain": "C4",
        "priority": "default",
    },
    "resume_parse": {
        "asynqTask": "resume.parse",
        "apiFacing": True,
        "triggerEvent": "api:register_resume",
        "ownerDomain": "C7",
        "priority": "default",
    },
    "report_generate": {
        "asynqTask": "report.generate",
        "apiFacing": True,
        "triggerEvent": "practice.session.completed",
        "ownerDomain": "C6",
        "priority": "critical",
    },
    "resume_tailor": {
        "asynqTask": "resume.tailor",
        "apiFacing": True,
        "triggerEvent": "api:request_tailor",
        "ownerDomain": "C7",
        "priority": "default",
    },
    "debrief_generate": {
        "asynqTask": "debrief.generate",
        "apiFacing": True,
        "triggerEvent": "debrief.created",
        "ownerDomain": "C9",
        "priority": "default",
    },
    "source_refresh": {
        "asynqTask": "source.refresh",
        "apiFacing": False,
        "triggerEvent": "target.parsed",
        "ownerDomain": "C13",
        "priority": "low",
    },
    "embedding_upsert": {
        "asynqTask": "embedding.upsert",
        "apiFacing": False,
        "triggerEvent": "target.parsed",
        "ownerDomain": "C11",
        "priority": "low",
    },
    "privacy_export": {
        "asynqTask": "privacy.export",
        "apiFacing": True,
        "triggerEvent": "privacy.request.created",
        "ownerDomain": "C12",
        "priority": "low",
    },
    "privacy_delete": {
        "asynqTask": "privacy.delete",
        "apiFacing": True,
        "triggerEvent": "privacy.request.created",
        "ownerDomain": "C8",
        "priority": "critical",
    },
    "email_dispatch": {
        "asynqTask": "email.dispatch",
        "apiFacing": False,
        "triggerEvent": "api:auth_email_start",
        "ownerDomain": "C1+C8",
        "priority": "low",
    },
}
EXPECTED_API_FACING_SUBSET = [
    "target_import",
    "resume_parse",
    "report_generate",
    "resume_tailor",
    "debrief_generate",
    "privacy_export",
    "privacy_delete",
]
EMAIL_DISPATCH_PAYLOAD = {
    "authChallengeId": "uuidv7",
    "userId": "uuidv7",
    "templateKey": "slug",
    "locale": "bcp47",
    "deliverySecretRef": "opaque_ref",
    "dedupeKey": "string",
}
EMAIL_DISPATCH_REDACTED_FIELDS = [
    "rawMagicLinkToken",
    "magicLinkUrl",
    "recipientEmail",
    "recipientEmailHash",
    "emailBody",
    "emailSubject",
]


def _field_map(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    fields = ((data.get("envelope") or {}).get("fields") or [])
    if not isinstance(fields, list):
        return {}
    out: dict[str, dict[str, Any]] = {}
    for field in fields:
        if isinstance(field, dict) and isinstance(field.get("name"), str):
            out[field["name"]] = field
    return out


def validate_events_yaml(data: dict[str, Any], conventions_data: dict[str, Any] | None = None) -> list[str]:
    errors: list[str] = []
    fields = _field_map(data)

    missing = set(EXPECTED_ENVELOPE_FIELDS) - set(fields)
    if missing:
        errors.append(f"envelope.fields missing required fields: {sorted(missing)}")

    for name, expected in EXPECTED_ENVELOPE_FIELDS.items():
        field = fields.get(name)
        if not field:
            continue
        for key, want in expected.items():
            got = field.get(key)
            if got != want:
                errors.append(f"envelope field {name}.{key} must be {want!r}, got {got!r}")

    for uuid_field in ("eventId", "aggregateId"):
        field = fields.get(uuid_field)
        if field and field.get("type") != "$ref:b1.UUIDv7":
            errors.append(f"envelope field {uuid_field} must use $ref:b1.UUIDv7 instead of copying UUIDv7 regex")

    producer = fields.get("producer") or {}
    if producer.get("values") != EXPECTED_PRODUCERS:
        errors.append(
            "envelope producer values must be "
            f"{EXPECTED_PRODUCERS!r}, got {producer.get('values')!r}"
        )

    errors.extend(_validate_event_local_enums(data.get("eventLocalEnums") or [], conventions_data))
    errors.extend(_validate_events(data.get("events") or []))
    return errors


def _validate_event_local_enums(enums: Any, conventions_data: dict[str, Any] | None) -> list[str]:
    errors: list[str] = []
    if not isinstance(enums, list):
        return ["eventLocalEnums must be a list"]

    by_name: dict[str, dict[str, Any]] = {}
    for enum in enums:
        if not isinstance(enum, dict) or not isinstance(enum.get("name"), str):
            errors.append("each eventLocalEnums entry must be a mapping with name")
            continue
        name = enum["name"]
        if name in by_name:
            errors.append(f"duplicate eventLocalEnums name: {name}")
        by_name[name] = enum

    missing = set(EXPECTED_EVENT_LOCAL_ENUMS) - set(by_name)
    extra = set(by_name) - set(EXPECTED_EVENT_LOCAL_ENUMS)
    if missing:
        errors.append(f"eventLocalEnums missing expected names: {sorted(missing)}")
    if extra:
        errors.append(f"eventLocalEnums include unexpected names: {sorted(extra)}")

    for name, expected_values in EXPECTED_EVENT_LOCAL_ENUMS.items():
        enum = by_name.get(name)
        if not enum:
            continue
        if enum.get("values") != expected_values:
            errors.append(f"eventLocalEnums.{name}.values must be {expected_values!r}, got {enum.get('values')!r}")
        description = str(enum.get("description") or "")
        if "B1" not in description or "B2" not in description:
            errors.append(f"eventLocalEnums.{name}.description must state it does not enter B1 / B2 public enums")

    if conventions_data:
        b1_enum_names = {
            enum.get("name")
            for enum in (conventions_data.get("enums") or [])
            if isinstance(enum, dict)
        }
        collisions = set(EXPECTED_EVENT_LOCAL_ENUMS) & b1_enum_names
        if collisions:
            errors.append(f"event-local enum names must not collide with B1 enums: {sorted(collisions)}")

    return errors


def _validate_events(events: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(events, list):
        return ["events must be a list"]

    by_name: dict[str, dict[str, Any]] = {}
    for event in events:
        if not isinstance(event, dict) or not isinstance(event.get("name"), str):
            errors.append("each event must be a mapping with name")
            continue
        name = event["name"]
        errors.extend(_validate_event_name(name))
        if name in by_name:
            errors.append(f"duplicate event name: {name}")
        by_name[name] = event

    missing = set(EXPECTED_EVENTS) - set(by_name)
    extra = set(by_name) - set(EXPECTED_EVENTS)
    if missing:
        errors.append(f"events missing expected names: {sorted(missing)}")
    if extra:
        errors.append(f"events include unexpected names: {sorted(extra)}")

    for name, expected in EXPECTED_EVENTS.items():
        event = by_name.get(name)
        if not event:
            continue
        if event.get("version") != 1:
            errors.append(f"{name}.version must be 1, got {event.get('version')!r}")
        if _producer_list(event.get("producer")) != _producer_list(expected["producer"]):
            errors.append(f"{name}.producer must be {expected['producer']!r}, got {event.get('producer')!r}")
        if event.get("aggregateType") != expected["aggregateType"]:
            errors.append(
                f"{name}.aggregateType must be {expected['aggregateType']!r}, got {event.get('aggregateType')!r}"
            )
        optional_payload = event.get("optionalPayload")
        if optional_payload != {}:
            errors.append(f"{name}.optionalPayload must be empty object in v1, got {optional_payload!r}")
        if not event.get("piiBoundary"):
            errors.append(f"{name}.piiBoundary must be non-empty")
        errors.extend(_validate_payload(name, event.get("requiredPayload"), expected["requiredPayload"]))

    return errors


def _validate_event_name(name: str) -> list[str]:
    errors: list[str] = []
    if not EVENT_NAME_RE.match(name):
        errors.append(f"{name} must be dot.case with 2 or 3 lowercase alphanumeric segments")
        return errors
    parts = name.split(".")
    if parts[0] not in ALLOWED_EVENT_DOMAINS:
        errors.append(f"{name} domain must be one of {sorted(ALLOWED_EVENT_DOMAINS)}")
    if parts[-1] not in ALLOWED_PAST_TENSE_VERBS:
        errors.append(f"{name} final segment must be a whitelisted past-tense verb")
    return errors


def _producer_list(value: Any) -> list[str]:
    if isinstance(value, str):
        return [value]
    if isinstance(value, list):
        return value
    return []


def _validate_payload(name: str, actual_payload: Any, expected_payload: dict[str, str]) -> list[str]:
    errors: list[str] = []
    if not isinstance(actual_payload, dict):
        return [f"{name}.requiredPayload must be a mapping"]
    actual_fields = set(actual_payload)
    expected_fields = set(expected_payload)
    if actual_fields != expected_fields:
        errors.append(
            f"{name}.requiredPayload fields must be {sorted(expected_fields)}, got {sorted(actual_fields)}"
        )
    for field_name, expected_type in expected_payload.items():
        field = actual_payload.get(field_name)
        if not isinstance(field, dict):
            continue
        got_type = field.get("type")
        if got_type != expected_type:
            errors.append(f"{name}.requiredPayload.{field_name}.type must be {expected_type!r}, got {got_type!r}")
        if not field.get("source"):
            errors.append(f"{name}.requiredPayload.{field_name}.source must be non-empty")
    return errors


def validate_jobs_yaml(data: dict[str, Any], events_data: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    jobs = data.get("jobs") or []
    if not isinstance(jobs, list):
        return ["jobs must be a list"]

    by_canonical: dict[str, dict[str, Any]] = {}
    for job in jobs:
        if not isinstance(job, dict) or not isinstance(job.get("canonical"), str):
            errors.append("each job must be a mapping with canonical")
            continue
        canonical = job["canonical"]
        if canonical in by_canonical:
            errors.append(f"duplicate canonical job: {canonical}")
        by_canonical[canonical] = job

    missing = set(EXPECTED_JOBS) - set(by_canonical)
    extra = set(by_canonical) - set(EXPECTED_JOBS)
    if missing:
        errors.append(f"jobs missing expected canonical values: {sorted(missing)}")
    if extra:
        errors.append(f"jobs include unexpected canonical values: {sorted(extra)}")

    subset = data.get("apiFacingSubset")
    if subset != EXPECTED_API_FACING_SUBSET:
        errors.append(
            f"apiFacingSubset must be {EXPECTED_API_FACING_SUBSET!r}, got {subset!r}"
        )

    event_names = {
        event.get("name")
        for event in (events_data.get("events") or [])
        if isinstance(event, dict)
    }
    for canonical, expected in EXPECTED_JOBS.items():
        job = by_canonical.get(canonical)
        if not job:
            continue
        for key, want in expected.items():
            got = job.get(key)
            if got != want:
                errors.append(f"{canonical}.{key} must be {want!r}, got {got!r}")
        if canonical in EXPECTED_API_FACING_SUBSET and job.get("apiFacing") is not True:
            errors.append(f"{canonical}.apiFacing must be true because it is in apiFacingSubset")
        if canonical not in EXPECTED_API_FACING_SUBSET and job.get("apiFacing") is not False:
            errors.append(f"{canonical}.apiFacing must be false because it is internal-only")
        trigger = job.get("triggerEvent")
        if isinstance(trigger, str) and not trigger.startswith("api:") and trigger not in event_names:
            errors.append(f"{canonical}.triggerEvent must reference a known eventName or api: source, got {trigger!r}")

    email_dispatch = by_canonical.get("email_dispatch")
    if email_dispatch:
        errors.extend(_validate_email_dispatch(email_dispatch))

    return errors


def _validate_email_dispatch(job: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    payload_schema = job.get("payloadSchema")
    if not isinstance(payload_schema, dict):
        return ["email_dispatch.payloadSchema must be a mapping"]
    actual_fields = set(payload_schema)
    expected_fields = set(EMAIL_DISPATCH_PAYLOAD)
    redacted_fields = set(EMAIL_DISPATCH_REDACTED_FIELDS)
    if actual_fields != expected_fields:
        errors.append(
            f"email_dispatch.payloadSchema fields must be {sorted(expected_fields)}, got {sorted(actual_fields)}"
        )
    forbidden = actual_fields & redacted_fields
    if forbidden:
        errors.append(f"email_dispatch.payloadSchema contains redacted fields: {sorted(forbidden)}")
    for field, expected_type in EMAIL_DISPATCH_PAYLOAD.items():
        entry = payload_schema.get(field)
        if isinstance(entry, dict) and entry.get("type") != expected_type:
            errors.append(f"email_dispatch.payloadSchema.{field}.type must be {expected_type!r}, got {entry.get('type')!r}")

    redacted = job.get("redactedFields")
    if redacted != EMAIL_DISPATCH_REDACTED_FIELDS:
        errors.append(
            f"email_dispatch.redactedFields must be {EMAIL_DISPATCH_REDACTED_FIELDS!r}, got {redacted!r}"
        )
    return errors


def _load_yaml(path: Path) -> dict[str, Any]:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def main() -> int:
    if len(sys.argv) < 2 or len(sys.argv) > 4:
        print("usage: events_inventory.py shared/events.yaml [shared/jobs.yaml] [shared/conventions.yaml]", file=sys.stderr)
        return 2

    events_path = Path(sys.argv[1])
    if not events_path.exists():
        print(f"FAIL: {events_path} does not exist", file=sys.stderr)
        return 2

    try:
        events = _load_yaml(events_path)
    except (OSError, ValueError, yaml.YAMLError) as exc:
        print(f"FAIL: {events_path}: {exc}", file=sys.stderr)
        return 2

    conventions = None
    jobs = None
    for raw_path in sys.argv[2:]:
        path = Path(raw_path)
        if not path.exists():
            print(f"FAIL: {path} does not exist", file=sys.stderr)
            return 2
        try:
            loaded = _load_yaml(path)
        except (OSError, ValueError, yaml.YAMLError) as exc:
            print(f"FAIL: {path}: {exc}", file=sys.stderr)
            return 2
        if path.name == "jobs.yaml":
            jobs = loaded
        elif path.name == "conventions.yaml":
            conventions = loaded
        else:
            print(f"FAIL: unsupported companion file {path}", file=sys.stderr)
            return 2

    errors = validate_events_yaml(events, conventions)
    if jobs is not None:
        errors.extend(validate_jobs_yaml(jobs, events))
    if errors:
        for err in errors:
            print(f"FAIL: {err}", file=sys.stderr)
        return 1

    print(f"OK: {events_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
