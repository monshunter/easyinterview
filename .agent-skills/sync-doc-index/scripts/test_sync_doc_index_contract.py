"""Contract checks for sync-doc-index handling of template assets and migration."""

from __future__ import annotations

import importlib.util
from pathlib import Path

import pytest


SCRIPT_PATH = Path(__file__).resolve().parent / "sync-doc-index.py"


def _load_module():
    spec = importlib.util.spec_from_file_location("sync_doc_index", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def _write_doc(path: Path, status: str, version: str = "1.0", date: str = "2026-04-26") -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        f"# {path.parent.name}\n"
        f"\n"
        f"> **版本**: {version}\n"
        f"> **状态**: {status}\n"
        f"> **更新日期**: {date}\n",
        encoding="utf-8",
    )


def _write_plan_triplet(plans_dir: Path, plan_name: str, status: str) -> None:
    """Create plan.md and checklist.md with the same status under <plans_dir>/<plan_name>/."""
    plan_dir = plans_dir / plan_name
    _write_doc(plan_dir / "plan.md", status)
    _write_doc(plan_dir / "checklist.md", status)


def _bootstrap_subspec(tmp_path: Path, subspec: str = "demo-spec") -> Path:
    """Bootstrap a docs/spec/<subspec>/ tree with valid spec.md/history.md/plans/INDEX.md."""
    spec_dir = tmp_path / "docs" / "spec" / subspec
    plans_dir = spec_dir / "plans"
    plans_dir.mkdir(parents=True)
    _write_doc(spec_dir / "spec.md", "active")
    _write_doc(spec_dir / "history.md", "active")
    # Top-level docs/spec/INDEX.md to satisfy run_check
    (tmp_path / "docs" / "spec" / "INDEX.md").write_text(
        f"# 索引\n\n## 1\n\n"
        f"| Subject | 版本 | 状态 | 更新日期 | Plans |\n"
        f"|---------|------|------|----------|-------|\n"
        f"| [{subspec}](./{subspec}/spec.md) | 1.0 | active | 2026-04-26 | [plans](./{subspec}/plans/) |\n",
        encoding="utf-8",
    )
    return plans_dir


def test_run_check_ignores_templates_assets(tmp_path):
    module = _load_module()

    spec_dir = tmp_path / "docs" / "spec"
    spec_dir.mkdir(parents=True)

    (spec_dir / "INDEX.md").write_text(
        "# 设计文档索引\n\n"
        "| 文档 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|----------|\n",
        encoding="utf-8",
    )
    (spec_dir / "TEMPLATES.md").write_text("# 模板资产\n", encoding="utf-8")

    report = module.run_check(tmp_path)

    assert report["header_violations"] == []
    assert report["index_drifts"] == []
    assert report["orphans"]["missing_from_index"] == []


def test_fix_spec_index_preserves_extra_plans_column(tmp_path):
    """Column fixes must not drop the spec INDEX `Plans` projection cell."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path, subspec="demo-spec")
    spec_dir = plans_dir.parent
    _write_doc(spec_dir / "spec.md", "active", version="1.1", date="2026-04-27")

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=False)

    assert skipped == []
    assert any(f.get("index") == "docs/spec/INDEX.md" for f in fixes)
    text = (tmp_path / "docs" / "spec" / "INDEX.md").read_text(encoding="utf-8")
    assert (
        "| [demo-spec](./demo-spec/spec.md) | 1.1 | active | 2026-04-27 | [plans](./demo-spec/plans/) |"
        in text
    )

    report = module.run_check(tmp_path)
    assert report["summary"]["drifts"] == 0


def test_migrate_active_to_existing_completed_section(tmp_path):
    """A row whose linked plan flipped to `completed` should move into the
    existing `已完成` section, leaving the source `进行中` table empty."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="completed")
    _write_plan_triplet(plans_dir, "002-onboarding", status="active")

    (plans_dir / "INDEX.md").write_text(
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | completed | 2026-04-26 |\n"
        "| [002-onboarding](./002-onboarding/plan.md) | [plan](./002-onboarding/plan.md) / [checklist](./002-onboarding/checklist.md) | 1.0 | active | 2026-04-26 |\n"
        "\n## 2 已完成（Completed）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 完成日期 |\n"
        "|------|------|------|------|----------|\n",
        encoding="utf-8",
    )

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=False)
    assert skipped == []
    migrate_fixes = [f for f in fixes if f.get("action") == "migrate_row"]
    assert len(migrate_fixes) == 1
    assert migrate_fixes[0]["to_status"] == "completed"

    text = (plans_dir / "INDEX.md").read_text(encoding="utf-8")
    completed_idx = text.index("## 2 已完成")
    active_idx = text.index("## 1 进行中")
    assert text.index("001-bootstrap", completed_idx) > completed_idx
    # The active section should no longer contain 001-bootstrap
    assert "001-bootstrap" not in text[active_idx:completed_idx]
    # 002-onboarding remains in active
    assert "002-onboarding" in text[active_idx:completed_idx]

    # Post-fix check: zero drifts
    report = module.run_check(tmp_path)
    assert report["summary"]["drifts"] == 0


def test_migrate_creates_missing_completed_section(tmp_path):
    """When `已完成` section doesn't exist yet, the migration should create one
    at the end of the file with the standard header (`完成日期` column)."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="completed")

    (plans_dir / "INDEX.md").write_text(
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | completed | 2026-04-26 |\n",
        encoding="utf-8",
    )

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=False)
    assert skipped == []
    text = (plans_dir / "INDEX.md").read_text(encoding="utf-8")
    assert "## 2 已完成（Completed）" in text
    assert "完成日期" in text  # the new section uses the canonical column header
    completed_idx = text.index("## 2 已完成")
    assert "001-bootstrap" in text[completed_idx:]

    report = module.run_check(tmp_path)
    assert report["summary"]["drifts"] == 0


def test_migrate_skips_superseded_transitions(tmp_path):
    """Migrations involving the superseded section have a different column
    schema (no version/date) and must be left for the LLM, not auto-moved."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="superseded")

    (plans_dir / "INDEX.md").write_text(
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | superseded | 2026-04-26 |\n",
        encoding="utf-8",
    )

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=False)
    migrate_fixes = [f for f in fixes if f.get("action") == "migrate_row"]
    assert migrate_fixes == []
    assert len(skipped) == 1
    text = (plans_dir / "INDEX.md").read_text(encoding="utf-8")
    # Row stays put because we refuse to fabricate the superseded shape.
    assert "001-bootstrap" in text.split("## 1 进行中")[1]


def test_no_migration_when_status_matches_section(tmp_path):
    """Rows whose status already matches the section heading must not be touched."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="active")

    original = (
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | active | 2026-04-26 |\n"
    )
    (plans_dir / "INDEX.md").write_text(original, encoding="utf-8")

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=False)
    assert [f for f in fixes if f.get("action") == "migrate_row"] == []
    assert skipped == []
    assert (plans_dir / "INDEX.md").read_text(encoding="utf-8") == original


def test_dry_run_does_not_modify_index(tmp_path):
    """`--fix-index --dry-run` reports the migration but never touches disk."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="completed")

    original = (
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | completed | 2026-04-26 |\n"
    )
    (plans_dir / "INDEX.md").write_text(original, encoding="utf-8")

    fixes, skipped = module.fix_index_columns(tmp_path, dry_run=True)
    migrate_fixes = [f for f in fixes if f.get("action") == "migrate_row"]
    assert len(migrate_fixes) == 1
    assert (plans_dir / "INDEX.md").read_text(encoding="utf-8") == original


def test_check_reports_status_group_drift_as_auto_fixable(tmp_path):
    """`run_check` should mark active↔completed group drifts as auto_fixable=True
    so the report no longer pushes them into the LLM bucket."""
    module = _load_module()
    plans_dir = _bootstrap_subspec(tmp_path)

    _write_plan_triplet(plans_dir, "001-bootstrap", status="completed")
    (plans_dir / "INDEX.md").write_text(
        "# Plans 索引\n\n"
        "## 1 进行中（Active）\n\n"
        "| 计划 | 文件 | 版本 | 状态 | 更新日期 |\n"
        "|------|------|------|------|----------|\n"
        "| [001-bootstrap](./001-bootstrap/plan.md) | [plan](./001-bootstrap/plan.md) / [checklist](./001-bootstrap/checklist.md) | 1.0 | completed | 2026-04-26 |\n",
        encoding="utf-8",
    )

    report = module.run_check(tmp_path)
    group_drifts = [d for d in report["index_drifts"] if d["field"] == "状态(group)"]
    assert group_drifts, "expected a 状态(group) drift in the report"
    assert all(d["auto_fixable"] for d in group_drifts)
