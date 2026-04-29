#!/usr/bin/env python3
"""Contract tests for scripts/lint/conventions_drift.py."""

from __future__ import annotations

import importlib.util
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("conventions_drift.py")


def load_linter():
    spec = importlib.util.spec_from_file_location("conventions_drift_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class ConventionsDriftTest(unittest.TestCase):
    def setUp(self) -> None:
        self.linter = load_linter()

    def test_generated_output_set_covers_enum_error_and_ai_go_ts(self) -> None:
        outputs = set(self.linter.GENERATED_OUTPUTS)

        self.assertIn("backend/internal/shared/types/enums.go", outputs)
        self.assertIn("backend/internal/shared/errors/codes.go", outputs)
        self.assertIn("backend/internal/shared/ai/vocabulary.go", outputs)
        self.assertIn("frontend/src/lib/conventions/enums.ts", outputs)
        self.assertIn("frontend/src/lib/conventions/errors.ts", outputs)
        self.assertIn("frontend/src/lib/conventions/ai.ts", outputs)

    def test_reports_go_ai_vocabulary_drift_with_direction(self) -> None:
        actual = {"backend/internal/shared/ai/vocabulary.go": "hand edited go"}
        expected = {"backend/internal/shared/ai/vocabulary.go": "canonical go"}

        errs = self.linter.compare_generated_outputs(actual, expected)

        self.assertTrue(
            any(
                "AI vocabulary drift" in err
                and "shared/conventions.yaml -> backend/internal/shared/ai/vocabulary.go" in err
                for err in errs
            ),
            errs,
        )

    def test_reports_ts_ai_vocabulary_drift_with_direction(self) -> None:
        actual = {"frontend/src/lib/conventions/ai.ts": "hand edited ts"}
        expected = {"frontend/src/lib/conventions/ai.ts": "canonical ts"}

        errs = self.linter.compare_generated_outputs(actual, expected)

        self.assertTrue(
            any(
                "AI vocabulary drift" in err
                and "shared/conventions.yaml -> frontend/src/lib/conventions/ai.ts" in err
                for err in errs
            ),
            errs,
        )

    def test_reports_missing_generated_output_with_direction(self) -> None:
        actual: dict[str, str] = {}
        expected = {"frontend/src/lib/conventions/ai.ts": "canonical ts"}

        errs = self.linter.compare_generated_outputs(actual, expected)

        self.assertTrue(
            any(
                "missing generated output" in err
                and "shared/conventions.yaml -> frontend/src/lib/conventions/ai.ts" in err
                for err in errs
            ),
            errs,
        )

    def test_make_codegen_check_invokes_drift_wrapper_and_ai_paths(self) -> None:
        makefile = Path(__file__).resolve().parents[2] / "Makefile"
        src = makefile.read_text(encoding="utf-8")

        self.assertIn("scripts/lint/conventions_drift.py", src)
        self.assertIn("backend/internal/shared/ai", src)
        self.assertIn("frontend/src/lib/conventions/ai.ts", src)


if __name__ == "__main__":
    unittest.main()
