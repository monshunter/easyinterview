"""Tests for path- and owner-evidence based change-intake matching."""

from __future__ import annotations

import importlib.util
import os
from pathlib import Path

import yaml


SCRIPT_PATH = Path(__file__).resolve().parent / "match_change_context.py"
OBSOLETE_EN_STATUS = "de" + "precated"
OBSOLETE_ZH_STATUS = "废" + "弃"


def _load_module():
    spec = importlib.util.spec_from_file_location("match_change_context", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def _base_context() -> dict:
    return {
        "apiVersion": "plancontext.agent.dev/v1alpha1",
        "kind": "PlanContext",
        "metadata": {"name": "001-backend"},
        "spec": {
            "defaultTarget": "backend",
            "targets": {
                "backend": {
                    "plan": "./plan.md",
                    "checklist": "./checklist.md",
                    "spec": "../../spec.md",
                }
            },
        },
    }


def _write_plan(
    root: Path,
    subject: str,
    status: str,
    context: dict | None = None,
    *,
    with_context: bool = True,
    with_checklist: bool = True,
):
    spec_dir = root / subject
    plan_dir = spec_dir / "plans" / "001-backend"
    plan_dir.mkdir(parents=True)
    (plan_dir / "plan.md").write_text(
        "# Plan\n\n"
        "> **版本**: 1.0\n"
        f"> **状态**: {status}\n"
        "> **更新日期**: 2026-04-04\n",
        encoding="utf-8",
    )
    if with_checklist:
        (plan_dir / "checklist.md").write_text("# Checklist\n", encoding="utf-8")
    (spec_dir / "spec.md").write_text(f"# {subject}\n", encoding="utf-8")
    if with_context:
        with open(plan_dir / "context.yaml", "w", encoding="utf-8") as f:
            yaml.safe_dump(context or _base_context(), f, sort_keys=False, allow_unicode=True)
    return plan_dir


def _write_scenario_owner(root: Path, scenario_id: str, owner_plan: Path):
    slug = scenario_id.lower().replace(".", "-")
    scenario_dir = root / "test" / "scenarios" / "e2e" / f"{slug}-owner-regression"
    scenario_dir.mkdir(parents=True)
    owner_href = os.path.relpath(owner_plan, scenario_dir).replace(os.sep, "/")
    (scenario_dir / "README.md").write_text(
        f"# E2E.{scenario_id} owner regression\n\n"
        f"> Owner: [`owner`]({owner_href})\n",
        encoding="utf-8",
    )


def test_matches_subject_and_plan_names_derived_from_path(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    _write_plan(plan_root, "local-auth", "active")
    _write_plan(plan_root, "wiki-sync", "active")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="local auth backend is broken",
    )

    assert result["recommended"]["displayPlan"] == "local-auth/001-backend"
    assert any(reason.startswith("displayName=") for reason in result["recommended"]["reasons"])


def test_legacy_discovery_and_references_do_not_affect_matching(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    poisoned = _base_context()
    poisoned["spec"]["discovery"] = {
        "aliases": ["resume create flow"],
        "keywords": ["resume create flow"],
    }
    poisoned["spec"]["targets"]["backend"]["discovery"] = {
        "apiNames": ["resume create flow"],
    }
    poisoned["spec"]["targets"]["backend"]["references"] = ["../../spec.md"]
    _write_plan(plan_root, "unrelated-owner", "active", poisoned)
    _write_plan(plan_root, "resume-create-flow", "active")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="resume create flow",
    )

    assert result["recommended"]["displayPlan"] == "resume-create-flow/001-backend"
    assert all(
        not reason.startswith(
            ("aliases=", "keywords=", "apiNames=", "packages=", "uiRoutes=", "references=")
        )
        for candidate in result["candidates"]
        for reason in candidate["reasons"]
    )


def test_completed_candidate_marks_in_place_revision(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    _write_plan(plan_root, "sealed-secrets", "completed")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="sealed secrets regression",
    )

    assert result["recommended"]["displayPlan"] == "sealed-secrets/001-backend"
    assert result["recommended"]["status"] == "completed"
    assert result["recommended"]["reviseInPlace"] is True


def test_exact_scenario_readme_owner_outranks_path_terms(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    owner_dir = _write_plan(plan_root, "resume-detail-owner", "completed")
    _write_plan(plan_root, "p0-037-generic", "active")
    _write_scenario_owner(tmp_path, "P0.037", owner_dir / "plan.md")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="P0.037 remains loading",
    )

    assert result["recommended"]["displayPlan"] == "resume-detail-owner/001-backend"
    assert "scenarioOwner=E2E.P0.037" in result["recommended"]["reasons"]


def test_unrelated_query_returns_no_candidate(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    _write_plan(plan_root, "local-auth", "active")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="completely unrelated vocabulary",
    )

    assert result["confidence"] == "none"
    assert result["recommended"] is None


def test_generic_manifest_vocabulary_does_not_create_false_candidate(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    _write_plan(plan_root, "frontend-workspace-and-practice", "active")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="context yaml spec discovery metadata references name",
    )

    assert result["confidence"] == "none"
    assert result["recommended"] is None


def test_ignores_active_plan_without_required_context_manifest(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"
    _write_plan(
        plan_root,
        "harness-simplification",
        "active",
        with_context=False,
        with_checklist=False,
    )

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="optimize harness workflow",
    )

    assert result["confidence"] == "none"
    assert result["recommended"] is None


def test_obsolete_lifecycle_status_is_unknown():
    module = _load_module()
    assert module.normalize_status(OBSOLETE_EN_STATUS) == "unknown"
    assert module.normalize_status(OBSOLETE_ZH_STATUS) == "unknown"
    assert module.normalize_status("super" + "seded") == "unknown"


def test_tokenize_filters_stop_words_and_adds_action_inflections():
    module = _load_module()
    tokens = module.tokenize("created opening returns and of")
    assert {"create", "open", "return"}.issubset(tokens)
    assert tokens.isdisjoint({"and", "of"})
