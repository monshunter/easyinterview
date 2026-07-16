"""Minimal contracts for the real-environment E2E framework."""

from __future__ import annotations

import os
import re
import subprocess
import sys
import threading
from pathlib import Path
from urllib.parse import quote

import yaml


ROOT = Path(__file__).resolve().parents[2]
SCENARIO_ROOT = ROOT / "test" / "scenarios"
E2E_ROOT = SCENARIO_ROOT / "e2e"

REQUIRED_SCENARIO_FILES = (
    "README.md",
    "data/seed-input.md",
    "data/expected-outcome.md",
    "scripts/setup.sh",
    "scripts/trigger.sh",
    "scripts/verify.sh",
    "scripts/cleanup.sh",
)
FORBIDDEN_CODE_RUNNERS = re.compile(
    r"\bgo\s+test\b|\bvitest\b|\b(?:npm|pnpm)\s+(?:run\s+)?test\b|"
    r"\bpytest\b|\bunittest\b|source[_ -]contract|script[_ -]contract|"
    r"frontend-real-backend-(?:gate|verify)|\b(?:vite|go)\s+build\b",
    re.IGNORECASE,
)
PLAYWRIGHT_INTERCEPTION = re.compile(
    r"\bpage\.route\s*\(|\broute\.(?:fulfill|abort)\s*\(",
)


def scenario_directories() -> list[Path]:
    return sorted(path for path in E2E_ROOT.glob("p*-*") if path.is_dir())


def test_e2e_index_matches_real_scenario_directories() -> None:
    index = (E2E_ROOT / "INDEX.md").read_text(encoding="utf-8")
    listed = set(re.findall(r"`(p\d+-\d{3}-[^`]+/)`", index))
    actual = {f"{path.name}/" for path in scenario_directories()}

    assert listed == actual


def test_e2e_scenarios_have_one_real_environment_contract() -> None:
    for scenario in scenario_directories():
        for relative in REQUIRED_SCENARIO_FILES:
            assert (scenario / relative).is_file(), f"{scenario.name}: missing {relative}"

        scripts = "\n".join(
            (scenario / relative).read_text(encoding="utf-8")
            for relative in ("scripts/trigger.sh", "scripts/verify.sh")
        )
        assert FORBIDDEN_CODE_RUNNERS.search(scripts) is None, scenario.name

        contract = (scenario / "README.md").read_text(encoding="utf-8").lower()
        assert "real" in contract or "真实" in contract, scenario.name


def test_e2e_playwright_specs_do_not_intercept_application_requests() -> None:
    referenced_specs: set[Path] = set()
    for trigger in E2E_ROOT.glob("p*-*/scripts/trigger.sh"):
        text = trigger.read_text(encoding="utf-8")
        for name in re.findall(r"([a-z0-9-]+\.spec\.ts)", text):
            referenced_specs.add(ROOT / "frontend" / "tests" / "e2e" / name)

    for spec in referenced_specs:
        assert spec.is_file(), f"missing Playwright spec: {spec}"
        assert PLAYWRIGHT_INTERCEPTION.search(spec.read_text(encoding="utf-8")) is None, spec


def test_p0101_failure_output_redacts_synthetic_email_before_logging(
    tmp_path: Path,
) -> None:
    scenario = E2E_ROOT / "p0-101-auth-email-code-profile-setup"
    redactor = scenario / "scripts" / "redact_stream.py"
    trigger = (scenario / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    email = "auth-email-code-red@example.test"
    encoded_email = quote(email, safe="")
    failure_output = (
        f'Error: expect(received).toBe(expected)\nExpected: "{email}"\n'
        f'Received: "wrong@example.test"\nattachment={encoded_email}\n'
    )
    log = tmp_path / "trigger.log"
    pipeline = (
        "set -o pipefail; "
        "{ printf '%s' \"$1\"; exit 23; } | "
        "\"$2\" \"$3\" \"$4\" | tee \"$5\""
    )

    assert redactor.is_file()
    assert 'redact_stream.py" "$AUTH_EMAIL" | tee "$LOG"' in trigger
    result = subprocess.run(
        [
            "bash",
            "-c",
            pipeline,
            "p0101-redaction-probe",
            failure_output,
            sys.executable,
            str(redactor),
            email,
            str(log),
        ],
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode == 23
    for output in (result.stdout, log.read_text(encoding="utf-8")):
        assert email not in output
        assert encoded_email not in output
        assert output.count("<redacted-synthetic-email>") == 2


def test_scenario_shell_scripts_are_syntax_valid() -> None:
    scripts = sorted(SCENARIO_ROOT.glob("**/*.sh"))
    assert scripts
    for script in scripts:
        subprocess.run(["bash", "-n", str(script)], check=True)


def test_root_make_test_owns_full_code_regression() -> None:
    makefile = (ROOT / "Makefile").read_text(encoding="utf-8")
    test_target = makefile.split("\ntest:", 1)[1].split("\nbuild:", 1)[0]

    assert "go test ./..." in test_target
    assert "pnpm --filter @easyinterview/frontend test" in test_target
    assert "test/scenarios/e2e" not in test_target


def test_top_level_scenario_environment_entrypoints_remain_independent() -> None:
    expected = {
        "env-setup.sh": "scenario-env-setup",
        "env-status.sh": "scenario-env-status",
        "env-verify.sh": "scenario-env-verify",
        "env-cleanup.sh": "scenario-env-cleanup",
        "env-redeploy.sh": "scenario-env-redeploy",
    }
    makefile = (ROOT / "Makefile").read_text(encoding="utf-8")

    for script, target in expected.items():
        assert (SCENARIO_ROOT / script).is_file()
        assert f"{target}:" in makefile


def test_optional_full_container_runtime_contract() -> None:
    compose_path = ROOT / "deploy" / "dev-stack" / "docker-compose.yaml"
    compose = yaml.safe_load(compose_path.read_text(encoding="utf-8"))

    services = compose["services"]
    for service_name in ("migrate-dev", "backend-dev", "frontend-dev"):
        assert services[service_name]["profiles"] == ["full-container"]
    assert services["backend-dev"]["depends_on"]["migrate-dev"]["condition"] == (
        "service_completed_successfully"
    )
    assert services["backend-dev"]["ports"] == [
        "127.0.0.1:${FULL_CONTAINER_API_HOST_PORT:-10901}:8080"
    ]
    assert services["frontend-dev"]["ports"] == [
        "127.0.0.1:${FULL_CONTAINER_FRONTEND_HOST_PORT:-10900}:8080"
    ]
    for service_name in ("backend-dev", "frontend-dev"):
        assert services[service_name]["labels"]["easyinterview.dev-stack.role"] == "app"
        assert "healthcheck" in services[service_name]
    for relative in (
        "deploy/dev-stack/Dockerfile.backend",
        "deploy/dev-stack/Dockerfile.frontend",
        "deploy/dev-stack/frontend-nginx.conf",
    ):
        assert (ROOT / relative).is_file(), relative

    nginx = (ROOT / "deploy" / "dev-stack" / "frontend-nginx.conf").read_text(
        encoding="utf-8"
    )
    assert "try_files $uri $uri/ /index.html" in nginx
    assert "proxy_pass http://backend-dev:8080" in nginx


def test_full_container_email_provider_environment_contract() -> None:
    compose = yaml.safe_load(
        (ROOT / "deploy" / "dev-stack" / "docker-compose.yaml").read_text(
            encoding="utf-8"
        )
    )
    environment = compose["x-backend-environment"]
    assert environment["EMAIL_SMTP_HOST"] == "${EMAIL_SMTP_HOST:-mailpit-dev}"
    assert environment["EMAIL_SMTP_PORT"] == "${EMAIL_SMTP_PORT:-1025}"
    assert environment["EMAIL_SMTP_USERNAME"] == "${EMAIL_SMTP_USERNAME:-}"
    assert environment["EMAIL_SMTP_PASSWORD"] == "${EMAIL_SMTP_PASSWORD:-}"
    assert environment["EMAIL_SMTP_TLS_MODE"] == "${EMAIL_SMTP_TLS_MODE:-none}"
    assert "EMAIL_PROVIDER_API_KEY" not in environment

    stack_makefile = (ROOT / "deploy" / "dev-stack" / "Makefile").read_text(
        encoding="utf-8"
    )
    container_up = stack_makefile.split("\ncontainer-up:", 1)[1].split(
        "\ncontainer-down:", 1
    )[0]
    assert 'provider="$${EMAIL_PROVIDER:-$$(awk' in container_up
    assert 'if [ "$$provider" = "mailpit" ]' in container_up
    assert "EMAIL_SMTP_HOST=mailpit-dev EMAIL_SMTP_PORT=1025" in container_up
    assert '$(MAKE) -s -C "$(DEV_STACK_DIR)" _stop_host_runtimes' in container_up
    assert "_stop_host_runtimes:" in stack_makefile
    assert ".test-output/local-dev/backend.pid" in stack_makefile
    assert ".test-output/local-dev/frontend.pid" in stack_makefile


def test_stop_host_runtimes_does_not_kill_reused_unowned_pid(tmp_path: Path) -> None:
    process = subprocess.Popen(["sleep", "30"], start_new_session=True)
    backend_pid = tmp_path / "backend.pid"
    frontend_pid = tmp_path / "frontend.pid"
    backend_pid.write_text(str(process.pid), encoding="utf-8")
    try:
        result = subprocess.run(
            [
                "make",
                "-s",
                "-C",
                str(ROOT / "deploy" / "dev-stack"),
                "_stop_host_runtimes",
                f"HOST_BACKEND_PID={backend_pid}",
                f"HOST_FRONTEND_PID={frontend_pid}",
            ],
            text=True,
            capture_output=True,
            check=False,
        )

        assert result.returncode == 0, result.stderr
        assert process.poll() is None, "stale pidfile killed an unrelated process"
        assert not backend_pid.exists()
    finally:
        if process.poll() is None:
            os.killpg(process.pid, 15)
            process.wait(timeout=5)


def test_host_runtime_pidfile_requires_repo_owned_cwd(tmp_path: Path) -> None:
    process = subprocess.Popen(
        [
            sys.executable,
            "-c",
            "import time; time.sleep(30)",
            "go run ./backend/cmd/api -config-dir config",
        ],
        cwd=tmp_path,
        start_new_session=True,
    )
    pid_file = tmp_path / "backend.pid"
    pid_file.write_text(str(process.pid), encoding="utf-8")
    helper = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    probe = r'''
set -euo pipefail
REPO_ROOT="$1"
. "$2"
pid_command_matches_role() { return 0; }
stop_pidfile_process_group "$3" backend-dev
'''
    try:
        result = subprocess.run(
            [
                "bash",
                "-c",
                probe,
                "repo-cwd-probe",
                str(ROOT),
                str(helper),
                str(pid_file),
            ],
            text=True,
            capture_output=True,
            check=False,
        )

        assert result.returncode == 0, result.stderr
        assert process.poll() is None, "same-name process outside the repo was stopped"
        assert not pid_file.exists()
    finally:
        if process.poll() is None:
            os.killpg(process.pid, 15)
            process.wait(timeout=5)


def test_stop_host_runtimes_stops_only_inspectable_owned_backend_pid(
    tmp_path: Path,
) -> None:
    process = subprocess.Popen(
        [
            sys.executable,
            "-c",
            "import time; time.sleep(30)",
            "go run ./backend/cmd/api -config-dir config",
        ],
        start_new_session=True,
    )
    waiter = threading.Thread(target=process.wait, daemon=True)
    waiter.start()
    backend_pid = tmp_path / "backend.pid"
    frontend_pid = tmp_path / "frontend.pid"
    backend_pid.write_text(str(process.pid), encoding="utf-8")
    try:
        ownership_probe = subprocess.run(
            ["ps", "-p", str(process.pid), "-o", "command="],
            text=True,
            capture_output=True,
            check=False,
        )
    except PermissionError:
        ownership_is_inspectable = False
    else:
        command_is_inspectable = (
            ownership_probe.returncode == 0
            and "go run ./backend/cmd/api" in ownership_probe.stdout
        )
        try:
            inspected_cwd = Path(os.readlink(f"/proc/{process.pid}/cwd"))
        except OSError:
            try:
                cwd_probe = subprocess.run(
                    ["lsof", "-a", "-p", str(process.pid), "-d", "cwd", "-Fn"],
                    text=True,
                    capture_output=True,
                    check=False,
                )
            except (FileNotFoundError, PermissionError):
                inspected_cwd = None
            else:
                cwd_lines = [
                    line[1:] for line in cwd_probe.stdout.splitlines() if line.startswith("n")
                ]
                inspected_cwd = Path(cwd_lines[0]) if cwd_lines else None
        ownership_is_inspectable = (
            command_is_inspectable
            and inspected_cwd is not None
            and (inspected_cwd == ROOT or ROOT in inspected_cwd.parents)
        )
    try:
        result = subprocess.run(
            [
                "make",
                "-s",
                "-C",
                str(ROOT / "deploy" / "dev-stack"),
                "_stop_host_runtimes",
                f"HOST_BACKEND_PID={backend_pid}",
                f"HOST_FRONTEND_PID={frontend_pid}",
            ],
            text=True,
            capture_output=True,
            check=False,
        )

        assert result.returncode == 0, result.stderr
        waiter.join(timeout=1)
        if ownership_is_inspectable:
            assert process.returncode is not None, "owned backend process was not stopped"
        else:
            assert process.poll() is None, "unverifiable process must not be stopped"
        assert not backend_pid.exists()
    finally:
        if process.poll() is None:
            os.killpg(process.pid, 15)
            process.wait(timeout=5)


def test_full_container_reuses_existing_redis_for_delivery_secrets() -> None:
    compose = yaml.safe_load(
        (ROOT / "deploy" / "dev-stack" / "docker-compose.yaml").read_text(
            encoding="utf-8"
        )
    )
    services = compose["services"]
    redis_services = [
        name
        for name, service in services.items()
        if str(service.get("image", "")).startswith("redis:")
    ]
    assert redis_services == ["redis-dev"]
    assert compose["x-backend-environment"]["REDIS_URL"] == "redis://redis-dev:6379/0"
    assert services["backend-dev"]["depends_on"]["redis-dev"]["condition"] == (
        "service_healthy"
    )

    runbook = (ROOT / "deploy" / "dev-stack" / "README.md").read_text(
        encoding="utf-8"
    )
    assert "共享加密 delivery secret" in runbook
    assert "投递正确性不依赖停止 host-run backend" in runbook


def test_optional_full_container_lifecycle_contract() -> None:
    root_makefile = (ROOT / "Makefile").read_text(encoding="utf-8")
    stack_makefile = (ROOT / "deploy" / "dev-stack" / "Makefile").read_text(
        encoding="utf-8"
    )
    env_example = (ROOT / "deploy" / "dev-stack" / ".env.example").read_text(
        encoding="utf-8"
    )

    targets = (
        "dev-container-up",
        "dev-container-down",
        "dev-container-doctor",
        "dev-container-logs",
    )
    for target in targets:
        assert f"{target}:" in root_makefile
    for target in ("container-up", "container-down", "container-doctor", "container-logs"):
        assert f"{target}:" in stack_makefile

    default_up = stack_makefile.split("\nup:", 1)[1].split("\ncontainer-up:", 1)[0]
    assert "full-container" not in default_up

    assert "FULL_CONTAINER_FRONTEND_HOST_PORT=10900" in env_example
    assert "FULL_CONTAINER_API_HOST_PORT=10901" in env_example

    for relative in (
        "deploy/dev-stack/README.md",
        "test/scenarios/README.md",
    ):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert "dev-container-up" in text, relative
        assert "10900" in text, relative
        assert "10901" in text, relative

    for relative in (
        ".agent-skills/scenario-env/SKILL.md",
        ".agent-skills/scenario-redeploy/SKILL.md",
    ):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert "dev-container-up" in text, relative
        assert "command output" in text or "current README" in text, relative
        for coupled_port in ("10800", "10801", "10900", "10901"):
            assert coupled_port not in text, relative


def test_host_and_full_container_defaults_share_10900_10901_ports() -> None:
    root_env = (ROOT / ".env.example").read_text(encoding="utf-8")
    stack_env = (
        ROOT / "deploy" / "dev-stack" / ".env.example"
    ).read_text(encoding="utf-8")
    base_config = (ROOT / "config" / "config.yaml").read_text(encoding="utf-8")
    compose = (ROOT / "deploy" / "dev-stack" / "docker-compose.yaml").read_text(
        encoding="utf-8"
    )
    vite = (ROOT / "frontend" / "vite.config.ts").read_text(encoding="utf-8")
    runtime = (
        SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    ).read_text(encoding="utf-8")

    for env_text in (root_env, stack_env):
        assert "APP_LISTEN_ADDR=:10901" in env_text
        assert "EMAIL_VERIFY_BASE_URL=http://127.0.0.1:10900/auth/verify" in env_text
    assert "API_HOST_PORT=10901" in stack_env
    assert "FRONTEND_HOST_PORT=10900" in stack_env
    assert "FULL_CONTAINER_FRONTEND_HOST_PORT=10900" in stack_env
    assert "FULL_CONTAINER_API_HOST_PORT=10901" in stack_env
    assert "VITE_EI_API_BASE_URL=http://127.0.0.1:10901/api/v1" in stack_env
    assert 'listenAddr: ":10901"' in base_config
    assert 'verifyBaseURL: "http://127.0.0.1:10900/auth/verify"' in base_config
    assert "${FULL_CONTAINER_FRONTEND_HOST_PORT:-10900}" in compose
    assert "${FULL_CONTAINER_API_HOST_PORT:-10901}" in compose
    assert 'envPort("FRONTEND_HOST_PORT", 10900)' in vite
    assert 'APP_LISTEN_ADDR:-:10901' in runtime
    assert 'FRONTEND_HOST_PORT:-10900' in runtime


def test_local_ai_raw_capture_config_and_backend_only_bind_contract() -> None:
    base = yaml.safe_load((ROOT / "config" / "config.yaml").read_text(encoding="utf-8"))
    dev = yaml.safe_load((ROOT / "config" / "dev.yaml").read_text(encoding="utf-8"))
    test = yaml.safe_load((ROOT / "config" / "test.yaml").read_text(encoding="utf-8"))
    staging = yaml.safe_load(
        (ROOT / "config" / "staging.yaml").read_text(encoding="utf-8")
    )
    prod = yaml.safe_load((ROOT / "config" / "prod.yaml").read_text(encoding="utf-8"))

    assert base["ai"]["debugCaptureRawIO"] is False
    assert base["ai"]["debugRawIOPath"] == (
        ".test-output/local-dev/ai-raw.ndjson"
    )
    for local in (dev, test):
        assert local["ai"]["debugCaptureRawIO"] is True
        assert local["ai"]["debugRawIOPath"] == (
            ".test-output/local-dev/ai-raw.ndjson"
        )
    for deployment in (staging, prod):
        assert deployment.get("ai", {}).get("debugCaptureRawIO", False) is False

    env_example = (ROOT / "deploy" / "dev-stack" / ".env.example").read_text(
        encoding="utf-8"
    )
    assert "AI_DEBUG_CAPTURE_RAW_IO=true" in env_example
    assert "AI_DEBUG_RAW_IO_PATH=.test-output/local-dev/ai-raw.ndjson" in env_example

    compose = yaml.safe_load(
        (ROOT / "deploy" / "dev-stack" / "docker-compose.yaml").read_text(
            encoding="utf-8"
        )
    )
    environment = compose["x-backend-environment"]
    assert environment["AI_DEBUG_CAPTURE_RAW_IO"] == (
        "${AI_DEBUG_CAPTURE_RAW_IO:-true}"
    )
    assert str(environment["AI_DEBUG_RAW_IO_PATH"]).endswith(
        "/app/.test-output/local-dev/ai-raw.ndjson}"
    ) or environment["AI_DEBUG_RAW_IO_PATH"] == (
        "/app/.test-output/local-dev/ai-raw.ndjson"
    )

    raw_target = "/app/.test-output/local-dev"

    def volume_target(volume: object) -> str:
        if isinstance(volume, dict):
            return str(volume.get("target", ""))
        parts = str(volume).split(":")
        return parts[1] if len(parts) >= 2 else ""

    backend_volumes = compose["services"]["backend-dev"].get("volumes", [])
    assert any(
        volume_target(volume) == raw_target and str(volume).endswith(":rw")
        for volume in backend_volumes
    ), backend_volumes
    for service_name, service in compose["services"].items():
        if service_name == "backend-dev":
            continue
        assert all(
            volume_target(volume) != raw_target
            for volume in service.get("volumes", [])
        ), service_name


def test_raw_capture_preflight_rejects_filesystem_root_parent() -> None:
    helper = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    probe = r'''
set -euo pipefail
REPO_ROOT="$1"
. "$2"
AI_DEBUG_CAPTURE_RAW_IO=true
AI_DEBUG_RAW_IO_PATH=/ai-raw-contract-probe.ndjson
export AI_DEBUG_CAPTURE_RAW_IO AI_DEBUG_RAW_IO_PATH
secure_raw_capture_path
'''
    result = subprocess.run(
        ["bash", "-c", probe, "raw-root-probe", str(ROOT), str(helper)],
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode != 0
    assert "filesystem or volume root" in result.stderr


def test_dev_stack_pidfile_stop_requires_command_and_repo_cwd() -> None:
    makefile = (ROOT / "deploy" / "dev-stack" / "Makefile").read_text(
        encoding="utf-8"
    )
    stop_target = makefile.split("\n_stop_host_runtimes:", 1)[1].split(
        "\n_all_healthy:", 1
    )[0]

    assert 'ps -p "$$pid" -o command=' in stop_target
    assert 'lsof -a -p "$$pid" -d cwd' in stop_target
    assert 'repo_owned" != "yes"' in stop_target


def test_local_ai_raw_capture_old_stderr_key_has_zero_current_runtime_references() -> None:
    current_runtime_files = (
        ROOT / "config" / "config.yaml",
        ROOT / "config" / "dev.yaml",
        ROOT / "config" / "test.yaml",
        ROOT / "config" / "staging.yaml",
        ROOT / "config" / "prod.yaml",
        ROOT / "deploy" / "dev-stack" / ".env.example",
        ROOT / "deploy" / "dev-stack" / "docker-compose.yaml",
        ROOT / "backend" / "internal" / "platform" / "config" / "bindings.go",
        ROOT / "backend" / "cmd" / "api" / "main.go",
    )
    for path in current_runtime_files:
        text = path.read_text(encoding="utf-8")
        assert "AI_DEBUG_PRINT_RAW_OUTPUT" not in text, path
        assert "ai.debugPrintRawOutput" not in text, path
        assert "AI_RAW_OUTPUT_DEBUG_BEGIN" not in text, path
        assert "AI_RAW_OUTPUT_DEBUG_END" not in text, path


def test_p0099_requires_raw_capture_but_keeps_raw_file_outside_evidence() -> None:
    scenario = E2E_ROOT / "p0-099-report-generating-live-ui"
    setup = (scenario / "scripts" / "setup.sh").read_text(encoding="utf-8")
    trigger = (scenario / "scripts" / "trigger.sh").read_text(encoding="utf-8")
    verify = (scenario / "scripts" / "verify.sh").read_text(encoding="utf-8")

    assert "AI_DEBUG_CAPTURE_RAW_IO" in setup
    assert "AI_DEBUG_RAW_IO_PATH" in setup
    assert re.search(r"AI_DEBUG_CAPTURE_RAW_IO[^\n]*(?:true|=1)", setup)
    assert "resolve" in setup or "realpath" in setup
    assert "symlink" in setup.lower()
    assert "regular" in setup.lower()
    assert "AI_DEBUG_PRINT_RAW_OUTPUT" not in setup
    assert "raw debug disabled" not in setup.lower()

    # Trigger and verifier may inspect bounded status/digests, but must never
    # read, copy, summarize, or attach the dedicated raw Complete file.
    for script in (trigger, verify):
        assert "AI_DEBUG_RAW_IO_PATH" not in script
        assert re.search(r"\b(?:cat|cp|mv|tee|sed|awk|grep)\b[^\n]*ai-raw", script) is None


def test_host_and_full_container_app_runners_are_mutually_exclusive() -> None:
    helper = (
        SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    ).read_text(encoding="utf-8")
    redeploy = (SCENARIO_ROOT / "env-redeploy.sh").read_text(encoding="utf-8")
    verify = (SCENARIO_ROOT / "env-verify.sh").read_text(encoding="utf-8")
    combined_redeploy = helper + "\n" + redeploy
    combined_verify = helper + "\n" + verify

    assert re.search(
        r"(?:docker\s+compose|\$\{?COMPOSE\}?)[^\n]*(?:stop|rm)[^\n]*backend-dev",
        combined_redeploy,
    )
    assert re.search(
        r"(?:docker\s+compose|\$\{?COMPOSE\}?)[^\n]*(?:stop|rm)[^\n]*frontend-dev",
        combined_redeploy,
    )
    for role in ("backend-dev", "frontend-dev"):
        assert role in combined_verify
    assert re.search(
        r"(?:conflict|multiple|coexist|double|more than one|并存|冲突)",
        combined_verify,
        re.IGNORECASE,
    )


def test_all_redeploy_removes_full_container_roles_before_dependency_doctor() -> None:
    result = subprocess.run(
        [str(SCENARIO_ROOT / "env-redeploy.sh"), "all", "--dry-run"],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode == 0, result.stderr
    output = result.stdout
    backend_remove = output.index("docker compose rm -sf backend-dev")
    frontend_remove = output.index("docker compose rm -sf frontend-dev")
    dependency_up = output.index("make dev-up")
    dependency_doctor = output.index("make dev-doctor")
    assert backend_remove < dependency_up
    assert frontend_remove < dependency_up
    assert dependency_up < dependency_doctor


def test_single_runner_guard_fails_closed_without_stopping_processes() -> None:
    helper = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    probe = r'''
set -euo pipefail
REPO_ROOT="$1"
. "$2"
load_dev_stack_env() { :; }
compose_app_role_running() { [ "$1" = "backend-dev" ]; }
host_app_role_running() { [ "$1" = "backend-dev" ]; }
assert_single_app_runners
'''
    result = subprocess.run(
        ["bash", "-c", probe, "single-runner-probe", str(ROOT), str(helper)],
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode != 0
    assert "runner conflict" in result.stderr
    assert "backend-dev" in result.stderr


def test_host_mailpit_smtp_route_mismatch_fails_before_backend_start() -> None:
    helper = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    probe = r'''
set -euo pipefail
REPO_ROOT="$1"
. "$2"
EMAIL_PROVIDER=mailpit
EMAIL_SMTP_HOST=127.0.0.1
EMAIL_SMTP_PORT=1025
MAILPIT_SMTP_HOST_PORT=11025
export EMAIL_PROVIDER EMAIL_SMTP_HOST EMAIL_SMTP_PORT MAILPIT_SMTP_HOST_PORT
assert_host_mailpit_smtp_route
'''
    result = subprocess.run(
        ["bash", "-c", probe, "mailpit-route-probe", str(ROOT), str(helper)],
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode != 0
    assert "EMAIL_SMTP_PORT=1025" in result.stderr
    assert "MAILPIT_SMTP_HOST_PORT=11025" in result.stderr


def test_host_mailpit_smtp_route_accepts_matching_host_mapping() -> None:
    helper = SCENARIO_ROOT / "_shared" / "scripts" / "local-dev-runtime.sh"
    probe = r'''
set -euo pipefail
REPO_ROOT="$1"
. "$2"
EMAIL_PROVIDER=mailpit
EMAIL_SMTP_HOST=127.0.0.1
EMAIL_SMTP_PORT=11025
MAILPIT_SMTP_HOST_PORT=11025
export EMAIL_PROVIDER EMAIL_SMTP_HOST EMAIL_SMTP_PORT MAILPIT_SMTP_HOST_PORT
assert_host_mailpit_smtp_route
'''
    result = subprocess.run(
        ["bash", "-c", probe, "mailpit-route-probe", str(ROOT), str(helper)],
        text=True,
        capture_output=True,
        check=False,
    )

    assert result.returncode == 0, result.stderr
