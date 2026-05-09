#!/usr/bin/env python3
"""Unit tests for scripts/lint/prompt_lint.py.

Covers:
- Happy path: linting `config/prompts/` baseline must succeed.
- TestCanonicalHashAgainstReadme: the canonical hash matches the README §3
  description verbatim (cross-tool source of truth).
- Negative fixture (hash drift): body changed without hash bump.
- Negative fixture (field order): reordered yaml fields.
"""
from __future__ import annotations

import hashlib
import importlib.util
import json
import pathlib
import subprocess
import sys
import textwrap

THIS_DIR = pathlib.Path(__file__).resolve().parent
SCRIPT = THIS_DIR / "prompt_lint.py"
REPO_ROOT = THIS_DIR.parents[1]


def _load_module():
    spec = importlib.util.spec_from_file_location("prompt_lint_under_test", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _run(prompts_dir: pathlib.Path, migrations_dir: pathlib.Path) -> subprocess.CompletedProcess:
    return subprocess.run(
        [
            sys.executable,
            str(SCRIPT),
            "--prompts-dir",
            str(prompts_dir),
            "--migrations-dir",
            str(migrations_dir),
        ],
        capture_output=True,
        text=True,
        check=False,
    )


def test_baseline_passes():
    result = _run(REPO_ROOT / "config/prompts", REPO_ROOT / "migrations")
    assert result.returncode == 0, f"stdout={result.stdout!r} stderr={result.stderr!r}"


def TestCanonicalHashAgainstReadme():
    """Plan §1.1 verification gate.

    Recompute the canonical hash by hand exactly as `config/prompts/README.md`
    §3 describes and confirm `prompt_lint.expected_hash` matches.
    """
    module = _load_module()
    body = "Hello {{var}}\n"
    meta = {
        "feature_key": "fixture.canonical",
        "version": "v0.1.0",
        "language": "multi",
        "status": "active",
        "created_at": "2026-05-09T12:00:00Z",
    }
    body_bytes = body.encode("utf-8")
    actual = module.expected_hash(body_bytes, meta)

    canonical = (
        json.dumps(
            {k: v for k, v in meta.items() if k != "template_hash"},
            sort_keys=True,
            ensure_ascii=False,
            separators=(",", ":"),
        )
        + "\n"
    ).encode("utf-8")
    expected = hashlib.sha256(body_bytes + canonical).hexdigest()
    assert actual == expected


def test_canonical_hash_against_readme():
    """pytest-discoverable alias for TestCanonicalHashAgainstReadme."""
    TestCanonicalHashAgainstReadme()


def _write_baseline_pair(tmp_path: pathlib.Path, feature_key: str, body: str, hash_value: str | None = None):
    """Create one valid prompt yaml/md pair under tmp_path."""
    module = _load_module()
    meta = {
        "feature_key": feature_key,
        "version": "v0.1.0",
        "language": "multi",
        "status": "active",
        "created_at": "2026-05-09T12:00:00Z",
    }
    if hash_value is None:
        hash_value = module.expected_hash(body.encode("utf-8"), meta)

    feature_dir = tmp_path / "config" / "prompts" / feature_key
    feature_dir.mkdir(parents=True)
    (feature_dir / "v0.1.0.md").write_text(body, encoding="utf-8")

    yaml_text = textwrap.dedent(
        f"""\
        feature_key: "{feature_key}"
        version: "v0.1.0"
        language: "multi"
        template_hash: "{hash_value}"
        status: "active"
        created_at: "2026-05-09T12:00:00Z"
        """
    )
    (feature_dir / "v0.1.0.yaml").write_text(yaml_text, encoding="utf-8")
    return feature_dir


def test_hash_drift_negative(tmp_path):
    """Editing the body without refreshing template_hash must fail lint."""
    body = "original body\n"
    feature_dir = _write_baseline_pair(tmp_path, "drift.fixture", body)
    # Mutate the body but leave the yaml hash unchanged.
    (feature_dir / "v0.1.0.md").write_text("mutated body\n", encoding="utf-8")

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "template_hash drift" in result.stderr


def test_field_order_negative(tmp_path):
    """Reordering top-level fields must fail lint."""
    body = "ordered body\n"
    _write_baseline_pair(tmp_path, "order.fixture", body)

    # Overwrite yaml with reshuffled field order (status moved before language)
    # but a hash that still matches the canonical algorithm — order check
    # must fail independently of hash drift.
    module = _load_module()
    meta = {
        "feature_key": "order.fixture",
        "version": "v0.1.0",
        "language": "multi",
        "status": "active",
        "created_at": "2026-05-09T12:00:00Z",
    }
    correct_hash = module.expected_hash(body.encode("utf-8"), meta)

    yaml_text = textwrap.dedent(
        f"""\
        feature_key: "order.fixture"
        version: "v0.1.0"
        status: "active"
        language: "multi"
        template_hash: "{correct_hash}"
        created_at: "2026-05-09T12:00:00Z"
        """
    )
    (tmp_path / "config" / "prompts" / "order.fixture" / "v0.1.0.yaml").write_text(
        yaml_text, encoding="utf-8"
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "field order" in result.stderr


if __name__ == "__main__":
    import unittest

    unittest.main(verbosity=2)
