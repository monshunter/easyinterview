"""Tests for change-intake plan matching helpers."""

from __future__ import annotations

import importlib.util
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
