"""Tests for change-intake plan matching helpers."""

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


def _write_plan(root: Path, name: str, status: str, context: dict):
    spec_dir = root / name
    plan_dir = spec_dir / "plans" / "001-backend"
    plan_dir.mkdir(parents=True)
    (plan_dir / "plan.md").write_text(
        "# Plan\n\n"
        "> **版本**: 1.0\n"
        f"> **状态**: {status}\n"
        "> **更新日期**: 2026-04-04\n"
        "## 1 Goal\n\nText.\n",
        encoding="utf-8",
    )
    (plan_dir / "checklist.md").write_text(
        "# Checklist\n\n"
        "> **版本**: 1.0\n"
        f"> **状态**: {status}\n"
        "> **更新日期**: 2026-04-04\n\n"
        "## Phase 1\n\n- [ ] 1.1 Item\n",
        encoding="utf-8",
    )
    (spec_dir / "spec.md").write_text("# Spec\n", encoding="utf-8")
    with open(plan_dir / "context.yaml", "w", encoding="utf-8") as f:
        yaml.safe_dump(context, f, sort_keys=False, allow_unicode=True)
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


def _base_context(name: str) -> dict:
    return {
        "apiVersion": "plancontext.agent.dev/v1alpha1",
        "kind": "PlanContext",
        "metadata": {
            "subspec": name,
            "name": "001-backend",
            "sequence": 1,
            "specVersion": {"from": None, "to": 1.0},
        },
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


def test_prefers_active_candidate_with_discovery(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    local_auth = _base_context("local-auth")
    local_auth["spec"]["discovery"] = {
        "aliases": ["auth", "login"],
        "keywords": ["sign in", "session refresh"],
    }
    local_auth["spec"]["targets"]["backend"]["discovery"] = {
        "apiNames": ["RefreshSession"],
    }
    _write_plan(plan_root, "local-auth", "active", local_auth)

    other = _base_context("wiki-sync")
    other["spec"]["discovery"] = {"aliases": ["wiki"], "keywords": ["sync page"]}
    _write_plan(plan_root, "wiki-sync", "active", other)

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="login session refresh is broken after sign in",
    )

    assert result["confidence"] in {"high", "medium"}
    assert result["recommended"]["plan"] == "001-backend"
    assert result["recommended"]["displayPlan"] == "local-auth/001-backend"
    assert "aliases=login" in result["recommended"]["reasons"]


def test_fallback_to_plan_and_target_names_when_discovery_missing(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    context = _base_context("manifest-commit-ordering-fix")
    _write_plan(plan_root, "manifest-commit-ordering-fix", "active", context)

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="need a fix for manifest commit ordering",
    )

    assert result["recommended"]["displayPlan"] == "manifest-commit-ordering-fix/001-backend"
    assert any(reason.startswith("displayName=") for reason in result["recommended"]["reasons"])


def test_completed_candidate_marks_in_place_revision(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    sealed = _base_context("sealed-secrets")
    sealed["spec"]["discovery"] = {
        "aliases": ["sealed secret"],
        "keywords": ["intentblocks", "selector mismatch"],
        "relatedBugs": ["BUG-0037"],
    }
    _write_plan(plan_root, "sealed-secrets", "completed", sealed)

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="BUG-0037 selector mismatch on sealed secret edit",
    )

    assert result["recommended"]["displayPlan"] == "sealed-secrets/001-backend"
    assert result["recommended"]["status"] == "completed"
    assert result["recommended"]["reviseInPlace"] is True


def test_ignores_unsupported_commands_discovery(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    commands_only = _base_context("commands-only")
    commands_only["spec"]["targets"]["backend"]["discovery"] = {
        "commands": ["make redeploy-local"],
    }
    _write_plan(plan_root, "commands-only", "active", commands_only)

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="make redeploy local",
    )

    assert result["confidence"] == "none"
    assert result["recommended"] is None
    assert result["candidates"] == []


def test_obsolete_lifecycle_status_is_unknown(tmp_path):
    module = _load_module()
    assert module.normalize_status(OBSOLETE_EN_STATUS) == "unknown"
    assert module.normalize_status(OBSOLETE_ZH_STATUS) == "unknown"
    assert module.normalize_status("super" + "seded") == "unknown"


def test_tokenize_filters_stop_words_and_adds_action_inflections():
    module = _load_module()

    tokens = module.tokenize("created opening returns and of")

    assert {"create", "open", "return"}.issubset(tokens)
    assert tokens.isdisjoint({"and", "of"})


def test_partial_field_score_counts_each_query_token_once():
    module = _load_module()
    query = "resume detail"

    score, reasons = module.score_discovery_values(
        query,
        module.tokenize(query),
        "keywords",
        ["resume detail", "resume detail route", "resume"],
    )

    assert score == 6
    assert reasons


def test_stop_word_only_exact_value_does_not_score():
    module = _load_module()
    query = "and of"

    score, reasons = module.score_discovery_values(
        query,
        module.tokenize(query),
        "keywords",
        ["and", "of"],
    )

    assert score == 0
    assert reasons == []


def test_precise_action_vocabulary_beats_repeated_generic_values(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    generic = _base_context("generic-resume-owner")
    generic["spec"]["discovery"] = {
        "aliases": ["resume workshop", "resume list", "resume detail"],
        "keywords": [
            "resume workshop route shell",
            "flat resume list table",
            "resume detail readonly",
            "resume detail waiting state",
            "resume workshop detail",
        ],
    }
    _write_plan(plan_root, "generic-resume-owner", "completed", generic)

    create_flow = _base_context("resume-create-flow-owner")
    create_flow["spec"]["discovery"] = {
        "aliases": ["resume-create-flow"],
        "keywords": [
            "static prototype Save and open",
            "create-to-detail waiting ready state",
        ],
    }
    _write_plan(plan_root, "resume-create-flow-owner", "active", create_flow)

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query=(
            "Resume Workshop static prototype Save and open returns to resume list "
            "instead of opening newly created resume detail waiting state"
        ),
    )

    assert result["recommended"]["displayPlan"] == "resume-create-flow-owner/001-backend"


def test_exact_scenario_readme_owner_outranks_generic_api_terms(tmp_path):
    module = _load_module()
    plan_root = tmp_path / "docs" / "spec"

    owner = _base_context("resume-detail-owner")
    owner["spec"]["discovery"] = {"aliases": ["resume detail"]}
    owner["spec"]["targets"]["backend"]["discovery"] = {
        "apiNames": ["getResume"],
    }
    owner_dir = _write_plan(plan_root, "resume-detail-owner", "completed", owner)

    generic = _base_context("report-dashboard")
    generic["spec"]["discovery"] = {
        "aliases": ["report dashboard"],
        "keywords": ["feedback report getResume polling"],
    }
    generic["spec"]["targets"]["backend"]["discovery"] = {
        "apiNames": ["getResume"],
    }
    _write_plan(plan_root, "report-dashboard", "active", generic)
    _write_scenario_owner(tmp_path, "P0.037", owner_dir / "plan.md")

    result = module.match_change_contexts(
        plan_root=str(plan_root),
        query="P0.037 getResume report polling remains loading",
    )

    assert result["recommended"]["displayPlan"] == "resume-detail-owner/001-backend"
    assert "scenarioOwner=E2E.P0.037" in result["recommended"]["reasons"]
