#!/usr/bin/env python3
"""Tests for scripts/codegen/sync_fixtures_from_prototype.py.

Covers Phase 2.2 contract:
- Running the sync writes a `scenarios.prototype-baseline` to the 8 P0 P0
  closed-loop endpoints listed in plan 2.4.
- Re-running is idempotent (`git diff --exit-code` clean).
- After sync, `validate_fixtures.py` exits 0 (default + prototype-baseline
  both schema-valid).
- A mapping gap (data.jsx without a required section) makes the sync exit
  non-zero with a clear `Mapping gap:` message.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SYNC = REPO_ROOT / "scripts" / "codegen" / "sync_fixtures_from_prototype.py"
VALIDATE = REPO_ROOT / "scripts" / "lint" / "validate_fixtures.py"

P0_BASELINE_OPS = (
    ("Auth", "getMe"),
    ("Profile", "listExperienceCards"),
    ("TargetJobs", "listTargetJobs"),
    ("TargetJobs", "getTargetJob"),
    ("PracticeSessions", "getPracticeSession"),
    ("Reports", "getFeedbackReport"),
    ("Mistakes", "listMistakes"),
    ("Growth", "getGrowthOverview"),
)


def _run(script: Path, repo: Path, *extra: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["python3", str(script), "--repo-root", str(repo), *extra],
        capture_output=True,
        text=True,
        check=False,
    )


def _read_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


class SyncFixturesFromPrototypeTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Path(self.tmp.name) / "repo"
        for sub in ("openapi", "easyinterview-ui", "scripts"):
            shutil.copytree(REPO_ROOT / sub, self.repo / sub)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_sync_runs_clean_and_writes_p0_prototype_baseline(self) -> None:
        out = _run(SYNC, self.repo)
        self.assertEqual(out.returncode, 0, msg=f"stdout={out.stdout}\nstderr={out.stderr}")
        for tag, opid in P0_BASELINE_OPS:
            path = self.repo / "openapi/fixtures" / tag / f"{opid}.json"
            data = _read_json(path)
            scenarios = data.get("scenarios", {})
            with self.subTest(operationId=opid):
                self.assertIn(
                    "prototype-baseline", scenarios,
                    f"{opid}: prototype-baseline scenario must be present after sync",
                )
                # default must remain the first key.
                self.assertEqual(next(iter(scenarios)), "default")
                pb = scenarios["prototype-baseline"]
                response = pb.get("response", {})
                body = response.get("body")
                self.assertIsNotNone(body, f"{opid}: prototype-baseline.response.body required")
                self.assertNotEqual(body, {}, f"{opid}: prototype-baseline body must not be empty")

    def test_sync_is_idempotent(self) -> None:
        first = _run(SYNC, self.repo)
        self.assertEqual(first.returncode, 0)
        snapshots: dict[Path, str] = {}
        fixtures_root = self.repo / "openapi" / "fixtures"
        for tag, opid in P0_BASELINE_OPS:
            p = fixtures_root / tag / f"{opid}.json"
            snapshots[p] = p.read_text(encoding="utf-8")
        second = _run(SYNC, self.repo)
        self.assertEqual(second.returncode, 0)
        for path, before in snapshots.items():
            with self.subTest(path=str(path)):
                self.assertEqual(
                    before, path.read_text(encoding="utf-8"),
                    f"{path}: re-running sync must be idempotent",
                )

    def test_sync_output_passes_validate_fixtures(self) -> None:
        sync_out = _run(SYNC, self.repo)
        self.assertEqual(sync_out.returncode, 0)
        val = _run(VALIDATE, self.repo)
        self.assertEqual(
            val.returncode, 0,
            msg=f"validate-fixtures must pass after sync\nstderr={val.stderr}",
        )

    def test_sync_fails_fast_on_mapping_gap(self) -> None:
        # Drop the experiences section that listExperienceCards depends on.
        data_file = self.repo / "easyinterview-ui" / "src" / "data.jsx"
        text = data_file.read_text(encoding="utf-8")
        # Replace the experiences array with an empty list.
        replaced = text.replace("experiences: [", "experiences__missing: [", 1)
        self.assertNotEqual(text, replaced, "Test setup: data.jsx must contain `experiences: [`")
        data_file.write_text(replaced, encoding="utf-8")
        out = _run(SYNC, self.repo)
        self.assertNotEqual(out.returncode, 0)
        self.assertIn("Mapping gap", out.stderr + out.stdout)
        self.assertIn("listExperienceCards", out.stderr + out.stdout)


if __name__ == "__main__":
    unittest.main()
