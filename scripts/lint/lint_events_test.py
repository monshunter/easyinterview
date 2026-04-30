#!/usr/bin/env python3
"""Contract tests for scripts/lint/lint_events.py."""

from __future__ import annotations

import copy
import importlib.util
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


if __name__ == "__main__":
    unittest.main()
