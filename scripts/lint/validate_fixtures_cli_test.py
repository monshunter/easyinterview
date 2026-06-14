#!/usr/bin/env python3
"""CLI-level fault-injection tests for scripts/lint/validate_fixtures.py.

Each test copies the real openapi/ tree into a tempdir, mutates exactly one
fixture, and asserts that the validator (a) exits non-zero and (b) names the
specific operationId / rule that fired. Reverting the mutation makes the
validator pass again — this proves Phase 1.4's fault-injection contract.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT = REPO_ROOT / "scripts" / "lint" / "validate_fixtures.py"


def _run_validator(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(SCRIPT), "--repo-root", str(repo)],
        capture_output=True,
        text=True,
        check=False,
    )


class ValidatorCliTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self.tmp.name) / "repo"
        # Copy only the surfaces the validator reads.
        shutil.copytree(REPO_ROOT / "openapi", self.repo / "openapi")

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def _read(self, rel: str) -> dict:
        with (self.repo / rel).open("r", encoding="utf-8") as f:
            return json.load(f)

    def _write(self, rel: str, data: dict) -> None:
        with (self.repo / rel).open("w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            f.write("\n")

    # ---- baseline ----
    def test_clean_fixtures_exit_zero(self) -> None:
        out = _run_validator(self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        self.assertIn("43", out.stdout)

    # ---- §1.3.5 missing operation ----
    def test_missing_fixture_fails(self) -> None:
        target = self.repo / "openapi/fixtures/Auth/getMe.json"
        target.unlink()
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("getMe", out.stderr + out.stdout)

    # ---- §1.3.2 AI provenance ----
    def test_missing_provenance_field_fails(self) -> None:
        rel = "openapi/fixtures/Reports/getFeedbackReport.json"
        data = self._read(rel)
        del data["scenarios"]["default"]["response"]["body"]["provenance"]["modelId"]
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("getFeedbackReport", out.stderr + out.stdout)
        self.assertIn("modelId", out.stderr + out.stdout)

    # ---- §1.3.4 UUIDv7 / tmp_ scan ----
    def test_tmp_prefix_id_fails(self) -> None:
        rel = "openapi/fixtures/TargetJobs/getTargetJob.json"
        data = self._read(rel)
        data["scenarios"]["default"]["response"]["body"]["id"] = "tmp_not_a_uuid"
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("tmp_", out.stderr + out.stdout)
        self.assertIn("getTargetJob", out.stderr + out.stdout)

    # ---- §1.3.1 status code white-list ----
    def test_privacy_export_must_be_501(self) -> None:
        rel = "openapi/fixtures/Privacy/requestPrivacyExport.json"
        data = self._read(rel)
        data["scenarios"]["default"]["response"]["status"] = 202
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("requestPrivacyExport", out.stderr + out.stdout)

    # ---- §1.3.3 privacy email allowlist ----
    def test_real_email_domain_fails(self) -> None:
        rel = "openapi/fixtures/Auth/getMe.json"
        data = self._read(rel)
        data["scenarios"]["default"]["response"]["body"]["emailMasked"] = "alice@gmail.com"
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("getMe", out.stderr + out.stdout)

    # ---- §1.3.1 operationId mismatch ----
    def test_operation_id_filename_mismatch_fails(self) -> None:
        rel = "openapi/fixtures/Auth/getMe.json"
        data = self._read(rel)
        data["operationId"] = "imposter"
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("operationId", out.stderr + out.stdout)

    def test_d20_retired_resume_fixture_key_fails(self) -> None:
        rel = "openapi/fixtures/Debriefs/suggestDebriefQuestions.json"
        data = self._read(rel)
        body = data["scenarios"]["default"]["request"]["body"]
        body["resumeVersionId"] = body.pop("resumeId")
        self._write(rel, data)

        out = _run_validator(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("suggestDebriefQuestions", out.stderr + out.stdout)
        self.assertIn("resumeVersionId", out.stderr + out.stdout)
        self.assertIn("D-20 flat resume", out.stderr + out.stdout)

    def test_missing_required_practice_session_scenario_fails(self) -> None:
        rel = "openapi/fixtures/PracticeSessions/appendSessionEvent.json"
        data = self._read(rel)
        del data["scenarios"]["mismatch"]
        self._write(rel, data)
        out = _run_validator(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("appendSessionEvent", out.stderr + out.stdout)
        self.assertIn("missing required scenarios", out.stderr + out.stdout)

    def test_validate_fixtures_uses_openapi_as_operation_inventory(self) -> None:
        openapi_path = self.repo / "openapi/openapi.yaml"
        openapi_text = openapi_path.read_text(encoding="utf-8")
        insertion = """
  /diagnostics/ping:
    get:
      tags: [Auth]
      operationId: getDiagnosticsPing
      summary: Temporary test-only additive operation
      responses:
        '204':
          description: No content.
        default:
          $ref: '#/components/responses/ApiErrorResponse'
"""
        openapi_path.write_text(
            openapi_text.replace("\ncomponents:\n", f"\n{insertion}\ncomponents:\n"),
            encoding="utf-8",
        )
        fixture = {
            "operationId": "getDiagnosticsPing",
            "scenarios": {
                "default": {
                    "response": {
                        "status": 204,
                        "headers": {"X-Request-ID": "req_2026-04-28T13-45-12-abcdef"},
                    }
                }
            },
        }
        path = self.repo / "openapi/fixtures/Auth/getDiagnosticsPing.json"
        with path.open("w", encoding="utf-8") as f:
            json.dump(fixture, f, indent=2, ensure_ascii=False)
            f.write("\n")

        out = _run_validator(self.repo)

        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        self.assertIn("44", out.stdout)

    def test_fixture_without_openapi_operation_fails(self) -> None:
        extra = self.repo / "openapi/fixtures/Growth/getGrowthOverview.json"
        extra.parent.mkdir(exist_ok=True)
        with extra.open("w", encoding="utf-8") as f:
            json.dump(
                {
                    "operationId": "getGrowthOverview",
                    "scenarios": {"default": {"response": {"status": 200}}},
                },
                f,
                indent=2,
            )
            f.write("\n")

        out = _run_validator(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("getGrowthOverview", out.stderr + out.stdout)
        self.assertIn("not present in openapi.yaml", out.stderr + out.stdout)


if __name__ == "__main__":
    unittest.main()
