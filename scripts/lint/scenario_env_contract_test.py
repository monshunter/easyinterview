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
    helper_name = "_shared/scripts/local-dev-runtime.sh"
    for script_name in (*ENV_SCRIPTS.values(), helper_name):
        script = SCENARIO_ROOT / script_name
        assert script.exists(), f"{script_name} must be a top-level scenario env entrypoint"
        assert os.access(script, os.X_OK), f"{script_name} must be executable for skill usage"

        text = script.read_text(encoding="utf-8")
        if script_name != helper_name:
            assert "--dry-run" in text
        assert "p0-" not in text
        assert "manual-uat/full-funnel" not in text
        assert "backend/cmd/devsession" not in text
        assert "backend/internal/devsession" not in text

        subprocess.run(["bash", "-n", str(script)], cwd=REPO_ROOT, check=True)


def test_top_level_scenario_env_scripts_support_dry_run() -> None:
    cases = [
        (SCENARIO_ROOT / ENV_SCRIPTS["setup"], ["--dry-run"]),
        (SCENARIO_ROOT / ENV_SCRIPTS["setup"], ["--with-migrations", "--dry-run"]),
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
    helper = (SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh").read_text(
        encoding="utf-8"
    )

    assert "go build ./cmd/..." in text
    assert "pnpm --filter @easyinterview/frontend build" in text
    assert "deploy/dev-stack/.env" in text
    assert "VITE_EI_API_MODE" in text
    assert "VITE_EI_API_BASE_URL" in text
    assert "set -a" in text
    assert "make dev-up" in text
    assert "make dev-doctor" in text
    assert "restart_backend_runtime" in text
    assert "restart_frontend_runtime" in text
    assert "local_dev_summary" in text
    assert "go run ./backend/cmd/api" in helper
    assert "pnpm --filter @easyinterview/frontend dev" in helper
    assert "start_new_session=True" in helper
    assert ".test-output/local-dev/backend.log" in helper
    assert ".test-output/local-dev/frontend.log" in helper
    assert "backend API" in helper
    assert "frontend dev" in helper
    assert "Mailpit" in helper


def test_setup_migrations_load_dev_stack_env_and_derive_database_url() -> None:
    text = (SCENARIO_ROOT / ENV_SCRIPTS["setup"]).read_text(encoding="utf-8")

    assert "deploy/dev-stack/.env" in text
    assert "set -a" in text
    assert "POSTGRES_HOST_PORT" in text
    assert "POSTGRES_USER" in text
    assert "POSTGRES_PASSWORD" in text
    assert "POSTGRES_DB" in text
    assert "urllib.parse" in text
    assert "make migrate-up" in text
    assert 'DATABASE_URL="${DATABASE_URL:-' not in text
    assert "local_dev_summary" in text


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
    assert ".test-output/local-dev/backend.log" in dev_stack
    assert ".test-output/local-dev/frontend.log" in dev_stack
    assert "test/scenarios/env-redeploy.sh all" in dev_stack


def test_real_provider_hybrid_uat_is_registered_as_e2e_scenario() -> None:
    index = (SCENARIO_ROOT / "e2e" / "INDEX.md").read_text(encoding="utf-8")
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    readme = (scenario_dir / "README.md").read_text(encoding="utf-8")

    assert "E2E.P0.100" in index
    assert "`p0-100-real-provider-full-funnel-hybrid/`" in index
    assert "| hybrid | Ready |" in index
    assert "../../../../docs/spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md" in readme
    assert "](../../../docs/spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md)" not in readme

    assert scenario_dir.exists()
    for rel_path in (
        "README.md",
        "scripts/setup.sh",
        "scripts/trigger.sh",
        "scripts/verify.sh",
        "scripts/cleanup.sh",
        "data/seed-input.md",
        "data/expected-outcome.md",
    ):
        assert (scenario_dir / rel_path).exists(), f"{rel_path} is required for E2E.P0.100"

    assert not (SCENARIO_ROOT / "manual-uat").exists()
    assert not (scenario_dir / "env-template" / "dev-real.env.example").exists()


def test_real_provider_hybrid_uat_uses_dev_stack_env_as_single_source() -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    dev_env_example = (REPO_ROOT / "deploy" / "dev-stack" / ".env.example").read_text(
        encoding="utf-8"
    )

    for required in (
        "SESSION_COOKIE_SECRET=",
        "AUTH_CHALLENGE_TOKEN_PEPPER=",
        "AI_PROVIDER_BASE_URL=https://api.deepseek.com",
        "AI_PROVIDER_API_KEY=",
        "AI_DEBUG_PRINT_RAW_OUTPUT=true",
        "VITE_EI_API_MODE=real",
        "VITE_EI_API_BASE_URL=http://127.0.0.1:8080/api/v1",
    ):
        assert required in dev_env_example

    for path in (
        scenario_dir / "README.md",
        scenario_dir / "data" / "seed-input.md",
        scenario_dir / "scripts" / "setup.sh",
        scenario_dir / "scripts" / "trigger.sh",
    ):
        text = path.read_text(encoding="utf-8")
        assert "dev-real.env" not in text
        assert "env-template/dev-real.env.example" not in text

    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    assert 'DEV_STACK_ENV="$REPO_ROOT/deploy/dev-stack/.env"' in trigger
    assert "AI_DEBUG_PRINT_RAW_OUTPUT" in trigger
    assert 'AI_DEBUG_PRINT_RAW_OUTPUT:-' in trigger
    assert "LOCAL_ENV=" not in trigger


def test_real_provider_hybrid_evidence_must_be_current_and_redacted() -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    setup = (scenario_dir / "scripts" / "setup.sh").read_text(encoding="utf-8")
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
    readme = (scenario_dir / "README.md").read_text(encoding="utf-8")

    assert "evidence.md" in setup
    assert "rm -f" in setup
    assert "RUN_ID=" in setup
    assert "RUN_ID" in trigger
    assert "evidence run_id does not match setup run" in trigger
    assert "AI_PROVIDER_API_KEY" in trigger
    assert "prompt body" in trigger
    assert "response body" in trigger
    assert "sk-" in trigger
    assert "scan_evidence_redline" in trigger
    assert "scan_evidence_redline" in verify
    assert "run_id" in readme


def test_real_provider_hybrid_owner_docs_capture_current_evidence_gates() -> None:
    plan_dir = (
        REPO_ROOT
        / "docs"
        / "spec"
        / "e2e-scenarios-p0"
        / "plans"
        / "002-manual-uat-real-provider-full-funnel"
    )
    plan = (plan_dir / "plan.md").read_text(encoding="utf-8")
    checklist = (plan_dir / "checklist.md").read_text(encoding="utf-8")
    bdd_plan = (plan_dir / "bdd-plan.md").read_text(encoding="utf-8")
    bdd_checklist = (plan_dir / "bdd-checklist.md").read_text(encoding="utf-8")

    combined_docs = "\n".join((plan, checklist, bdd_plan, bdd_checklist))

    for required in (
        "RUN_ID",
        "run_id",
        "evidence redline",
        "evidence.md",
        "scan_evidence_redline",
        "env consumer gate",
        "env-setup.sh --with-migrations",
        "env-redeploy.sh frontend",
        "deploy/dev-stack/.env",
    ):
        assert required in combined_docs


def test_scenario_run_skill_requires_env_preflight_and_hybrid_results() -> None:
    skill = (REPO_ROOT / ".agent-skills" / "scenario-run" / "SKILL.md").read_text(
        encoding="utf-8"
    )

    assert "test/scenarios/env-setup.sh" in skill
    assert "test/scenarios/env-verify.sh" in skill
    assert "MANUAL_REQUIRED" in skill
    assert "hybrid" in skill
