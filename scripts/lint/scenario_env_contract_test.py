import os
import subprocess
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCENARIO_ROOT = REPO_ROOT / "test" / "scenarios"

ENV_SCRIPTS = {
    "setup": "env-setup.sh",
    "status": "env-status.sh",
    "verify": "env-verify.sh",
    "cleanup": "env-cleanup.sh",
    "redeploy": "env-redeploy.sh",
}


def run_script(script: Path, *args: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [str(script), *args],
        cwd=REPO_ROOT,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )


def test_top_level_scenario_env_scripts_exist_and_are_decoupled() -> None:
    for script_name in ENV_SCRIPTS.values():
        script = SCENARIO_ROOT / script_name
        assert script.exists(), f"{script_name} must be a top-level scenario env entrypoint"
        assert os.access(script, os.X_OK), f"{script_name} must be executable for skill usage"

        text = script.read_text(encoding="utf-8")
        assert "--dry-run" in text
        assert "p0-" not in text
        assert "manual-uat/full-funnel" not in text
        assert "backend/cmd/devsession" not in text
        assert "backend/internal/devsession" not in text

        subprocess.run(["bash", "-n", str(script)], cwd=REPO_ROOT, check=True)


def test_top_level_scenario_env_scripts_support_dry_run() -> None:
    cases = [
        (SCENARIO_ROOT / ENV_SCRIPTS["setup"], ["--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["status"], ["--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["verify"], ["--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["cleanup"], ["--dry-run", "--with-volumes"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["redeploy"], ["backend", "--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["redeploy"], ["frontend", "--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["redeploy"], ["all", "--dry-run"]),
    ]

    for script, args in cases:
        result = run_script(script, *args)
        assert result.returncode == 0, result.stderr
        output = result.stdout + result.stderr
        assert "dry-run" in output.lower()


def test_redeploy_script_documents_host_run_artifact_boundary() -> None:
    text = (SCENARIO_ROOT / ENV_SCRIPTS["redeploy"]).read_text(encoding="utf-8")

    assert "go build ./cmd/..." in text
    assert "pnpm --filter @easyinterview/frontend build" in text
    assert "make dev-up" in text
    assert "make dev-doctor" in text
    assert "go run ./backend/cmd/api" not in text
    assert "pnpm --filter @easyinterview/frontend dev" not in text


def test_root_makefile_exposes_scenario_env_targets() -> None:
    makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")
    target_to_script = {
        "scenario-env-setup": "test/scenarios/env-setup.sh",
        "scenario-env-status": "test/scenarios/env-status.sh",
        "scenario-env-verify": "test/scenarios/env-verify.sh",
        "scenario-env-cleanup": "test/scenarios/env-cleanup.sh",
        "scenario-env-redeploy": "test/scenarios/env-redeploy.sh",
    }

    for target, script in target_to_script.items():
        assert target in makefile
        assert script in makefile

    dry_run_cases = [
        ("scenario-env-setup", ["ARGS=--dry-run"], "test/scenarios/env-setup.sh"),
        ("scenario-env-status", ["ARGS=--dry-run"], "test/scenarios/env-status.sh"),
        ("scenario-env-verify", ["ARGS=--dry-run"], "test/scenarios/env-verify.sh"),
        ("scenario-env-cleanup", ["ARGS=--dry-run --with-volumes"], "test/scenarios/env-cleanup.sh"),
        ("scenario-env-redeploy", ["TARGET=backend", "ARGS=--dry-run"], "test/scenarios/env-redeploy.sh"),
    ]

    for target, variables, expected_script in dry_run_cases:
        result = subprocess.run(
            ["make", "-n", target, *variables],
            cwd=REPO_ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )
        assert result.returncode == 0, result.stderr
        assert expected_script in result.stdout


def test_scenario_env_skills_prefer_top_level_env_entrypoints() -> None:
    scenario_env = (REPO_ROOT / ".agent-skills" / "scenario-env" / "SKILL.md").read_text(
        encoding="utf-8"
    )
    scenario_redeploy = (
        REPO_ROOT / ".agent-skills" / "scenario-redeploy" / "SKILL.md"
    ).read_text(encoding="utf-8")

    for script_name in ENV_SCRIPTS.values():
        assert f"test/scenarios/{script_name}" in scenario_env

    assert "scenario-env-redeploy" in scenario_env
    assert "rebuild" in scenario_env.lower()
    assert "host-run" in scenario_env
    assert "specific scenario" in scenario_env.lower() or "具体场景" in scenario_env

    assert "test/scenarios/env-redeploy.sh" in scenario_redeploy
    assert "deps|backend|frontend|all" in scenario_redeploy
    assert "host-run" in scenario_redeploy
    assert "Kind" in scenario_redeploy and "Helm" in scenario_redeploy


def test_scenario_docs_describe_independent_env_lifecycle() -> None:
    framework = (SCENARIO_ROOT / "README.md").read_text(encoding="utf-8")
    suite = (SCENARIO_ROOT / "e2e" / "README.md").read_text(encoding="utf-8")
    dev_stack = (REPO_ROOT / "deploy" / "dev-stack" / "README.md").read_text(encoding="utf-8")

    for text in (framework, suite, dev_stack):
        assert "test/scenarios/env-setup.sh" in text
        assert "test/scenarios/env-redeploy.sh" in text
        assert "scenario-env" in text
        assert "具体场景" in text or "specific scenario" in text.lower()

    assert "手动引导" in suite
    assert "host-run" in dev_stack
    assert "make scenario-env-setup" in dev_stack
