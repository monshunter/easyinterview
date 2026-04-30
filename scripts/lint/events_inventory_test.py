#!/usr/bin/env python3
"""Contract tests for scripts/lint/events_inventory.py."""

from __future__ import annotations

import copy
import importlib.util
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("events_inventory.py")


def load_linter():
    spec = importlib.util.spec_from_file_location("events_inventory_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def valid_events_data() -> dict:
    return {
        "version": 1,
        "schemaVersion": 1,
        "envelope": {
            "fields": [
                {"name": "eventId", "type": "$ref:b1.UUIDv7", "required": True},
                {"name": "eventName", "type": "dot_case_event_name", "required": True},
                {"name": "eventVersion", "type": "int", "required": True},
                {"name": "aggregateType", "type": "snake_case", "required": True},
                {"name": "aggregateId", "type": "$ref:b1.UUIDv7", "required": True},
                {"name": "occurredAt", "type": "rfc3339", "required": True},
                {
                    "name": "producer",
                    "type": "enum",
                    "required": True,
                    "values": ["api", "worker", "dispatcher", "review"],
                },
                {"name": "traceId", "type": "string", "required": False, "softRequired": True},
                {"name": "payload", "type": "polymorphic", "required": True},
            ]
        },
        "eventLocalEnums": [
            {
                "name": "TargetImportSourceType",
                "description": "Event-local enum; does not enter B1 or B2 public enums.",
                "values": ["url", "text", "file"],
            },
            {
                "name": "ResumeTailorMode",
                "description": "Event-local enum; does not enter B1 or B2 public enums.",
                "values": ["inline", "rewrite", "mirror"],
            },
            {
                "name": "SourceFreshnessStatus",
                "description": "Event-local enum; does not enter B1 or B2 public enums.",
                "values": ["fresh", "stale", "failed"],
            },
        ],
        "events": valid_event_entries(),
    }


EXPECTED_EVENT_PAYLOADS = {
    "target.import.requested": ["targetJobId", "userId", "sourceType", "targetLanguage"],
    "target.parsed": ["targetJobId", "userId", "analysisStatus", "requirementCount", "coreThemes"],
    "target.analysis.failed": ["targetJobId", "errorCode", "retryable"],
    "practice.session.started": ["sessionId", "planId", "targetJobId", "goal", "mode", "language"],
    "practice.turn.completed": ["sessionId", "turnId", "turnIndex", "questionIntent", "followUpCount", "answerCharLength"],
    "practice.session.completed": ["sessionId", "planId", "targetJobId", "turnCount", "language"],
    "report.generation.requested": ["reportId", "sessionId", "targetJobId"],
    "report.generated": ["reportId", "sessionId", "targetJobId", "preparednessLevel", "mistakeCount", "promptVersion", "rubricVersion", "modelId"],
    "report.generation.failed": ["reportId", "sessionId", "errorCode", "retryable"],
    "mistake.created": ["mistakeId", "targetJobId", "sourceSessionId", "competencyCode", "status", "priority"],
    "mistake.status.changed": ["mistakeId", "fromStatus", "toStatus", "targetJobId"],
    "resume.parse.completed": ["resumeAssetId", "userId", "parseStatus"],
    "resume.tailor.completed": ["tailorRunId", "resumeAssetId", "targetJobId", "mode", "status"],
    "debrief.created": ["debriefId", "targetJobId", "roundType", "questionCount"],
    "debrief.completed": ["debriefId", "targetJobId", "riskItemCount", "generatedMistakeCount"],
    "source.refreshed": ["sourceRecordId", "ownerType", "ownerId", "freshnessStatus"],
    "privacy.request.created": ["privacyRequestId", "userId", "requestType"],
    "privacy.request.completed": ["privacyRequestId", "userId", "requestType", "status"],
}


def valid_event_entries() -> list[dict]:
    def fields(names: list[str]) -> dict[str, dict[str, str]]:
        out: dict[str, dict[str, str]] = {}
        for name in names:
            out[name] = {"type": payload_type(name), "source": "spec:3.1.4"}
        return out

    aggregate_types = {
        "target.import.requested": "target_job",
        "target.parsed": "target_job",
        "target.analysis.failed": "target_job",
        "practice.session.started": "practice_session",
        "practice.turn.completed": "practice_turn",
        "practice.session.completed": "practice_session",
        "report.generation.requested": "feedback_report",
        "report.generated": "feedback_report",
        "report.generation.failed": "feedback_report",
        "mistake.created": "mistake_entry",
        "mistake.status.changed": "mistake_entry",
        "resume.parse.completed": "resume_asset",
        "resume.tailor.completed": "resume_tailor_run",
        "debrief.created": "debrief",
        "debrief.completed": "debrief",
        "source.refreshed": "source_record",
        "privacy.request.created": "privacy_request",
        "privacy.request.completed": "privacy_request",
    }
    producers = {
        "target.import.requested": "api",
        "target.parsed": "worker",
        "target.analysis.failed": "worker",
        "practice.session.started": "api",
        "practice.turn.completed": "api",
        "practice.session.completed": "api",
        "report.generation.requested": ["api", "dispatcher"],
        "report.generated": "worker",
        "report.generation.failed": "worker",
        "mistake.created": ["worker", "review"],
        "mistake.status.changed": "review",
        "resume.parse.completed": "worker",
        "resume.tailor.completed": "worker",
        "debrief.created": "api",
        "debrief.completed": "worker",
        "source.refreshed": "worker",
        "privacy.request.created": "api",
        "privacy.request.completed": "worker",
    }
    return [
        {
            "name": name,
            "version": 1,
            "producer": producers[name],
            "aggregateType": aggregate_types[name],
            "requiredPayload": event_payload_fields(name, payload_fields),
            "optionalPayload": {},
            "piiBoundary": "ids/status/counts only",
        }
        for name, payload_fields in EXPECTED_EVENT_PAYLOADS.items()
    ]


def event_payload_fields(event_name: str, names: list[str]) -> dict[str, dict[str, str]]:
    out = {name: {"type": payload_type(name), "source": "spec:3.1.4"} for name in names}
    overrides = {
        ("target.import.requested", "sourceType"): "$ref:event.TargetImportSourceType",
        ("practice.session.started", "mode"): "$ref:b1.PracticeMode",
        ("mistake.created", "status"): "$ref:b1.MistakeStatus",
        ("resume.tailor.completed", "mode"): "$ref:event.ResumeTailorMode",
        ("resume.tailor.completed", "status"): "$ref:b1.ReportStatus",
        ("source.refreshed", "freshnessStatus"): "$ref:event.SourceFreshnessStatus",
        ("privacy.request.completed", "status"): "$ref:b1.PrivacyRequestStatus",
    }
    for (event, field), typ in overrides.items():
        if event == event_name and field in out:
            out[field]["type"] = typ
    return out


def payload_type(name: str) -> str:
    uuid_fields = {
        "targetJobId",
        "userId",
        "sessionId",
        "planId",
        "turnId",
        "reportId",
        "mistakeId",
        "sourceSessionId",
        "resumeAssetId",
        "tailorRunId",
        "debriefId",
        "sourceRecordId",
        "ownerId",
        "privacyRequestId",
    }
    int_fields = {
        "requirementCount",
        "turnIndex",
        "followUpCount",
        "answerCharLength",
        "turnCount",
        "mistakeCount",
        "priority",
        "questionCount",
        "riskItemCount",
        "generatedMistakeCount",
    }
    bool_fields = {"retryable"}
    enum_refs = {
        "analysisStatus": "$ref:b1.TargetJobParseStatus",
        "goal": "$ref:b1.PracticeGoal",
        "mode": "string",
        "preparednessLevel": "$ref:b1.ReadinessTier",
        "status": "$ref:b1.PrivacyRequestStatus",
        "fromStatus": "$ref:b1.MistakeStatus",
        "toStatus": "$ref:b1.MistakeStatus",
        "parseStatus": "$ref:b1.TargetJobParseStatus",
        "roundType": "$ref:b1.InterviewerRole",
        "requestType": "$ref:b1.PrivacyRequestType",
    }
    if name in uuid_fields:
        return "uuidv7"
    if name in int_fields:
        return "int"
    if name in bool_fields:
        return "bool"
    if name == "coreThemes":
        return "string[]"
    return enum_refs.get(name, "string")


class EventsInventoryEnvelopeTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()

    def test_valid_envelope_contract_passes(self) -> None:
        self.assertEqual([], self.linter.validate_events_yaml(valid_events_data()))

    def test_requires_all_envelope_fields(self) -> None:
        data = valid_events_data()
        data["envelope"]["fields"] = [
            field for field in data["envelope"]["fields"] if field["name"] != "aggregateId"
        ]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("aggregateId" in err for err in errs), errs)

    def test_rejects_copied_uuid_regex_in_envelope(self) -> None:
        data = copy.deepcopy(valid_events_data())
        data["envelope"]["fields"][0]["type"] = "^[0-9a-f]{8}-regex-copy$"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("eventId" in err and "$ref:b1.UUIDv7" in err for err in errs), errs)

    def test_rejects_invalid_producer_enum(self) -> None:
        data = copy.deepcopy(valid_events_data())
        producer = next(field for field in data["envelope"]["fields"] if field["name"] == "producer")
        producer["values"] = ["api", "worker", "dispatcher", "review", "cron"]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("producer" in err and "api" in err and "cron" in err for err in errs), errs)

    def test_requires_full_18_event_inventory(self) -> None:
        data = valid_events_data()
        data["events"] = [event for event in data["events"] if event["name"] != "report.generated"]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generated" in err and "missing" in err for err in errs), errs)

    def test_requires_exact_payload_field_set(self) -> None:
        data = valid_events_data()
        report_generated = next(event for event in data["events"] if event["name"] == "report.generated")
        report_generated["requiredPayload"].pop("mistakeCount")

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generated" in err and "mistakeCount" in err for err in errs), errs)

    def test_rejects_invalid_event_domain(self) -> None:
        data = valid_events_data()
        data["events"][0]["name"] = "unknown.import.requested"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("unknown.import.requested" in err and "domain" in err for err in errs), errs)

    def test_rejects_snake_segment_event_name(self) -> None:
        data = valid_events_data()
        data["events"][7]["name"] = "report.generation_failed"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generation_failed" in err and "dot.case" in err for err in errs), errs)

    def test_rejects_non_whitelisted_past_tense_verb(self) -> None:
        data = valid_events_data()
        data["events"][1]["name"] = "target.imported"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("target.imported" in err and "past-tense" in err for err in errs), errs)

    def test_requires_b1_enum_payload_fields_to_use_aliases(self) -> None:
        data = valid_events_data()
        report_generated = next(event for event in data["events"] if event["name"] == "report.generated")
        report_generated["requiredPayload"]["preparednessLevel"]["type"] = "string"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("preparednessLevel" in err and "$ref:b1.ReadinessTier" in err for err in errs), errs)

    def test_requires_event_local_enum_declarations(self) -> None:
        data = valid_events_data()
        data["eventLocalEnums"] = [
            enum for enum in data["eventLocalEnums"] if enum["name"] != "ResumeTailorMode"
        ]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("ResumeTailorMode" in err and "eventLocalEnums" in err for err in errs), errs)

    def test_rejects_event_local_enum_collision_with_b1(self) -> None:
        conventions = {"enums": [{"name": "TargetImportSourceType"}]}

        errs = self.linter.validate_events_yaml(valid_events_data(), conventions)

        self.assertTrue(any("TargetImportSourceType" in err and "B1" in err for err in errs), errs)

    def test_requires_event_local_payload_fields_to_use_aliases(self) -> None:
        data = valid_events_data()
        target_import = next(event for event in data["events"] if event["name"] == "target.import.requested")
        target_import["requiredPayload"]["sourceType"]["type"] = "string"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("sourceType" in err and "$ref:event.TargetImportSourceType" in err for err in errs), errs)


if __name__ == "__main__":
    unittest.main()
