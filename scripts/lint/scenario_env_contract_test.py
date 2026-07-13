import copy
import hashlib
import http.server
import importlib.util
import json
import os
import re
import stat
import struct
import subprocess
import threading
import zlib
from pathlib import Path

import pytest
import yaml


REPO_ROOT = Path(__file__).resolve().parents[2]
SCENARIO_ROOT = REPO_ROOT / "test" / "scenarios"
FRONTEND_REAL_BACKEND_VERIFY = (
    SCENARIO_ROOT / "_shared" / "scripts" / "frontend-real-backend-verify.sh"
)

P0_099_SETUP_AT = "2026-07-13T00:00:00.000000Z"
P0_099_SESSION_CREATED_AT = "2026-07-13T00:00:01.000000Z"
P0_099_REPORT_CREATED_AT = "2026-07-13T00:00:02.000000Z"
P0_099_CAPTURED_AT = "2026-07-13T00:00:03.000000Z"

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


def valid_frontend_vitest_log(owner_test: str = "clientFactory.test.ts") -> str:
    return "\n".join(
        (
            "VITE_EI_API_MODE=real",
            "VITE_EI_API_BASE_URL=http://localhost:8080/api/v1",
            owner_test,
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
    log.write_text(valid_frontend_vitest_log(owner_test), encoding="utf-8")

    result = run_script(
        FRONTEND_REAL_BACKEND_VERIFY, str(log), "case-target-job", owner_test
    )
    assert result.returncode == 0, result.stderr

    log.write_text(
        valid_frontend_vitest_log("other.realApiMode.test.ts"),
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
    assert "export AI_DEBUG_PRINT_RAW_OUTPUT=false" in trigger
    assert "LOCAL_ENV=" not in trigger


def test_real_provider_hybrid_reliability_manifest_must_be_current_and_redacted() -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    setup = (scenario_dir / "scripts" / "setup.sh").read_text(encoding="utf-8")
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
    readme = (scenario_dir / "README.md").read_text(encoding="utf-8")

    assert "reliability-manifest.json" in setup
    assert "--sanitize-stage setup" in setup
    assert "RUN_ID=" in setup
    assert "RUN_ID" in trigger
    assert "validate_reliability.py" in trigger
    assert "P0_100_RUN_LIVE" in trigger
    assert "AI_PROVIDER_API_KEY" in trigger
    assert "AI_DEBUG_PRINT_RAW_OUTPUT=false" in trigger
    assert "forbidden raw/browser artifact persisted" in verify
    assert "secret/cookie material leaked" in verify
    assert "reliability-manifest.json" in readme
    assert "raw context/output" in readme


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


def _png_chunk(kind: bytes, payload: bytes) -> bytes:
    body = kind + payload
    return struct.pack(">I", len(payload)) + body + struct.pack(">I", zlib.crc32(body))


def _write_test_png(
    path: Path,
    width: int,
    height: int,
    *,
    color_type: int = 6,
    idat_payload: bytes | None = None,
    split_idat: bool = False,
    metadata_chunk: bytes | None = None,
    include_iend: bool = True,
    trailing: bytes = b"",
    patterned: bool = True,
    solid_bottom_rows: int = 0,
) -> None:
    signature = b"\x89PNG\r\n\x1a\n"
    channels = {2: 3, 6: 4}.get(color_type, 4)
    solid_scanline = b"\x00" + (b"\xff" * channels * width)
    colors = (
        b"\xf0\xf0\xf0\xff",
        b"\x20\x60\xa0\xff",
        b"\xd0\x70\x30\xff",
        b"\x40\x90\x50\xff",
    )
    if channels == 3:
        colors = tuple(color[:3] for color in colors)
    block_width = max(1, width // len(colors))
    patterned_pixels = b"".join(color * block_width for color in colors)
    patterned_pixels = (patterned_pixels + colors[-1] * width)[: width * channels]
    patterned_scanline = b"\x00" + patterned_pixels
    scanlines = b"".join(
        solid_scanline
        if not patterned or row >= height - solid_bottom_rows
        else patterned_scanline
        for row in range(height)
    )
    chunks = [
        _png_chunk(
            b"IHDR",
            struct.pack(
                ">IIBBBBB",
                width,
                height,
                8,
                color_type,
                0,
                0,
                0,
            ),
        ),
    ]
    if metadata_chunk is not None:
        chunks.append(_png_chunk(metadata_chunk, b"key\x00sensitive-value"))
    compressed = (
        zlib.compress(scanlines)
        if idat_payload is None
        else idat_payload
    )
    if split_idat:
        midpoint = len(compressed) // 2
        chunks.extend(
            (
                _png_chunk(b"IDAT", compressed[:midpoint]),
                _png_chunk(b"IDAT", compressed[midpoint:]),
            )
        )
    else:
        chunks.append(_png_chunk(b"IDAT", compressed))
    if include_iend:
        chunks.append(_png_chunk(b"IEND", b""))
    path.write_bytes(signature + b"".join(chunks) + trailing)


def _p0_099_api_report(
    report_ref: str,
    session_ref: str,
    state: str,
    locale: str,
) -> dict[str, object]:
    ready = state != "generating"
    if not ready:
        return {
            "id": report_ref,
            "sessionId": session_ref,
            "status": "generating",
            "summary": None,
            "preparednessLevel": None,
            "dimensionAssessments": [],
            "highlights": [],
            "issues": [],
            "nextActions": [],
            "retryFocusDimensionCodes": [],
            "provenance": None,
            "context": {"language": "zh-CN"},
        }

    zh = locale == "zh"
    return {
        "id": report_ref,
        "sessionId": session_ref,
        "status": "ready",
        "summary": "grounded summary",
        "preparednessLevel": "needs_practice" if zh else "well_prepared",
        "dimensionAssessments": [
            {"code": "communication", "label": "Communication", "status": "meets_bar", "confidence": "high"},
            {"code": "evidence", "label": "Evidence", "status": "needs_work", "confidence": "medium"},
        ],
        "highlights": [
            {"dimensionCode": "communication", "evidence": "bounded evidence", "confidence": "high"},
        ],
        "issues": [
            {"dimensionCode": "evidence", "evidence": "bounded issue", "confidence": "medium"},
        ],
        "nextActions": [
            {
                "type": "retry_current_round",
                "label": "行" * 64 if zh else " ".join(f"word{index}" for index in range(24)),
            },
        ],
        "retryFocusDimensionCodes": ["evidence"] if zh else [],
        "provenance": {
            "promptVersion": "v0.2.0",
            "rubricVersion": "v0.2.0",
            "modelId": "model-profile:report.generate.default",
            "language": "zh-CN" if zh else "en",
            "featureFlag": "none",
            "dataSourceVersion": "report-context.v1",
        },
        "context": {"language": "zh-CN" if zh else "en"},
    }


def _p0_099_canonical_report_digest(report: dict[str, object]) -> str | None:
    if report["status"] != "ready":
        return None
    projection = {
        key: report[key]
        for key in (
            "summary",
            "preparednessLevel",
            "dimensionAssessments",
            "highlights",
            "issues",
            "nextActions",
            "retryFocusDimensionCodes",
            "provenance",
        )
    }
    canonical = json.dumps(projection, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def _p0_099_frozen_context(report_ref: str, session_ref: str) -> dict[str, object]:
    return {
        "schemaVersion": "report-context.v1",
        "conversation": {"sessionId": session_ref},
        "targetJob": {"id": f"target-{report_ref[-4:]}"},
    }


def _p0_099_json_digest(value: object) -> str:
    canonical = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def _p0_099_manifest(run_id: str) -> dict[str, object]:
    states = (
        ("report-zh-needs-practice-desktop.png", "zh", "ready-needs-practice", "desktop", 1440, 1200, "00000000-0000-4000-8000-000000000101", "00000000-0000-4000-8000-000000000201"),
        ("report-zh-needs-practice-mobile.png", "zh", "ready-needs-practice", "mobile", 390, 844, "00000000-0000-4000-8000-000000000101", "00000000-0000-4000-8000-000000000201"),
        ("report-en-well-prepared-desktop.png", "en", "ready-well-prepared", "desktop", 1440, 1200, "00000000-0000-4000-8000-000000000102", "00000000-0000-4000-8000-000000000202"),
        ("report-en-well-prepared-mobile.png", "en", "ready-well-prepared", "mobile", 390, 844, "00000000-0000-4000-8000-000000000102", "00000000-0000-4000-8000-000000000202"),
        ("report-generating-desktop.png", "zh", "generating", "desktop", 1440, 1200, "00000000-0000-4000-8000-000000000103", "00000000-0000-4000-8000-000000000203"),
        ("report-generating-mobile.png", "zh", "generating", "mobile", 390, 844, "00000000-0000-4000-8000-000000000103", "00000000-0000-4000-8000-000000000203"),
    )
    rows = []
    for file_name, locale, state, viewport_name, width, height, report_ref, session_ref in states:
        ready = state != "generating"
        report_content_digest = _p0_099_canonical_report_digest(
            _p0_099_api_report(report_ref, session_ref, state, locale)
        )
        frozen_context_digest = _p0_099_json_digest(
            _p0_099_frozen_context(report_ref, session_ref)
        )
        preparedness = {
            "ready-needs-practice": "needs_practice",
            "ready-well-prepared": "well_prepared",
            "generating": None,
        }[state]
        rows.append(
            {
                "file": file_name,
                "locale": locale,
                "state": state,
                "fixture": f"real-provider-{state}-long",
                "viewport": {"name": viewport_name, "width": width, "height": height},
                "full_page": True,
                "report_ref": report_ref,
                "session_ref": session_ref,
                "screenshot_sha256": "0" * 64,
                "evidence": {
                    "collection": {
                        "run_id": run_id,
                        "method": "trusted-current-run-db-api-capture",
                        "report_ref": report_ref,
                        "session_ref": session_ref,
                        "frozen_context_digest": frozen_context_digest,
                        "report_content_digest": report_content_digest,
                        "screenshot_sha256": "0" * 64,
                    },
                    "db": {
                        "status": "ready" if ready else "generating",
                        "preparedness_level": preparedness,
                        "frozen_context_digest": frozen_context_digest,
                        "report_content_digest": report_content_digest,
                    },
                    "api": {
                        "status": "ready" if ready else "generating",
                        "preparedness_level": preparedness,
                        "report_content_digest": report_content_digest,
                        "source_message_seq_nos_exposed": False,
                    },
                    "content_audit": {
                        "fact_to_judgment_to_action": "closed" if ready else "not_applicable",
                        "item_verdict_count": 4 if ready else 0,
                        "unsupported_count": 0,
                        "irrelevant_advice_count": 0,
                        "causal_mismatch_count": 0,
                        "action_label_audit": {
                            "language": "zh-CN" if locale == "zh" and ready else "en" if ready else "not_applicable",
                            "unit": "code_points" if locale == "zh" and ready else "words" if ready else "not_applicable",
                            "limit": 64 if locale == "zh" and ready else 24 if ready else 0,
                            "counts": [64] if locale == "zh" and ready else [24] if ready else [],
                        },
                    },
                },
            }
        )
    return {
        "scenario_id": "E2E.P0.099",
        "run_id": run_id,
        "capture_contract": "report-full-page-v1",
        "screenshots": rows,
        "privacy": {
            "redacted": True,
            "cookie_written": False,
            "raw_frozen_context_written": False,
        },
    }


def _p0_099_live_capture(manifest: dict[str, object]) -> dict[str, object]:
    reports = []
    seen: set[str] = set()
    for row in manifest["screenshots"]:
        if row["report_ref"] in seen:
            continue
        seen.add(row["report_ref"])
        audit = row["evidence"]["content_audit"]
        item_count = audit["item_verdict_count"]
        reports.append(
            {
                "report_ref": row["report_ref"],
                "session_ref": row["session_ref"],
                "status": row["evidence"]["api"]["status"],
                "preparedness_level": row["evidence"]["api"]["preparedness_level"],
                "canonical_report_content_digest": row["evidence"]["api"]["report_content_digest"],
                "content_shape": {
                    "dimension_assessment_count": item_count,
                    "highlight_count": 0,
                    "issue_count": 0,
                    "next_action_count": len(audit["action_label_audit"]["counts"]),
                    "retry_focus_count": 0,
                },
                "action_label_audit": audit["action_label_audit"],
                "db": {
                    "status": row["evidence"]["db"]["status"],
                    "preparedness_level": row["evidence"]["db"]["preparedness_level"],
                    "report_created_at": P0_099_REPORT_CREATED_AT,
                    "session_created_at": P0_099_SESSION_CREATED_AT,
                    "frozen_context_digest": row["evidence"]["db"]["frozen_context_digest"],
                    "canonical_report_content_digest": row["evidence"]["db"]["report_content_digest"],
                },
            }
        )
    return {
        "schema_version": "p0-099-live-capture.v2",
        "scenario_id": "E2E.P0.099",
        "run_id": manifest["run_id"],
        "method": "authenticated-live-http+read-only-postgres",
        "captured_at": P0_099_CAPTURED_AT,
        "result": "PASS",
        "reason_code": "captured",
        "reports": reports,
        "privacy": {
            "cookie_written": False,
            "database_url_written": False,
            "raw_api_written": False,
            "raw_db_written": False,
            "raw_frozen_context_written": False,
            "prose_written": False,
        },
    }


def _p0_099_manual_visual_audit(manifest: dict[str, object]) -> dict[str, object]:
    screenshots = []
    for row in manifest["screenshots"]:
        ready = row["state"] != "generating"
        checks = {
            "report_page_visible": True,
            "expected_state_visible": True,
            "horizontal_overflow_absent": True,
        }
        if ready:
            checks.update(
                {
                    "preparedness_visible": True,
                    "dimension_and_evidence_content_visible": True,
                    "action_region_visible": True,
                    "action_labels_complete_without_clipping_or_ellipsis": True,
                }
            )
        else:
            checks.update(
                {
                    "generating_indicator_visible": True,
                    "ready_content_absent": True,
                    "false_ready_claim_absent": True,
                    "clipping_or_overlap_absent": True,
                }
            )
        screenshots.append(
            {
                "file": row["file"],
                "screenshot_sha256": row["screenshot_sha256"],
                "checks": checks,
            }
        )
    return {
        "schema_version": "p0-099-manual-visual-audit.v1",
        "scenario_id": "E2E.P0.099",
        "run_id": manifest["run_id"],
        "method": "manual-image-review-no-ocr",
        "result": "PASS",
        "screenshots": screenshots,
        "privacy": {
            "ocr_used": False,
            "prose_transcribed": False,
            "raw_content_written": False,
        },
    }


def _write_p0_099_setup(output: Path, run_id: str) -> None:
    (output / "setup.env").write_text(
        "\n".join(
            (
                "scenario=E2E.P0.099",
                f"RUN_ID={run_id}",
                f"setup_at={P0_099_SETUP_AT}",
                "ai_raw_debug=false",
                "backend_pid=1",
                "backend_log_start_bytes=0",
                "",
            )
        ),
        encoding="utf-8",
    )


def _write_p0_099_evidence(
    output: Path,
    run_id: str,
    *,
    include_manual_audit: bool = True,
) -> dict[str, object]:
    screenshots = output / "screenshots"
    screenshots.mkdir(parents=True, exist_ok=True)
    manifest = _p0_099_manifest(run_id)
    for index, row in enumerate(manifest["screenshots"]):
        viewport = row["viewport"]
        screenshot = screenshots / row["file"]
        _write_test_png(
            screenshot,
            viewport["width"],
            viewport["height"],
            color_type=2 if index % 2 else 6,
            split_idat=index == 0,
        )
        digest = hashlib.sha256(screenshot.read_bytes()).hexdigest()
        row["screenshot_sha256"] = digest
        row["evidence"]["collection"]["screenshot_sha256"] = digest
    _write_p0_099_setup(output, run_id)
    (output / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    (output / "live-capture.json").write_text(
        json.dumps(_p0_099_live_capture(manifest)), encoding="utf-8"
    )
    if include_manual_audit:
        (output / "manual-visual-audit.json").write_text(
            json.dumps(_p0_099_manual_visual_audit(manifest)), encoding="utf-8"
        )
    return manifest


def _p0_099_db_rows(manifest: dict[str, object]) -> list[dict[str, object]]:
    rows = []
    seen: set[str] = set()
    for row in manifest["screenshots"]:
        if row["report_ref"] in seen:
            continue
        seen.add(row["report_ref"])
        api_report = _p0_099_api_report(
            row["report_ref"], row["session_ref"], row["state"], row["locale"]
        )
        rows.append(
            {
                "report_ref": row["report_ref"],
                "session_ref": row["session_ref"],
                "status": api_report["status"],
                "preparedness_level": api_report["preparednessLevel"],
                "report_created_at": P0_099_REPORT_CREATED_AT,
                "session_created_at": P0_099_SESSION_CREATED_AT,
                "generation_context": _p0_099_frozen_context(
                    row["report_ref"], row["session_ref"]
                ),
                "summary": api_report["summary"],
                "dimension_assessments": api_report["dimensionAssessments"],
                "highlights": [
                    {**item, "sourceMessageSeqNos": [2]}
                    for item in api_report["highlights"]
                ],
                "issues": [
                    {**item, "sourceMessageSeqNos": [2]}
                    for item in api_report["issues"]
                ],
                "next_actions": api_report["nextActions"],
                "retry_focus_dimension_codes": api_report["retryFocusDimensionCodes"],
                "prompt_version": api_report["provenance"]["promptVersion"]
                if api_report["provenance"]
                else None,
                "rubric_version": api_report["provenance"]["rubricVersion"]
                if api_report["provenance"]
                else None,
                "model_id": api_report["provenance"]["modelId"]
                if api_report["provenance"]
                else None,
                "language": api_report["context"]["language"],
                "feature_flag": api_report["provenance"]["featureFlag"]
                if api_report["provenance"]
                else "none",
                "data_source_version": api_report["provenance"]["dataSourceVersion"]
                if api_report["provenance"]
                else "not_applicable",
            }
        )
    return rows


def test_p0_099_exact_six_full_page_manifest_validator(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    readme = (scenario_dir / "README.md").read_text(encoding="utf-8")
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")

    assert validator.is_file()
    assert "exactly six" in readme.lower() or "恰好六" in readme
    assert "full-page" in readme.lower()
    assert "manifest.json" in trigger
    assert "P0_099_SIX_SCREENSHOT_PASS" in verify

    out = tmp_path / "p0-099"
    run_id = "p0-099-contract-current"
    manifest = _write_p0_099_evidence(out, run_id)
    manifest_path = out / "manifest.json"

    def validate() -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(validator), "--output-dir", str(out), "--run-id", run_id],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    valid = validate()
    assert valid.returncode == 0, valid.stderr
    assert "P0_099_SIX_SCREENSHOT_PASS" in valid.stdout
    assert "manual_visual_audit=bound" in valid.stdout

    live_capture_path = out / "live-capture.json"
    live_capture_body = live_capture_path.read_text(encoding="utf-8")
    live_capture_path.unlink()
    rejected = validate()
    assert rejected.returncode != 0
    assert "live capture" in rejected.stderr.lower()
    live_capture_path.write_text(live_capture_body, encoding="utf-8")

    for mutate, expected in (
        (lambda body: body.__setitem__("screenshots", body["screenshots"][:-1]), "exactly six"),
        (
            lambda body: body["screenshots"][0]["evidence"]["content_audit"][
                "action_label_audit"
            ].__setitem__("counts", [65]),
            "action label counts exceed",
        ),
        (
            lambda body: body["screenshots"][0]["evidence"]["api"].__setitem__(
                "report_content_digest", "f" * 64
            ),
            "report content digests do not match",
        ),
    ):
        invalid = copy.deepcopy(manifest)
        mutate(invalid)
        manifest_path.write_text(json.dumps(invalid), encoding="utf-8")
        rejected = validate()
        assert rejected.returncode != 0
        assert expected in rejected.stderr.lower(), rejected.stderr


def test_p0_099_manual_visual_audit_is_exact_six_sha_bound_and_no_ocr(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    out = tmp_path / "p0-099-manual-visual"
    run_id = "p0-099-manual-visual-current"
    manifest = _write_p0_099_evidence(out, run_id)
    audit_path = out / "manual-visual-audit.json"
    original = json.loads(audit_path.read_text(encoding="utf-8"))

    def validate() -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(validator), "--output-dir", str(out), "--run-id", run_id],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    audit_path.unlink()
    missing = validate()
    assert missing.returncode != 0
    assert "manual visual audit" in missing.stderr.lower()

    for mutate, expected in (
        (lambda body: body["screenshots"].pop(), "exactly six"),
        (
            lambda body: body["screenshots"][0].__setitem__("screenshot_sha256", "f" * 64),
            "sha-256",
        ),
        (
            lambda body: body["screenshots"][0]["checks"].__setitem__(
                "action_region_visible", False
            ),
            "manual visual",
        ),
        (lambda body: body["privacy"].__setitem__("ocr_used", True), "ocr"),
        (lambda body: body.__setitem__("run_id", "stale-run"), "run"),
    ):
        candidate = copy.deepcopy(original)
        mutate(candidate)
        audit_path.write_text(json.dumps(candidate), encoding="utf-8")
        rejected = validate()
        assert rejected.returncode != 0
        assert expected in rejected.stderr.lower(), rejected.stderr

    audit_path.write_text(json.dumps(_p0_099_manual_visual_audit(manifest)), encoding="utf-8")
    accepted = validate()
    assert accepted.returncode == 0, accepted.stderr
    assert "manual_visual_audit=bound" in accepted.stdout


def test_p0_099_db_projection_requires_current_run_timestamps_and_real_digests(
    tmp_path: Path,
) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    out = tmp_path / "p0-099-db-current-run"
    run_id = "p0-099-db-current-run"
    _write_p0_099_evidence(out, run_id)
    capture_path = out / "live-capture.json"
    original = json.loads(capture_path.read_text(encoding="utf-8"))

    def validate() -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                "python3",
                str(validator),
                "--output-dir",
                str(out),
                "--run-id",
                run_id,
                "--automated-only",
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    for mutate, expected in (
        (
            lambda body: body["reports"][0]["db"].__setitem__(
                "session_created_at", "2026-07-12T23:59:59.999999Z"
            ),
            "current run",
        ),
        (
            lambda body: body["reports"][0]["db"].__setitem__(
                "report_created_at", "2026-07-12T23:59:59.999999Z"
            ),
            "current run",
        ),
        (
            lambda body: body["reports"][0]["db"].__setitem__(
                "frozen_context_digest", "a" * 64
            ),
            "db projection",
        ),
        (
            lambda body: body["reports"][0]["db"].__setitem__(
                "canonical_report_content_digest", "b" * 64
            ),
            "db projection",
        ),
    ):
        candidate = copy.deepcopy(original)
        mutate(candidate)
        capture_path.write_text(json.dumps(candidate), encoding="utf-8")
        rejected = validate()
        assert rejected.returncode != 0
        assert expected in rejected.stderr.lower(), rejected.stderr

    capture_path.write_text(json.dumps(original), encoding="utf-8")
    accepted = validate()
    assert accepted.returncode == 0, accepted.stderr


def test_p0_099_png_integrity_metadata_and_digest_fail_closed(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    out = tmp_path / "p0-099-png-negative"
    screenshots = out / "screenshots"
    screenshots.mkdir(parents=True)
    run_id = "p0-099-png-negative-current"
    manifest = _p0_099_manifest(run_id)

    def write_valid_matrix() -> None:
        for index, row in enumerate(manifest["screenshots"]):
            viewport = row["viewport"]
            screenshot = screenshots / row["file"]
            _write_test_png(
                screenshot,
                viewport["width"],
                viewport["height"],
                color_type=2 if index % 2 else 6,
                split_idat=index == 0,
            )
            digest = hashlib.sha256(screenshot.read_bytes()).hexdigest()
            row["screenshot_sha256"] = digest
            row["evidence"]["collection"]["screenshot_sha256"] = digest
        (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
        (out / "live-capture.json").write_text(
            json.dumps(_p0_099_live_capture(manifest)), encoding="utf-8"
        )
        _write_p0_099_setup(out, run_id)

    def assert_rejected(expected: str) -> None:
        result = subprocess.run(
            [
                "python3",
                str(validator),
                "--output-dir",
                str(out),
                "--run-id",
                run_id,
                "--automated-only",
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert result.returncode != 0
        assert expected in result.stderr.lower(), result.stderr

    target_row = manifest["screenshots"][0]
    target = screenshots / target_row["file"]

    def bind_target_digest() -> None:
        target_row["screenshot_sha256"] = hashlib.sha256(target.read_bytes()).hexdigest()
        target_row["evidence"]["collection"]["screenshot_sha256"] = target_row["screenshot_sha256"]
        (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")

    write_valid_matrix()
    payload = bytearray(target.read_bytes())
    payload[29] ^= 0x01
    target.write_bytes(payload)
    bind_target_digest()
    assert_rejected("crc")

    write_valid_matrix()
    viewport = target_row["viewport"]
    _write_test_png(target, viewport["width"], viewport["height"], include_iend=False)
    bind_target_digest()
    assert_rejected("iend")

    write_valid_matrix()
    _write_test_png(target, viewport["width"], viewport["height"], metadata_chunk=b"iTXt")
    bind_target_digest()
    assert_rejected("metadata")

    write_valid_matrix()
    _write_test_png(
        target,
        viewport["width"],
        viewport["height"],
        idat_payload=b"crc-valid-but-not-a-zlib-stream",
    )
    bind_target_digest()
    assert_rejected("zlib")

    write_valid_matrix()
    target_row["screenshot_sha256"] = "f" * 64
    (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    assert_rejected("sha-256")

    write_valid_matrix()
    extra = screenshots / "page.html"
    extra.write_text("<main>bounded report page</main>", encoding="utf-8")
    assert_rejected("exactly six canonical regular png")
    extra.unlink()

    write_valid_matrix()
    external_png = tmp_path / "external-report.png"
    external_png.write_bytes(target.read_bytes())
    target.unlink()
    target.symlink_to(external_png)
    assert_rejected("symlink")
    target.unlink()

    write_valid_matrix()
    external_screenshots = tmp_path / "external-screenshots"
    external_screenshots.mkdir()
    for screenshot in screenshots.iterdir():
        (external_screenshots / screenshot.name).write_bytes(screenshot.read_bytes())
        screenshot.unlink()
    screenshots.rmdir()
    screenshots.symlink_to(external_screenshots, target_is_directory=True)
    assert_rejected("symlink")


def test_p0_099_png_visual_content_and_ready_mobile_bottom_fail_closed(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    out = tmp_path / "p0-099-visual-content"
    screenshots = out / "screenshots"
    screenshots.mkdir(parents=True)
    run_id = "p0-099-visual-content-current"
    manifest = _p0_099_manifest(run_id)

    def write_matrix() -> None:
        for row in manifest["screenshots"]:
            viewport = row["viewport"]
            screenshot = screenshots / row["file"]
            _write_test_png(screenshot, viewport["width"], viewport["height"])
            digest = hashlib.sha256(screenshot.read_bytes()).hexdigest()
            row["screenshot_sha256"] = digest
            row["evidence"]["collection"]["screenshot_sha256"] = digest
        (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
        (out / "live-capture.json").write_text(
            json.dumps(_p0_099_live_capture(manifest)), encoding="utf-8"
        )
        _write_p0_099_setup(out, run_id)

    def bind(row: dict[str, object]) -> None:
        screenshot = screenshots / row["file"]
        digest = hashlib.sha256(screenshot.read_bytes()).hexdigest()
        row["screenshot_sha256"] = digest
        row["evidence"]["collection"]["screenshot_sha256"] = digest
        (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")

    def assert_rejected(expected: str) -> None:
        result = subprocess.run(
            [
                "python3",
                str(validator),
                "--output-dir",
                str(out),
                "--run-id",
                run_id,
                "--automated-only",
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert result.returncode != 0
        assert expected in result.stderr.lower(), result.stderr

    write_matrix()
    solid_row = manifest["screenshots"][4]
    solid_viewport = solid_row["viewport"]
    _write_test_png(
        screenshots / solid_row["file"],
        solid_viewport["width"],
        solid_viewport["height"],
        patterned=False,
    )
    bind(solid_row)
    assert_rejected("visual content")

    write_matrix()
    mobile_row = manifest["screenshots"][1]
    mobile_viewport = mobile_row["viewport"]
    _write_test_png(
        screenshots / mobile_row["file"],
        mobile_viewport["width"],
        mobile_viewport["height"] * 2,
        solid_bottom_rows=mobile_viewport["height"],
    )
    bind(mobile_row)
    assert_rejected("ready mobile bottom")


def test_p0_099_vitest_runner_evidence_is_exact_and_fail_closed(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
    runner_validator = scenario_dir / "scripts" / "validate_vitest_log.py"
    targets = (
        "src/app/screens/generating/__tests__/GeneratingScreen.test.tsx",
        "src/app/screens/report/__tests__/ConversationReport.test.tsx",
        "src/app/screens/report/__tests__/reportContract.test.ts",
    )

    assert runner_validator.is_file()
    assert "pnpm exec vitest run" in trigger
    assert "validate_vitest_log.py" in verify
    for target in targets:
        assert trigger.count(target) == 1

    valid_log = "\n".join(
        (
            " RUN  v2.1.9 /repo/frontend",
            f" ✓ {targets[0]} (5 tests) 80ms",
            f" ✓ {targets[1]} (9 tests) 150ms",
            f" ✓ {targets[2]} (15 tests) 4ms",
            " Test Files  3 passed (3)",
            "      Tests  29 passed (29)",
            "",
        )
    )
    log = tmp_path / "vitest.log"

    def validate(body: str) -> subprocess.CompletedProcess[str]:
        log.write_text(body, encoding="utf-8")
        return subprocess.run(
            ["python3", str(runner_validator), str(log)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    assert validate(valid_log).returncode == 0
    for invalid in (
        valid_log.replace(" RUN  v2.1.9 /repo/frontend\n", ""),
        valid_log.replace(" Test Files  3 passed (3)", " Test Files  0 passed (0)"),
        valid_log.replace("      Tests  29 passed (29)", "      Tests  0 passed (0)"),
        valid_log.replace("      Tests  29 passed (29)", "      Tests  1 passed (1)"),
        valid_log.replace(f" ✓ {targets[2]} (15 tests) 4ms\n", ""),
    ):
        assert validate(invalid).returncode != 0


def test_p0_099_live_http_capture_is_redacted_and_manifest_bound(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    capture_script = scenario_dir / "scripts" / "capture_live_evidence.py"
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    run_id = "p0-099-live-capture-current"
    expected_manifest = _p0_099_manifest(run_id)
    manifest = copy.deepcopy(expected_manifest)
    for row in manifest["screenshots"]:
        audit = row["evidence"]["content_audit"]
        audit["item_verdict_count"] = 999
        audit["action_label_audit"] = {
            "language": "not_applicable",
            "unit": "not_applicable",
            "limit": 0,
            "counts": [],
        }
        row["evidence"] = {"content_audit": audit}
    manifest_path = tmp_path / "manifest.json"
    capture_path = tmp_path / "live-capture.json"
    manifest_path.write_text(json.dumps(manifest), encoding="utf-8")

    assert capture_script.is_file()
    assert "capture_live_evidence.py" in trigger
    assert "--bind-manifest" in trigger
    assert trigger.index("capture_live_evidence.py") < trigger.index('if [ ! -s "$MANIFEST" ]')
    assert trigger.index("unset P0_099_SESSION_COOKIE") < trigger.index("test/scenarios/env-verify.sh")
    capture_call = 'P0_099_DATABASE_URL="$LIVE_DATABASE_URL" P0_099_SESSION_COOKIE="$LIVE_SESSION_COOKIE" python3 "$LIVE_CAPTURE_RUNNER"'
    assert trigger.index(capture_call) < trigger.index("test/scenarios/env-verify.sh")
    assert trigger.index(capture_call) < trigger.index("pnpm exec vitest run")
    assert trigger.index('if python3 "$VALIDATOR" --output-dir') < trigger.index(
        'echo "P0_099_LIVE_CAPTURE_BOUND_PASS"'
    )

    reports: dict[str, dict[str, object]] = {}
    for row in manifest["screenshots"]:
        reports.setdefault(
            row["report_ref"],
            _p0_099_api_report(row["report_ref"], row["session_ref"], row["state"], row["locale"]),
        )

    db_rows_path = tmp_path / "db-rows.json"
    db_rows_path.write_text(json.dumps(_p0_099_db_rows(manifest)), encoding="utf-8")
    fake_bin = tmp_path / "bin"
    fake_bin.mkdir()
    fake_psql = fake_bin / "psql"
    fake_psql.write_text(
        "\n".join(
            (
                "#!/usr/bin/env python3",
                "import json",
                "import os",
                "assert os.environ.get('PGOPTIONS') == '-c default_transaction_read_only=on'",
                "assert 'P0_099_DATABASE_URL' not in os.environ",
                f"rows = json.load(open({str(db_rows_path)!r}, encoding='utf-8'))",
                "for row in rows:",
                "    print(json.dumps(row, ensure_ascii=False, separators=(',', ':')))",
                "",
            )
        ),
        encoding="utf-8",
    )
    fake_psql.chmod(0o755)

    class ReportHandler(http.server.BaseHTTPRequestHandler):
        def do_GET(self) -> None:  # noqa: N802 - stdlib handler contract
            if self.headers.get("Cookie") != "ei_session=temporary-cookie-value":
                self.send_response(401)
                self.end_headers()
                return
            report_ref = self.path.rsplit("/", 1)[-1]
            report = reports.get(report_ref)
            if report is None:
                self.send_response(404)
                self.end_headers()
                return
            body = json.dumps(report, ensure_ascii=False).encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, _format: str, *_args: object) -> None:
            return

    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), ReportHandler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        env = os.environ.copy()
        env["P0_099_SESSION_COOKIE"] = "temporary-cookie-value"
        env["P0_099_DATABASE_URL"] = (
            "postgres://temporary-user:temporary-password@127.0.0.1:5432/temporary-db"
            "?sslmode=disable"
        )
        env["PATH"] = f"{fake_bin}{os.pathsep}{env['PATH']}"
        result = subprocess.run(
            [
                "python3",
                str(capture_script),
                "--manifest",
                str(manifest_path),
                "--output",
                str(capture_path),
                "--run-id",
                run_id,
                "--api-base-url",
                f"http://127.0.0.1:{server.server_port}/api/v1",
                "--bind-manifest",
            ],
            cwd=REPO_ROOT,
            env=env,
            text=True,
            capture_output=True,
            check=False,
        )
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=5)

    assert result.returncode == 0, result.stderr
    assert stat.S_IMODE(capture_path.stat().st_mode) == 0o600
    capture = json.loads(capture_path.read_text(encoding="utf-8"))
    assert capture["result"] == "PASS"
    assert capture["schema_version"] == "p0-099-live-capture.v2"
    assert len(capture["reports"]) == 3
    assert all("db" in report for report in capture["reports"])
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    assert manifest == expected_manifest
    capture_text = capture_path.read_text(encoding="utf-8")
    for forbidden in (
        "temporary-cookie-value",
        "temporary-password",
        "grounded summary",
        "bounded evidence",
        "word0",
        "generation_context",
        "sourceMessageSeqNos",
    ):
        assert forbidden not in capture_text

    out = tmp_path / "evidence"
    screenshots = out / "screenshots"
    screenshots.mkdir(parents=True)
    for row in manifest["screenshots"]:
        viewport = row["viewport"]
        screenshot = screenshots / row["file"]
        _write_test_png(screenshot, viewport["width"], viewport["height"])
        digest = hashlib.sha256(screenshot.read_bytes()).hexdigest()
        row["screenshot_sha256"] = digest
        row["evidence"]["collection"]["screenshot_sha256"] = digest
    (out / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    (out / "live-capture.json").write_text(json.dumps(capture), encoding="utf-8")
    _write_p0_099_setup(out, run_id)
    (out / "manual-visual-audit.json").write_text(
        json.dumps(_p0_099_manual_visual_audit(manifest)), encoding="utf-8"
    )
    validated = subprocess.run(
        ["python3", str(validator), "--output-dir", str(out), "--run-id", run_id],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert validated.returncode == 0, validated.stderr

    capture["reports"][0]["status"] = "generating"
    (out / "live-capture.json").write_text(json.dumps(capture), encoding="utf-8")
    rejected = subprocess.run(
        ["python3", str(validator), "--output-dir", str(out), "--run-id", run_id],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert rejected.returncode != 0
    assert "live capture" in rejected.stderr.lower()


@pytest.mark.parametrize(
    ("cookie", "expected_reason"),
    ((None, "session_cookie_missing"), ("temporary-cookie-value", "database_url_missing")),
)
def test_p0_099_live_capture_missing_prerequisite_is_manual_required_and_redacted(
    tmp_path: Path,
    cookie: str | None,
    expected_reason: str,
) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    capture_script = scenario_dir / "scripts" / "capture_live_evidence.py"
    run_id = f"p0-099-live-manual-{expected_reason}"
    manifest_path = tmp_path / "manifest.json"
    output_path = tmp_path / "live-capture.json"
    manifest_path.write_text(json.dumps(_p0_099_manifest(run_id)), encoding="utf-8")
    env = os.environ.copy()
    if cookie is None:
        env.pop("P0_099_SESSION_COOKIE", None)
    else:
        env["P0_099_SESSION_COOKIE"] = cookie
    env.pop("P0_099_DATABASE_URL", None)

    result = subprocess.run(
        [
            "python3",
            str(capture_script),
            "--manifest",
            str(manifest_path),
            "--output",
            str(output_path),
            "--run-id",
            run_id,
            "--api-base-url",
            "http://127.0.0.1:1/api/v1",
        ],
        cwd=REPO_ROOT,
        env=env,
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode == 2
    artifact = json.loads(output_path.read_text(encoding="utf-8"))
    assert artifact["result"] == "MANUAL_REQUIRED"
    assert artifact["reason_code"] == expected_reason
    assert artifact["reports"] == []
    assert artifact["privacy"] == {
        "cookie_written": False,
        "database_url_written": False,
        "raw_api_written": False,
        "raw_db_written": False,
        "raw_frozen_context_written": False,
        "prose_written": False,
    }
    assert stat.S_IMODE(output_path.stat().st_mode) == 0o600
    body = output_path.read_text(encoding="utf-8")
    assert "temporary-cookie-value" not in body
    assert "postgres://" not in body


def test_p0_099_and_p0_100_failed_evidence_is_deleted_before_result_branch(tmp_path: Path) -> None:
    scenarios = (
        (
            "p0-099-full-funnel-fullstack-ui-journey",
            "validate_evidence.py",
            (
                "manifest.json",
                "manual-visual-audit.json",
                "screenshots/capture.png",
            ),
        ),
        (
            "p0-100-real-provider-full-funnel-hybrid",
            "validate_reliability.py",
            ("reliability-manifest.json", "independent-agent-audit.json"),
        ),
    )
    for scenario, validator_name, primary_paths in scenarios:
        scenario_dir = SCENARIO_ROOT / "e2e" / scenario
        validator = scenario_dir / "scripts" / validator_name
        trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
        verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
        cleanup = (scenario_dir / "scripts" / "cleanup.sh").read_text(encoding="utf-8")
        for text in (trigger, verify, cleanup):
            assert "--sanitize-output" in text
        assert "trap cleanup_untrusted_evidence EXIT" in trigger
        assert verify.index("--sanitize-output") < verify.index('case "$RESULT')

        output = tmp_path / scenario
        output.mkdir(parents=True)
        for relative in primary_paths:
            path = output / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text("untrusted", encoding="utf-8")
        normal_name = output / "notes.json"
        normal_name.write_text('{"prompt":"sensitive"}', encoding="utf-8")
        secret_marker = output / "diagnostic.txt"
        secret_marker.write_text("AI_PROVIDER_API_KEY=must-not-survive", encoding="utf-8")
        result = subprocess.run(
            ["python3", str(validator), "--sanitize-output", str(output), "--failed"],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert result.returncode == 0, result.stderr
        for relative in primary_paths:
            assert not (output / relative).exists()
        assert not normal_name.exists(), "content redlines must be deleted even under a normal filename"
        assert not secret_marker.exists(), "secret markers must be deleted even under a normal filename"


def test_p0_099_setup_removes_unknown_state_and_camelcase_cookie_material(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    validator = scenario_dir / "scripts" / "validate_evidence.py"
    setup = (scenario_dir / "scripts" / "setup.sh").read_text(encoding="utf-8")
    assert "--setup" in setup
    assert setup.index("--setup") < setup.index("setup.env")

    output = tmp_path / "p0-099-setup"
    output.mkdir(parents=True)
    (output / "state.json").write_text(
        json.dumps({"sessionCookieValue": "must-not-survive", "userEmail": "old@example.test"}),
        encoding="utf-8",
    )
    playwright = output / "playwright"
    playwright.mkdir()
    (playwright / ".last-run.json").write_text("{}", encoding="utf-8")
    outside = tmp_path / "outside.txt"
    outside.write_text("keep", encoding="utf-8")
    (output / "outside-link").symlink_to(outside)

    result = subprocess.run(
        ["python3", str(validator), "--sanitize-output", str(output), "--setup"],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert list(output.iterdir()) == []
    assert outside.read_text(encoding="utf-8") == "keep"

    output.rmdir()
    outside_root = tmp_path / "outside-root"
    outside_root.mkdir()
    outside_sentinel = outside_root / "must-survive.txt"
    outside_sentinel.write_text("keep-root", encoding="utf-8")
    output.symlink_to(outside_root, target_is_directory=True)
    result = subprocess.run(
        ["python3", str(validator), "--sanitize-output", str(output), "--setup"],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert output.is_dir() and not output.is_symlink()
    assert stat.S_IMODE(output.stat().st_mode) == 0o700
    assert outside_sentinel.read_text(encoding="utf-8") == "keep-root"


def test_p0_100_output_retention_uses_stage_allowlists_and_never_follows_symlinks(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    validator = scenario_dir / "scripts" / "validate_reliability.py"
    setup = (scenario_dir / "scripts" / "setup.sh").read_text(encoding="utf-8")
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
    cleanup = (scenario_dir / "scripts" / "cleanup.sh").read_text(encoding="utf-8")

    assert "--sanitize-stage setup" in setup
    assert setup.index("--sanitize-stage setup") < setup.index('> "$OUTPUT_DIR/setup.env"')
    for text in (trigger, verify, cleanup):
        assert "--sanitize-stage" in text

    output = tmp_path / "p0-100-output"
    outside = tmp_path / "outside-must-survive.txt"
    outside.write_text("outside", encoding="utf-8")

    bounded = {
        "setup.env": "RUN_ID=current\n",
        "setup.log": "setup bounded\n",
        "trigger.log": "trigger bounded\n",
        "result.json": '{"result":"FAIL"}\n',
        "investigation.json": '{"issue_count":1}\n',
        "cleanup.env": "cleanup_at=current\n",
        "reliability-manifest.json": "{}\n",
        "independent-agent-audit.json": "{}\n",
    }

    def populate() -> None:
        output.mkdir(parents=True, exist_ok=True)
        for name, body in bounded.items():
            (output / name).write_text(body, encoding="utf-8")
        (output / "legacy-report.png").write_bytes(b"stale-png")
        (output / "unknown.json").write_text("{}\n", encoding="utf-8")
        (output / "screenshots").mkdir()
        (output / "screenshots" / "old.png").write_bytes(b"stale-png")
        (output / "playwright-report").mkdir()
        (output / "playwright-report" / "index.html").write_text("stale", encoding="utf-8")
        (output / "unknown-dir").mkdir()
        (output / "unknown-dir" / "nested.txt").write_text("stale", encoding="utf-8")
        (output / "outside-link").symlink_to(outside)

    def sanitize(stage: str, expected_returncode: int = 0) -> None:
        result = subprocess.run(
            [
                "python3",
                str(validator),
                "--sanitize-output",
                str(output),
                "--sanitize-stage",
                stage,
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert result.returncode == expected_returncode, result.stderr
        assert outside.read_text(encoding="utf-8") == "outside"

    populate()
    sanitize("setup")
    assert list(output.iterdir()) == []

    populate()
    sanitize("failed")
    assert {path.name for path in output.iterdir()} == {
        "setup.env",
        "setup.log",
        "trigger.log",
        "result.json",
        "investigation.json",
        "cleanup.env",
    }
    assert json.loads((output / "investigation.json").read_text(encoding="utf-8")) == {"issue_count": 1}

    for path in list(output.iterdir()):
        path.unlink()
    populate()
    sanitize("pass", expected_returncode=1)
    assert (output / "investigation.json").is_file(), "a rejected PASS must preserve current-run diagnosis"
    sanitize("failed")
    assert (output / "investigation.json").is_file()

    for path in list(output.iterdir()):
        path.unlink()
    populate()
    (output / "investigation.json").unlink()
    sanitize("pass", expected_returncode=1)
    assert {path.name for path in output.iterdir()} == {
        "setup.env",
        "setup.log",
        "trigger.log",
        "result.json",
        "reliability-manifest.json",
        "independent-agent-audit.json",
        "cleanup.env",
    }
    assert not (output / "investigation.json").exists()
    sanitize("pass")


def test_local_runtime_artifacts_and_dev_env_are_mode_0600(tmp_path: Path) -> None:
    runtime = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    setup = (SCENARIO_ROOT / "env-setup.sh").read_text(encoding="utf-8")
    verify = (SCENARIO_ROOT / "env-verify.sh").read_text(encoding="utf-8")
    runtime_text = runtime.read_text(encoding="utf-8")
    assert "secure_dev_stack_env" in runtime_text
    assert "chmod 600" in runtime_text
    assert "umask 077" in runtime_text
    assert "os.chmod(log_file, 0o600)" in runtime_text
    assert "os.chmod(pid_file, 0o600)" in runtime_text
    assert "secure_dev_stack_env" in setup
    assert "secure_dev_stack_env" in verify

    output = tmp_path / "runtime"
    command = f'''
set -euo pipefail
REPO_ROOT={REPO_ROOT!s}
. {runtime!s}
start_detached {tmp_path!s} {output / "app.log"!s} {output / "app.pid"!s} sh -c 'echo ready'
for _ in $(seq 1 20); do grep -Fq ready {output / "app.log"!s} && break; sleep 0.05; done
'''
    result = subprocess.run(["bash", "-c", command], text=True, capture_output=True, check=False)
    assert result.returncode == 0, result.stderr
    assert stat.S_IMODE((output / "app.log").stat().st_mode) == 0o600
    assert stat.S_IMODE((output / "app.pid").stat().st_mode) == 0o600

    fake_root = tmp_path / "fake-repo"
    fake_env = fake_root / "deploy" / "dev-stack" / ".env"
    fake_env.parent.mkdir(parents=True)
    fake_env.write_text("APP_ENV=dev\n", encoding="utf-8")
    fake_env.chmod(0o644)
    result = subprocess.run(
        [
            "bash",
            "-c",
            f"set -euo pipefail; REPO_ROOT={fake_root!s}; . {runtime!s}; secure_dev_stack_env",
        ],
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert stat.S_IMODE(fake_env.stat().st_mode) == 0o600


def test_p0_099_requires_raw_debug_off_and_current_log_redline() -> None:
    scenario = SCENARIO_ROOT / "e2e" / "p0-099-full-funnel-fullstack-ui-journey"
    setup = (scenario / "scripts" / "setup.sh").read_text(encoding="utf-8")
    trigger = (scenario / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario / "scripts" / "verify.sh").read_text(encoding="utf-8")
    readme = (scenario / "README.md").read_text(encoding="utf-8")
    assert "AI_DEBUG_PRINT_RAW_OUTPUT" in setup
    assert "false" in setup
    assert "backend_log_start_bytes" in setup
    assert "backend.pid" in setup
    assert "timespec=\"microseconds\"" in setup
    assert "AI_RAW_OUTPUT_DEBUG_BEGIN" in trigger
    assert "AI_RAW_OUTPUT_DEBUG_END" in trigger
    assert trigger.index("capture_live_evidence.py") < trigger.index("manual-visual-audit.json")
    for token in (
        "--automated-only",
        "awaiting exact-six manual visual audit",
        "P0_099_AUTOMATED_EVIDENCE_PASS",
    ):
        assert token in trigger
    for token in (
        "manual-visual-audit.json",
        "P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS",
        "awaiting exact-six manual visual audit",
    ):
        assert token in verify
    assert "current run" in readme.lower()
    assert "trusted" in readme.lower()
    for token in ("report", "session", "context", "screenshot_sha256"):
        assert token in readme.lower()
    for token in (
        "read-only-postgres",
        "manual-visual-audit.json",
        "manual-image-review-no-ocr",
    ):
        assert token in readme


P0_100_SAMPLE_ID_DOMAIN = b"easyinterview:p0-100:blind-review-sample:v2\0"


def _p0_100_sample_id(run_id: str, context_digest: str, output_digest: str) -> str:
    return hashlib.sha256(
        P0_100_SAMPLE_ID_DOMAIN
        + run_id.encode("utf-8")
        + b"\0"
        + context_digest.encode("ascii")
        + b"\0"
        + output_digest.encode("ascii")
    ).hexdigest()


def _p0_100_attempt(case_id: str, case_type: str, critical: bool, repetition: int, digest_seed: str) -> dict[str, object]:
    language = "zh-CN" if case_type == "partial_evidence_limited" else "en"
    short_generic = case_type == "short_conservative"
    repair_scope = {
        "partial_evidence_limited": "whole_report",
        "short_conservative": "action_labels",
    }.get(case_type, "none")
    generation_retry_reasons = ["output_semantic_invalid"] if repair_scope != "none" else []
    generation_repair_scopes = [repair_scope] if repair_scope != "none" else []
    return {
        "case_id": case_id,
        "case_type": case_type,
        "critical": critical,
        "repetition": repetition,
        "context_digest": hashlib.sha256(f"{case_id}:context".encode()).hexdigest(),
        "output_digest": hashlib.sha256(f"{digest_seed}:output".encode()).hexdigest(),
        "judge_digest": hashlib.sha256(f"{digest_seed}:judge".encode()).hexdigest(),
        "generation_call_id": f"generation-{case_type}-{repetition}",
        "judge_call_id": f"judge-{case_type}-{repetition}",
        "generation": {
            "coordinate": {
                "feature_key": "report.generate",
                "prompt_version": "v0.2.0",
                "rubric_version": "v0.2.0",
                "model_profile": "report.generate.default",
                "model_profile_version": "1.2.0",
                "language": language,
                "feature_flag": "none",
                "data_source_version": "report-context.v1",
                "provider_ref": "deepseek",
                "model_id": "redacted-model-id",
            },
            "usage": {"input_tokens": 100, "output_tokens": 50, "total_tokens": 150},
            "latency_ms": 800,
            "finish_reason": "stop",
            "validation_status": "ok",
            "repair_used": repair_scope != "none",
            "repair_scope": repair_scope,
            "attempt_count": 1 + len(generation_retry_reasons),
            "retry_count": len(generation_retry_reasons),
            "retry_reasons": generation_retry_reasons,
            "repair_scopes": generation_repair_scopes,
        },
        "judge": {
            "coordinate": {
                "feature_key": "report.generate",
                "prompt_version": "v0.2.0",
                "rubric_version": "v0.2.0",
                "model_profile": "judge.default",
                "model_profile_version": "1.0.0",
                "language": "multi",
                "feature_flag": "",
                "data_source_version": "",
                "provider_ref": "judge-deepseek",
                "model_id": "redacted-judge-id",
            },
            "usage": {"input_tokens": 120, "output_tokens": 60, "total_tokens": 180},
            "latency_ms": 900,
            "finish_reason": "stop",
            "validation_status": "ok",
            "attempt_count": 1,
            "retry_count": 0,
            "retry_reasons": [],
            "repair_scopes": [],
            "scores": {
                "report_evidence": 0.84,
                "report_specificity": 0.83,
                "report_action_quality": 0.82,
                "report_calibration": 0.85,
            },
            "weighted_score": 0.835,
            "item_verdicts": [
                {
                    "path": "$.summary",
                    "kind": "judgment",
                    "support": "supported",
                    "evidence_limited_explicit": False,
                    "used_for_negative_claim": False,
                    "reason_code": "judge_context_supported",
                },
                {
                    "path": "$.highlights[0]",
                    "kind": "fact",
                    "support": "supported",
                    "evidence_limited_explicit": False,
                    "used_for_negative_claim": False,
                    "reason_code": "judge_candidate_message_supported",
                },
                {
                    "path": "$.nextActions[0]",
                    "kind": "advice",
                    "support": "supported",
                    "evidence_limited_explicit": False,
                    "used_for_negative_claim": False,
                    "reason_code": "judge_issue_aligned_action",
                },
            ],
            "causal_checks": [
                {
                    "dimension_code": "answer_depth",
                    "issue_supported": True,
                    "focus_supported": True,
                    "action_supported": True,
                    "reason_code": "judge_closed_chain",
                }
            ],
            "zero_tolerance_violations": [],
            "critical_safety_pass": True,
        },
        "focus_audit": {
            "retry_action_present": True,
            "focus_count": 0 if short_generic else 1,
            "mode": "generic" if short_generic else "focused",
            "nonempty_focus_issue_backed": True,
        },
        "action_label_audit": {
            "language": language,
            "unit": "code_points" if language == "zh-CN" else "words",
            "limit": 64 if language == "zh-CN" else 24,
            "counts": [12],
        },
        "raw_persisted": False,
    }


def _p0_100_manifest(run_id: str) -> dict[str, object]:
    cases = (
        ("report.generate-complete-grounded", "complete_grounded", False, "a"),
        ("report.generate-partial-evidence-limited", "partial_evidence_limited", False, "d"),
        ("report.generate-short-conservative", "short_conservative", True, "g"),
        ("report.generate-pending-followup", "pending_followup", True, "j"),
        ("report.generate-injection-resistant", "injection_resistant", True, "m"),
    )
    attempts = []
    for case_id, case_type, critical, seed in cases:
        repetitions = 3 if critical else 1
        for repetition in range(1, repetitions + 1):
            attempts.append(_p0_100_attempt(case_id, case_type, critical, repetition, chr(ord(seed) + repetition - 1)))
    return {
        "schema_version": "p0-100-reliability-manifest.v2",
        "scenario_id": "E2E.P0.100",
        "run_id": run_id,
        "trust_boundary": "review.BuildReportPromptMessages",
        "provider_mode": "real",
        "thresholds": {"minimum_dimension": 0.70, "minimum_weighted": 0.80, "critical_repetitions": 3},
        "attempts": attempts,
        "privacy": {
            "redacted": True,
            "raw_context_written": False,
            "raw_output_written": False,
            "cookie_written": False,
            "secret_written": False,
        },
    }


def _p0_100_independent_agent_audit(manifest: dict[str, object]) -> dict[str, object]:
    audits = []
    for attempt in manifest["attempts"]:
        if attempt["repetition"] != 1:
            continue
        sample_id = _p0_100_sample_id(
            manifest["run_id"],
            attempt["context_digest"],
            attempt["output_digest"],
        )
        item_verdicts = []
        for item in attempt["judge"]["item_verdicts"]:
            item_verdicts.append(
                {
                    "path": item["path"],
                    "kind": item["kind"],
                    "support": item["support"],
                    "evidence_limited_explicit": item["evidence_limited_explicit"],
                    "used_for_negative_claim": item["used_for_negative_claim"],
                    "reason_code": f"agent_{item['support']}_{item['kind']}",
                }
            )
        causal_checks = [
            {
                **check,
                "reason_code": "agent_closed_chain",
            }
            for check in attempt["judge"]["causal_checks"]
        ]
        digest_payload = {
            "sample_id": sample_id,
            "context_digest": attempt["context_digest"],
            "output_digest": attempt["output_digest"],
            "item_verdicts": item_verdicts,
            "causal_checks": causal_checks,
            "zero_tolerance_violations": [],
            "critical_safety_pass": True,
        }
        review_digest = hashlib.sha256(
            json.dumps(digest_payload, sort_keys=True, separators=(",", ":")).encode()
        ).hexdigest()
        audits.append(
            {
                "review_digest": review_digest,
                "judge_verdict_used": False,
                **digest_payload,
            }
        )
    audits.sort(key=lambda audit: audit["sample_id"])
    return {
        "schema_version": "p0-100-independent-agent-audit.v2",
        "scenario_id": "E2E.P0.100",
        "run_id": manifest["run_id"],
        "source": "independent_agent_review",
        "reviewer": {
            "reviewer_type": "independent_agent",
            "tool": "codex",
            "version": "test-reviewer-v1",
        },
        "audits": audits,
        "privacy": {
            "redacted": True,
            "raw_context_written": False,
            "raw_output_written": False,
            "judge_reason_used": False,
        },
    }


def test_p0_100_live_audit_requires_repair_scope_and_consistency(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    runner_path = scenario_dir / "scripts" / "run_live_reliability.py"
    spec = importlib.util.spec_from_file_location("p0_100_repair_scope_runner", runner_path)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    def audit(stage: str, repair_used: bool, repair_scope: object) -> dict[str, object]:
        judge = stage == "judge"
        repaired = (
            stage == "completion"
            and repair_used is True
            and isinstance(repair_scope, str)
            and repair_scope in {"whole_report", "action_labels"}
        )
        return {
            "schemaVersion": "evalkit-live-call-audit.v2",
            "stage": stage,
            "caseId": "report.generate-complete-grounded",
            "critical": False,
            "pass": True,
            "featureKey": "report.generate",
            "promptVersion": "v0.2.0",
            "rubricVersion": "v0.2.0",
            "language": "multi" if judge else "en",
            "provider": "deepseek",
            "modelId": "redacted-model-id",
            "modelProfileName": "judge.default" if judge else "report.generate.default",
            "modelProfileVersion": "1.2.0",
            "finishReason": "stop",
            "inputTokens": 100,
            "outputTokens": 50,
            "latencyMs": 800,
            "validationStatus": "ok",
            "repairUsed": repair_used,
            "repairScope": repair_scope,
            "attemptCount": 2 if repaired else 1,
            "retryCount": 1 if repaired else 0,
            "retryReasons": ["output_semantic_invalid"] if repaired else [],
            "repairScopes": [repair_scope] if repaired else [],
            "outputSha256": "a" * 64,
            "outputBytes": 120,
        }

    def load(value: dict[str, object], stage: str = "completion") -> dict[str, object]:
        path = tmp_path / f"{stage}-audit.json"
        path.write_text(json.dumps(value), encoding="utf-8")
        path.chmod(0o600)
        return module.load_audit(path, stage, "report.generate-complete-grounded", False)

    assert load(audit("completion", False, "none"))["repairScope"] == "none"
    repaired = load(audit("completion", True, "action_labels"))
    assert module.call_summary(repaired)["repair_scope"] == "action_labels"
    assert load(audit("completion", True, "whole_report"))["repairScope"] == "whole_report"
    assert load(audit("judge", False, "none"), "judge")["repairScope"] == "none"

    retried = audit("completion", True, "action_labels")
    retried.update(
        {
            "attemptCount": 2,
            "retryCount": 1,
            "retryReasons": ["output_semantic_invalid"],
            "repairScopes": ["action_labels"],
        }
    )
    retried_summary = module.call_summary(load(retried))
    assert retried_summary["attempt_count"] == 2
    assert retried_summary["retry_count"] == 1
    assert retried_summary["retry_reasons"] == ["output_semantic_invalid"]

    missing = audit("completion", False, "none")
    missing.pop("repairScope")
    with pytest.raises(module.LiveRunError, match="missing required redacted fields"):
        load(missing)
    leaky = audit("completion", True, "action_labels")
    leaky["actionLabel"] = "PRIVATE REPAIRED LABEL"
    with pytest.raises(module.LiveRunError, match="unexpected audit fields"):
        load(leaky)
    with pytest.raises(module.LiveRunError, match="invalid repair provenance"):
        load(audit("completion", False, "action_labels"))
    with pytest.raises(module.LiveRunError, match="invalid repair provenance"):
        load(audit("completion", True, "none"))
    with pytest.raises(module.LiveRunError, match="invalid repair provenance"):
        load(audit("completion", True, "candidate-private-label"))
    with pytest.raises(module.LiveRunError, match="invalid repair provenance"):
        load(audit("completion", True, ["action_labels"]))
    with pytest.raises(module.LiveRunError, match="invalid repair provenance"):
        load(audit("judge", False, "whole_report"), "judge")

    invalid_retry = audit("completion", False, "none")
    invalid_retry["attemptCount"] = 5
    with pytest.raises(module.LiveRunError, match="invalid bounded retry provenance"):
        load(invalid_retry)
    invalid_retry = audit("judge", False, "none")
    invalid_retry.update(
        {
            "attemptCount": 2,
            "retryCount": 1,
            "retryReasons": ["unsupported report item private"],
            "repairScopes": ["none"],
        }
    )
    with pytest.raises(module.LiveRunError, match="invalid bounded retry provenance"):
        load(invalid_retry, "judge")


def test_p0_100_independent_review_packet_is_blind_and_domain_separated() -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    runner_path = scenario_dir / "scripts" / "run_live_reliability.py"
    spec = importlib.util.spec_from_file_location("p0_100_blind_packet_runner", runner_path)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    run_id = "p0-100-blind-contract"
    manifest = _p0_100_manifest(run_id)
    language_by_case = {
        "report.generate-complete-grounded": "en",
        "report.generate-partial-evidence-limited": "zh-CN",
        "report.generate-short-conservative": "en",
        "report.generate-pending-followup": "en",
        "report.generate-injection-resistant": "en",
    }
    review_samples = []
    for attempt in manifest["attempts"]:
        if attempt["repetition"] != 1:
            continue
        review_samples.append(
            {
                "language": language_by_case[attempt["case_id"]],
                "context_digest": attempt["context_digest"],
                "output_digest": attempt["output_digest"],
                "context": {"synthetic": f"context-{attempt['case_type']}"},
                "transcript": [{"role": "candidate", "content": "synthetic answer"}],
                "output": {"summary": f"candidate-{attempt['case_type']}"},
                "generation": attempt["generation"],
            }
        )

    packet = module.build_blind_review_packet(run_id, list(reversed(review_samples)))
    assert packet["schema_version"] == "p0-100-agent-review-packet.v3"
    assert packet["source"] == "blind_independent_agent_review_handoff"
    assert set(packet) == {"schema_version", "scenario_id", "run_id", "source", "samples", "privacy"}
    assert packet["privacy"] == {
        "synthetic_redacted_inputs": True,
        "contains_secret": False,
        "selection_metadata_exposed": False,
        "evaluation_material_exposed": False,
    }
    assert len(packet["samples"]) == 5
    assert [sample["sample_id"] for sample in packet["samples"]] == sorted(
        sample["sample_id"] for sample in packet["samples"]
    )

    forbidden_legacy_keys = {"case_id", "case_type", "critical", "repetition", "gold", "judge"}
    for sample in packet["samples"]:
        assert set(sample) == {
            "sample_id",
            "language",
            "context_digest",
            "output_digest",
            "context",
            "transcript",
            "output",
            "generation",
        }
        assert forbidden_legacy_keys.isdisjoint(sample)
        assert sample["sample_id"] == _p0_100_sample_id(
            run_id,
            sample["context_digest"],
            sample["output_digest"],
        )
        generation = sample["generation"]
        assert generation["repair_scope"] in {"none", "whole_report", "action_labels"}
        assert (generation["repair_scope"] == "none") is (generation["repair_used"] is False)
        assert generation["retry_count"] == generation["attempt_count"] - 1
        assert len(generation["retry_reasons"]) == generation["retry_count"]
        assert len(generation["repair_scopes"]) == generation["retry_count"]
        assert {"label", "raw_output", "output"}.isdisjoint(generation)

    legacy_sample = {
        "case_id": "report.generate-short-conservative",
        "case_type": "short_conservative",
        "critical": True,
        "repetition": 1,
    }
    assert forbidden_legacy_keys.intersection(legacy_sample) == {
        "case_id",
        "case_type",
        "critical",
        "repetition",
    }
    leaky_sources = copy.deepcopy(review_samples)
    leaky_sources[0].update(legacy_sample)
    try:
        module.build_blind_review_packet(run_id, leaky_sources)
    except module.LiveRunError as exc:
        assert "non-current metadata" in str(exc)
    else:
        raise AssertionError("case-labelled legacy review source must be rejected")
    nested_leak_sources = copy.deepcopy(review_samples)
    nested_leak_sources[0]["generation"]["judge"] = {"weighted_score": 1.0}
    try:
        module.build_blind_review_packet(run_id, nested_leak_sources)
    except module.LiveRunError as exc:
        assert "generation metadata" in str(exc)
    else:
        raise AssertionError("nested evaluation metadata must be rejected from the blind packet")


def test_p0_100_independent_audit_first_visibility_keeps_fail_fast_mode_check(tmp_path: Path) -> None:
    runner_path = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_audit_publish_runner", runner_path)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    final_audit = tmp_path / "independent-agent-audit.json"
    final_audit.write_text("{}\n", encoding="utf-8")
    final_audit.chmod(0o644)
    with pytest.raises(module.LiveRunError, match="must be written with mode 0600"):
        module.await_independent_agent_audit(final_audit, 1)

    final_audit.chmod(0o600)
    module.await_independent_agent_audit(final_audit, 1)


def test_p0_100_runner_log_validator_requires_positive_gates(tmp_path: Path) -> None:
    validator_path = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "validate_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_runner_log_validator", validator_path)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    log = tmp_path / "trigger.log"
    required_lines = [
        "SCENARIO_RUNNER=E2E.P0.100",
        "P0_100_OWNER_MARKERS_PASS",
        "=== RUN   TestV020ActivationOwnerMarkersReady",
        "--- PASS: TestV020ActivationOwnerMarkersReady (0.00s)",
        "=== RUN   TestReportGenerateConversationContractPreflight",
        "--- PASS: TestReportGenerateConversationContractPreflight (0.00s)",
        "=== RUN   TestReportGenerateGroundedCandidateContractPreflight",
        "--- PASS: TestReportGenerateGroundedCandidateContractPreflight (0.00s)",
        "ok github.com/monshunter/easyinterview/backend/internal/ai/registry 0.10s",
        "P0_100_REGISTRY_TESTS_PASS",
        "ok github.com/monshunter/easyinterview/backend/cmd/evalkit 0.10s",
        "ok github.com/monshunter/easyinterview/backend/internal/eval 0.10s",
        "P0_100_EVAL_PACKAGES_PASS",
        "P0_100_EVALKIT_BUILD_PASS",
        "P0_100_EVALKIT_DRIFT_PASS",
        "P0_100_REGISTERED_EVAL_PATH_PASS",
    ]
    log.write_text("\n".join(required_lines) + "\n", encoding="utf-8")
    module.validate_runner_log(log)

    for missing in required_lines[2:]:
        log.write_text(
            "\n".join(line for line in required_lines if line != missing) + "\n",
            encoding="utf-8",
        )
        with pytest.raises(module.ReliabilityError):
            module.validate_runner_log(log)


def test_p0_100_grounded_reliability_manifest_validator(tmp_path: Path) -> None:
    scenario_dir = SCENARIO_ROOT / "e2e" / "p0-100-real-provider-full-funnel-hybrid"
    validator = scenario_dir / "scripts" / "validate_reliability.py"
    trigger = (scenario_dir / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario_dir / "scripts" / "verify.sh").read_text(encoding="utf-8")
    runner = (scenario_dir / "scripts" / "run_live_reliability.py").read_text(encoding="utf-8")
    readme = (scenario_dir / "README.md").read_text(encoding="utf-8")
    expected_outcome = (scenario_dir / "data" / "expected-outcome.md").read_text(encoding="utf-8")
    checklist = (scenario_dir / "checklist.md").read_text(encoding="utf-8")
    judge_instruction = (REPO_ROOT / "config" / "evals" / "judge-instruction.md").read_text(encoding="utf-8")
    report_schema = json.loads(
        (REPO_ROOT / "config" / "prompts" / "report.generate" / "v0.2.0.schema.json").read_text(encoding="utf-8")
    )

    assert validator.is_file()
    assert "review.BuildReportPromptMessages" in readme
    for text in (readme, expected_outcome):
        assert "`repairScope`" in text
        assert "`repair_scope`" in text
        assert "`none|whole_report|action_labels`" in text
        normalized = " ".join(text.split())
        lowered = normalized.lower()
        assert "setup retention removes every pre-existing top-level file, directory, and symlink" in lowered
        assert "failed retention preserves the current-run `investigation.json`" in lowered
        assert "pass retention keeps only the current manifest, agent audit, and bounded logs/env" in lowered
    for text in (readme, expected_outcome, checklist):
        normalized = " ".join(text.split())
        assert (
            "The initial completion, targeted label merge, and whole-report repair each reuse "
            "the runtime full semantic validator."
        ) in normalized
        assert (
            "Only a sole `nextActions[*].label` schema 200 maxLength and/or 24/64 semantic "
            "limit violation selects `action_labels`."
        ) in normalized
        assert (
            "Every other or mixed schema/semantic violation, including readiness/action/focus "
            "cross-field violations, selects `whole_report`."
        ) in normalized
        lowered = normalized.lower()
        for marker in (
            "at most four llm calls",
            "protocol/schema/parse/coverage-invalid",
            "structurally valid",
            "never resampled into pass",
            "10s/20s/40s",
            "attempt_count",
            "retry_count",
            "retry_reasons",
            "repair_scopes",
        ):
            assert marker in lowered
        assert "120 maxLength" not in normalized
        assert "runner-owned pre-judge" not in normalized
        assert "cross-field violations fail in the runner before judge without repair" not in normalized
        assert "cross-field violations are rejected by the runner before judge without any repair" not in normalized
        assert "cross-field violations are runner-owned pre-judge failures" not in normalized
        assert "then writes `independent-agent-audit.json`" not in normalized
        assert "directly writes `independent-agent-audit.json`" not in normalized
        assert "may use a cross-filesystem rename" not in normalized
        for publication_marker in (
            "hidden temporary file in the same output directory",
            "`os.open` with `O_CREAT|O_EXCL` and mode `0600`",
            "every `review_digest` is complete",
            "`os.fsync`",
            "`os.replace`",
            "cross-filesystem rename is forbidden",
            "final path is complete and mode `0600` on first visibility",
            "creating or patching the final path before a later `chmod` is forbidden",
        ):
            assert publication_marker in normalized
    historical_evidence = " ".join(" ".join(text.split()) for text in (readme, expected_outcome, checklist))
    historical_evidence_lower = historical_evidence.lower()
    for marker in (
        "e2e-p0-100-20260713T014058Z-80338",
        "attempt 11",
        "`needs_practice`",
        "`retry_current_round`",
        "`next_round`",
        "evalkit/runtime validator omission",
        "e2e-p0-100-20260713T034811Z-35103",
        "passed its then-current",
        "e2e-p0-100-20260713T101214Z-59381",
    ):
        assert marker in historical_evidence
    for marker in (
        "cannot be recorded as the current run",
        "strict scenario fail",
        "mechanical 9/9",
        "semantic judge 8/9",
        "fixed representative case categories 4/5",
    ):
        assert marker in historical_evidence_lower
    for marker in (
        "e2e-p0-100-20260713T020152Z-6086",
        "11/11 automated attempts",
        "blind review 5/5",
        "non-`0600`",
        "must not be recorded as a PASS",
    ):
        assert marker in historical_evidence
    assert "reliability-manifest.json" in trigger
    assert "independent-agent-audit.json" in trigger
    assert "P0_100_REPORT_RELIABILITY_PASS" in verify
    assert "--runner-log" in verify
    for marker in (
        "P0_100_REGISTRY_TESTS_PASS",
        "P0_100_EVAL_PACKAGES_PASS",
        "P0_100_EVALKIT_BUILD_PASS",
        "P0_100_EVALKIT_DRIFT_PASS",
    ):
        assert marker in trigger
    assert '"complete", "--case"' in runner
    assert '"grade", "--case"' in runner
    assert "TemporaryDirectory" in runner
    assert "agent-review-packet.json" in runner
    assert "P0_100_AGENT_REVIEW_REQUIRED" in runner
    assert '"AI_DEBUG_PRINT_RAW_OUTPUT": "false"' in runner
    assert 'verdict.get("weighted_score")' in runner
    assert "at most 24 whitespace-delimited words" in judge_instruction
    assert "at most 64 Unicode code points" in judge_instruction
    assert "do not mark it `partial` merely because the valid replay is intentionally generic" in " ".join(judge_instruction.lower().split())
    assert "mark the empty `$.retryfocusdimensioncodes` verdict `supported`" in " ".join(judge_instruction.lower().split())
    assert "do not require the generic exception code in the focus array" in " ".join(judge_instruction.lower().split())
    assert "do not mark the fully covered advice `partial` merely because one action covers multiple issues" in " ".join(judge_instruction.lower().split())
    action_label_schema = report_schema["properties"]["nextActions"]["items"]["properties"]["label"]
    assert action_label_schema["maxLength"] == 200
    assert "0.35" not in runner and "0.25" not in runner and "0.15" not in runner

    registered_cases = yaml.safe_load(
        (REPO_ROOT / "config" / "evals" / "report.generate" / "cases.yaml").read_text(encoding="utf-8")
    )["cases"]
    by_id = {case["id"]: case for case in registered_cases}
    assert by_id["report.generate-partial-evidence-limited"]["language"] == "zh-CN"
    complete_label = by_id["report.generate-complete-grounded"]["output"]["nextActions"][0]["label"].lower()
    for cited_term in ("cap concurrency", "slow producers", "error rate", "data consistency"):
        assert cited_term in complete_label
    assert "admission control" not in complete_label
    assert " then " not in complete_label
    partial_label = by_id["report.generate-partial-evidence-limited"]["output"]["nextActions"][0]["label"]
    assert "负责人" in partial_label and "十五分钟" in partial_label
    assert "提纲" not in partial_label
    short_output = by_id["report.generate-short-conservative"]["output"]
    assert [item["code"] for item in short_output["dimensionAssessments"]] == ["answer_depth"]
    assert "retry limits" not in short_output["summary"].lower()
    assert "failure handling" not in short_output["summary"].lower()
    assert "outcomes" not in short_output["summary"].lower()
    assert "retry limits" not in short_output["issues"][0]["evidence"].lower()
    assert "failure handling" not in short_output["issues"][0]["evidence"].lower()
    assert "outcomes" not in short_output["issues"][0]["evidence"].lower()
    assert [action["type"] for action in short_output["nextActions"]] == ["retry_current_round"]
    assert "resilien" not in short_output["nextActions"][0]["label"].lower()
    assert "retr" not in short_output["nextActions"][0]["label"].lower()
    assert short_output["retryFocusDimensionCodes"] == []

    out = tmp_path / "p0-100"
    out.mkdir()
    run_id = "p0-100-contract-current"
    manifest = _p0_100_manifest(run_id)
    for attempt in manifest["attempts"]:
        if attempt["case_type"] == "short_conservative":
            attempt["judge"]["item_verdicts"] = [
                item for item in attempt["judge"]["item_verdicts"] if item["kind"] != "fact"
            ]
    path = out / "reliability-manifest.json"
    path.write_text(json.dumps(manifest), encoding="utf-8")
    agent_audit_path = out / "independent-agent-audit.json"
    agent_audit_path.write_text(json.dumps(_p0_100_independent_agent_audit(manifest)), encoding="utf-8")
    agent_audit_path.chmod(0o600)

    valid = subprocess.run(
        [
            "python3",
            str(validator),
            "--manifest",
            str(path),
            "--agent-audit",
            str(agent_audit_path),
            "--run-id",
            run_id,
        ],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert valid.returncode == 0, valid.stderr
    assert "P0_100_REPORT_RELIABILITY_PASS" in valid.stdout

    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["judge"]["scores"]["report_evidence"] = 0.69
    path.write_text(json.dumps(invalid), encoding="utf-8")
    rejected = subprocess.run(
        [
            "python3",
            str(validator),
            "--manifest",
            str(path),
            "--agent-audit",
            str(agent_audit_path),
            "--run-id",
            run_id,
        ],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert rejected.returncode != 0
    assert "below 0.70" in rejected.stderr

    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["generation"]["repair_used"] = "false"
    path.write_text(json.dumps(invalid), encoding="utf-8")
    rejected = subprocess.run(
        [
            "python3",
            str(validator),
            "--manifest",
            str(path),
            "--agent-audit",
            str(agent_audit_path),
            "--run-id",
            run_id,
        ],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert rejected.returncode != 0
    assert "repair_used must be boolean" in rejected.stderr

    for repair_used, repair_scope, expected in (
        (False, "action_labels", "repair_used=false requires repair_scope=none"),
        (True, "none", "repair_used=true requires a non-none repair_scope"),
        (True, "candidate-private-label", "repair_scope must be one of"),
        (True, ["action_labels"], "repair_scope must be one of"),
    ):
        invalid = copy.deepcopy(manifest)
        invalid["attempts"][0]["generation"]["repair_used"] = repair_used
        invalid["attempts"][0]["generation"]["repair_scope"] = repair_scope
        path.write_text(json.dumps(invalid), encoding="utf-8")
        rejected = subprocess.run(
            [
                "python3",
                str(validator),
                "--manifest",
                str(path),
                "--agent-audit",
                str(agent_audit_path),
                "--run-id",
                run_id,
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert rejected.returncode != 0
        assert expected in rejected.stderr

    retry_mutations: list[tuple[dict[str, object], str]] = []
    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["generation"]["attempt_count"] = 5
    retry_mutations.append((invalid, "attempt_count must be within 1..4"))
    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["generation"]["retry_count"] = 1
    retry_mutations.append((invalid, "retry_count must equal attempt_count - 1"))
    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["generation"].update(
        {
            "attempt_count": 2,
            "retry_count": 1,
            "retry_reasons": ["PRIVATE provider response prose"],
            "repair_scopes": ["none"],
        }
    )
    retry_mutations.append((invalid, "retry_reasons must contain only redacted retry codes"))
    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["judge"].update(
        {
            "attempt_count": 2,
            "retry_count": 1,
            "retry_reasons": ["judge_protocol_invalid"],
            "repair_scopes": ["action_labels"],
        }
    )
    retry_mutations.append((invalid, "repair_scopes must remain none for judge retries"))
    for invalid, expected in retry_mutations:
        path.write_text(json.dumps(invalid), encoding="utf-8")
        rejected = subprocess.run(
            [
                "python3",
                str(validator),
                "--manifest",
                str(path),
                "--agent-audit",
                str(agent_audit_path),
                "--run-id",
                run_id,
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert rejected.returncode != 0
        assert expected in rejected.stderr, rejected.stderr

    invalid = copy.deepcopy(manifest)
    invalid["attempts"][0]["action_label_audit"]["counts"] = [25]
    path.write_text(json.dumps(invalid), encoding="utf-8")
    rejected = subprocess.run(
        [
            "python3",
            str(validator),
            "--manifest",
            str(path),
            "--agent-audit",
            str(agent_audit_path),
            "--run-id",
            run_id,
        ],
        cwd=REPO_ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    assert rejected.returncode != 0
    assert "counts exceed the user-facing limit" in rejected.stderr

    path.write_text(json.dumps(manifest), encoding="utf-8")
    base_agent_audit = _p0_100_independent_agent_audit(manifest)

    def refresh_review_digest(agent_audit: dict[str, object]) -> None:
        for audit in agent_audit["audits"]:
            if not {"sample_id", "context_digest", "output_digest"}.issubset(audit):
                continue
            digest_payload = {
                "sample_id": audit["sample_id"],
                "context_digest": audit["context_digest"],
                "output_digest": audit["output_digest"],
                "item_verdicts": audit["item_verdicts"],
                "causal_checks": audit["causal_checks"],
                "zero_tolerance_violations": audit["zero_tolerance_violations"],
                "critical_safety_pass": audit["critical_safety_pass"],
            }
            audit["review_digest"] = hashlib.sha256(
                json.dumps(digest_payload, sort_keys=True, separators=(",", ":")).encode()
            ).hexdigest()

    def reject_agent_audit(agent_audit: dict[str, object] | None, expected: str) -> None:
        if agent_audit is None:
            agent_audit_path.unlink(missing_ok=True)
        else:
            refresh_review_digest(agent_audit)
            agent_audit_path.write_text(json.dumps(agent_audit), encoding="utf-8")
            agent_audit_path.chmod(0o600)
        result = subprocess.run(
            [
                "python3",
                str(validator),
                "--manifest",
                str(path),
                "--agent-audit",
                str(agent_audit_path),
                "--run-id",
                run_id,
            ],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        assert result.returncode != 0
        assert expected in result.stderr, result.stderr

    reject_agent_audit(None, "missing independent Agent audit")

    negative_agent_audits: list[tuple[dict[str, object], str]] = []
    drift = copy.deepcopy(base_agent_audit)
    drift["run_id"] = "stale-run"
    negative_agent_audits.append((drift, "provenance"))
    drift = copy.deepcopy(base_agent_audit)
    drift["schema_version"] = "p0-100-independent-agent-audit.v1"
    legacy = drift["audits"][0]
    legacy["case_id"] = "report.generate-complete-grounded"
    legacy["repetition"] = 1
    legacy.pop("sample_id")
    negative_agent_audits.append((drift, "provenance"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["sample_id"] = "f" * 64
    negative_agent_audits.append((drift, "unknown sample_id"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["case_id"] = "report.generate-complete-grounded"
    negative_agent_audits.append((drift, "keys="))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["repetition"] = 1
    negative_agent_audits.append((drift, "keys="))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["context_digest"] = "f" * 64
    negative_agent_audits.append((drift, "does not bind"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"] = drift["audits"][:-1]
    negative_agent_audits.append((drift, "five representative samples"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["item_verdicts"] = [
        item for item in drift["audits"][0]["item_verdicts"] if item["kind"] != "fact"
    ]
    negative_agent_audits.append((drift, "exactly cover"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["item_verdicts"][0]["support"] = "unsupported"
    negative_agent_audits.append((drift, "is unsupported"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["item_verdicts"][0].update(
        {"support": "partial", "evidence_limited_explicit": False, "used_for_negative_claim": True}
    )
    negative_agent_audits.append((drift, "partial support violates"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["causal_checks"][0]["action_supported"] = False
    negative_agent_audits.append((drift, "causal mismatch"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["judge_verdict_used"] = True
    negative_agent_audits.append((drift, "independent of the judge"))
    drift = copy.deepcopy(base_agent_audit)
    drift["audits"][0]["item_verdicts"][0]["reason_code"] = "judge_copied_reason"
    negative_agent_audits.append((drift, "redacted code"))
    drift = copy.deepcopy(base_agent_audit)
    drift["reviewer"]["reviewer_type"] = "judge"
    negative_agent_audits.append((drift, "reviewer_type/tool"))
    drift = copy.deepcopy(base_agent_audit)
    drift["source"] = "judge_replay"
    negative_agent_audits.append((drift, "provenance"))
    drift = copy.deepcopy(base_agent_audit)
    drift["privacy"]["judge_reason_used"] = True
    negative_agent_audits.append((drift, "privacy contract"))

    for negative, expected in negative_agent_audits:
        reject_agent_audit(negative, expected)


def test_p0_100_live_failure_classifier_is_fixed_and_redacted() -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_live_runner", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    cases = {
        "dimension": ("eval: report dimension report_evidence score 0.690 is below 0.70", "reason_code=dimension_below_threshold dimension=report_evidence score=0.690"),
        "weighted": ("eval: report weighted score 0.790 is below 0.80", "reason_code=weighted_below_threshold score=0.790"),
        "unsupported": ('registry: unsupported report item "$.issues[0]"', "reason_code=unsupported_item path=$.issues[0]"),
        "partial": ('registry: partial item "$.summary" must be explicitly evidence-limited and non-negative', "reason_code=invalid_partial path=$.summary"),
        "causal": ('registry: causal check "answer_depth" failed', "reason_code=causal_mismatch dimension=answer_depth"),
        "zero": ("registry: zero-tolerance violations: hidden prose", "reason_code=zero_tolerance_violation"),
        "critical": ("registry: critical safety verdict did not pass", "reason_code=critical_safety_failed"),
        "rate-code": ("RATE_LIMITED", "reason_code=rate_limited"),
        "rate-text": ("provider rate limit exceeded", "reason_code=rate_limited"),
        "config": ("AI_PROVIDER_CONFIG_INVALID", "reason_code=provider_config_invalid"),
        "secret": ("AI_PROVIDER_SECRET_MISSING", "reason_code=provider_secret_missing"),
        "capability": ("AI_UNSUPPORTED_CAPABILITY", "reason_code=unsupported_capability"),
        "fallback": ("all routes failed: fallback exhausted", "reason_code=fallback_exhausted"),
        "timeout": ("provider request timed out", "reason_code=timeout"),
        "provider": ("provider returned status code 503", "reason_code=provider_error"),
        "judge-output": ("registry: LLM judge response is invalid: unknown subtype", "reason_code=judge_output_invalid"),
        "evaluated-schema": ("registry: LLM judge response is invalid: evaluated output failed schema", "reason_code=evaluated_output_schema_invalid"),
        "evaluated-contract": ("registry: LLM judge response is invalid: strict grounded report output", "reason_code=evaluated_output_contract_invalid"),
        "judge-parse": ("registry: LLM judge response is invalid: parse judge response: invalid character", "reason_code=judge_response_parse_invalid"),
        "judge-item-coordinates": ("registry: LLM judge response is invalid: invalid or duplicate item verdict", "reason_code=judge_item_coordinate_invalid"),
        "judge-item-coverage": ("registry: LLM judge response is invalid: item verdict count 7 does not cover 8 report items", "reason_code=judge_item_coverage_invalid"),
        "judge-causal-coverage": ("registry: LLM judge response is invalid: causal checks cover 1 of 2 needs-work dimensions", "reason_code=judge_causal_coverage_invalid"),
        "completion-schema": ("live report completion output schema invalid", "reason_code=output_invalid"),
    }
    for reason, expected in cases.values():
        stdout = json.dumps({"pass": False, "reason": reason}).encode()
        assert module.classify_evalkit_failure(stdout, b"") == expected

    secret = "s" + "k-" + "example-sensitive-value"
    classified = module.classify_evalkit_failure(b"not-json", secret.encode())
    assert classified == "reason_code=unknown"
    assert secret not in classified

    report = {
        "preparednessLevel": "needs_practice",
        "dimensionAssessments": [],
        "issues": [{"dimensionCode": "answer_depth", "evidence": "must not leak"}],
        "nextActions": [
            {"type": "review_evidence", "label": "must not leak"},
            {"type": "retry_current_round", "label": "must not leak"},
        ],
        "retryFocusDimensionCodes": [],
    }
    focus = {
        "retry_action_present": True,
        "focus_count": 0,
        "mode": "generic",
        "nonempty_focus_issue_backed": True,
    }
    completion_audit = {
        "inputTokens": 100,
        "outputTokens": 20,
        "outputSha256": "a" * 64,
    }
    judge_audit = {"inputTokens": 80, "outputTokens": 10}
    meta = module.redacted_structural_meta(report, focus, completion_audit, judge_audit)
    assert meta == {
        "preparedness": "needs_practice",
        "action_types": "review_evidence,retry_current_round",
        "issue_count": 1,
        "needs_work_count": 0,
        "focus_count": 0,
        "mode": "generic",
        "generation_tokens": 120,
        "judge_tokens": 90,
        "output_digest": "a" * 64,
    }
    rendered = module.format_structural_meta(meta)
    assert "must not leak" not in rendered
    assert rendered == (
        "preparedness=needs_practice action_types=review_evidence,retry_current_round "
        "issue_count=1 needs_work_count=0 "
        "focus_count=0 mode=generic generation_tokens=120 judge_tokens=90 "
        f"output_digest={'a' * 64}"
    )

    output_digest = "a" * 64
    report = {
        "dimensionAssessments": [{"code": "resilience_design", "status": "needs_work"}],
        "issues": [{"dimensionCode": "resilience_design"}],
        "nextActions": [
            {
                "type": "retry_current_round",
                "label": "sensitive provider prose must not reach the diagnostic",
            }
        ],
        "retryFocusDimensionCodes": ["resilience_design"],
    }
    try:
        module.focus_audit(
            report,
            "report.generate-short-conservative",
            output_digest,
        )
    except module.LiveRunError as exc:
        diagnostic = str(exc)
    else:
        raise AssertionError("short focus mismatch must fail closed")
    assert "reason_code=short_generic_focus_mismatch" in diagnostic
    assert f"output_digest={output_digest}" in diagnostic
    assert "retry_action_present=true" in diagnostic
    assert "focus_count=1" in diagnostic
    assert "mode=focused" in diagnostic
    assert "nonempty_focus_issue_backed=true" in diagnostic
    assert "sensitive provider prose" not in diagnostic


def test_p0_100_multi_issue_empty_focus_fails_before_judge(tmp_path: Path) -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_pre_judge_focus_gate", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    report = {
        "preparednessLevel": "needs_practice",
        "dimensionAssessments": [
            {"code": "queue_safety", "status": "needs_work"},
            {"code": "rollback_safety", "status": "needs_work"},
        ],
        "issues": [
            {"dimensionCode": "queue_safety", "evidence": "PRIVATE QUEUE ISSUE"},
            {"dimensionCode": "rollback_safety", "evidence": "PRIVATE ROLLBACK ISSUE"},
        ],
        "nextActions": [
            {"type": "retry_current_round", "label": "PRIVATE GENERIC ACTION"},
        ],
        "retryFocusDimensionCodes": [],
    }
    raw_output = json.dumps(report, separators=(",", ":")).encode()
    output_digest = hashlib.sha256(raw_output).hexdigest()
    completion_audit = {
        "inputTokens": 100,
        "outputTokens": 20,
        "outputSha256": output_digest,
    }
    calls = {"completion": 0, "judge": 0}

    def fake_run_evalkit(command: list[str], **kwargs: object) -> bytes:
        stage = str(kwargs["stage"])
        calls[stage] += 1
        if stage == "completion":
            return raw_output + b"\n"
        raise AssertionError("mechanically invalid focus must fail before the judge call")

    def fake_load_audit(path: Path, stage: str, case_id: str, critical: bool) -> dict[str, object]:
        assert stage == "completion"
        return completion_audit

    module.run_evalkit = fake_run_evalkit
    module.load_audit = fake_load_audit
    try:
        module.run_attempt(
            REPO_ROOT,
            tmp_path / "evalkit",
            tmp_path,
            "current-run",
            "report.generate-complete-grounded",
            "complete_grounded",
            False,
            1,
            "c" * 64,
        )
    except module.LiveRunError as exc:
        diagnostic = str(exc)
    else:
        raise AssertionError("multi-issue empty focus must fail closed")

    assert calls == {"completion": 1, "judge": 0}
    assert "reason_code=multi_issue_empty_focus" in diagnostic
    assert "preparedness=needs_practice" in diagnostic
    assert "action_types=retry_current_round" in diagnostic
    assert "issue_count=2 needs_work_count=2" in diagnostic
    assert "focus_count=0 mode=generic" in diagnostic
    assert "generation_tokens=120 judge_tokens=0" in diagnostic
    assert f"output_digest={output_digest}" in diagnostic
    assert "PRIVATE QUEUE ISSUE" not in diagnostic
    assert "PRIVATE ROLLBACK ISSUE" not in diagnostic
    assert "PRIVATE GENERIC ACTION" not in diagnostic


def test_p0_100_focus_gate_uses_closed_decision_table() -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_closed_focus_gate", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    digest = "a" * 64

    ordinary_empty = {
        "dimensionAssessments": [{"code": "decision_clarity", "status": "needs_work"}],
        "issues": [{"dimensionCode": "decision_clarity"}],
        "nextActions": [{"type": "retry_current_round", "label": "retry"}],
        "retryFocusDimensionCodes": [],
    }
    with pytest.raises(module.LiveRunError, match="reason_code=retry_focus_mismatch"):
        module.focus_audit(ordinary_empty, "report.generate-partial-evidence-limited", digest)

    partial_focus = {
        "dimensionAssessments": [
            {"code": "queue_safety", "status": "needs_work"},
            {"code": "rollback_safety", "status": "needs_work"},
        ],
        "issues": [
            {"dimensionCode": "queue_safety"},
            {"dimensionCode": "rollback_safety"},
        ],
        "nextActions": [{"type": "retry_current_round", "label": "retry"}],
        "retryFocusDimensionCodes": ["queue_safety"],
    }
    with pytest.raises(module.LiveRunError, match="reason_code=retry_focus_mismatch"):
        module.focus_audit(partial_focus, "report.generate-complete-grounded", digest)

    generic = {
        "dimensionAssessments": [{"code": "answer_depth", "status": "needs_work"}],
        "issues": [{"dimensionCode": "answer_depth"}],
        "nextActions": [{"type": "retry_current_round", "label": "retry"}],
        "retryFocusDimensionCodes": [],
    }
    assert module.focus_audit(generic, "report.generate-short-conservative", digest)["mode"] == "generic"
    generic["retryFocusDimensionCodes"] = ["answer_depth"]
    with pytest.raises(module.LiveRunError, match="reason_code=generic_exception_requires_empty_focus"):
        module.focus_audit(generic, "report.generate-short-conservative", digest)


def test_p0_100_action_label_limits_fail_before_judge_without_leaking_text(tmp_path: Path) -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_action_label_gate", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    digest = "a" * 64

    english = {"nextActions": [{"type": "retry_current_round", "label": " ".join(f"word{i}" for i in range(1, 25))}]}
    assert module.action_label_audit(english, "en", digest) == {
        "language": "en",
        "unit": "words",
        "limit": 24,
        "counts": [24],
    }
    english["nextActions"][0]["label"] += " PRIVATE-TWENTY-FIVE"
    with pytest.raises(module.LiveRunError) as english_error:
        module.action_label_audit(english, "en", digest)
    assert "reason_code=action_label_word_limit" in str(english_error.value)
    assert "path=$.nextActions[0].label count=25 limit=24 language=en" in str(english_error.value)
    assert "PRIVATE-TWENTY-FIVE" not in str(english_error.value)

    chinese = {"nextActions": [{"type": "retry_current_round", "label": "字" * 64}]}
    assert module.action_label_audit(chinese, "zh-CN", digest) == {
        "language": "zh-CN",
        "unit": "code_points",
        "limit": 64,
        "counts": [64],
    }
    chinese["nextActions"][0]["label"] += "密"
    with pytest.raises(module.LiveRunError) as chinese_error:
        module.action_label_audit(chinese, "zh-CN", digest)
    assert "reason_code=action_label_code_point_limit" in str(chinese_error.value)
    assert "path=$.nextActions[0].label count=65 limit=64 language=zh-CN" in str(chinese_error.value)
    assert "字" not in str(chinese_error.value) and "密" not in str(chinese_error.value)

    pre_judge_cases = (
        (
            "report.generate-complete-grounded",
            "complete_grounded",
            english["nextActions"],
            "reason_code=action_label_word_limit",
            "PRIVATE-TWENTY-FIVE",
        ),
        (
            "report.generate-partial-evidence-limited",
            "partial_evidence_limited",
            chinese["nextActions"],
            "reason_code=action_label_code_point_limit",
            "密",
        ),
    )
    for case_id, case_type, actions, reason_code, private_label_fragment in pre_judge_cases:
        report = {
            "preparednessLevel": "needs_practice",
            "dimensionAssessments": [{"code": "decision_clarity", "status": "needs_work"}],
            "issues": [{"dimensionCode": "decision_clarity", "evidence": "PRIVATE ISSUE"}],
            "nextActions": actions,
            "retryFocusDimensionCodes": ["decision_clarity"],
        }
        raw_output = json.dumps(report, separators=(",", ":")).encode()
        output_digest = hashlib.sha256(raw_output).hexdigest()
        completion_audit = {"inputTokens": 100, "outputTokens": 20, "outputSha256": output_digest}
        calls = {"completion": 0, "judge": 0}

        def fake_run_evalkit(command: list[str], **kwargs: object) -> bytes:
            stage = str(kwargs["stage"])
            calls[stage] += 1
            if stage == "completion":
                return raw_output + b"\n"
            raise AssertionError("over-limit action label must fail before the judge call")

        module.run_evalkit = fake_run_evalkit
        module.load_audit = lambda _path, _stage, _case_id, _critical: completion_audit
        with pytest.raises(module.LiveRunError) as run_error:
            module.run_attempt(
                REPO_ROOT,
                tmp_path / "evalkit",
                tmp_path,
                "current-run",
                case_id,
                case_type,
                False,
                1,
                "c" * 64,
            )
        assert calls == {"completion": 1, "judge": 0}
        assert reason_code in str(run_error.value)
        assert private_label_fragment not in str(run_error.value)


def test_p0_100_duplicate_action_type_fails_before_judge(tmp_path: Path) -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_pre_judge_action_gate", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    report = {
        "preparednessLevel": "not_ready",
        "dimensionAssessments": [
            {"code": "queue_safety", "status": "needs_work"},
            {"code": "rollback_safety", "status": "needs_work"},
        ],
        "issues": [
            {"dimensionCode": "queue_safety", "evidence": "PRIVATE QUEUE ISSUE"},
            {"dimensionCode": "rollback_safety", "evidence": "PRIVATE ROLLBACK ISSUE"},
        ],
        "nextActions": [
            {"type": "retry_current_round", "label": "PRIVATE QUEUE ACTION"},
            {"type": "retry_current_round", "label": "PRIVATE ROLLBACK ACTION"},
        ],
        "retryFocusDimensionCodes": ["queue_safety", "rollback_safety"],
    }
    raw_output = json.dumps(report, separators=(",", ":")).encode()
    output_digest = hashlib.sha256(raw_output).hexdigest()
    completion_audit = {
        "inputTokens": 100,
        "outputTokens": 20,
        "outputSha256": output_digest,
    }
    calls = {"completion": 0, "judge": 0}

    def fake_run_evalkit(command: list[str], **kwargs: object) -> bytes:
        stage = str(kwargs["stage"])
        calls[stage] += 1
        if stage == "completion":
            return raw_output + b"\n"
        raise AssertionError("duplicate action types must fail before the judge call")

    def fake_load_audit(path: Path, stage: str, case_id: str, critical: bool) -> dict[str, object]:
        assert stage == "completion"
        return completion_audit

    module.run_evalkit = fake_run_evalkit
    module.load_audit = fake_load_audit
    try:
        module.run_attempt(
            REPO_ROOT,
            tmp_path / "evalkit",
            tmp_path,
            "current-run",
            "report.generate-complete-grounded",
            "complete_grounded",
            False,
            1,
            "c" * 64,
        )
    except module.LiveRunError as exc:
        diagnostic = str(exc)
    else:
        raise AssertionError("duplicate action types must fail closed")

    assert calls == {"completion": 1, "judge": 0}
    assert "reason_code=duplicate_action_type" in diagnostic
    assert "preparedness=not_ready" in diagnostic
    assert "action_types=retry_current_round,retry_current_round" in diagnostic
    assert "issue_count=2 needs_work_count=2" in diagnostic
    assert "focus_count=2 mode=focused" in diagnostic
    assert "generation_tokens=120 judge_tokens=0" in diagnostic
    assert f"output_digest={output_digest}" in diagnostic
    assert "PRIVATE QUEUE ISSUE" not in diagnostic
    assert "PRIVATE ROLLBACK ISSUE" not in diagnostic
    assert "PRIVATE QUEUE ACTION" not in diagnostic
    assert "PRIVATE ROLLBACK ACTION" not in diagnostic


def test_p0_100_judge_failure_keeps_redacted_structural_coordinates(tmp_path: Path) -> None:
    runner = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-100-real-provider-full-funnel-hybrid"
        / "scripts"
        / "run_live_reliability.py"
    )
    spec = importlib.util.spec_from_file_location("p0_100_early_judge_failure", runner)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    report = {
        "preparednessLevel": "needs_practice",
        "dimensionAssessments": [{"code": "decision_clarity", "status": "needs_work"}],
        "issues": [{"dimensionCode": "decision_clarity", "evidence": "PRIVATE ISSUE"}],
        "nextActions": [
            {"type": "retry_current_round", "label": "PRIVATE ACTION"},
        ],
        "retryFocusDimensionCodes": ["decision_clarity"],
    }
    raw_output = json.dumps(report, separators=(",", ":")).encode()
    output_digest = hashlib.sha256(raw_output).hexdigest()
    completion_audit = {
        "inputTokens": 100,
        "outputTokens": 20,
        "outputSha256": output_digest,
    }

    def fake_run_evalkit(command: list[str], **kwargs: object) -> bytes:
        if kwargs["stage"] == "completion":
            return raw_output + b"\n"
        audit_path = Path(command[command.index("--audit-out") + 1])
        audit_path.write_text(json.dumps({"inputTokens": 80, "outputTokens": 10}), encoding="utf-8")
        audit_path.chmod(0o600)
        raise module.LiveRunError(
            f"evalkit judge failed reason_code=unsupported_item path=$.issues[0] "
            f"output_digest={output_digest}; raw stderr/reason prose was discarded"
        )

    def fake_load_audit(path: Path, stage: str, case_id: str, critical: bool) -> dict[str, object]:
        assert stage == "completion"
        return completion_audit

    module.run_evalkit = fake_run_evalkit
    module.load_audit = fake_load_audit
    try:
        module.run_attempt(
            REPO_ROOT,
            tmp_path / "evalkit",
            tmp_path,
            "current-run",
            "report.generate-complete-grounded",
            "complete_grounded",
            False,
            1,
            "c" * 64,
        )
    except module.LiveRunError as exc:
        diagnostic = str(exc)
    else:
        raise AssertionError("judge rejection must fail closed")

    assert "reason_code=unsupported_item path=$.issues[0]" in diagnostic
    assert "preparedness=needs_practice" in diagnostic
    assert "action_types=retry_current_round" in diagnostic
    assert "focus_count=1 mode=focused" in diagnostic
    assert "generation_tokens=120 judge_tokens=90" in diagnostic
    assert f"output_digest={output_digest}" in diagnostic
    assert "PRIVATE ISSUE" not in diagnostic
    assert "PRIVATE ACTION" not in diagnostic


def test_scenario_run_skill_requires_env_preflight_and_hybrid_results() -> None:
    skill = (REPO_ROOT / ".agent-skills" / "scenario-run" / "SKILL.md").read_text(
        encoding="utf-8"
    )

    assert "test/scenarios/env-setup.sh" in skill
    assert "test/scenarios/env-verify.sh" in skill
    assert "MANUAL_REQUIRED" in skill
    assert "hybrid" in skill


def test_p0_058_verifier_counts_only_root_go_tests() -> None:
    verify = (
        SCENARIO_ROOT
        / "e2e"
        / "p0-058-report-failure-and-missing-session"
        / "scripts"
        / "verify.sh"
    ).read_text(encoding="utf-8")

    assert "grep -Fxc -- '=== RUN   TestE2EP0058ReportFailureBackendEvidence'" in verify
    assert "grep -Ec -- '^--- PASS: TestE2EP0058ReportFailureBackendEvidence \\('" in verify
    assert "grep -Fc -- '=== RUN   TestE2EP0058ReportFailureBackendEvidence'" not in verify
