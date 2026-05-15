from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
API_PACKAGE = "github.com/monshunter/easyinterview/backend/cmd/api"
SCENARIOS = {
    "p0-048-practice-hint-assisted-across-goals": "TestE2EP0048PracticeHintAssistedAcrossGoals",
    "p0-049-practice-hint-strict-refusal": "TestE2EP0049PracticeHintStrictRefusalAcrossGoals",
    "p0-050-practice-hint-provenance-task-runs": "TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns",
    "p0-051-practice-hint-degrade-privacy": "TestE2EP0051PracticeHintDegradeAndPrivacy",
}


def test_backend_practice_003_trigger_scripts_preserve_go_test_exit_status() -> None:
    for scenario in SCENARIOS:
        trigger = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/trigger.sh"
        text = trigger.read_text(encoding="utf-8")

        assert "| tee" not in text
        assert "go_test_status=$?" in text
        assert 'exit "$go_test_status"' in text


def test_backend_practice_003_verify_scripts_require_passing_test_and_package() -> None:
    for scenario, test_name in SCENARIOS.items():
        verify = REPO_ROOT / "test/scenarios/e2e" / scenario / "scripts/verify.sh"
        text = verify.read_text(encoding="utf-8")

        assert f"=== RUN   {test_name}" in text
        assert f"--- PASS: {test_name}" in text
        assert f"^ok[[:space:]]+{API_PACKAGE}" in text
        assert "--- FAIL:" in text
        assert "no tests to run" in text
