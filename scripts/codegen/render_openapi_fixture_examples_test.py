#!/usr/bin/env python3
"""Tests for scripts/codegen/render_openapi_fixture_examples.py.

Phase 3.1 contract per `002-fixtures-and-mock-source` plan:
- Reads openapi.yaml + fixtures, emits `openapi/.generated/openapi-with-fixtures.yaml`.
- 37 default named examples (one per spec §3.1.1 operation).
- Each example body is byte-equal to the fixture's `scenarios.default.response.body`.
- Re-running is idempotent.
"""

from __future__ import annotations

import hashlib
import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT = REPO_ROOT / "scripts" / "codegen" / "render_openapi_fixture_examples.py"
OUTPUT_REL = "openapi/.generated/openapi-with-fixtures.yaml"


def _run(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(SCRIPT), "--repo-root", str(repo)],
        capture_output=True, text=True, check=False,
    )


def _read_fixture_default_body(repo: Path, tag: str, opid: str):
    with (repo / "openapi" / "fixtures" / tag / f"{opid}.json").open(
        "r", encoding="utf-8"
    ) as f:
        return json.load(f)["scenarios"]["default"]["response"]["body"]


def _hash(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class RenderOpenapiFixtureExamplesTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self.tmp.name) / "repo"
        for sub in ("openapi", "scripts"):
            shutil.copytree(REPO_ROOT / sub, self.repo / sub)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def _output(self) -> Path:
        return self.repo / OUTPUT_REL

    def _load_output(self) -> dict:
        with self._output().open("r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    def test_run_clean(self) -> None:
        out = _run(self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        self.assertTrue(self._output().is_file(), "output file must exist")

    def test_37_default_examples_present(self) -> None:
        _run(self.repo)
        spec = self._load_output()
        count = 0
        for path, methods in (spec.get("paths") or {}).items():
            if not isinstance(methods, dict):
                continue
            for method, op in methods.items():
                if not isinstance(op, dict) or "operationId" not in op:
                    continue
                examples_found = False
                for status, resp in (op.get("responses") or {}).items():
                    content = (resp or {}).get("content") or {}
                    js = content.get("application/json") or {}
                    examples = js.get("examples") or {}
                    if "default" in examples:
                        examples_found = True
                        break
                if examples_found:
                    count += 1
        self.assertEqual(count, 37, "expected 37 operations to carry default named examples")

    def test_example_byte_equal_to_fixture_body(self) -> None:
        _run(self.repo)
        spec = self._load_output()

        # Spot-check across tags / status families.
        targets = [
            ("Auth", "getMe", "200"),
            ("Auth", "deleteMe", "202"),
            ("TargetJobs", "listTargetJobs", "200"),
            ("PracticeSessions", "getPracticeSession", "200"),
            ("Reports", "getFeedbackReport", "200"),
            ("Privacy", "requestPrivacyExport", "501"),
        ]
        index = {}
        for path, methods in (spec.get("paths") or {}).items():
            for method, op in methods.items():
                if isinstance(op, dict) and "operationId" in op:
                    index[op["operationId"]] = op
        for tag, opid, status in targets:
            with self.subTest(operationId=opid):
                op = index[opid]
                resp = op["responses"][status]
                example = resp["content"]["application/json"]["examples"]["default"]["value"]
                fixture_body = _read_fixture_default_body(self.repo, tag, opid)
                self.assertEqual(example, fixture_body)

    def test_missing_fixture_source_fails(self) -> None:
        (self.repo / "openapi" / "fixtures" / "Auth" / "deleteMe.json").unlink()
        out = _run(self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("deleteMe", out.stderr + out.stdout)

    def test_idempotent(self) -> None:
        first = _run(self.repo)
        self.assertEqual(first.returncode, 0)
        h1 = _hash(self._output())
        second = _run(self.repo)
        self.assertEqual(second.returncode, 0)
        h2 = _hash(self._output())
        self.assertEqual(h1, h2, "re-running render must produce byte-identical output")


if __name__ == "__main__":
    unittest.main()
