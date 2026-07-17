from __future__ import annotations

from pathlib import Path

from scripts.harness_index import (
    REQUIRED_QUALITY_BOUNDARIES,
    StaleIndexError,
    audit_legacy_structure,
    build_index,
    collect_legacy_baseline,
    load_index_cache,
    route_query,
    write_index_cache,
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


def test_build_index_uses_only_canonical_documents() -> None:
    index = build_index(REPO_ROOT)

    assert index["schema"] == "easyinterview.harness-index/v1"
    assert index["git_commit"]
    documents = {document["path"]: document for document in index["documents"]}
    owner = documents["docs/spec/harness-simplification/spec.md"]
    assert owner["kind"] == "spec"
    assert owner["subject"] == "harness-simplification"
    assert owner["title"] == "轻量级 Harness 上下文与技能体系"
    assert all("/plans/" not in path for path in documents)
    assert all(not path.endswith("context.yaml") for path in documents)


def test_index_cache_is_rebuildable_and_commit_bound(tmp_path: Path) -> None:
    index = build_index(REPO_ROOT)
    cache_path = tmp_path / "harness-index.json"

    write_index_cache(index, cache_path)
    assert load_index_cache(cache_path, expected_commit=index["git_commit"]) == index

    try:
        load_index_cache(cache_path, expected_commit="0" * 40)
    except StaleIndexError as error:
        assert index["git_commit"] in str(error)
    else:
        raise AssertionError("stale cache must fail visibly")


def test_route_query_prefers_exact_subject_with_auditable_reason() -> None:
    result = route_query(build_index(REPO_ROOT), "harness-simplification")

    assert result["confidence"] == "high"
    assert result["candidates"][0]["path"] == "docs/spec/harness-simplification/spec.md"
    assert "exact subject" in result["candidates"][0]["reasons"]


def test_route_query_cannot_turn_generic_words_into_high_confidence() -> None:
    result = route_query(build_index(REPO_ROOT), "spec plan context docs")

    assert result["confidence"] == "low"
    assert len(result["candidates"]) <= 3
    assert result["requires_user_choice"] is True


def test_route_query_uses_exact_identifiers_before_semantic_tokens() -> None:
    index = {
        "schema": "easyinterview.harness-index/v1",
        "git_commit": "a" * 40,
        "documents": [
            {
                "path": "docs/spec/alpha/spec.md",
                "subject": "alpha",
                "kind": "spec",
                "title": "Shared Runtime",
                "identifiers": ["report.generate"],
            },
            {
                "path": "docs/spec/report/spec.md",
                "subject": "report",
                "kind": "spec",
                "title": "Report UI",
                "identifiers": [],
            },
        ],
    }

    result = route_query(index, "where is report.generate owned")

    assert result["confidence"] == "high"
    assert result["candidates"][0]["subject"] == "alpha"
    assert "exact identifier: report.generate" in result["candidates"][0]["reasons"]
