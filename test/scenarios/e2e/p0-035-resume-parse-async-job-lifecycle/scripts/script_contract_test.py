#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import shutil
import subprocess
import unittest


SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[4]
OUT = ROOT / ".test-output" / "e2e" / "p0-035-resume-parse-async-job-lifecycle"
REQUIRED_TESTS = (
    "TestResumeParseRunnerHTTPScenario",
    "TestResumeParseRunnerRetryableFailureScenario",
    "TestBuildResumeRuntimeWiresRoutesRunnerAndDeterministicAI",
    "TestCatalogKeepsResumeParseOutputBudget",
    "TestParseHandlerRejectsDOCXUploadText",
    "TestParseHandlerRejectsUnreadablePDFText",
    "TestParseHandlerExtractsReadableUploadText",
    "TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox",
    "TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox",
    "TestParseHandlerRetriesFailedAssetBackToProcessing",
    "TestParseHandlerObservedAIWritesResumeTaskRunColumns",
    "TestParseHandlerPIIRedlineForLogsAuditTaskRunsAndOutbox",
    "TestParseHandlerPreservesInlineHeadingWordsInSourceSnapshot",
    "TestParseHandlerPreservesLongInputTailWithStructuredOnlyResponse",
    "TestParseHandlerRejectsLengthFinishReasonAndPreservesSourceSnapshot",
    "TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady",
    "TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically",
    "TestCompleteParseFailureCanPersistExtractedTextSnapshot",
    "TestCompleteParseFailureMarksFailedWithoutCompletedOutbox",
    "TestResumesIntegrationCRUDStateIsolationPaginationAndRollback",
)


class ScenarioScriptContractTest(unittest.TestCase):
    def setUp(self) -> None:
        OUT.mkdir(parents=True, exist_ok=True)
        self._backup = OUT / "trigger.log.backup-for-script-contract-test"
        self._trigger = OUT / "trigger.log"
        self._verify_backup = OUT / "verify.log.backup-for-script-contract-test"
        self._verify = OUT / "verify.log"
        if self._trigger.exists():
            shutil.copyfile(self._trigger, self._backup)
        if self._verify.exists():
            shutil.copyfile(self._verify, self._verify_backup)

    def tearDown(self) -> None:
        if self._backup.exists():
            shutil.move(self._backup, self._trigger)
        elif self._trigger.exists():
            self._trigger.unlink()
        if self._verify_backup.exists():
            shutil.move(self._verify_backup, self._verify)
        elif self._verify.exists():
            self._verify.unlink()

    def test_verify_rejects_failed_named_test_evidence(self) -> None:
        failed_test = "TestParseHandlerPreservesLongInputTailWithStructuredOnlyResponse"
        lines = [
            (
                f"--- FAIL: {test_name} (0.00s)"
                if test_name == failed_test
                else f"--- PASS: {test_name} (0.00s)"
            )
            for test_name in REQUIRED_TESTS
        ]
        lines.append("FAIL")
        self._trigger.write_text("\n".join(lines) + "\n", encoding="utf-8")

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
        self.assertIn("failed test evidence", result.stdout)


if __name__ == "__main__":
    unittest.main()
