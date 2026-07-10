import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
API_PACKAGE = "github.com/monshunter/easyinterview/backend/cmd/api"
CMD_API_EXACT_GO_SCENARIOS = {
    "p0-048-practice-hint-assisted-across-goals": "TestE2EP0048PracticeHintAssistedAcrossGoals",
    "p0-050-practice-hint-provenance-task-runs": "TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns",
    "p0-051-practice-hint-degrade-privacy": "TestE2EP0051PracticeHintDegradeAndPrivacy",
    "p0-070-practice-derived-plan-create-read-replay": "TestE2EP0070PracticeDerivedPlanCreateReadReplay",
    "p0-072-practice-derived-source-isolation-privacy": "TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy",
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


def test_resume_tailor_negative_gates_are_contextual_and_ignore_tests() -> None:
    contextual = "(tailor|mode).*(inline|rewrite|mirror)|(inline|rewrite|mirror).*(tailor|mode)"
    scenarios = REPO_ROOT / "test/scenarios/e2e"
    for scenario in (
        "p0-077-resume-tailor-async-dispatch-and-ready",
        "p0-078-resume-tailor-failure-and-retry",
        "p0-080-resume-tailor-privacy-negative",
    ):
        verify = (scenarios / scenario / "scripts/verify.sh").read_text(encoding="utf-8")
        assert contextual in verify
        assert "--glob '!**/*_test.go'" in verify
