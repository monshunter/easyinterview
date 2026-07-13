#!/usr/bin/env python3
"""Unit tests for scripts/lint/prompt_lint.py.

Covers:
- Happy path: linting `config/prompts/` baseline must succeed.
- Canonical hash parity: the canonical hash matches the README §3
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


def test_report_v020_direct_semantics_and_old_keys_absent():
    schema_path = REPO_ROOT / "config/prompts/report.generate/v0.2.0.schema.json"
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    module = _load_module()
    required = module.collect_required_paths(schema)
    assert required == module.REPORT_V020_REQUIRED_PATHS

    property_names: set[str] = set()

    def collect_names(node: dict) -> None:
        for name, child in (node.get("properties") or {}).items():
            property_names.add(name)
            if isinstance(child, dict):
                collect_names(child)
        item = node.get("items")
        if isinstance(item, dict):
            collect_names(item)

    collect_names(schema)
    for forbidden in (
        "dimension_scores",
        "retry_focus_competency_codes",
        "score",
        "reasoning",
        "supporting_observations",
    ):
        assert forbidden not in property_names

    body = (REPO_ROOT / "config/prompts/report.generate/v0.2.0.md").read_text(encoding="utf-8")
    assert "{{rubric_dimensions}}" not in body
    assert "{{frozen_context}}" in body
    assert "{{conversation_messages}}" in body
    assert "{{language}}" in body
    normalized_body = " ".join(body.lower().split())
    assert "draft the summary only after the evidence items" in normalized_body
    assert "split it into factual clauses" in normalized_body
    assert "delete any clause that cannot be fully mapped" in normalized_body
    assert "do not upgrade a concrete candidate action into an outcome or quality property" in normalized_body
    assert "safe, reliable, resilient, reversible, isolated, effective, or successful" in normalized_body
    assert "set `w = true` if and only if" in normalized_body
    assert "when `w` is true, `preparednesslevel` must be `well_prepared`" in normalized_body
    assert "`basically_ready` is invalid in that case" in normalized_body
    assert "evidence being partial, rehearsed, or merely not covered is not itself a deficiency" in normalized_body
    assert "a brief assertion that only names a mechanism without concrete supporting detail" in normalized_body
    assert "not as evidence of a topic-specific capability gap" in normalized_body
    assert "use the exact dimension code `answer_depth`" in normalized_body
    assert "may state only that the answer provides no concrete supporting detail" in normalized_body
    assert "must not enumerate unmentioned expected details" in normalized_body
    assert "do not use the assistant question or its topic to create a more specific issue or retry focus" in normalized_body
    assert "emit `retry_current_round` with an empty focus array and a generic label" in normalized_body
    assert "the generic label must not repeat the assistant question or name its topic or mechanism" in normalized_body
    assert "a corrective `retry_current_round` label may only turn the cited missing behavior" in normalized_body
    assert "`review_evidence` must only ask the user to revisit cited positive or explicitly evidence-limited content" in normalized_body
    assert "do not invent an artifact, corrective gap, new scenario, or transfer task" in normalized_body
    assert "for every selected focus code, the first retry label must name at least one directly cited missing behavior" in normalized_body
    assert "umbrella labels such as `add a backpressure mechanism`, `add a safety check`, `add detail`, or `improve the answer` are invalid" in normalized_body
    assert "the schema's 200-character bound is only the outer malformed-output safety cap" in normalized_body
    assert "one semicolon-separated cited missing behavior per selected focus code" in normalized_body
    assert "in `en`, each label has 1-24 whitespace-delimited words" in normalized_body
    assert "in `zh-cn`, each label has 1-64 unicode code points" in normalized_body
    assert "do not add an introduction such as `retry the answer by adding`" in normalized_body
    assert "an action label may only turn the cited missing behavior" not in normalized_body
    assert "synthetic paired candidate input for the example below" in normalized_body
    assert "candidate user message seq 2" in normalized_body
    assert "ranked the options by user impact and delivery effort" in normalized_body
    assert "did not explain the tie-breaking rule" in normalized_body
    assert "demonstrate only json format and cross-field coherence" in normalized_body
    assert "never reuse any example fact, dimension, preparedness level, wording, or action" in normalized_body
    assert "regenerate every field from the current frozen context and cited candidate messages" in normalized_body
    assert "example complete json output:" in normalized_body
    assert "empty focus is allowed only for exactly one `answer_depth` issue" in normalized_body
    assert "or exactly one `answer_relevance` issue" in normalized_body
    assert "must equal the ascending unique dimension codes of all issues whose declared dimension status is `needs_work`" in normalized_body
    assert "when retry is present it may be empty" not in normalized_body
    assert "use non-empty focus only when" not in normalized_body
    assert "before returning json, set `i = len(issues)`" in normalized_body
    assert "if `i >= 2`, empty focus is invalid" in normalized_body

    focus_description = schema["properties"]["retryFocusDimensionCodes"]["description"].lower()
    assert "empty only for the exact single-issue answer_depth or answer_relevance generic exceptions" in focus_description
    assert schema["properties"]["nextActions"]["maxItems"] == 2
    action_description = schema["properties"]["nextActions"]["items"]["properties"]["label"]["description"].lower()
    assert "english: at most 24 whitespace-delimited words" in action_description
    assert "zh-cn: at most 64 unicode code points" in action_description
    assert "outer malformed-output safety cap, not a writing target" in action_description
    assert "semicolon-separated cited missing-behavior fragments" in action_description


def test_report_v020_mixed_answer_keeps_substance_grounding_contract():
    body = (REPO_ROOT / "config/prompts/report.generate/v0.2.0.md").read_text(encoding="utf-8")
    normalized_body = " ".join(body.lower().split())

    assert "classify candidate messages in this order" in normalized_body
    assert "first ignore only control fragments" in normalized_body
    assert "any remaining statement of experience, motivation, mechanism, decision, action, constraint, metric, example" in normalized_body
    assert "or an explicit limitation or missing detail is substantive interview content" in normalized_body
    assert "a direct statement such as `i do not know` or `i have no approach` is also substantive candidate content" in normalized_body
    assert "a mixed message with any such content is not control-only" in normalized_body
    assert "using the same message sequence number" in normalized_body
    assert "never use `answer_relevance` or claim that no substantive answer was provided" in normalized_body
    assert "if the remaining answer only names a mechanism without supporting detail" in normalized_body
    assert "apply the `answer_depth` branch in rule 8" in normalized_body
    assert "the issue, focus, and retry action may name only those candidate-stated missing details" in normalized_body
    assert "state that the candidate explicitly said those details were not explained" in normalized_body
    assert "do not characterize the rest of a mixed answer with totalizing qualifiers" in normalized_body
    assert "`only`, `merely`, `nothing`, `no substantive content`, `仅`, `只`, `任何`, or `完全`" in normalized_body
    assert "classification example: `我希望继续做分布式系统，但没有说明项目规模。 请结束本轮。`" in normalized_body
    assert "a supported issue is `候选人明确表示未说明项目规模。`" in normalized_body
    assert "`候选人仅要求结束本轮。`, and `候选人未提供实质性回答。` are invalid" in normalized_body
    assert "only when removing the control fragments leaves no interview content" in normalized_body


def test_report_v020_output_contract_keeps_paired_anti_copy_example():
    module = _load_module()
    body = (REPO_ROOT / "config/prompts/report.generate/v0.2.0.md").read_text(encoding="utf-8")
    schema = json.loads(
        (REPO_ROOT / "config/prompts/report.generate/v0.2.0.schema.json").read_text(encoding="utf-8")
    )

    paired_input = (
        "Synthetic paired candidate input for the example below:\n"
        "- Candidate user message seq 2: \"I ranked the options by user impact and delivery effort. "
        "I did not explain the tie-breaking rule.\""
    )
    anti_copy = (
        "The paired input and output demonstrate only JSON format and cross-field coherence. "
        "Never reuse any example fact, dimension, preparedness level, wording, or action. "
        "Regenerate every field from the current frozen context and cited candidate messages."
    )
    assert paired_input in body
    assert anti_copy in body
    assert "Example complete JSON output:" in body
    assert body.index(paired_input) < body.index("Example complete JSON output:")
    assert module.extract_output_contract_block(body) == module.render_output_contract(schema)

    rendered = module.render_output_contract(schema)
    assert '"summary": "The candidate gave a usable prioritization approach but explicitly said the tie-breaking rule was not explained."' in rendered
    assert '"preparednessLevel": "needs_practice"' in rendered
    assert '"code": "decision_clarity"' in rendered
    assert '"sourceMessageSeqNos": [\n        2\n      ]' in rendered
    assert '"label": "Retry the prioritization answer by explaining the tie-breaking rule"' in rendered


def test_practice_chat_v020_uses_structured_semantic_focus_and_canonical_hash():
    module = _load_module()
    body_path = REPO_ROOT / "config/prompts/practice.session.chat/v0.2.0.md"
    yaml_path = REPO_ROOT / "config/prompts/practice.session.chat/v0.2.0.yaml"
    body = body_path.read_text(encoding="utf-8")
    meta, _ = module._read_yaml_with_order(yaml_path)
    schema_path = REPO_ROOT / "config/prompts/practice.session.chat/v0.2.0.schema.json"
    schema = json.loads(schema_path.read_text(encoding="utf-8"))

    assert module.validate_practice_chat_context(body_path, body) == []
    assert module.validate_practice_chat_schema(schema_path, schema) == []
    assert module.collect_property_paths(schema) == {"$.messageText"}
    assert module.expected_hash(body.encode("utf-8"), meta) == meta["template_hash"]
    assert meta["status"] == "active"


def test_practice_chat_v020_rejects_legacy_focus_context():
    module = _load_module()
    path = pathlib.Path("practice.session.chat/v0.2.0.md")
    body = textwrap.dedent(
        """\
        <untrusted_interview_context_json>
        {"focusCompetencies": {{focus_competencies_json}}}
        </untrusted_interview_context_json>
        """
    )

    errors = "\n".join(module.validate_practice_chat_context(path, body))
    assert "semanticFocus" in errors
    assert "legacy focus token" in errors


def test_canonical_hash_against_readme():
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

def _hint_schema() -> dict:
    return {
        "type": "object",
        "description": "Practice conversation assistant response.",
        "required": ["messageText"],
        "properties": {
            "messageText": {"type": "string", "description": "Assistant reply text."},
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
    feature_dir = _write_baseline_pair(tmp_path, "practice.session.chat")
    # Mutate the body but leave the yaml hash unchanged.
    with (feature_dir / "v0.1.0.md").open("a", encoding="utf-8") as f:
        f.write("mutated body\n")

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "template_hash drift" in result.stderr


def test_field_order_negative(tmp_path):
    """Reordering top-level fields must fail lint."""
    feature_key = "practice.session.chat"
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
    feature_key = "practice.session.chat"
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
    feature_key = "practice.session.chat"
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("Respond in {{language}}.\n\n", "")
    _write_baseline_pair(tmp_path, feature_key, body, schema=schema)

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "runtime language instruction" in result.stderr


def test_output_schema_illegal_keyword_negative(tmp_path):
    schema = _hint_schema()
    schema["$ref"] = "#/forbidden"
    _write_baseline_pair(
        tmp_path,
        "practice.session.chat",
        _body_with_contract(schema),
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "unsupported schema keys" in result.stderr


def test_grounded_report_schema_requires_recursive_closure_and_bounds():
    module = _load_module()
    path = REPO_ROOT / "config/prompts/report.generate/v0.2.0.schema.json"
    schema = json.loads(path.read_text(encoding="utf-8"))
    assert module.validate_grounded_report_schema(path, schema) == []

    open_schema = json.loads(json.dumps(schema))
    del open_schema["properties"]["dimensionAssessments"]["items"]["additionalProperties"]
    assert "additionalProperties=false" in "\n".join(
        module.validate_grounded_report_schema(path, open_schema)
    )

    unbounded_schema = json.loads(json.dumps(schema))
    del unbounded_schema["properties"]["summary"]["maxLength"]
    assert "summary" in "\n".join(module.validate_grounded_report_schema(path, unbounded_schema))


def test_output_schema_required_not_in_prompt_negative(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("`$.messageText`", "`$.hintText`")
    _write_baseline_pair(tmp_path, "practice.session.chat", body, schema=schema)

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "output contract block drift" in result.stderr


def test_output_schema_struct_mismatch_negative(tmp_path):
    schema = _hint_schema()
    schema["required"] = ["hint"]
    schema["properties"]["hint"] = {"type": "string", "description": "Wrong parser key."}
    _write_baseline_pair(
        tmp_path,
        "practice.session.chat",
        _body_with_contract(schema),
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "required paths missing" in result.stderr
    assert "not parser/struct-owned" in result.stderr


def test_prompt_contract_block_drift_negative(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema).replace("Assistant reply text.", "Mutated reply text.")
    _write_baseline_pair(tmp_path, "practice.session.chat", body, schema=schema)

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


def test_rendered_report_example_uses_a_non_blocking_coherent_calibration():
    module = _load_module()
    schema = {
        "type": "object",
        "description": "Grounded report.",
        "required": ["summary", "preparednessLevel"],
        "properties": {
            "summary": {"type": "string", "description": "Summary."},
            "preparednessLevel": {
                "type": "string",
                "description": "Readiness.",
                "enum": ["needs_practice"],
            },
        },
    }

    example = module.example_for_schema(schema)
    assert example["summary"] == (
        "The candidate gave a usable prioritization approach but explicitly said "
        "the tie-breaking rule was not explained."
    )
    assert "queue backpressure" not in json.dumps(example)
    assert "rollback verification" not in json.dumps(example)
    assert "incident" not in json.dumps(example).lower()


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


def test_numeric_schema_bounds_are_supported_and_enforced():
    module = _load_module()
    schema = {
        "type": "number",
        "description": "Candidate score from 1.0 to 5.0.",
        "minimum": 1.0,
        "maximum": 5.0,
    }
    assert module.validate_schema_subset(pathlib.Path("fixture.schema.json"), schema) == []

    errors: list[str] = []
    module.validate_value_against_schema(0.9, schema, "$", errors)
    assert "must be >= 1.0" in "\n".join(errors)


def test_missing_schema_description_reports_lint_error_without_traceback(tmp_path):
    schema = _hint_schema()
    body = _body_with_contract(schema)
    del schema["description"]
    _write_baseline_pair(
        tmp_path,
        "practice.session.chat",
        body=body,
        schema=schema,
    )

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")
    assert result.returncode == 1
    assert "missing non-empty description" in result.stderr
    assert "Traceback" not in result.stderr


def test_out_of_scope_jd_match_prompt_contract_is_absent():
    assert not (REPO_ROOT / "config/prompts/jd_match.recommendation").exists()
    assert not (REPO_ROOT / "config/prompts/jd_match.search").exists()


def test_out_of_scope_jd_match_prompt_key_is_rejected(tmp_path):
    _write_baseline_pair(tmp_path, "jd_match.search")

    result = _run(tmp_path / "config/prompts", tmp_path / "migrations")

    assert result.returncode == 1
    assert "feature_key 'jd_match.search' is out-of-scope" in result.stderr


if __name__ == "__main__":
    import unittest

    unittest.main(verbosity=2)
