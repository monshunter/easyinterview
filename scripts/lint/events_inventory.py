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
    "mistake",
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
EXPECTED_PRODUCERS = ["api", "worker", "dispatcher", "review"]
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
        "producer": "worker",
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
        "producer": "worker",
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
        "producer": "worker",
        "aggregateType": "feedback_report",
        "requiredPayload": {
            "reportId": "uuidv7",
            "sessionId": "uuidv7",
            "targetJobId": "uuidv7",
            "preparednessLevel": "$ref:b1.ReadinessTier",
            "mistakeCount": "int",
            "promptVersion": "string",
            "rubricVersion": "string",
            "modelId": "string",
        },
    },
    "report.generation.failed": {
        "producer": "worker",
        "aggregateType": "feedback_report",
        "requiredPayload": {
            "reportId": "uuidv7",
            "sessionId": "uuidv7",
            "errorCode": "string",
            "retryable": "bool",
        },
    },
    "mistake.created": {
        "producer": ["worker", "review"],
        "aggregateType": "mistake_entry",
        "requiredPayload": {
            "mistakeId": "uuidv7",
            "targetJobId": "uuidv7",
            "sourceSessionId": "uuidv7",
            "competencyCode": "string",
            "status": "$ref:b1.MistakeStatus",
            "priority": "int",
        },
    },
    "mistake.status.changed": {
        "producer": "review",
        "aggregateType": "mistake_entry",
        "requiredPayload": {
            "mistakeId": "uuidv7",
            "fromStatus": "$ref:b1.MistakeStatus",
            "toStatus": "$ref:b1.MistakeStatus",
            "targetJobId": "uuidv7",
        },
    },
    "resume.parse.completed": {
        "producer": "worker",
        "aggregateType": "resume_asset",
        "requiredPayload": {
            "resumeAssetId": "uuidv7",
            "userId": "uuidv7",
            "parseStatus": "$ref:b1.TargetJobParseStatus",
        },
    },
    "resume.tailor.completed": {
        "producer": "worker",
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
        "producer": "worker",
        "aggregateType": "debrief",
        "requiredPayload": {
            "debriefId": "uuidv7",
            "targetJobId": "uuidv7",
            "riskItemCount": "int",
            "generatedMistakeCount": "int",
        },
    },
    "source.refreshed": {
        "producer": "worker",
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
        "producer": "worker",
        "aggregateType": "privacy_request",
        "requiredPayload": {
            "privacyRequestId": "uuidv7",
            "userId": "uuidv7",
            "requestType": "$ref:b1.PrivacyRequestType",
            "status": "$ref:b1.PrivacyRequestStatus",
        },
    },
}


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


def _load_yaml(path: Path) -> dict[str, Any]:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def main() -> int:
    if len(sys.argv) not in (2, 3):
        print("usage: events_inventory.py shared/events.yaml [shared/jobs.yaml]", file=sys.stderr)
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
    if len(sys.argv) == 3:
        conventions_path = Path(sys.argv[2])
        if not conventions_path.exists():
            print(f"FAIL: {conventions_path} does not exist", file=sys.stderr)
            return 2
        try:
            conventions = _load_yaml(conventions_path)
        except (OSError, ValueError, yaml.YAMLError) as exc:
            print(f"FAIL: {conventions_path}: {exc}", file=sys.stderr)
            return 2

    errors = validate_events_yaml(events, conventions)
    if errors:
        for err in errors:
            print(f"FAIL: {err}", file=sys.stderr)
        return 1

    print(f"OK: {events_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
