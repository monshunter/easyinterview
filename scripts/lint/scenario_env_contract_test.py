import os
import re
import subprocess
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCENARIO_ROOT = REPO_ROOT / "test" / "scenarios"
FRONTEND_REAL_BACKEND_VERIFY = (
    SCENARIO_ROOT / "_shared" / "scripts" / "frontend-real-backend-verify.sh"
)

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


def valid_frontend_vitest_log() -> str:
    return "\n".join(
        (
            "VITE_EI_API_MODE=real",
            "VITE_EI_API_BASE_URL=http://localhost:8080/api/v1",
            "frontendOwners.realApiMode.test.ts",
            " RUN  v2.1.9 /repo/frontend",
            " Test Files  2 passed (2)",
            "      Tests  8 passed (8)",
            "",
        )
    )


def test_frontend_real_backend_verify_owns_vitest_summary_contract(tmp_path: Path) -> None:
    cases = {
        "valid": (valid_frontend_vitest_log(), 0, ""),
        "missing-runner": (
            valid_frontend_vitest_log().replace(" RUN  v2.1.9 /repo/frontend\n", ""),
            1,
            "vitest runner marker missing",
        ),
        "no-tests": (
            valid_frontend_vitest_log() + "No test files found\n",
            1,
            "no-test marker found",
        ),
        "failed-summary": (
            valid_frontend_vitest_log().replace(
                " Test Files  2 passed (2)", " Test Files  1 failed (1)"
            ),
            1,
            "failing vitest summary found",
        ),
        "missing-tests-summary": (
            valid_frontend_vitest_log().replace("      Tests  8 passed (8)\n", ""),
            1,
            "no passing tests",
        ),
    }

    for name, (body, expected_status, expected_error) in cases.items():
        log = tmp_path / f"{name}.log"
        log.write_text(body, encoding="utf-8")
        result = run_script(FRONTEND_REAL_BACKEND_VERIFY, str(log), f"case-{name}")
        assert result.returncode == expected_status, result.stderr
        if expected_error:
            assert expected_error in result.stderr


def test_frontend_real_backend_verify_accepts_configurable_owner_test(
    tmp_path: Path,
) -> None:
    owner_test = "targetJob.realApiMode.test.ts"
    log = tmp_path / "target-job.log"
    log.write_text(
        valid_frontend_vitest_log().replace(
            "frontendOwners.realApiMode.test.ts", owner_test
        ),
        encoding="utf-8",
    )

    result = run_script(
        FRONTEND_REAL_BACKEND_VERIFY, str(log), "case-target-job", owner_test
    )
    assert result.returncode == 0, result.stderr

    log.write_text(
        valid_frontend_vitest_log().replace(
            "frontendOwners.realApiMode.test.ts", "other.realApiMode.test.ts"
        ),
        encoding="utf-8",
    )
    result = run_script(
        FRONTEND_REAL_BACKEND_VERIFY, str(log), "case-target-job", owner_test
    )
    assert result.returncode == 1
    assert f"{owner_test} did not run" in result.stderr


def test_frontend_real_backend_verify_callers_do_not_duplicate_vitest_parsing() -> None:
    callers = sorted(
        path
        for path in (SCENARIO_ROOT / "e2e").glob("*/scripts/verify.sh")
        if "frontend-real-backend-verify.sh" in path.read_text(encoding="utf-8")
    )
    assert callers

    forbidden = (
        "No test files found",
        "vitest runner marker missing",
        "failing vitest summary found",
        "Test Files",
    )
    offenders = {
        path.relative_to(REPO_ROOT).as_posix(): [token for token in forbidden if token in text]
        for path in callers
        if (text := path.read_text(encoding="utf-8"))
        if any(token in text for token in forbidden)
    }
    assert offenders == {}


def test_home_parse_real_backend_verify_callers_use_shared_helper() -> None:
    relative_callers = (
        "e2e/p0-014-home-default-render/scripts/verify.sh",
        "e2e/p0-015-jd-import-and-parse/scripts/verify.sh",
        "e2e/p0-016-parse-confirm-to-workspace/scripts/verify.sh",
    )
    forbidden = (
        "grep -Fq 'VITE_EI_API_MODE=real'",
        "grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1'",
        "grep -Fq 'targetJob.realApiMode.test.ts'",
        "grep -Eq 'Test Files +",
        "grep -Eq 'Tests +",
        "grep -Eq '[0-9]+ passed'",
        "grep -q 'Tests.*passed'",
    )
    offenders: dict[str, list[str]] = {}

    for relative in relative_callers:
        path = SCENARIO_ROOT / relative
        text = path.read_text(encoding="utf-8")
        missing = []
        if "frontend-real-backend-verify.sh" not in text:
            missing.append("shared helper call")
        if '"targetJob.realApiMode.test.ts"' not in text:
            missing.append("targetJob owner test argument")
        missing.extend(token for token in forbidden if token in text)
        if missing:
            offenders[path.relative_to(REPO_ROOT).as_posix()] = missing

    assert offenders == {}


def test_scenario_lifecycle_scripts_do_not_keep_zero_read_path_metadata() -> None:
    cases = {
        "e2e/p0-018-workspace-default-render/scripts/setup.sh": ("SCENARIO_DIR=",),
        "e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/setup.sh": (
            "SCENARIO_DIR=",
        ),
        "e2e/p0-084-resume-flat-ui-regression/scripts/cleanup.sh": (
            "SCRIPT_DIR=",
            "REPO_ROOT=",
            "SCENARIO_ID=",
        ),
    }
    offenders = {
        relative: [token for token in forbidden if token in text]
        for relative, forbidden in cases.items()
        if (text := (SCENARIO_ROOT / relative).read_text(encoding="utf-8"))
        if any(token in text for token in forbidden)
    }

    assert offenders == {}


def test_duplicate_scenario_evidence_lifecycle_uses_shared_helpers() -> None:
    setup_helper = SCENARIO_ROOT / "_shared" / "scripts" / "scenario-evidence-setup.sh"
    cleanup_helper = (
        SCENARIO_ROOT / "_shared" / "scripts" / "scenario-evidence-cleanup.sh"
    )
    setup_callers = (
        "e2e/p0-021-workspace-handoff/scripts/setup.sh",
        "e2e/p0-081-resume-create-flow-upload-paste-direct-detail/scripts/setup.sh",
        "e2e/p0-082-resume-create-flow-direct-detail-only/scripts/setup.sh",
        "e2e/p0-083-resume-create-flow-direct-create-handoff/scripts/setup.sh",
        "e2e/p0-084-resume-flat-ui-regression/scripts/setup.sh",
    )
    cleanup_callers = tuple(path.replace("setup.sh", "cleanup.sh") for path in setup_callers[:4])

    offenders: dict[str, list[str]] = {}
    if not setup_helper.is_file():
        offenders[setup_helper.relative_to(REPO_ROOT).as_posix()] = ["missing"]
    if not cleanup_helper.is_file():
        offenders[cleanup_helper.relative_to(REPO_ROOT).as_posix()] = ["missing"]

    for relative in setup_callers:
        text = (SCENARIO_ROOT / relative).read_text(encoding="utf-8")
        issues = []
        if "scenario-evidence-setup.sh" not in text:
            issues.append("shared setup helper call")
        issues.extend(token for token in ("mkdir -p", "printf 'scenario=") if token in text)
        if issues:
            offenders[relative] = issues

    for relative in cleanup_callers:
        text = (SCENARIO_ROOT / relative).read_text(encoding="utf-8")
        issues = []
        if "scenario-evidence-cleanup.sh" not in text:
            issues.append("shared cleanup helper call")
        if 'rm -f "$OUTPUT_DIR/setup.env"' in text:
            issues.append("direct setup.env cleanup")
        if issues:
            offenders[relative] = issues

    assert offenders == {}


def test_scenario_readmes_only_reference_existing_shared_scripts() -> None:
    inventories = [
        (
            SCENARIO_ROOT / "README.md",
            r"`(test/scenarios/_shared/scripts/[A-Za-z0-9._-]+\.sh)`",
            REPO_ROOT,
        ),
        (
            SCENARIO_ROOT / "_shared" / "README.md",
            r"`(scripts/[A-Za-z0-9._-]+\.sh)`",
            SCENARIO_ROOT / "_shared",
        ),
    ]
    references: set[Path] = set()

    for readme, pattern, base in inventories:
        text = readme.read_text(encoding="utf-8")
        references.update(base / raw for raw in re.findall(pattern, text))

    assert references, "scenario README shared-script inventory must not be empty"
    missing = sorted(path.relative_to(REPO_ROOT).as_posix() for path in references if not path.is_file())
    assert not missing, f"scenario READMEs reference missing shared scripts: {missing}"


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
    assert "backend_listen_addr()" in helper
    assert "127.0.0.1:%s" in helper
    assert 'APP_LISTEN_ADDR="$(backend_listen_addr)"' in helper
    assert "export APP_LISTEN_ADDR" in helper
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


def test_root_makefile_exposes_scenario_env_reset_redeploy_target() -> None:
    makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")

    assert "scenario-env-reset-redeploy" in makefile
    assert "SCENARIO_ENV_CLEANUP" in makefile
    assert "SCENARIO_ENV_SETUP" in makefile
    assert "SCENARIO_ENV_REDEPLOY" in makefile
    assert "SCENARIO_ENV_VERIFY" in makefile
    assert "--with-volumes" in makefile
    assert "--with-migrations" in makefile

    result = subprocess.run(
        ["make", "scenario-env-reset-redeploy", "ARGS=--dry-run"],
        cwd=REPO_ROOT,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )

    assert result.returncode == 0, result.stderr
    output = result.stdout + result.stderr
    ordered_markers = [
        "dry-run: DEV_RESET_FORCE=1 make dev-reset",
        "dry-run: make dev-up",
        "make migrate-up",
        "dry-run: cd backend && go build ./cmd/...",
        "dry-run: restart backend host-run process",
        "pnpm --filter @easyinterview/frontend build",
        "dry-run: restart frontend host-run process",
    ]
    positions = []
    start = 0
    for marker in ordered_markers:
        position = output.find(marker, start)
        assert position != -1, output
        positions.append(position)
        start = position + len(marker)

    final_verify = output.rfind("dry-run: make dev-doctor")
    assert final_verify > positions[-1], output
    assert positions == sorted(positions), output


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
        assert "scenario-env-reset-redeploy" in text
        assert "具体场景" in text or "specific scenario" in text.lower()

    assert "手动引导" in suite
    assert "host-run" in dev_stack
    assert "127.0.0.1:${API_HOST_PORT:-8080}" in dev_stack
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
