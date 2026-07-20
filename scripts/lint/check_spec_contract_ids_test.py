#!/usr/bin/env python3
"""Contract tests for scripts/lint/check_spec_contract_ids.py."""

from __future__ import annotations

import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("check_spec_contract_ids.py")
MODULE_NAME = "check_spec_contract_ids_under_test"
REPO_ROOT = Path(__file__).resolve().parents[2]


def load_linter():
    spec = importlib.util.spec_from_file_location(MODULE_NAME, SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    sys.modules[MODULE_NAME] = module
    spec.loader.exec_module(module)
    return module


def write_spec(root: Path, subject: str, body: str) -> Path:
    path = root / subject / "spec.md"
    path.parent.mkdir(parents=True)
    path.write_text(body, encoding="utf-8")
    return path


class CheckSpecContractIdsTest(unittest.TestCase):
    def test_clean_subjects_allow_the_same_id_in_different_specs(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_spec(root, "alpha", "| ID | Decision |\n|---|---|\n| D-1 | A |\n| C-1 | A |\n")
            write_spec(root, "beta", "| ID | Decision |\n|---|---|\n| D-1 | B |\n| C-1 | B |\n")

            self.assertEqual(linter.scan_directory(root), [])

    def test_duplicate_decision_and_acceptance_ids_report_both_lines(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = write_spec(
                root,
                "alpha",
                "# Spec\n| D-2 | first |\n| C-3 | first |\n| D-2 | duplicate |\n| C-3 | duplicate |\n",
            )

            findings = linter.scan_directory(root)

            self.assertEqual(
                [(item.contract_id, item.first_line, item.duplicate_line) for item in findings],
                [("D-2", 2, 4), ("C-3", 3, 5)],
            )
            self.assertEqual({item.source for item in findings}, {source})

    def test_fenced_examples_and_nested_non_subject_specs_are_ignored(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_spec(
                root,
                "alpha",
                "```markdown\n| D-4 | example |\n| D-4 | example |\n```\n| C-1 | live |\n",
            )
            nested = root / "alpha" / "plans" / "001" / "spec.md"
            nested.parent.mkdir(parents=True)
            nested.write_text("| C-9 | one |\n| C-9 | two |\n", encoding="utf-8")

            self.assertEqual(linter.scan_directory(root), [])

    def test_explicit_legacy_baseline_does_not_hide_new_subject_duplicates(self):
        linter = load_linter()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write_spec(
                root,
                "ai-provider-and-model-routing",
                "| C-14 | legacy |\n| C-14 | legacy |\n",
            )
            write_spec(root, "new-subject", "| C-14 | new |\n| C-14 | duplicate |\n")

            findings = linter.scan_directory(root)

            self.assertEqual(len(findings), 1)
            self.assertEqual(findings[0].source, root / "new-subject" / "spec.md")
            self.assertEqual(findings[0].contract_id, "C-14")

    def test_make_docs_check_runs_the_contract_id_linter(self):
        makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")
        docs_recipe = makefile.split("\ndocs-check:", 1)[1].split("\ncodegen-check:", 1)[0]

        self.assertIn("scripts/lint/check_spec_contract_ids.py", docs_recipe)


if __name__ == "__main__":
    unittest.main()
