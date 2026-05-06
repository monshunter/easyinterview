#!/usr/bin/env python3
"""Contract tests for scripts/lint/runtime_topology.py."""
from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path


REPO = Path(__file__).resolve().parents[2]
SCRIPT = REPO / "scripts" / "lint" / "runtime_topology.py"


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


class RuntimeTopologyLintTest(unittest.TestCase):
    def test_passes_backend_internal_runner_terms(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(
                repo / "backend" / "internal" / "migrations" / "privacy.go",
                "package migrations\n\n// Used by the backend internal privacy runner.\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "event-and-outbox-contract"
                / "plans"
                / "001-bootstrap"
                / "plan.md",
                "producer enum (`api` / `backend_async` / `dispatcher` / `review`)\n",
            )

            result = run_lint(repo)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("runtime_topology: OK", result.stdout)

    def test_rejects_retired_worker_process_terms_in_active_surfaces(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(
                repo / "backend" / "internal" / "migrations" / "privacy.go",
                "package migrations\n\n// Used by downstream privacy workers.\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "event-and-outbox-contract"
                / "plans"
                / "001-bootstrap"
                / "plan.md",
                "producer enum (`api` / `worker` / `dispatcher` / `review`)\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "secrets-and-config"
                / "plans"
                / "001-bootstrap"
                / "checklist.md",
                "验证: go test ./internal/platform/config ./cmd/worker -count=1\n",
            )

            result = run_lint(repo)

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("privacy worker wording", result.stderr)
            self.assertIn("worker producer enum", result.stderr)
            self.assertIn("cmd/worker entrypoint", result.stderr)

    def test_allows_owner_negative_assertions_history_and_tests(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(
                repo
                / "docs"
                / "spec"
                / "backend-runtime-topology"
                / "spec.md",
                "`cmd/worker`, `WORKER_LISTEN_ADDR`, `worker.listenAddr` must be removed.\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "secrets-and-config"
                / "history.md",
                "Historical cmd/worker wording.\n",
            )
            write(
                repo / "backend" / "internal" / "platform" / "config" / "validator_test.go",
                "package config_test\n\nfunc TestWorkerRemoved() { _ = \"WORKER_LISTEN_ADDR\" }\n",
            )

            result = run_lint(repo)

            self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
