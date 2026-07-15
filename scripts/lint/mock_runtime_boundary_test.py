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

    def test_scoped_out_of_scope_contract_token_fails_with_owner_hint(self) -> None:
        rel = "openapi/fixtures/Auth/getRuntimeConfig.json"
        path = self.repo / rel
        data = json.loads(path.read_text(encoding="utf-8"))
        data["scenarios"]["default"]["response"]["body"]["featureFlags"]["ai.gateway_route"] = True
        path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

        out = _run_lint(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("ai.gateway", out.stderr + out.stdout)
        self.assertIn("docs/spec/mock-contract-suite/spec.md", out.stderr + out.stdout)

    def test_session_scoped_voice_operation_is_allowed_but_standalone_voice_route_fails(self) -> None:
        client = self.repo / "frontend/src/api/voice-session.ts"
        client.write_text(
            textwrap.dedent(
                """\
                export const path = "/practice/sessions/{sessionId}/voice-turns";
                export interface PracticeVoiceTurnResult { voiceTurnId: string }
                """
            ),
            encoding="utf-8",
        )

        out = _run_lint(self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")

        client.write_text('export const path = "/api/v1/voice/sessions";\n', encoding="utf-8")
        out = _run_lint(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("out-of-scope mock/API token '/voice'", out.stderr + out.stdout)

    def test_out_of_scope_voice_tag_requires_word_boundary(self) -> None:
        client = self.repo / "frontend/src/api/voice-types.ts"
        client.write_text("export type PracticeVoiceTurnResult = { voiceTurnId: string };\n", encoding="utf-8")
        out = _run_lint(self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")

        client.write_text('export const tag = "Voice";\n', encoding="utf-8")
        out = _run_lint(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("out-of-scope mock/API token 'Voice'", out.stderr + out.stdout)

    def test_out_of_scope_fixture_tag_directory_fails_even_when_empty(self) -> None:
        out_of_scope = self.repo / "openapi/fixtures/Growth"
        if out_of_scope.exists():
            shutil.rmtree(out_of_scope)
        out_of_scope.mkdir()

        out = _run_lint(self.repo)

        self.assertNotEqual(out.returncode, 0)
        self.assertIn("unexpected fixture tag directory 'Growth'", out.stderr + out.stdout)
        self.assertIn("docs/spec/mock-contract-suite/spec.md", out.stderr + out.stdout)


if __name__ == "__main__":
    unittest.main()
