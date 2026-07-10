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
                    "values": ["api", "backend_async", "dispatcher", "review"],
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
                "values": ["gap_review", "bullet_suggestions"],
            },
            {
                "name": "SourceFreshnessStatus",
                "description": "Event-local enum; does not enter B1 or B2 public enums.",
                "values": ["fresh", "stale", "failed"],
            },
        ],
        "events": valid_event_entries(),
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
        "triggerEventSemantic": "source_event_only",
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
    "source_refresh": {
        "asynqTask": "source.refresh",
        "apiFacing": False,
        "triggerEvent": "target.parsed",
        "ownerDomain": "C13",
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
EMAIL_DISPATCH_PAYLOAD = {
    "authChallengeId": "uuidv7",
    "userId": "uuidv7",
    "templateKey": "slug",
    "locale": "bcp47",
    "deliverySecretRef": "opaque_ref",
    "dedupeKey": "string",
}
EMAIL_DISPATCH_REDACTED = [
    "rawEmailCode",
    "emailVerificationUrl",
    "recipientEmail",
    "recipientEmailHash",
    "emailBody",
    "emailSubject",
]


def valid_jobs_data() -> dict:
    jobs = [
        {"canonical": canonical, **attrs}
        for canonical, attrs in EXPECTED_JOBS.items()
    ]
    email_dispatch = next(job for job in jobs if job["canonical"] == "email_dispatch")
    email_dispatch["payloadSchema"] = {
        field: {"type": typ}
        for field, typ in EMAIL_DISPATCH_PAYLOAD.items()
    }
    email_dispatch["redactedFields"] = EMAIL_DISPATCH_REDACTED.copy()
    return {
        "version": 1,
        "schemaVersion": 1,
        "apiFacingSubset": [
            "target_import",
            "resume_parse",
            "report_generate",
            "resume_tailor",
                    "privacy_export",
            "privacy_delete",
        ],
        "jobs": jobs,
    }


EXPECTED_EVENT_PAYLOADS = {
    "target.import.requested": ["targetJobId", "userId", "sourceType", "targetLanguage"],
    "target.parsed": ["targetJobId", "userId", "analysisStatus", "requirementCount", "coreThemes"],
    "target.analysis.failed": ["targetJobId", "errorCode", "retryable"],
    "practice.session.started": ["sessionId", "planId", "targetJobId", "goal", "mode", "language"],
    "practice.turn.completed": ["sessionId", "turnId", "turnIndex", "questionIntent", "followUpCount", "answerCharLength"],
    "practice.session.completed": ["sessionId", "planId", "targetJobId", "turnCount", "language"],
    "report.generation.requested": ["reportId", "sessionId", "targetJobId"],
    "report.generated": ["reportId", "sessionId", "targetJobId", "preparednessLevel", "questionIssueCount", "promptVersion", "rubricVersion", "modelId"],
    "report.generation.failed": ["reportId", "sessionId", "errorCode", "retryable"],
    "resume.parse.completed": ["resumeId", "userId", "parseStatus"],
    "resume.tailor.completed": ["tailorRunId", "resumeId", "targetJobId", "mode", "status"],
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
        "resume.parse.completed": "resume",
        "resume.tailor.completed": "resume",
        "source.refreshed": "source_record",
        "privacy.request.created": "privacy_request",
        "privacy.request.completed": "privacy_request",
    }
    producers = {
        "target.import.requested": "api",
        "target.parsed": "backend_async",
        "target.analysis.failed": "backend_async",
        "practice.session.started": "api",
        "practice.turn.completed": "api",
        "practice.session.completed": "api",
        "report.generation.requested": ["api", "dispatcher"],
        "report.generated": "backend_async",
        "report.generation.failed": "backend_async",
        "resume.parse.completed": "backend_async",
        "resume.tailor.completed": "backend_async",
        "source.refreshed": "backend_async",
        "privacy.request.created": "api",
        "privacy.request.completed": "backend_async",
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
        "resumeId",
        "tailorRunId",
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
        "questionIssueCount",
    }
    bool_fields = {"retryable"}
    enum_refs = {
        "analysisStatus": "$ref:b1.TargetJobParseStatus",
        "goal": "$ref:b1.PracticeGoal",
        "mode": "string",
        "preparednessLevel": "$ref:b1.ReadinessTier",
        "status": "$ref:b1.PrivacyRequestStatus",
        "parseStatus": "$ref:b1.TargetJobParseStatus",
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
        producer["values"] = ["api", "backend_async", "dispatcher", "review", "cron"]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("producer" in err and "api" in err and "cron" in err for err in errs), errs)

    def test_requires_full_14_event_inventory(self) -> None:
        data = valid_events_data()
        data["events"] = [event for event in data["events"] if event["name"] != "report.generated"]

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generated" in err and "missing" in err for err in errs), errs)

    def test_requires_exact_payload_field_set(self) -> None:
        data = valid_events_data()
        report_generated = next(event for event in data["events"] if event["name"] == "report.generated")
        report_generated["requiredPayload"].pop("questionIssueCount")

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generated" in err and "questionIssueCount" in err for err in errs), errs)

    def test_rejects_invalid_event_domain(self) -> None:
        data = valid_events_data()
        data["events"][0]["name"] = "unknown.import.requested"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("unknown.import.requested" in err and "domain" in err for err in errs), errs)

    def test_rejects_non_whitelisted_final_event_verb(self) -> None:
        data = valid_events_data()
        data["events"][7]["name"] = "report.generation_failed"

        errs = self.linter.validate_events_yaml(data)

        self.assertTrue(any("report.generation_failed" in err and "past-tense" in err for err in errs), errs)

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


class EventsInventoryJobsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()

    def test_shared_jobs_marks_report_generate_as_source_event_only(self) -> None:
        repo_root = SCRIPT.parents[2]
        data = self.linter._load_yaml(repo_root / "shared/jobs.yaml")
        report_generate = next(job for job in data["jobs"] if job["canonical"] == "report_generate")

        self.assertEqual("source_event_only", report_generate.get("triggerEventSemantic"))

    def test_valid_jobs_contract_passes(self) -> None:
        self.assertEqual([], self.linter.validate_jobs_yaml(valid_jobs_data(), valid_events_data()))

    def test_requires_all_8_canonical_jobs(self) -> None:
        data = valid_jobs_data()
        data["jobs"] = [job for job in data["jobs"] if job["canonical"] != "email_dispatch"]

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("email_dispatch" in err and "missing" in err for err in errs), errs)

    def test_rejects_wrong_asynq_task_mapping(self) -> None:
        data = valid_jobs_data()
        target_import = next(job for job in data["jobs"] if job["canonical"] == "target_import")
        target_import["asynqTask"] = "target.imports"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("target_import" in err and "target.import" in err for err in errs), errs)

    def test_rejects_invalid_job_priority(self) -> None:
        data = valid_jobs_data()
        report_generate = next(job for job in data["jobs"] if job["canonical"] == "report_generate")
        report_generate["priority"] = "urgent"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("report_generate" in err and "critical" in err for err in errs), errs)

    def test_rejects_unknown_trigger_event(self) -> None:
        data = valid_jobs_data()
        report_generate = next(job for job in data["jobs"] if job["canonical"] == "report_generate")
        report_generate["triggerEvent"] = "report.created"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("report_generate" in err and "triggerEvent" in err for err in errs), errs)

    def test_rejects_unknown_trigger_event_semantic(self) -> None:
        data = valid_jobs_data()
        report_generate = next(job for job in data["jobs"] if job["canonical"] == "report_generate")
        report_generate["triggerEventSemantic"] = "event_dispatches_job"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("report_generate" in err and "triggerEventSemantic" in err for err in errs), errs)

    def test_requires_report_generate_source_event_only_semantic(self) -> None:
        data = valid_jobs_data()
        report_generate = next(job for job in data["jobs"] if job["canonical"] == "report_generate")
        report_generate.pop("triggerEventSemantic")

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("report_generate" in err and "source_event_only" in err for err in errs), errs)

    def test_source_event_only_requires_api_facing_job(self) -> None:
        data = valid_jobs_data()
        source_refresh = next(job for job in data["jobs"] if job["canonical"] == "source_refresh")
        source_refresh["triggerEventSemantic"] = "source_event_only"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("source_refresh" in err and "apiFacing" in err for err in errs), errs)

    def test_source_event_only_requires_real_event_trigger(self) -> None:
        data = valid_jobs_data()
        resume_parse = next(job for job in data["jobs"] if job["canonical"] == "resume_parse")
        resume_parse["triggerEventSemantic"] = "source_event_only"

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("resume_parse" in err and "known eventName" in err for err in errs), errs)

    def test_requires_fixed_api_facing_subset(self) -> None:
        data = valid_jobs_data()
        data["apiFacingSubset"] = data["apiFacingSubset"] + ["email_dispatch"]

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("apiFacingSubset" in err and "email_dispatch" in err for err in errs), errs)

    def test_rejects_internal_only_job_marked_api_facing(self) -> None:
        data = valid_jobs_data()
        email_dispatch = next(job for job in data["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["apiFacing"] = True

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("email_dispatch" in err and "apiFacing" in err for err in errs), errs)

    def test_requires_email_dispatch_payload_schema(self) -> None:
        data = valid_jobs_data()
        email_dispatch = next(job for job in data["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["payloadSchema"].pop("deliverySecretRef")

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("deliverySecretRef" in err and "email_dispatch.payloadSchema" in err for err in errs), errs)

    def test_rejects_redacted_field_in_email_dispatch_payload_schema(self) -> None:
        data = valid_jobs_data()
        email_dispatch = next(job for job in data["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["payloadSchema"]["recipientEmail"] = {"type": "string"}

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("recipientEmail" in err and "redacted" in err for err in errs), errs)

    def test_requires_email_dispatch_redacted_fields(self) -> None:
        data = valid_jobs_data()
        email_dispatch = next(job for job in data["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["redactedFields"] = ["rawEmailCode"]

        errs = self.linter.validate_jobs_yaml(data, valid_events_data())

        self.assertTrue(any("redactedFields" in err and "emailBody" in err for err in errs), errs)


if __name__ == "__main__":
    unittest.main()
