"""Contract checks for sync-doc-index handling of template assets."""

from __future__ import annotations

import importlib.util
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().parent / "sync-doc-index.py"


def _load_module():
    spec = importlib.util.spec_from_file_location("sync_doc_index", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


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
