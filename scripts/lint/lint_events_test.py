#!/usr/bin/env python3
"""Contract tests for scripts/lint/lint_events.py."""

from __future__ import annotations

import copy
import importlib.util
import tempfile
import textwrap
import unittest
from pathlib import Path

from events_inventory_test import valid_events_data, valid_jobs_data


SCRIPT = Path(__file__).with_name("lint_events.py")


def load_linter():
    spec = importlib.util.spec_from_file_location("lint_events_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class LintEventsBaselineTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()
        self.events = valid_events_data()
        self.jobs = valid_jobs_data()

    def test_detects_deleted_event_name(self) -> None:
        current = copy.deepcopy(self.events)
        current["events"] = [event for event in current["events"] if event["name"] != "report.generated"]

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertTrue(any("report.generated" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_detects_required_payload_type_change(self) -> None:
        current = copy.deepcopy(self.events)
        report_generated = next(event for event in current["events"] if event["name"] == "report.generated")
        report_generated["requiredPayload"]["mistakeCount"]["type"] = "string"

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertTrue(any("mistakeCount" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_detects_required_payload_deleted(self) -> None:
        current = copy.deepcopy(self.events)
        report_generated = next(event for event in current["events"] if event["name"] == "report.generated")
        report_generated["requiredPayload"].pop("mistakeCount")

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertTrue(any("mistakeCount" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_detects_dot_case_event_renamed_to_snake(self) -> None:
        current = copy.deepcopy(self.events)
        report_generated = next(event for event in current["events"] if event["name"] == "report.generated")
        report_generated["name"] = "report_generated"

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertTrue(any("report.generated" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_detects_event_local_enum_member_removed(self) -> None:
        current = copy.deepcopy(self.events)
        source_type = next(enum for enum in current["eventLocalEnums"] if enum["name"] == "TargetImportSourceType")
        source_type["values"] = ["url", "text"]

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertTrue(any("TargetImportSourceType" in err and "file" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_allows_additive_optional_payload_field(self) -> None:
        current = copy.deepcopy(self.events)
        report_generated = next(event for event in current["events"] if event["name"] == "report.generated")
        report_generated["optionalPayload"]["reviewerNote"] = {"type": "string", "source": "spec:3.1.4"}

        errs = self.linter.compare_events_baseline(current, self.events)

        self.assertEqual([], errs)

    def test_detects_deleted_job_type(self) -> None:
        current = copy.deepcopy(self.jobs)
        current["jobs"] = [job for job in current["jobs"] if job["canonical"] != "email_dispatch"]

        errs = self.linter.compare_jobs_baseline(current, self.jobs)

        self.assertTrue(any("email_dispatch" in err and "breaking change requires eventVersion + 1" in err for err in errs), errs)

    def test_detects_internal_job_made_api_facing(self) -> None:
        current = copy.deepcopy(self.jobs)
        email_dispatch = next(job for job in current["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["apiFacing"] = True

        errs = self.linter.compare_jobs_baseline(current, self.jobs)

        self.assertTrue(any("email_dispatch" in err and "apiFacing" in err for err in errs), errs)


class LintEventsSourceScanTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()
        self.events = valid_events_data()
        self.jobs = valid_jobs_data()
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def write(self, relpath: str, content: str) -> None:
        path = self.root / relpath
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(textwrap.dedent(content).lstrip(), encoding="utf-8")

    def write_generated_events(self, omitted: set[str] | None = None) -> None:
        omitted = omitted or set()
        lines = [
            "// Code generated by backend/cmd/codegen/events; DO NOT EDIT.",
            "package events",
            "const (",
        ]
        for event in self.events["events"]:
            if event["name"] not in omitted:
                lines.append(f'\tEventName{constant_suffix(event["name"])} EventName = "{event["name"]}"')
        lines.append(")")
        self.write("backend/internal/shared/events/events.go", "\n".join(lines))

    def write_generated_jobs(self, api_facing: list[str] | None = None) -> None:
        api_facing = api_facing or self.jobs["apiFacingSubset"]
        lines = [
            "// Code generated by backend/cmd/codegen/events; DO NOT EDIT.",
            "package jobs",
            "const (",
        ]
        for job in self.jobs["jobs"]:
            lines.append(f'\tJobType{constant_suffix(job["canonical"])} JobType = "{job["canonical"]}"')
        lines.extend([")", "var APIFacingJobTypes = []JobType{"])
        for canonical in api_facing:
            lines.append(f"\tJobType{constant_suffix(canonical)},")
        lines.append("}")
        self.write("backend/internal/shared/jobs/jobs.go", "\n".join(lines))

    def test_rejects_backend_and_frontend_naked_literals(self) -> None:
        self.write("backend/internal/service/publisher.go", 'package service\nconst name = "target.import.requested"\n')
        self.write("frontend/src/features/jobs/job.ts", 'export const task = "email.dispatch";\n')

        errs = self.linter.scan_source_literals(self.root, self.events, self.jobs)

        self.assertTrue(any("publisher.go" in err and "target.import.requested" in err for err in errs), errs)
        self.assertTrue(any("job.ts" in err and "email.dispatch" in err for err in errs), errs)

    def test_allows_generated_docs_and_fixture_literals(self) -> None:
        self.write("backend/internal/shared/events/events.go", 'const EventNameReportGenerated = "report.generated"\n')
        self.write("frontend/src/lib/jobs/jobs.ts", 'export const JOB_TYPE_TARGET_IMPORT = "target_import" as const;\n')
        self.write("backend/internal/service/testdata/event.json", '{"eventName":"target.import.requested"}\n')
        self.write("docs/spec/event-and-outbox-contract/example.md", "`report.generated`\n")

        errs = self.linter.scan_source_literals(self.root, self.events, self.jobs)

        self.assertEqual([], errs)

    def test_rejects_handwritten_event_name_constant(self) -> None:
        self.write("backend/internal/service/names.go", 'package service\nconst EventNameCustomCreated = "custom.event.created"\n')

        errs = self.linter.scan_source_literals(self.root, self.events, self.jobs)

        self.assertTrue(any("EventNameCustomCreated" in err for err in errs), errs)

    def test_validates_generated_event_collection_matches_yaml(self) -> None:
        self.write_generated_events(omitted={"report.generated"})
        self.write_generated_jobs()

        errs = self.linter.validate_generated_contracts(self.root, self.events, self.jobs)

        self.assertTrue(any("event names" in err and "18" in err and "report.generated" in err for err in errs), errs)

    def test_rejects_missing_generated_contract_files(self) -> None:
        errs = self.linter.validate_generated_contracts(self.root, self.events, self.jobs)

        self.assertTrue(any("backend/internal/shared/events/events.go" in err and "missing" in err for err in errs), errs)
        self.assertTrue(any("frontend/src/lib/events/events.ts" in err and "missing" in err for err in errs), errs)
        self.assertTrue(any("backend/internal/shared/jobs/jobs.go" in err and "missing" in err for err in errs), errs)
        self.assertTrue(any("frontend/src/lib/jobs/jobs.ts" in err and "missing" in err for err in errs), errs)

    def test_validates_generated_api_facing_job_types_match_yaml(self) -> None:
        self.write_generated_events()
        self.write_generated_jobs(api_facing=self.jobs["apiFacingSubset"] + ["email_dispatch"])

        errs = self.linter.validate_generated_contracts(self.root, self.events, self.jobs)

        self.assertTrue(any("APIFacingJobTypes" in err and "email_dispatch" in err for err in errs), errs)

    def test_rejects_api_facing_subset_when_job_is_not_api_facing(self) -> None:
        jobs = copy.deepcopy(self.jobs)
        jobs["apiFacingSubset"] = jobs["apiFacingSubset"] + ["email_dispatch"]

        errs = self.linter.validate_jobs_contract_shape(jobs)

        self.assertTrue(any("email_dispatch" in err and "apiFacing" in err for err in errs), errs)

    def test_rejects_redacted_field_added_to_email_dispatch_payload_schema(self) -> None:
        jobs = copy.deepcopy(self.jobs)
        email_dispatch = next(job for job in jobs["jobs"] if job["canonical"] == "email_dispatch")
        email_dispatch["payloadSchema"]["recipientEmail"] = {"type": "string"}

        errs = self.linter.validate_jobs_contract_shape(jobs)

        self.assertTrue(any("recipientEmail" in err and "redacted" in err for err in errs), errs)


def constant_suffix(value: str) -> str:
    return "".join(part.capitalize() for part in value.replace(".", "_").split("_"))


if __name__ == "__main__":
    unittest.main()
