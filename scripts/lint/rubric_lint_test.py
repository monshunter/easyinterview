#!/usr/bin/env python3
"""Unit tests for scripts/lint/rubric_lint.py."""
from __future__ import annotations

import pathlib
import subprocess
import sys

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
        'status: "active"\n'
        "dimensions:\n"
    )
    body = header + dimensions_yaml
    if not body.endswith("\n"):
        body += "\n"
    yaml_path = feature_dir / "v0.1.0.yaml"
    yaml_path.write_text(body, encoding="utf-8")
    return yaml_path


def test_baseline_passes():
    """Linting `config/rubrics/` against the baseline files must succeed."""
    result = _run(REPO_ROOT / "config/rubrics")
    assert result.returncode == 0, f"stdout={result.stdout!r} stderr={result.stderr!r}"


def test_unknown_status_is_rejected(tmp_path):
    dimensions = (
        '  - name: "report_evidence"\n'
        '    weight: 1.0\n'
        '    description: "Grounding quality."\n'
        '    score_levels:\n'
        '      - label: "weak"\n'
        '        threshold: 0.0\n'
        '        description: "Unsupported."\n'
        '      - label: "proficient"\n'
        '        threshold: 0.7\n'
        '        description: "Supported."\n'
        '      - label: "strong"\n'
        '        threshold: 0.9\n'
        '        description: "Fully supported."\n'
    )
    path = _write_baseline(tmp_path, "report.generate", dimensions)
    path.write_text(path.read_text(encoding="utf-8").replace('status: "active"', 'status: "retired"'), encoding="utf-8")

    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "status" in result.stderr
    assert "active|inactive" in result.stderr


def test_report_v020_locked_grounding_dimensions_and_weights():
    path = REPO_ROOT / "config/rubrics/report.generate/v0.2.0.yaml"
    import yaml

    rubric = yaml.safe_load(path.read_text(encoding="utf-8"))
    assert rubric["status"] == "active"
    assert {
        dimension["name"]: dimension["weight"] for dimension in rubric["dimensions"]
    } == {
        "report_evidence": 0.35,
        "report_specificity": 0.25,
        "report_action_quality": 0.25,
        "report_calibration": 0.15,
    }
    rendered = str(rubric)
    for term in (
        "candidate user messages",
        "supported",
        "executable",
        "causal",
    ):
        assert term in rendered
    action_quality = next(
        dimension for dimension in rubric["dimensions"]
        if dimension["name"] == "report_action_quality"
    )
    action_contract = " ".join(
        [action_quality["description"]]
        + [level["description"] for level in action_quality["score_levels"]]
    ).lower()
    assert "answer_depth" in action_contract
    assert "answer_relevance" in action_contract
    assert "intentionally generic" in action_contract


def test_practice_chat_v020_rubric_is_content_identical_to_v010():
    import yaml

    v010 = yaml.safe_load(
        (REPO_ROOT / "config/rubrics/practice.session.chat/v0.1.0.yaml").read_text(
            encoding="utf-8"
        )
    )
    v020 = yaml.safe_load(
        (REPO_ROOT / "config/rubrics/practice.session.chat/v0.2.0.yaml").read_text(
            encoding="utf-8"
        )
    )

    assert v010.pop("version") == "v0.1.0"
    assert v020.pop("version") == "v0.2.0"
    assert v010.pop("status") == "inactive"
    assert v020.pop("status") == "active"
    assert v020 == v010


def test_weight_sum_tolerance(tmp_path):
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

def test_dimension_name_allowlist(tmp_path):
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

def test_out_of_scope_jd_match_dimension_names_are_rejected(tmp_path):
    """JD-Match D-12 rubric dimensions are out-of-scope."""
    dimensions = (
        '  - name: "relevance_to_profile"\n'
        '    weight: 0.2\n'
        '    description: "Profile fit"\n'
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
        '  - name: "risk_clarity"\n'
        '    weight: 0.2\n'
        '    description: "Risk clarity"\n'
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
        '  - name: "actionability"\n'
        '    weight: 0.2\n'
        '    description: "Next-step usefulness"\n'
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
        '  - name: "query_alignment"\n'
        '    weight: 0.2\n'
        '    description: "Search query fit"\n'
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
        '  - name: "diversity"\n'
        '    weight: 0.1\n'
        '    description: "Result diversity"\n'
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
        '  - name: "privacy_compliance"\n'
        '    weight: 0.1\n'
        '    description: "Privacy compliance"\n'
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
    _write_baseline(tmp_path, "jd_match.search", dimensions)
    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "relevance_to_profile" in result.stderr
    assert "not in allowlist" in result.stderr


def test_language_override_without_allowlist_negative(tmp_path):
    """Baseline rubrics are language-independent unless an override is explicitly allowlisted."""
    dimensions = (
        '  - name: "language_consistency"\n'
        '    weight: 1.0\n'
        '    description: "Language consistency"\n'
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
    feature_key = "practice.turn.lightweight_observe"
    feature_dir = _write_baseline(tmp_path, feature_key, dimensions).parent
    (feature_dir / "v0.1.0.en.yaml").write_text(
        f'feature_key: "{feature_key}"\n'
        'version: "v0.1.0"\n'
        'language: "en"\n'
        "dimensions:\n"
        + dimensions,
        encoding="utf-8",
    )

    result = _run(tmp_path / "config/rubrics")
    assert result.returncode == 1
    assert "language override" in result.stderr
    assert "not allowlisted" in result.stderr


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
