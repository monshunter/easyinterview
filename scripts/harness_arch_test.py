"""Owner-level contract tests for the Project Arch v1 bootstrap tool.

The tests intentionally exercise ``harness_arch.py`` through its public CLI.
Every repository under test lives below pytest's ``tmp_path``; the real checkout
is only used to locate the production script.

The minimum CLI contract frozen by these tests is::

    python3 scripts/harness_arch.py MODE --repo-root PATH [--goal TEXT] [--apply]

Commands emit one JSON document on stdout.  ``status`` is one of ``ready``,
``spec_required``, ``decision_required``, or ``conflict``.  The environment
variable ``HARNESS_ARCH_TEST_FAIL_AFTER_WRITES`` is a test-only fault injection
hook used to prove transaction rollback and resumability.
"""

from __future__ import annotations

import json
import os
from pathlib import Path
import stat
import subprocess
import sys
from typing import Any


SCRIPT = Path(__file__).with_name("harness_arch.py")
ARCH_MARKER_V0 = "<!-- project-arch: v0 -->"
ARCH_MARKER_V1 = "<!-- project-arch: v1 -->"

ENV_ENTRYPOINTS = (
    "env-setup.sh",
    "env-status.sh",
    "env-verify.sh",
    "env-redeploy.sh",
    "env-cleanup.sh",
)

OPTIONAL_DOC_DIRECTORIES = (
    "apis",
    "bugs",
    "reports",
    "discuss",
    "work-journal",
    "ui-design",
)

FORBIDDEN_GENERATED_FILES = {
    "context.yaml",
    "checklist.md",
    "bdd-plan.md",
    "bdd-checklist.md",
    "test-plan.md",
    "test-checklist.md",
    "history.md",
    "TEMPLATES.md",
}


def _write(path: Path, body: str | bytes, *, executable: bool = False) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if isinstance(body, bytes):
        path.write_bytes(body)
    else:
        path.write_text(body, encoding="utf-8")
    if executable:
        path.chmod(path.stat().st_mode | stat.S_IXUSR)


def _seed_ready_python_repo(root: Path) -> Path:
    root.mkdir(parents=True, exist_ok=True)
    _write(
        root / "README.md",
        "# Sample CLI\n\nA standard-library Python command-line project.\n",
    )
    _write(
        root / "pyproject.toml",
        """[project]
name = "sample-cli"
version = "0.1.0"
requires-python = ">=3.11"

[tool.pytest.ini_options]
testpaths = ["tests"]
""",
    )
    _write(
        root / "src/sample_cli.py",
        "def main() -> int:\n    return 0\n",
    )
    _write(
        root / "tests/test_sample_cli.py",
        "from sample_cli import main\n\n\ndef test_main():\n    assert main() == 0\n",
    )
    return root


def _seed_legacy_v0_repo(root: Path) -> Path:
    _seed_ready_python_repo(root)
    _write(
        root / "docs/README.md",
        "# Project documentation\n\n"
        f"{ARCH_MARKER_V0}\n\n"
        "This human-maintained legacy owner note must survive upgrades.\n",
    )
    return root


def _run(
    root: Path,
    mode: str,
    *,
    apply: bool = False,
    goal: str | None = None,
    extra_env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    command = [
        sys.executable,
        str(SCRIPT),
        mode,
        "--repo-root",
        str(root),
    ]
    if goal is not None:
        command.extend(("--goal", goal))
    if apply:
        command.append("--apply")

    environment = os.environ.copy()
    if extra_env:
        environment.update(extra_env)

    return subprocess.run(
        command,
        cwd=root,
        env=environment,
        capture_output=True,
        text=True,
        check=False,
    )


def _diagnostic(result: subprocess.CompletedProcess[str]) -> str:
    return (
        f"returncode={result.returncode}\n"
        f"stdout={result.stdout!r}\n"
        f"stderr={result.stderr!r}"
    )


def _payload(result: subprocess.CompletedProcess[str]) -> dict[str, Any]:
    assert result.stdout.strip(), _diagnostic(result)
    try:
        value = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise AssertionError(
            "harness_arch.py must emit exactly one JSON document on stdout\n"
            + _diagnostic(result)
        ) from exc
    assert isinstance(value, dict), value
    return value


def _assert_status(
    result: subprocess.CompletedProcess[str],
    expected: str,
    *,
    successful: bool,
) -> dict[str, Any]:
    if successful:
        assert result.returncode == 0, _diagnostic(result)
    else:
        assert result.returncode != 0, _diagnostic(result)
    payload = _payload(result)
    assert payload.get("status") == expected, payload
    return payload


def _snapshot(root: Path) -> dict[str, tuple[Any, ...]]:
    snapshot: dict[str, tuple[Any, ...]] = {}
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root).as_posix()
        metadata = path.lstat()
        mode = stat.S_IMODE(metadata.st_mode)
        if path.is_symlink():
            snapshot[relative] = (
                "symlink",
                os.readlink(path),
                mode,
                metadata.st_mtime_ns,
            )
        elif path.is_dir():
            snapshot[relative] = ("directory", mode, metadata.st_mtime_ns)
        else:
            snapshot[relative] = (
                "file",
                path.read_bytes(),
                mode,
                metadata.st_mtime_ns,
            )
    return snapshot


def _inventory_paths(payload: dict[str, Any], classification: str) -> set[str]:
    inventory = payload.get("inventory")
    assert isinstance(inventory, dict), payload
    entries = inventory.get(classification)
    assert isinstance(entries, list), inventory

    paths: set[str] = set()
    for entry in entries:
        if isinstance(entry, str):
            paths.add(entry)
        elif isinstance(entry, dict) and isinstance(entry.get("path"), str):
            paths.add(entry["path"])
        else:
            raise AssertionError(
                f"inventory.{classification} entry has no path: {entry!r}"
            )
    return paths


def _assert_minimal_project_arch(root: Path) -> None:
    required_files = (
        "AGENTS.md",
        "docs/README.md",
        "docs/agent-workflow.md",
        "docs/development.md",
        "docs/spec/INDEX.md",
        "test/README.md",
        "test/scenarios/README.md",
        "test/scenarios/_shared/README.md",
        "scripts/harness_arch.py",
    )
    for relative in required_files:
        path = root / relative
        assert path.is_file(), f"missing Project Arch v1 interface: {relative}"
        assert path.stat().st_size > 0, f"empty Project Arch v1 interface: {relative}"

    docs_readme = (root / "docs/README.md").read_text(encoding="utf-8")
    assert docs_readme.count(ARCH_MARKER_V1) == 1

    subject_specs = list((root / "docs/spec").glob("*/spec.md"))
    assert len(subject_specs) == 1, subject_specs
    assert subject_specs[0].stat().st_size > 0
    plans_indexes = list((root / "docs/spec").glob("*/plans/INDEX.md"))
    assert len(plans_indexes) == 1, plans_indexes

    scenarios = root / "test/scenarios"
    for name in ENV_ENTRYPOINTS:
        entrypoint = scenarios / name
        assert entrypoint.is_file(), f"missing arch.env interface: {name}"
        assert entrypoint.stat().st_mode & stat.S_IXUSR, name

    assert not (scenarios / "e2e").exists(), "do not pre-create an empty E2E tree"
    for name in OPTIONAL_DOC_DIRECTORIES:
        assert not (root / "docs" / name).exists(), f"unexpected empty docs/{name}/"

    generated_names = {path.name for path in root.rglob("*") if path.is_file()}
    assert generated_names.isdisjoint(FORBIDDEN_GENERATED_FILES), generated_names & FORBIDDEN_GENERATED_FILES


def test_fresh_init_installs_minimal_four_layer_arch_and_runtime_interfaces(
    tmp_path: Path,
) -> None:
    repo = _seed_ready_python_repo(tmp_path / "fresh")

    result = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(result, "ready", successful=True)
    _assert_minimal_project_arch(repo)


def test_check_is_strictly_read_only(tmp_path: Path) -> None:
    repo = _seed_ready_python_repo(tmp_path / "check")
    initialized = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    _assert_status(initialized, "ready", successful=True)
    before = _snapshot(repo)

    checked = _run(repo, "check")

    _assert_status(checked, "ready", successful=True)
    assert _snapshot(repo) == before


def test_upgrade_moves_legacy_v0_to_v1_without_losing_owner_text(
    tmp_path: Path,
) -> None:
    repo = _seed_legacy_v0_repo(tmp_path / "legacy")
    legacy_note = "This human-maintained legacy owner note must survive upgrades."

    upgraded = _run(
        repo,
        "upgrade",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(upgraded, "ready", successful=True)
    docs_readme = (repo / "docs/README.md").read_text(encoding="utf-8")
    assert ARCH_MARKER_V0 not in docs_readme
    assert docs_readme.count(ARCH_MARKER_V1) == 1
    assert legacy_note in docs_readme
    _assert_minimal_project_arch(repo)


def test_repair_restores_only_confirmed_arch_owned_drift(tmp_path: Path) -> None:
    repo = _seed_ready_python_repo(tmp_path / "repair")
    initialized = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    _assert_status(initialized, "ready", successful=True)

    owned = repo / "docs/agent-workflow.md"
    original_owned = owned.read_bytes()
    custom = repo / "docs/project-notes.md"
    _write(custom, "human-owned notes\n")
    original_custom = custom.read_bytes()
    _write(owned, "corrupted arch-owned workflow\n")

    drift = _run(repo, "check")
    _assert_status(drift, "conflict", successful=False)

    repaired = _run(repo, "repair", apply=True)

    _assert_status(repaired, "ready", successful=True)
    assert owned.read_bytes() == original_owned
    assert custom.read_bytes() == original_custom


def test_same_version_init_is_a_true_zero_diff_noop(tmp_path: Path) -> None:
    repo = _seed_ready_python_repo(tmp_path / "idempotent")
    first = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    _assert_status(first, "ready", successful=True)
    before = _snapshot(repo)

    second = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(second, "ready", successful=True)
    assert _snapshot(repo) == before


def test_inventory_preserves_custom_files_and_fails_closed_on_conflicts(
    tmp_path: Path,
) -> None:
    custom_repo = _seed_ready_python_repo(tmp_path / "custom")
    custom_path = custom_repo / "docs/project-notes.md"
    _write(custom_path, "human-owned project knowledge\n")
    custom_bytes = custom_path.read_bytes()

    initialized = _run(
        custom_repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    custom_payload = _assert_status(initialized, "ready", successful=True)
    assert "docs/project-notes.md" in _inventory_paths(custom_payload, "custom")
    assert custom_path.read_bytes() == custom_bytes

    conflict_repo = _seed_ready_python_repo(tmp_path / "conflict")
    conflicting_path = conflict_repo / "docs/development.md"
    _write(conflicting_path, "human-owned incompatible contract\n")
    before = _snapshot(conflict_repo)

    conflicted = _run(
        conflict_repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    conflict_payload = _assert_status(conflicted, "conflict", successful=False)
    assert "docs/development.md" in _inventory_paths(
        conflict_payload, "conflicting"
    )
    assert _snapshot(conflict_repo) == before


def test_partial_write_failure_rolls_back_and_can_resume(tmp_path: Path) -> None:
    repo = _seed_ready_python_repo(tmp_path / "rollback")
    before = _snapshot(repo)

    failed = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
        extra_env={"HARNESS_ARCH_TEST_FAIL_AFTER_WRITES": "3"},
    )

    assert failed.returncode != 0, _diagnostic(failed)
    failure_payload = _payload(failed)
    assert failure_payload.get("status") != "ready", failure_payload
    recovery = failure_payload.get("recovery")
    assert isinstance(recovery, dict), failure_payload
    assert recovery.get("rolled_back") is True, recovery
    assert isinstance(recovery.get("resume_condition"), str), recovery
    assert recovery["resume_condition"].strip(), recovery
    assert _snapshot(repo) == before

    resumed = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )
    _assert_status(resumed, "ready", successful=True)
    _assert_minimal_project_arch(repo)


def test_upgrade_preserves_unknown_custom_and_secret_bytes_without_leaking(
    tmp_path: Path,
) -> None:
    repo = _seed_legacy_v0_repo(tmp_path / "preservation")
    protected = {
        repo / "docs/notes/customer-contract.md": b"custom business contract\n",
        repo / "vendor/opaque-state.bin": b"\x00\x01unknown-project-state\xff",
        repo / ".env.local": b"API_TOKEN=fixture-secret-never-print\n",
    }
    for path, body in protected.items():
        _write(path, body)

    upgraded = _run(
        repo,
        "upgrade",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(upgraded, "ready", successful=True)
    for path, expected in protected.items():
        assert path.read_bytes() == expected
    combined_output = upgraded.stdout + upgraded.stderr
    assert "fixture-secret-never-print" not in combined_output


def test_discovery_reports_ready_for_sufficient_repository_facts(
    tmp_path: Path,
) -> None:
    repo = _seed_ready_python_repo(tmp_path / "discovery-ready")

    result = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(result, "ready", successful=True)


def test_discovery_reports_spec_required_for_an_unimplemented_runtime(
    tmp_path: Path,
) -> None:
    repo = tmp_path / "discovery-spec-required"
    repo.mkdir()
    _write(repo / "README.md", "# HTTP service\n")
    _write(repo / "src/service.py", "# runtime not implemented yet\n")

    result = _run(
        repo,
        "init",
        apply=True,
        goal="run a long-lived HTTP service with a health endpoint",
    )

    _assert_status(result, "spec_required", successful=False)


def test_discovery_reports_decision_required_when_project_goal_is_unknown(
    tmp_path: Path,
) -> None:
    repo = tmp_path / "discovery-decision-required"
    repo.mkdir()
    _write(repo / "README.md", "# Unconfigured repository\n")

    result = _run(repo, "init", apply=True)

    _assert_status(result, "decision_required", successful=False)


def test_discovery_reports_conflict_for_an_incompatible_canonical_owner(
    tmp_path: Path,
) -> None:
    repo = _seed_ready_python_repo(tmp_path / "discovery-conflict")
    _write(repo / "docs/development.md", "incompatible human owner\n")
    before = _snapshot(repo)

    result = _run(
        repo,
        "init",
        apply=True,
        goal="maintain a standard-library Python command-line project",
    )

    _assert_status(result, "conflict", successful=False)
    assert _snapshot(repo) == before
