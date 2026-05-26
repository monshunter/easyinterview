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


def _hint_schema() -> dict:
    return {
        "type": "object",
        "description": "Lightweight real-time interview observation cue.",
        "required": ["cue"],
        "properties": {
            "cue": {"type": "string", "description": "Short cue."},
            "answerSummary": {"type": "string", "description": "Short answer summary."},
            "severity": {
                "type": "string",
                "description": "Optional urgency.",
                "enum": ["info", "nudge", "alert"],
            },
        },
    }


def _body_with_contract(schema: dict) -> str:
    module = _load_module()
    return "Fixture prompt.\nRespond in {{language}}.\n\n" + module.render_output_contract(schema) + "\n"


def _write_baseline_pair(
    tmp_path: pathlib.Path,
    feature_key: str,
    body: str | None = None,
    hash_value: str | None = None,
    schema: dict | None = None,
):
    """Create one valid prompt yaml/md pair under tmp_path."""
    module = _load_module()
    if schema is None and feature_key not in module.OUTPUT_SCHEMA_EXEMPT_FEATURE_KEYS:
        schema = _hint_schema()
    if body is None:
        body = _body_with_contract(schema) if schema is not None else "voice fixture body\n"
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
    if schema is not None:
        (feature_dir / "v0.1.0.schema.json").write_text(
            json.dumps(schema, indent=2) + "\n",
            encoding="utf-8",
        )

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
    feature_dir = _write_baseline_pair(tmp_path, "practice.turn.lightweight_observe")
    # Mutate the body but leave the yaml hash unchanged.
    with (feature_dir / "v0.1.0.md").open("a", encoding="utf-8") as f:
        f.write("mutated body\n")

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "template_hash drift" in result.stderr


def test_field_order_negative(tmp_path):
    """Reordering top-level fields must fail lint."""
    feature_key = "practice.turn.lightweight_observe"
    schema = _hint_schema()
    body = _body_with_contract(schema)
    _write_baseline_pair(tmp_path, feature_key, body, schema=schema)

    # Overwrite yaml with reshuffled field order (status moved before language)
    # but a hash that still matches the canonical algorithm — order check
    # must fail independently of hash drift.
    module = _load_module()
    meta = {
        "feature_key": feature_key,
        "version": "v0.1.0",
        "language": "multi",
        "status": "active",
        "created_at": "2026-05-09T12:00:00Z",
    }
    correct_hash = module.expected_hash(body.encode("utf-8"), meta)

    yaml_text = textwrap.dedent(
        f"""\
        feature_key: "{feature_key}"
        version: "v0.1.0"
        status: "active"
        language: "multi"
        template_hash: "{correct_hash}"
        created_at: "2026-05-09T12:00:00Z"
        """
    )
    (tmp_path / "config" / "prompts" / feature_key / "v0.1.0.yaml").write_text(
        yaml_text, encoding="utf-8"
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "field order" in result.stderr


def test_language_override_without_allowlist_negative(tmp_path):
    """Baseline storage is canonical multi; duplicate language variants need rationale."""
    feature_key = "practice.turn.lightweight_observe"
    schema = _hint_schema()
    body = _body_with_contract(schema)
    feature_dir = _write_baseline_pair(tmp_path, feature_key, body, schema=schema)

    module = _load_module()
    override_meta = {
        "feature_key": feature_key,
        "version": "v0.1.0",
        "language": "en",
        "status": "active",
        "created_at": "2026-05-09T12:00:00Z",
    }
    override_hash = module.expected_hash(body.encode("utf-8"), override_meta)
    (feature_dir / "v0.1.0.en.md").write_text(body, encoding="utf-8")
    (feature_dir / "v0.1.0.en.yaml").write_text(
        textwrap.dedent(
            f"""\
            feature_key: "{feature_key}"
            version: "v0.1.0"
            language: "en"
            template_hash: "{override_hash}"
            status: "active"
            created_at: "2026-05-09T12:00:00Z"
            """
        ),
        encoding="utf-8",
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "language override" in result.stderr
    assert "not allowlisted" in result.stderr


def test_multi_prompt_without_runtime_language_instruction_negative(tmp_path):
    feature_key = "practice.turn.lightweight_observe"
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("Respond in {{language}}.\n\n", "")
    _write_baseline_pair(tmp_path, feature_key, body, schema=schema)

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "runtime language instruction" in result.stderr


def test_output_schema_illegal_keyword_negative(tmp_path):
    schema = _hint_schema()
    schema["additionalProperties"] = False
    _write_baseline_pair(
        tmp_path,
        "practice.turn.lightweight_observe",
        _body_with_contract(schema),
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "unsupported schema keys" in result.stderr


def test_output_schema_required_not_in_prompt_negative(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("`$.cue`", "`$.hintText`")
    _write_baseline_pair(tmp_path, "practice.turn.lightweight_observe", body, schema=schema)

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "output contract block drift" in result.stderr


def test_output_schema_struct_mismatch_negative(tmp_path):
    schema = _hint_schema()
    schema["required"] = ["hint"]
    schema["properties"]["hint"] = {"type": "string", "description": "Wrong parser key."}
    _write_baseline_pair(
        tmp_path,
        "practice.turn.lightweight_observe",
        _body_with_contract(schema),
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "required paths missing" in result.stderr
    assert "not parser/struct-owned" in result.stderr


def test_prompt_contract_block_drift_negative(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("Short cue.", "Mutated cue text.")
    _write_baseline_pair(tmp_path, "practice.turn.lightweight_observe", body, schema=schema)

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "output contract block drift" in result.stderr


def test_rendered_example_is_schema_valid_for_nested_array_enum():
    module = _load_module()
    schema = {
        "type": "array",
        "description": "Array root.",
        "items": {
            "type": "object",
            "description": "Item.",
            "required": ["level", "fit"],
            "properties": {
                "level": {
                    "type": "string",
                    "description": "Seniority.",
                    "enum": ["junior", "senior"],
                },
                "fit": {
                    "type": "object",
                    "description": "Fit tuple.",
                    "required": ["must"],
                    "properties": {
                        "must": {"type": "integer", "description": "Must-have count."}
                    },
                },
            },
        },
    }
    block = module.render_output_contract(schema)
    assert "`$[]` (required, object)" in block
    assert "`$[].level` (required, string enum(junior, senior))" in block
    errors: list[str] = []
    module.validate_value_against_schema(module.example_for_schema(schema), schema, "$", errors)
    assert errors == []


def test_rendered_example_includes_optional_properties_and_representative_values():
    module = _load_module()
    schema = {
        "type": "object",
        "description": "First mock-interview question.",
        "required": ["questionText", "questionIntent"],
        "properties": {
            "questionText": {
                "type": "string",
                "description": "Question text shown to the candidate.",
            },
            "questionIntent": {
                "type": "string",
                "description": "Short intent label for why this question is asked.",
            },
            "focusDimension": {
                "type": "string",
                "description": "Optional rubric dimension the question is designed to probe.",
            },
            "expectedSignals": {
                "type": "array",
                "description": "Optional expected answer signals for later evaluator context.",
                "items": {
                    "type": "string",
                    "description": "One expected signal.",
                },
            },
            "timeBudgetSeconds": {
                "type": "integer",
                "description": "Optional suggested answer time budget in seconds.",
            },
        },
    }
    example = module.example_for_schema(schema)

    assert set(example) == {
        "questionText",
        "questionIntent",
        "focusDimension",
        "expectedSignals",
        "timeBudgetSeconds",
    }
    assert example["questionText"] != "string"
    assert example["timeBudgetSeconds"] != 1
    errors: list[str] = []
    module.validate_value_against_schema(example, schema, "$", errors)
    assert errors == []


def test_schema_description_required_negative():
    module = _load_module()
    schema = {"type": "object", "required": [], "properties": {}}
    errors = module.validate_schema_subset(pathlib.Path("fixture.schema.json"), schema)
    assert "missing non-empty description" in "\n".join(errors)


def test_missing_schema_description_reports_lint_error_without_traceback(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema)
    del schema["description"]
    _write_baseline_pair(
        tmp_path,
        "practice.turn.lightweight_observe",
        body=body,
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "missing non-empty description" in result.stderr
    assert "Traceback" not in result.stderr


def test_jd_match_recommendation_posted_is_optional_contract():
    schema_path = REPO_ROOT / "config/prompts/jd_match.recommendation/v0.1.0.schema.json"
    schema = json.loads(schema_path.read_text(encoding="utf-8"))

    assert "posted" not in schema["items"].get("required", [])


if __name__ == "__main__":
    import unittest

    unittest.main(verbosity=2)
