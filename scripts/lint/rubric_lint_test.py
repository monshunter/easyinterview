#!/usr/bin/env python3
"""Unit tests for scripts/lint/rubric_lint.py."""
from __future__ import annotations

import pathlib
import subprocess
import sys
import textwrap

THIS_DIR = pathlib.Path(__file__).resolve().parent
SCRIPT = THIS_DIR / "rubric_lint.py"
REPO_ROOT = THIS_DIR.parents[1]


def _run(rubrics_dir: pathlib.Path) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, str(SCRIPT), "--rubrics-dir", str(rubrics_dir)],
        capture_output=True,
        text=True,
        check=False,
    )


def _write_baseline(tmp_path: pathlib.Path, feature_key: str, dimensions_yaml: str) -> pathlib.Path:
    """Build a complete rubric yaml with the supplied dimensions block.

    `dimensions_yaml` is the raw text that follows the `dimensions:` key,
    including its leading 2-space indentation. We concatenate manually rather
    than interpolating through an f-string so the indentation does not get
    mangled by textwrap.dedent.
    """
    feature_dir = tmp_path / "config" / "rubrics" / feature_key
    feature_dir.mkdir(parents=True)
    header = (
        f'feature_key: "{feature_key}"\n'
        'version: "v0.1.0"\n'
        'language: "multi"\n'
        "dimensions:\n"
    )
    body = header + dimensions_yaml
    if not body.endswith("\n"):
        body += "\n"
    yaml_path = feature_dir / "v0.1.0.yaml"
    yaml_path.write_text(body, encoding="utf-8")
    return yaml_path


def test_baseline_passes():
    """Linting `config/rubrics/` against the 20 baseline files must succeed."""
    result = _run(REPO_ROOT / "config/rubrics")
    assert result.returncode == 0, f"stdout={result.stdout!r} stderr={result.stderr!r}"


def TestWeightSumTolerance(tmp_path):
    """Plan §1.3 + §1.5 verification: sum(weight) outside +/-0.001 must fail."""
    bad = (
        '  - name: "language_consistency"\n'
        '    weight: 0.5\n'
        '    description: "L"\n'
        '    score_levels:\n'
        '      - label: "weak"\n'
        '        threshold: 0.0\n'
        '        description: "x"\n'
        '      - label: "ok"\n'
        '        threshold: 0.5\n'
        '        description: "y"\n'
        '      - label: "strong"\n'
        '        threshold: 0.9\n'
        '        description: "z"\n'
        '  - name: "report_specificity"\n'
        '    weight: 0.6\n'
        '    description: "R"\n'
        '    score_levels:\n'
        '      - label: "weak"\n'
        '        threshold: 0.0\n'
        '        description: "x"\n'
        '      - label: "ok"\n'
        '        threshold: 0.5\n'
        '        description: "y"\n'
        '      - label: "strong"\n'
        '        threshold: 0.9\n'
        '        description: "z"\n'
    )
    _write_baseline(tmp_path, "weightfail.fixture", bad)
    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "weight sum" in result.stderr


def test_weight_sum_tolerance(tmp_path):
    """pytest alias for TestWeightSumTolerance."""
    TestWeightSumTolerance(tmp_path)


def TestDimensionNameAllowlist(tmp_path):
    """Plan §1.3 + §1.5 verification: unknown dimension name must fail."""
    bad = (
        '  - name: "made_up_metric"\n'
        '    weight: 1.0\n'
        '    description: "Not in allowlist"\n'
        '    score_levels:\n'
        '      - label: "weak"\n'
        '        threshold: 0.0\n'
        '        description: "x"\n'
        '      - label: "ok"\n'
        '        threshold: 0.5\n'
        '        description: "y"\n'
        '      - label: "strong"\n'
        '        threshold: 0.9\n'
        '        description: "z"\n'
    )
    _write_baseline(tmp_path, "namefail.fixture", bad)
    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "not in allowlist" in result.stderr


def test_dimension_name_allowlist(tmp_path):
    """pytest alias for TestDimensionNameAllowlist."""
    TestDimensionNameAllowlist(tmp_path)


def test_missing_weight_negative(tmp_path):
    """Negative fixture: dimension missing weight must fail lint."""
    bad = (
        '  - name: "language_consistency"\n'
        '    description: "no weight"\n'
        '    score_levels:\n'
        '      - label: "weak"\n'
        '        threshold: 0.0\n'
        '        description: "x"\n'
        '      - label: "ok"\n'
        '        threshold: 0.5\n'
        '        description: "y"\n'
        '      - label: "strong"\n'
        '        threshold: 0.9\n'
        '        description: "z"\n'
    )
    _write_baseline(tmp_path, "missingweight.fixture", bad)
    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "weight must be a non-negative number" in result.stderr


if __name__ == "__main__":
    import unittest

    unittest.main(verbosity=2)
