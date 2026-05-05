#!/usr/bin/env python3
"""Tests for mock runtime boundary lint."""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import textwrap
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT = REPO_ROOT / "scripts" / "lint" / "mock_runtime_boundary.py"


def _run_lint(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(SCRIPT), "--repo-root", str(repo)],
        capture_output=True,
        text=True,
        check=False,
    )


class MockRuntimeBoundaryTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self.tmp.name) / "repo"
        shutil.copytree(REPO_ROOT / "openapi", self.repo / "openapi")
        shutil.copytree(REPO_ROOT / "frontend" / "src", self.repo / "frontend" / "src")

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_clean_runtime_boundary_passes(self) -> None:
        out = _run_lint(self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")

    def test_frontend_runtime_importing_ui_design_data_fails(self) -> None:
        leak = self.repo / "frontend/src/api/leak.ts"
        leak.write_text(
            textwrap.dedent(
                """\
                import { mockData } from "../../../ui-design/src/data.jsx";
                export const leaked = mockData;
                """
            ),
            encoding="utf-8",
        )

        out = _run_lint(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("ui-design/src/data.jsx", out.stderr + out.stdout)
        self.assertIn("frontend/src/api/leak.ts", out.stderr + out.stdout)

    def test_fixture_response_prototype_only_display_field_fails(self) -> None:
        rel = "openapi/fixtures/Auth/getMe.json"
        path = self.repo / rel
        data = json.loads(path.read_text(encoding="utf-8"))
        data["scenarios"]["default"]["response"]["body"]["statusTone"] = "green"
        path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

        out = _run_lint(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("getMe", out.stderr + out.stdout)
        self.assertIn("statusTone", out.stderr + out.stdout)


if __name__ == "__main__":
    unittest.main()
