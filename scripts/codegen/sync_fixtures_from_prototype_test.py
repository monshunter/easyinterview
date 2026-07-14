#!/usr/bin/env python3
"""Tests for scripts/codegen/sync_fixtures_from_prototype.py.

Covers Phase 2.2 contract:
- Running the sync writes a `scenarios.prototype-baseline` to the 5 P0
  closed-loop endpoints listed in plan 2.4.
- Re-running is idempotent (`git diff --exit-code` clean).
- After sync, `validate_fixtures.py` exits 0 (default + prototype-baseline
  both schema-valid).
- A mapping gap (data.jsx without a required section) makes the sync exit
  non-zero with a clear `Mapping gap:` message.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SYNC = REPO_ROOT / "scripts" / "codegen" / "sync_fixtures_from_prototype.py"
VALIDATE = REPO_ROOT / "scripts" / "lint" / "validate_fixtures.py"

P0_BASELINE_OPS = (
    ("Auth", "getMe"),
    ("TargetJobs", "listTargetJobs"),
    ("TargetJobs", "getTargetJob"),
    ("PracticeSessions", "getPracticeSession"),
    ("Reports", "getFeedbackReport"),
)


def _run(script: Path, repo: Path, *extra: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(script), "--repo-root", str(repo), *extra],
        capture_output=True,
        text=True,
        check=False,
    )


def _read_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


class SyncFixturesFromPrototypeTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self.tmp.name) / "repo"
        for sub in ("openapi", "ui-design", "scripts"):
            shutil.copytree(REPO_ROOT / sub, self.repo / sub)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_sync_runs_clean_and_writes_p0_prototype_baseline(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        for tag, opid in P0_BASELINE_OPS:
            path = self.repo / "openapi/fixtures" / tag / f"{opid}.json"
            data = _read_json(path)
            scenarios = data.get("scenarios", {})
            with self.subTest(operationId=opid):
                self.assertIn(
                    "prototype-baseline", scenarios,
                    f"{opid}: prototype-baseline scenario must be present after sync",
                )
                # default must remain the first key.
                self.assertEqual(next(iter(scenarios)), "default")
                pb = scenarios["prototype-baseline"]
                response = pb.get("response", {})
                body = response.get("body")
                self.assertIsNotNone(body, f"{opid}: prototype-baseline.response.body required")
                self.assertNotEqual(body, {}, f"{opid}: prototype-baseline body must not be empty")

    def test_sync_is_idempotent(self) -> None:
        first = _run(SYNC, self.repo)
        self.assertEqual(first.returncode, 0)
        snapshots: dict[Path, str] = {}
        fixtures_root = self.repo / "openapi" / "fixtures"
        for tag, opid in P0_BASELINE_OPS:
            p = fixtures_root / tag / f"{opid}.json"
            snapshots[p] = p.read_text(encoding="utf-8")
        second = _run(SYNC, self.repo)
        self.assertEqual(second.returncode, 0)
        for path, before in snapshots.items():
            with self.subTest(path=str(path)):
                self.assertEqual(
                    before, path.read_text(encoding="utf-8"),
                    f"{path}: re-running sync must be idempotent",
                )

    def test_sync_output_passes_validate_fixtures(self) -> None:
        sync_out = _run(SYNC, self.repo)
        self.assertEqual(sync_out.returncode, 0)
        val = _run(VALIDATE, self.repo)
        self.assertEqual(
            val.returncode, 0,
            msg=f"validate-fixtures must pass after sync\nstderr={val.stderr}",
        )

    def test_list_target_jobs_progress_keeps_the_round_summary_needed_by_consumers(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0)
        fixture = _read_json(
            self.repo / "openapi/fixtures/TargetJobs/listTargetJobs.json"
        )
        items = fixture["scenarios"]["prototype-baseline"]["response"]["body"]["items"]
        for item in items:
            with self.subTest(targetJobId=item["id"]):
                progress = item.get("practiceProgress")
                if progress is None:
                    continue
                rounds = item.get("summary", {}).get("interviewRounds", [])
                refs = {
                    (f"round-{round_['sequence']}-{round_['type']}", round_["sequence"])
                    for round_ in rounds
                }
                self.assertGreaterEqual(len(rounds), 2)
                for completed in progress["completedRounds"]:
                    self.assertIn(
                        (completed["roundId"], completed["roundSequence"]), refs
                    )
                current = progress["currentRound"]
                if current is not None:
                    self.assertIn((current["roundId"], current["roundSequence"]), refs)

    def test_target_job_projection_never_restores_source_provenance(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")

        for operation_id in ("listTargetJobs", "getTargetJob"):
            fixture = _read_json(
                self.repo
                / "openapi/fixtures/TargetJobs"
                / f"{operation_id}.json"
            )
            body = fixture["scenarios"]["prototype-baseline"]["response"]["body"]
            encoded = json.dumps(body, ensure_ascii=False)
            with self.subTest(operationId=operation_id):
                self.assertNotIn('"sourceType"', encoded)
                self.assertNotIn('"sourceUrl"', encoded)

    def test_target_job_projection_never_restores_latest_report_pointer(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")

        for operation_id in ("listTargetJobs", "getTargetJob"):
            fixture = _read_json(
                self.repo
                / "openapi/fixtures/TargetJobs"
                / f"{operation_id}.json"
            )
            body = fixture["scenarios"]["prototype-baseline"]["response"]["body"]
            with self.subTest(operationId=operation_id):
                self.assertNotIn("latestReportId", json.dumps(body))

    def test_practice_projection_generates_deterministic_role_recovery_fields(self) -> None:
        first = _run(SYNC, self.repo)
        self.assertEqual(
            first.returncode,
            0,
            msg=f"stdout={first.stdout}\nstderr={first.stderr}",
        )
        fixture_path = (
            self.repo
            / "openapi/fixtures/PracticeSessions/getPracticeSession.json"
        )
        first_messages = _read_json(fixture_path)["scenarios"]["prototype-baseline"][
            "response"
        ]["body"]["messages"]
        users = [message for message in first_messages if message["role"] == "user"]
        assistants = [
            message for message in first_messages if message["role"] == "assistant"
        ]
        self.assertEqual(
            ["complete", "complete", "pending"],
            [message["replyStatus"] for message in users],
        )
        self.assertEqual(len(users), len({message["clientMessageId"] for message in users}))
        for message in users:
            self.assertRegex(
                message["clientMessageId"],
                r"^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
            )
        for message in assistants:
            self.assertNotIn("clientMessageId", message)
            self.assertNotIn("replyStatus", message)

        first_client_ids = [message["clientMessageId"] for message in users]
        second = _run(SYNC, self.repo)
        self.assertEqual(second.returncode, 0)
        second_messages = _read_json(fixture_path)["scenarios"]["prototype-baseline"][
            "response"
        ]["body"]["messages"]
        self.assertEqual(
            first_client_ids,
            [
                message["clientMessageId"]
                for message in second_messages
                if message["role"] == "user"
            ],
        )

    def test_feedback_report_projection_uses_direct_contract(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        fixture = _read_json(
            self.repo / "openapi/fixtures/Reports/getFeedbackReport.json"
        )
        body = fixture["scenarios"]["prototype-baseline"]["response"]["body"]

        self.assertEqual("ready", body["status"])
        self.assertIsNone(body["errorCode"])
        self.assertIsInstance(body["summary"], str)
        self.assertTrue(body["summary"])
        self.assertIn("context", body)
        self.assertIn("code", body["dimensionAssessments"][0])
        self.assertIn("label", body["dimensionAssessments"][0])
        self.assertIn("dimensionCode", body["highlights"][0])
        self.assertIn("retryFocusDimensionCodes", body)
        for old_field in (
            "questionAssessments",
            "retryFocusTurnIds",
            "retryFocusCompetencyCodes",
        ):
            self.assertNotIn(old_field, body)

    def test_sync_fails_fast_on_mapping_gap(self) -> None:
        # Drop the targetJobs section that listTargetJobs depends on.
        data_file = self.repo / "ui-design" / "src" / "data.jsx"
        text = data_file.read_text(encoding="utf-8")
        replaced = text.replace("targetJobs: [", "targetJobs__missing: [", 1)
        self.assertNotEqual(text, replaced, "Test setup: data.jsx must contain `targetJobs: [`")
        data_file.write_text(replaced, encoding="utf-8")
        out = _run(SYNC, self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("Mapping gap", out.stderr + out.stdout)
        self.assertIn("listTargetJobs", out.stderr + out.stdout)


if __name__ == "__main__":
    unittest.main()
