from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
API_PACKAGE = "github.com/monshunter/easyinterview/backend/cmd/api"
CMD_API_EXACT_GO_SCENARIOS = {
    "p0-048-practice-hint-assisted-across-goals": "TestE2EP0048PracticeHintAssistedAcrossGoals",
    "p0-049-practice-hint-strict-refusal": "TestE2EP0049PracticeHintStrictRefusalAcrossGoals",
    "p0-050-practice-hint-provenance-task-runs": "TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns",
    "p0-051-practice-hint-degrade-privacy": "TestE2EP0051PracticeHintDegradeAndPrivacy",
    "p0-070-practice-derived-plan-create-read-replay": "TestE2EP0070PracticeDerivedPlanCreateReadReplay",
    "p0-071-practice-debrief-start-source-question": "TestE2EP0071PracticeDebriefStartUsesSourceQuestion",
    "p0-072-practice-derived-source-isolation-privacy": "TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy",
    "p0-073-practice-debrief-mode-regression": "TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative",
}

DEBRIEF_MULTI_PACKAGE_GO_SCENARIOS = {
    "p0-060-debrief-create-worker-happy": {
        "tests": (
            "TestCreateDebrief_HappyResponse",
            "TestGenerateHandler_HappyResolution",
            "TestStoreUpdateDebriefCompleted_HappyTransaction",
        ),
        "packages": (
            "github.com/monshunter/easyinterview/backend/internal/api/debriefs",
            "github.com/monshunter/easyinterview/backend/internal/debrief",
            "github.com/monshunter/easyinterview/backend/internal/store/debrief",
        ),
    },
    "p0-061-debrief-get-isolation": {
        "tests": (
            "TestStoreGetDebrief_DraftPartial",
            "TestServiceGetDebrief_ProvenanceWireOnly",
            "TestGetDebrief_CrossUser404",
        ),
        "packages": (
            "github.com/monshunter/easyinterview/backend/internal/store/debrief",
            "github.com/monshunter/easyinterview/backend/internal/debrief",
            "github.com/monshunter/easyinterview/backend/internal/api/debriefs",
        ),
    },
    "p0-062-debrief-worker-retry-failure": {
        "tests": (
            "TestGenerateHandler_A3Timeout",
            "TestGenerateHandler_PermanentFailAt5Attempts",
            "TestRetryPolicy_BackoffBelowMax",
        ),
        "packages": (
            "github.com/monshunter/easyinterview/backend/internal/debrief",
            "github.com/monshunter/easyinterview/backend/internal/targetjob",
        ),
    },
    "p0-063-debrief-suggest-questions": {
        "tests": (
            "TestServiceSuggestQuestions_Happy",
            "TestServiceSuggestQuestions_A3Timeout",
            "TestSuggestDebriefQuestions_CountBoundary",
        ),
        "packages": (
            "github.com/monshunter/easyinterview/backend/internal/debrief",
            "github.com/monshunter/easyinterview/backend/internal/api/debriefs",
        ),
    },
    "p0-064-debrief-privacy-legacy": {
        "tests": (
            "TestOutboxPayload_NoRawText",
            "TestGenerateHandler_OutboxPayloadSchema",
            "TestAuditEvents_NoRawText",
        ),
        "packages": (
            "github.com/monshunter/easyinterview/backend/internal/store/debrief",
            "github.com/monshunter/easyinterview/backend/internal/debrief",
        ),
    },
}


def test_cmd_api_exact_go_trigger_scripts_preserve_go_test_exit_status() -> None:
    for scenario in CMD_API_EXACT_GO_SCENARIOS:
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")

        assert "| tee" not in text
        assert "go_test_status=$?" in text
        assert 'exit "$go_test_status"' in text


def test_cmd_api_exact_go_verify_scripts_require_passing_test_and_package() -> None:
    for scenario, test_name in CMD_API_EXACT_GO_SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        assert f"=== RUN   {test_name}" in text
        assert f"--- PASS: {test_name}" in text
        assert f"^ok[[:space:]]+{API_PACKAGE}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text


def test_backend_debrief_trigger_scripts_preserve_go_test_exit_status() -> None:
    for scenario in DEBRIEF_MULTI_PACKAGE_GO_SCENARIOS:
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")

        assert "set -euo pipefail" in text
        assert "| tee" in text


def test_backend_debrief_verify_scripts_require_passing_tests_and_packages() -> None:
    for scenario, evidence in DEBRIEF_MULTI_PACKAGE_GO_SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        for test_name in evidence["tests"]:
            assert f"=== RUN   {test_name}" in text
            assert f"--- PASS: {test_name}" in text
        for package in evidence["packages"]:
            assert f"^ok[[:space:]]+{package}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text
