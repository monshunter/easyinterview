"""Minimal contracts for the real-environment E2E framework."""

from __future__ import annotations

import re
import subprocess
import sys
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
        "127.0.0.1:${FULL_CONTAINER_API_HOST_PORT:-10801}:8080"
    ]
    assert services["frontend-dev"]["ports"] == [
        "127.0.0.1:${FULL_CONTAINER_FRONTEND_HOST_PORT:-10800}:8080"
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

    assert "FULL_CONTAINER_FRONTEND_HOST_PORT=10800" in env_example
    assert "FULL_CONTAINER_API_HOST_PORT=10801" in env_example

    for relative in (
        "deploy/dev-stack/README.md",
        "test/scenarios/README.md",
        ".agent-skills/scenario-env/SKILL.md",
        ".agent-skills/scenario-redeploy/SKILL.md",
    ):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert "dev-container-up" in text, relative
        assert "10800" in text, relative
        assert "10801" in text, relative
