import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
PRACTICE_SERVICE_SCENARIOS = {
    "p0-070-practice-derived-plan-create-read-replay": (
        "TestCreateDerivedPracticePlanIdempotencyMismatchHasNoSecondInsertOrLeak",
        "TestPracticeChatV020CandidateUsesSemanticFocus",
        "TestE2EP0070PracticeDerivedPlanCreateReadReplay",
    ),
    "p0-072-practice-derived-source-isolation-privacy": (
        "TestCreateDerivedPracticePlanRejectsCopiedServerFields",
        "TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy",
    ),
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
    for scenario, test_names in PRACTICE_SERVICE_SCENARIOS.items():
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")
        for test_name in test_names:
            assert test_name in text
        assert "set -euo pipefail" in text


def test_practice_service_verifiers_require_passing_named_tests() -> None:
    for scenario, test_names in PRACTICE_SERVICE_SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        for test_name in test_names:
            assert f"--- PASS: {test_name}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text


def test_report_derived_practice_scenarios_consume_f3_and_legacy_negative_markers() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    p0070 = (
        scenarios
        / "p0-070-practice-derived-plan-create-read-replay/scripts/verify.sh"
    ).read_text(encoding="utf-8")
    p0072 = (
        scenarios
        / "p0-072-practice-derived-source-isolation-privacy/scripts/verify.sh"
    ).read_text(encoding="utf-8")

    assert "PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS" in p0070
    assert "F3_PRACTICE_SEMANTIC_FOCUS_MARKER_CONSUMED_PASS" in p0070
    assert "REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS" in p0072
    assert "PROTOTYPE_MAPPING.md" in p0072
    for legacy in (
        "focusCompetencyCodes",
        "focus_competency_codes",
        "retryFocusCompetencyCodes",
        "retry_focus_competency_codes",
        "retryFocusTurnIds",
        "retry_focus_turn_ids",
        "retry_round",
        "semantic_focus",
    ):
        assert legacy in p0072


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
            "TestE2EP0047RejectsZeroAnswerCompletion",
            "TestE2EP0047FreezesReportContext",
            "TestE2EP0047CompletionReplayPreservesReportContext",
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


def test_p0047_verifier_owns_redacted_completion_evidence_artifact() -> None:
    scenario = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
    )
    setup = (scenario / "scripts/setup.sh").read_text(encoding="utf-8")
    trigger = (scenario / "scripts/trigger.sh").read_text(encoding="utf-8")
    verify = (scenario / "scripts/verify.sh").read_text(encoding="utf-8")
    cleanup = (scenario / "scripts/cleanup.sh").read_text(encoding="utf-8")

    assert "run_correlation_id" in setup
    assert "TestIntegrationE2EP0047RejectsZeroAnswerCompletion" in trigger
    assert "practice-completion-evidence.v1" in verify
    assert "completion-backend-evidence.json" in verify
    for key in (
        "schemaVersion",
        "scenarioId",
        "command",
        "tests",
        "markers",
        "database",
        "result",
    ):
        assert key in verify
    for marker in (
        "ZERO_ANSWER_COMPLETION_REJECTED_PASS",
        "REPORT_CONTEXT_SNAPSHOT_PASS",
        "REPORT_CONTEXT_REPLAY_PASS",
        "concurrent_mutation_blocked=true",
        "snapshot_replay_equal=true",
        "mismatch_side_effect_count=0",
    ):
        assert marker in verify
    assert "PracticeScreen.test.tsx" in trigger
    assert "TestE2EP0047RejectsZeroAnswerCompletion" in trigger
    assert "*.log" in cleanup
    assert "completion-backend-evidence.json" not in cleanup


def test_report_scenarios_require_current_backend_evidence_contract() -> None:
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    required = {
        "p0-056-generating-to-report-happy-path": {
            "test": "TestE2EP0056ReportBackendEvidence",
            "scenario_id": "E2E.P0.056",
            "evidence_schema": "report-backend-evidence.v1",
            "markers": (
                "REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS",
                "REPORT_DIRECT_READY_PASS",
                "REPORT_FROZEN_CONTEXT_READ_PASS",
                "REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS",
            ),
        },
        "p0-058-report-failure-and-missing-session": {
            "test": "TestE2EP0058ReportFailureBackendEvidence",
            "scenario_id": "E2E.P0.058",
            "evidence_schema": "report-backend-evidence.v3",
            "markers": (
                "REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS",
                "REPORT_CONTEXT_TOO_LARGE_PASS",
                "REPORT_OUTPUT_RETRY_PASS",
                "REPORT_FOUR_INVALID_FAIL_CLOSED_PASS",
                "REPORT_ACTION_RETRY_RESET_PASS",
                "REPORT_RETRY_LAYER_SEPARATION_PASS",
            ),
        },
    }

    for scenario, contract in required.items():
        root = scenarios / scenario
        setup = (root / "scripts/setup.sh").read_text(encoding="utf-8")
        trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
        verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
        cleanup = (root / "scripts/cleanup.sh").read_text(encoding="utf-8")

        assert "completion-backend-evidence.json" in setup
        assert "practice-completion-evidence.v1" in setup
        for owner_marker in (
            "ZERO_ANSWER_COMPLETION_REJECTED_PASS",
            "REPORT_CONTEXT_SNAPSHOT_PASS",
            "REPORT_CONTEXT_REPLAY_PASS",
        ):
            assert owner_marker in setup

        assert "PRACTICE_COMPLETION_EVIDENCE_PATH" in trigger
        assert "./internal/review ./internal/store/review ./internal/api/reports" in trigger
        assert f"-run '^{contract['test']}$'" in trigger
        assert "-count=1 -v" in trigger

        assert f"=== RUN   {contract['test']}" in verify
        assert f"--- PASS: {contract['test']}" in verify
        for marker in contract["markers"]:
            assert marker in verify
        for field in (
            "schemaVersion",
            "scenarioId",
            "command",
            "tests",
            "consumedOwnerEvidence",
            "markers",
            "database",
            "result",
        ):
            assert field in verify
        assert contract["evidence_schema"] in verify
        assert contract["scenario_id"] in verify
        assert "backend-evidence.json" in verify
        assert "keys ==" in verify
        assert "--- FAIL:" in verify
        assert "no tests to run" in verify
        assert "raw_(cookie|jd|resume|transcript|prompt|output)" in verify

        if scenario == "p0-058-report-failure-and-missing-session":
            for runtime_field in (
                "runtime",
                "firstActionCallCount",
                "secondActionInitialAttempt",
                "retryStateDestroyedAfterAction",
                "actionRetryScheduleSeconds",
                "asyncAttemptsAffectProductAttempt",
            ):
                assert runtime_field in verify

        assert 'ARTIFACT="$OUT/backend-evidence.json"' not in trigger
        assert 'ARTIFACT="$OUT/backend-evidence.json"' not in cleanup
        assert "*.log" in cleanup


def test_p0059_requires_semantic_geometry_and_pixel_difference_evidence() -> None:
    root = (
        REPO_ROOT
        / "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative"
    )
    readme = (root / "README.md").read_text(encoding="utf-8")
    seed = (root / "data/seed-input.md").read_text(encoding="utf-8")
    expected = (root / "data/expected-outcome.md").read_text(encoding="utf-8")
    trigger = (root / "scripts/trigger.sh").read_text(encoding="utf-8")
    verify = (root / "scripts/verify.sh").read_text(encoding="utf-8")
    contract = "\n".join((readme, seed, expected, trigger, verify))

    for required in (
        "DOM/style/bbox",
        "pixelmatch",
        "changed-pixel ratio",
        "0.5%",
        "1440",
        "390",
        "prototype/formal/diff",
        "threshold 0.1",
        "tests/pixel-parity/generating.spec.ts",
        "tests/pixel-parity/report.spec.ts",
    ):
        assert required in contract

    stale = contract.lower()
    for forbidden in (
        "non-empty in-memory screenshot",
        "non-empty screenshot",
        "missing-session",
        "mobile-overflow",
    ):
        assert forbidden not in stale

    assert "E2E.P0.059: running Playwright pixel parity" in trigger
    assert "E2E.P0.059: Playwright pixel parity complete" in trigger
    assert "Playwright pixel parity pass marker missing" in verify
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
