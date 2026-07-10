#!/usr/bin/env python3
"""Contract tests for scripts/lint/ai_provider_terminology.py."""
from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path


REPO = Path(__file__).resolve().parents[2]
SCRIPT = REPO / "scripts" / "lint" / "ai_provider_terminology.py"


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def run_lint(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )


class AIProviderTerminologyLintTest(unittest.TestCase):
    def test_passes_provider_neutral_active_surface(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(repo / ".env.example", "AI_PROVIDER_BASE_URL=\nAI_PROVIDER_API_KEY=\n")
            write(
                repo / "backend" / "internal" / "ai" / "aiclient" / "config.go",
                "package aiclient\n\ntype Config struct { ProviderBaseURL string }\n",
            )
            write(
                repo / "docs" / "spec" / "ai-provider-and-model-routing" / "spec.md",
                "# AI Provider and Model Routing Spec\n",
            )

            result = run_lint(repo)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("ai_provider_terminology: OK", result.stdout)

    def test_rejects_out_of_scope_active_env_and_schema_terms(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(repo / ".env.example", "AI_GATEWAY_BASE_URL=\n")
            write(
                repo / "config" / "ai-profiles" / "review.report.default.yaml",
                "gateway_route: review.report\n",
            )

            result = run_lint(repo)

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("AI_GATEWAY env key", result.stderr)
            self.assertIn("gateway_route schema key", result.stderr)

    def test_ignores_evidence_paths(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(repo / "docs" / "work-journal" / "2026-05-05.md", "AI_GATEWAY_BASE_URL\n")
            write(repo / "docs" / "reports" / "old.md", "gateway_route\n")
            write(repo / "docs" / "bugs" / "BUG-0001.md", "GatewayRoute\n")
            write(
                repo
                / "docs"
                / "spec"
                / "ai-provider-and-model-routing"
                / "history.md",
                "AI gateway prior wording\n",
            )

            result = run_lint(repo)

            self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
