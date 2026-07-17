from __future__ import annotations

from pathlib import Path

from scripts.harness_index import (
    REQUIRED_QUALITY_BOUNDARIES,
    audit_legacy_structure,
    collect_legacy_baseline,
)


REPO_ROOT = Path(__file__).resolve().parents[1]


def test_collect_legacy_baseline_captures_migration_cost() -> None:
    baseline = collect_legacy_baseline(REPO_ROOT)

    assert baseline["subjects"] >= 25
    assert baseline["context_files"] >= 49
    assert baseline["plan_files"] >= 49
    assert baseline["checklist_files"] >= 49
    assert baseline["bdd_files"] >= 40
    assert baseline["layer_indexes"] >= 29
    assert baseline["spec_bytes"] > 100_000
    assert baseline["git_commit"]


def test_collect_legacy_baseline_is_repository_relative() -> None:
    baseline = collect_legacy_baseline(REPO_ROOT)

    assert Path(baseline["repo_root"]) == REPO_ROOT
    assert not Path(baseline["repo_root"]).is_relative_to(REPO_ROOT / "docs")


def test_audit_legacy_structure_reports_every_removed_wrapper_type() -> None:
    violations = audit_legacy_structure(REPO_ROOT)

    kinds = {violation["kind"] for violation in violations}
    assert {
        "context-manifest",
        "plan-wrapper",
        "checklist-wrapper",
        "bdd-wrapper",
        "history-wrapper",
        "layer-index",
    } <= kinds
    assert all(violation["path"].startswith("docs/spec/") for violation in violations)


def test_migration_contract_preserves_non_negotiable_quality_boundaries() -> None:
    assert REQUIRED_QUALITY_BOUNDARIES == {
        "executable-tests",
        "real-e2e",
        "contract-owner",
        "backend-persistence",
        "current-evidence",
        "high-risk-confirmation",
        "security-privacy",
        "failure-recovery",
    }
