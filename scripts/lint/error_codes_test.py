#!/usr/bin/env python3
"""Contract tests for scripts/lint/error_codes.py."""

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("error_codes.py")


def load_linter():
    spec = importlib.util.spec_from_file_location("error_codes_under_test", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class ErrorCodesLintTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.go_codes = self.root / "backend/internal/shared/errors/codes.go"
        self.ts_codes = self.root / "frontend/src/lib/conventions/errors.ts"
        self.ts_src_root = self.root / "frontend/src"
        self.go_codes.parent.mkdir(parents=True)
        self.ts_codes.parent.mkdir(parents=True)

        self.go_codes.write_text(
            'package errors\n\nconst (\n\tCodeAuthUnauthorized = "AUTH_UNAUTHORIZED"\n)\n',
            encoding="utf-8",
        )
        self.ts_codes.write_text(
            "export const ERROR_CODES = {\n"
            "  AUTH_UNAUTHORIZED: 'AUTH_UNAUTHORIZED',\n"
            "} as const;\n",
            encoding="utf-8",
        )

        self.linter = load_linter()
        self.linter.ROOT = self.root
        self.linter.GO_CODES = self.go_codes
        self.linter.TS_CODES = self.ts_codes
        self.linter.TS_SRC_ROOT = self.ts_src_root

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_clean_generated_shape_passes(self) -> None:
        self.assertEqual([], self.linter.check_ts_codes())

    def test_rejects_lowercase_ts_error_code_entry(self) -> None:
        self.ts_codes.write_text(
            "export const ERROR_CODES = {\n"
            "  AUTH_UNAUTHORIZED: 'AUTH_UNAUTHORIZED',\n"
            "  auth_unauthorized: 'auth_unauthorized',\n"
            "} as const;\n",
            encoding="utf-8",
        )

        errs = self.linter.check_ts_codes()

        self.assertTrue(
            any("auth_unauthorized" in err and "UPPER_SNAKE_CASE" in err for err in errs),
            errs,
        )

    def test_rejects_ts_error_code_value_with_valid_key(self) -> None:
        self.ts_codes.write_text(
            "export const ERROR_CODES = {\n"
            "  AUTH_UNAUTHORIZED: 'auth_unauthorized',\n"
            "} as const;\n",
            encoding="utf-8",
        )

        errs = self.linter.check_ts_codes()

        self.assertTrue(
            any("auth_unauthorized" in err and "UPPER_SNAKE_CASE" in err for err in errs),
            errs,
        )


if __name__ == "__main__":
    unittest.main()
