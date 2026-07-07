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

    def test_rejects_non_current_worker_process_terms_in_active_surfaces(self) -> None:
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
                "round-trip fixtures cover `api` / `worker` producer.\n",
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
            write(
                repo
                / "docs"
                / "spec"
                / "secrets-and-config"
                / "plans"
                / "001-bootstrap"
                / "plan.md",
                "validator.go covers app/worker listen addr.\nworker bindings stayed in the remediation note.\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "engineering-roadmap"
                / "decisions"
                / "ADR-Q3-analytics-platform.md",
                "C8 backend-async-runtime owns analytics_dispatch.\n",
            )
            write(
                repo / "scripts" / "lint" / "env_dict.py",
                'scan_roots = ["backend/cmd/api", "backend/cmd/worker"]\n',
            )
            write(
                repo / "scripts" / "lint" / "getenv_boundary.go",
                'package main\n\nvar keys = []string{"APP_LISTEN_ADDR", "WORKER_LISTEN_ADDR"}\n',
            )
            write(
                repo / "shared" / "events.yaml",
                "events:\n  - name: target.parsed\n    producer: worker\n",
            )
            write(
                repo / "shared" / "events" / "baseline" / "events.v1.json",
                '{"events":[{"name":"target.parsed","producer":"worker"}]}\n',
            )

            result = run_lint(repo)

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("privacy worker wording", result.stderr)
            self.assertIn("worker producer enum", result.stderr)
            self.assertIn("cmd/worker entrypoint", result.stderr)
            self.assertIn("worker listen addr config", result.stderr)
            self.assertIn("worker config bindings", result.stderr)
            self.assertIn("backend async runner subject shorthand", result.stderr)
            self.assertIn("scripts/lint/env_dict.py", result.stderr)
            self.assertIn("scripts/lint/getenv_boundary.go", result.stderr)
            self.assertIn("shared/events.yaml", result.stderr)
            self.assertIn("shared/events/baseline/events.v1.json", result.stderr)

    def test_rejects_structured_multiline_worker_producer_values(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(
                repo / "shared" / "events.yaml",
                "envelope:\n"
                "  fields:\n"
                "    - name: producer\n"
                "      values:\n"
                "        - api\n"
                "        - worker\n"
                "events:\n"
                "  - name: target.parsed\n"
                "    producer:\n"
                "      - worker\n",
            )
            write(
                repo / "shared" / "events" / "schemas" / "target.parsed.v1.json",
                '{\n'
                '  "properties": {\n'
                '    "producer": {\n'
                '      "enum": [\n'
                '        "api",\n'
                '        "worker"\n'
                "      ]\n"
                "    }\n"
                "  }\n"
                "}\n",
            )

            result = run_lint(repo)

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("worker producer enum", result.stderr)
            self.assertIn("shared/events.yaml", result.stderr)
            self.assertIn("shared/events/schemas/target.parsed.v1.json", result.stderr)

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
                / "backend-runtime-topology"
                / "plans"
                / "001-worker-consolidation"
                / "checklist.md",
                "删除 `cmd/worker`，确认 `WORKER_LISTEN_ADDR` 不再作为 current config。\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "secrets-and-config"
                / "history.md",
                "Prior cmd/worker wording.\n",
            )
            write(
                repo / "backend" / "internal" / "platform" / "config" / "validator_test.go",
                "package config_test\n\nfunc TestWorkerRemoved() { _ = \"WORKER_LISTEN_ADDR\" }\n",
            )
            write(
                repo / "scripts" / "lint" / "runtime_topology.py",
                "NON_CURRENT_PATTERNS = ['cmd/worker', 'WORKER_LISTEN_ADDR', 'producer: worker']\n",
            )

            result = run_lint(repo)

            self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_owner_current_handoff_worker_process_terms(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            write(
                repo
                / "docs"
                / "spec"
                / "backend-runtime-topology"
                / "plans"
                / "001-worker-consolidation"
                / "plan.md",
                "Current handoff: build backend/cmd/worker as the runtime entry.\n",
            )
            write(
                repo
                / "docs"
                / "spec"
                / "backend-runtime-topology"
                / "plans"
                / "001-worker-consolidation"
                / "checklist.md",
                "Verification command: go test ./internal/platform/config ./cmd/worker -count=1\n",
            )

            result = run_lint(repo)

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("owner current handoff", result.stderr)
            self.assertIn("docs/spec/backend-runtime-topology/plans/001-worker-consolidation/plan.md", result.stderr)
            self.assertIn("docs/spec/backend-runtime-topology/plans/001-worker-consolidation/checklist.md", result.stderr)


if __name__ == "__main__":
    unittest.main()
