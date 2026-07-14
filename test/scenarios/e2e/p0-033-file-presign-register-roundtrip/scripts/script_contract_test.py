#!/usr/bin/env python3
from __future__ import annotations

import os
import pathlib
import shutil
import subprocess
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]
OUT = ROOT / ".test-output" / "e2e" / "p0-033-file-presign-register-roundtrip"


class ScenarioScriptContractTest(unittest.TestCase):
    def setUp(self) -> None:
        OUT.mkdir(parents=True, exist_ok=True)
        self._backup = OUT / "trigger.log.backup-for-script-contract-test"
        self._trigger = OUT / "trigger.log"
        if self._trigger.exists():
            shutil.copyfile(self._trigger, self._backup)

    def tearDown(self) -> None:
        if self._backup.exists():
            shutil.move(self._backup, self._trigger)
        elif self._trigger.exists():
            self._trigger.unlink()

    def test_trigger_requires_live_database_and_object_storage_env(self) -> None:
        env = os.environ.copy()
        for key in (
            "DATABASE_URL",
            "OBJECT_STORAGE_ENDPOINT",
            "OBJECT_STORAGE_BUCKET",
            "OBJECT_STORAGE_ACCESS_KEY",
            "OBJECT_STORAGE_SECRET_KEY",
        ):
            env.pop(key, None)

        result = subprocess.run(
            [str(SCRIPT_DIR / "trigger.sh")],
            cwd=ROOT,
            env=env,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=60,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("DATABASE_URL", result.stdout)
        self.assertIn("OBJECT_STORAGE_ENDPOINT", result.stdout)

    def test_trigger_runs_live_roundtrip_gate(self) -> None:
        trigger = (SCRIPT_DIR / "trigger.sh").read_text(encoding="utf-8")

        self.assertIn("TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip", trigger)
        self.assertIn("-tags=integration", trigger)

    def test_verify_rejects_skipped_live_integration_checks(self) -> None:
        self._trigger.write_text(
            "\n".join(
                [
                    "=== RUN   TestCreateUploadPresign",
                    "=== RUN   TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject",
                    "=== RUN   TestRepositoryRegisterUploadedChecksObjectWhileRowLocked",
                    "=== RUN   TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware",
                    "=== RUN   TestDeleteFileObjectsForUser",
                    "=== RUN   TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey",
                    "--- SKIP: TestMinIO (0.00s)",
                    "PASS",
                ]
            ),
            encoding="utf-8",
        )

        result = subprocess.run(
            [str(SCRIPT_DIR / "verify.sh")],
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=60,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("live integration skip", result.stdout)

    def test_verify_rejects_missing_live_roundtrip_test_evidence(self) -> None:
        self._trigger.write_text(
            "\n".join(
                [
                    "=== RUN   TestCreateUploadPresign",
                    "=== RUN   TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject",
                    "=== RUN   TestRepositoryRegisterUploadedChecksObjectWhileRowLocked",
                    "=== RUN   TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware",
                    "=== RUN   TestDeleteFileObjectsForUser",
                    "=== RUN   TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey",
                    "PASS",
                ]
            ),
            encoding="utf-8",
        )

        result = subprocess.run(
            [str(SCRIPT_DIR / "verify.sh")],
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=60,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_verify_requires_phase7_boundary_evidence(self) -> None:
        self._trigger.write_text(
            "\n".join(
                [
                    "=== RUN   TestCreateUploadPresign",
                    "=== RUN   TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject",
                    "=== RUN   TestRepositoryRegisterUploadedChecksObjectWhileRowLocked",
                    "=== RUN   TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware",
                    "=== RUN   TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip",
                    "=== RUN   TestDeleteFileObjectsForUser",
                    "=== RUN   TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey",
                    "PASS",
                ]
            ),
            encoding="utf-8",
        )

        result = subprocess.run(
            [str(SCRIPT_DIR / "verify.sh")],
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=60,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_verify_rejects_no_tests_to_run(self) -> None:
        self._trigger.write_text(
            "\n".join(
                [
                    "=== RUN   TestCreateUploadPresign",
                    "testing: warning: no tests to run",
                    "PASS",
                ]
            ),
            encoding="utf-8",
        )

        result = subprocess.run(
            [str(SCRIPT_DIR / "verify.sh")],
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=60,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("matched no tests", result.stdout)


if __name__ == "__main__":
    unittest.main()
