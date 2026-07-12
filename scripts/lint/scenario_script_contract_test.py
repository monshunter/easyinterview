import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
PRACTICE_SERVICE_SCENARIOS = {
    "p0-070-practice-derived-plan-create-read-replay": "TestCreateDerivedPracticePlanPassesReportSourceAndCompetencyFocus",
    "p0-072-practice-derived-source-isolation-privacy": "TestDerivedPracticePlanRequiresSourceReport",
}
FRONTEND_TEST_PATH_RE = re.compile(
    r"(?<![A-Za-z0-9_./-])((?:frontend/)?(?:src|tests)/[A-Za-z0-9_./-]+\.(?:test|spec)\.(?:tsx|ts))"
)


def test_explicit_frontend_test_paths_in_scenario_triggers_exist() -> None:
    missing: list[str] = []
    triggers = sorted((REPO_ROOT / "test/scenarios/e2e").glob("*/scripts/trigger.sh"))

    for trigger in triggers:
        text = trigger.read_text(encoding="utf-8")
        for reference in sorted(set(FRONTEND_TEST_PATH_RE.findall(text))):
            candidate = (
                REPO_ROOT / reference
                if reference.startswith("frontend/")
                else REPO_ROOT / "frontend" / reference
            )
            if not candidate.is_file():
                missing.append(f"{trigger.relative_to(REPO_ROOT)} -> {reference}")

    assert not missing, "scenario triggers reference missing frontend tests:\n" + "\n".join(
        missing
    )


def test_practice_service_triggers_run_current_named_tests() -> None:
    for scenario, test_name in PRACTICE_SERVICE_SCENARIOS.items():
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")
        assert test_name in text
        assert "set -euo pipefail" in text


def test_practice_service_verifiers_require_passing_named_tests() -> None:
    for scenario, test_name in PRACTICE_SERVICE_SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        assert f"--- PASS: {test_name}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text


def test_practice_failure_and_completion_scenarios_require_regression_markers() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-046-practice-text-loop-failure-and-recovery": (
            "TestSendPracticeMessageProviderFailureKeepsReservationUncommitted",
            "TestSendPracticeMessageExactReplayReturnsOriginalResultWithoutAICall",
            "TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess",
            "TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage",
            "TestSQLRepositoryReservePracticeMessageRejectsNewMessageWhileReplyPending",
        ),
        "p0-047-practice-text-loop-complete-and-generating-handoff": (
            "TestSQLRepositoryCommitPracticeMessageRejectsClosedSession",
        ),
    }

    for scenario, markers in required.items():
        trigger = (scenarios / scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (scenarios / scenario / "scripts/verify.sh").read_text(encoding="utf-8")
        for marker in markers:
            assert marker in trigger
            assert f"--- PASS: {marker}" in verify
        assert "--- FAIL:" in verify
        assert "no tests to run" in verify


def test_report_scenarios_require_candidate_score_regression_markers() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-056-generating-to-report-happy-path": "TestReadinessFromContentUsesCandidateScoreBoundaries",
        "p0-058-report-failure-and-missing-session": "TestGenerateReportRejectsInvalidDimensionScoresBeforePersistence",
    }

    for scenario, marker in required.items():
        trigger = (scenarios / scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (scenarios / scenario / "scripts/verify.sh").read_text(encoding="utf-8")
        assert marker in trigger
        assert f"--- PASS: {marker}" in verify
        assert "--- FAIL:" in verify
        assert "no tests to run" in verify


def test_resume_runtime_negative_gate_is_shared_and_ignores_tests() -> None:
    mode_pattern = "(tailor|mode).*(inline|rewrite|mirror)|(inline|rewrite|mirror).*(tailor|mode)"
    module_pattern = "mistakes|growth|drill|inline-debrief-record"
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    invocation = '"$ROOT/test/scenarios/_shared/scripts/resume-runtime-negative-gate.sh"'
    consumers = []
    for scenario in (
        "p0-075-resume-update-flat-fields-and-ik",
        "p0-076-resume-duplicate-save-as-new",
        "p0-077-resume-tailor-async-dispatch-and-ready",
        "p0-078-resume-tailor-failure-and-retry",
        "p0-079-resume-rewrites-accept-only-save",
        "p0-080-resume-tailor-privacy-negative",
    ):
        consumers.append(scenarios / scenario / "scripts/verify.sh")
    consumers.append(
        scenarios
        / "p0-080-resume-tailor-privacy-negative"
        / "scripts/trigger.sh"
    )

    for consumer in consumers:
        text = consumer.read_text(encoding="utf-8")
        assert invocation in text
        assert mode_pattern not in text
        assert module_pattern not in text

    gate = (
        REPO_ROOT
        / "test/scenarios/_shared/scripts/resume-runtime-negative-gate.sh"
    ).read_text(encoding="utf-8")
    assert mode_pattern in gate
    assert module_pattern in gate
    assert "--glob '!**/*_test.go'" in gate
    assert not (
        REPO_ROOT
        / "test/scenarios/_shared/scripts/resume-mode-negative-gate.sh"
    ).exists()
