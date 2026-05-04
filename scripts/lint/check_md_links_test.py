#!/usr/bin/env python3
"""Contract tests for scripts/lint/check_md_links.py."""

from __future__ import annotations

import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("check_md_links.py")
MODULE_NAME = "check_md_links_under_test"


def load_linter():
    spec = importlib.util.spec_from_file_location(MODULE_NAME, SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    sys.modules[MODULE_NAME] = module
    spec.loader.exec_module(module)
    return module


class CheckMdLinksTest(unittest.TestCase):
    def test_clean_directory_returns_no_findings(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\nlink to [B](b.md)\n", encoding="utf-8")
            (root / "b.md").write_text("# B\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_broken_relative_link_is_reported(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\nlink to [Missing](missing.md)\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(len(findings), 1)
            self.assertEqual(findings[0].target, "missing.md")
            self.assertTrue(findings[0].source.name == "a.md")

    def test_anchor_only_link_is_skipped(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[Top](#top)\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_external_link_is_skipped(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text(
                "# A\n[ext](https://example.com)\n[mail](mailto:x@y)\n", encoding="utf-8"
            )
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_link_with_anchor_strips_fragment(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[B Section](b.md#section)\n", encoding="utf-8")
            (root / "b.md").write_text("# B\n## Section\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_check_fragments_reports_missing_heading_anchor(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[B Missing](b.md#missing-heading)\n", encoding="utf-8")
            (root / "b.md").write_text("# B\n## Present Heading\n", encoding="utf-8")
            findings = linter.scan_directory(root, check_fragments=True)
            self.assertEqual(len(findings), 1)
            self.assertEqual(findings[0].target, "b.md#missing-heading")
            self.assertEqual(findings[0].kind, "fragment")

    def test_check_fragments_validates_in_page_anchor(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[Missing](#missing-section)\n", encoding="utf-8")
            findings = linter.scan_directory(root, check_fragments=True)
            self.assertEqual(len(findings), 1)
            self.assertEqual(findings[0].target, "#missing-section")

    def test_check_fragments_accepts_github_style_heading_slugs(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text(
                "# A\n[Chinese](b.md#63-s2--backend-domain-implementation)\n"
                "[JobType](b.md#311-dbc8-canonical-job_type--asynq-dotted-task-name-映射)\n"
                "[Duplicate](b.md#重复-heading-1)\n",
                encoding="utf-8",
            )
            (root / "b.md").write_text(
                "# B\n"
                "## 6.3 S2 · Backend domain implementation\n"
                "### 3.1.1 DB/C8 canonical job_type ↔ Asynq dotted task name 映射\n"
                "## 重复 Heading\n"
                "## 重复 Heading\n",
                encoding="utf-8",
            )
            findings = linter.scan_directory(root, check_fragments=True)
            self.assertEqual(findings, [])

    def test_fragments_are_not_checked_unless_enabled(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[B Missing](b.md#missing-heading)\n", encoding="utf-8")
            (root / "b.md").write_text("# B\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_relative_parent_link_resolves(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            sub = root / "sub"
            sub.mkdir()
            (sub / "a.md").write_text("# A\n[Up](../top.md)\n", encoding="utf-8")
            (root / "top.md").write_text("# Top\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_link_to_directory_is_valid(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text("# A\n[Subdir](sub/)\n", encoding="utf-8")
            (root / "sub").mkdir()
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_html_comment_block_is_skipped(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text(
                "# A\n<!-- example: [Bad](missing.md) -->\nReal: [B](b.md)\n",
                encoding="utf-8",
            )
            (root / "b.md").write_text("# B\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_inline_code_span_is_skipped(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text(
                "# A\nSee `[plans](...)` for example syntax. Real link: [B](b.md)\n",
                encoding="utf-8",
            )
            (root / "b.md").write_text("# B\n", encoding="utf-8")
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_code_fence_links_are_skipped(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "a.md").write_text(
                "# A\n```\n[Sample](missing.md)\n```\n",
                encoding="utf-8",
            )
            findings = linter.scan_directory(root)
            self.assertEqual(findings, [])

    def test_ignore_pattern_excludes_template_files(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "TEMPLATES.md").write_text("# T\n[X](xxx-placeholder.md)\n", encoding="utf-8")
            (root / "real.md").write_text("# Real\n[Y](missing.md)\n", encoding="utf-8")
            findings = linter.scan_directory(root, ignores=["**/TEMPLATES.md"])
            self.assertEqual(len(findings), 1)
            self.assertEqual(findings[0].source.name, "real.md")


if __name__ == "__main__":
    unittest.main()
