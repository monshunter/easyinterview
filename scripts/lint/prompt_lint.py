#!/usr/bin/env python3
"""scripts/lint/prompt_lint.py - F3 prompt registry linter.

Validates `config/prompts/<feature_key>/<version>[.<language>].{yaml,md}`
against the schema and canonical hash algorithm fixed by
`config/prompts/README.md` and `docs/spec/prompt-rubric-registry/spec.md` v2.9.

The canonical algorithm is shared with the Go loader at
`backend/internal/ai/registry/loader.go`. Both implementations must agree
byte-for-byte; the algorithm description in `config/prompts/README.md` §3 is
the cross-tool source of truth.

Run: `python3 scripts/lint/prompt_lint.py [--prompts-dir DIR] [--migrations-dir DIR]`
Exit: 0 on success, 1 on any violation.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import re
import sys
from typing import Iterable

import yaml

REQUIRED_FIELD_ORDER = [
    "feature_key",
    "version",
    "language",
    "template_hash",
    "status",
    "created_at",
]
STATUS_ENUM = {"draft", "active"}
SEMVER_RE = re.compile(r"^v\d+\.\d+\.\d+(-[A-Za-z0-9\.-]+)?(\+[A-Za-z0-9\.-]+)?$")
LANGUAGE_RE = re.compile(r"^multi$|^[a-z]{2,3}$")
LANGUAGE_OVERRIDE_ALLOWLIST: set[tuple[str, str, str]] = set()

# Forbidden tokens reserved for the lint gate. They live here, not in the
# README, so `grep -rE "\bTBD\b|placeholder" config/prompts/` stays clean
# while the lint script still rejects them inside Markdown bodies.
FORBIDDEN_BODY_TOKEN_RE = re.compile(r"\bTBD\b|\bplaceholder\b", re.IGNORECASE)
NON_CURRENT_MODULE_RE = re.compile(r"\bmistakes\b|\bgrowth\b|\bdrill\b|mistake\.extract")
NON_CURRENT_FEATURE_KEY_PREFIXES = ("jd_match.",)

SCHEMA_ALLOWED_KEYS = {"type", "required", "properties", "items", "enum", "description"}
SCHEMA_ALLOWED_TYPES = {"object", "array", "string", "number", "integer", "boolean", "null"}
OUTPUT_CONTRACT_START = "<!-- output-schema-contract:start -->"
OUTPUT_CONTRACT_END = "<!-- output-schema-contract:end -->"

# Voice feature keys do not produce JSON chat content. They are not present in
# config/prompts today, but keeping the explicit exemption here prevents future
# STT/TTS prompt metadata from being forced into output_schema.
OUTPUT_SCHEMA_EXEMPT_FEATURE_KEYS = {
    "practice.voice.stt",
    "practice.voice.tts",
    "practice.dictation.stt",
}

FEATURE_CONTRACTS = {
    "target.import.parse": {
        "type": "object",
        "required_paths": {
            "$.title",
            "$.companyName",
            "$.coreThemes",
            "$.interviewRounds",
            "$.interviewRounds[].sequence",
            "$.interviewRounds[].type",
            "$.interviewRounds[].name",
            "$.interviewRounds[].durationMinutes",
            "$.interviewRounds[].focus",
            "$.strengths",
            "$.gaps",
            "$.riskSignals",
            "$.requirements",
            "$.requirements[].kind",
            "$.requirements[].label",
        },
    },
    "practice.session.first_question": {
        "type": "object",
        "required_paths": {"$.questionText", "$.questionIntent"},
    },
    "practice.session.follow_up": {
        "type": "object",
        "required_paths": {"$.questionText", "$.questionIntent"},
    },
    "practice.turn.lightweight_observe": {
        "type": "object",
        "required_paths": {"$.cue"},
    },
    "report.generate": {
        "type": "object",
        "required_paths": {
            "$.summary",
            "$.dimension_scores",
            "$.dimension_scores[].name",
            "$.dimension_scores[].score",
            "$.dimension_scores[].reasoning",
            "$.dimension_scores[].supporting_observations",
            "$.highlights",
            "$.highlights[].dimension",
            "$.highlights[].evidence",
            "$.highlights[].confidence",
            "$.issues",
            "$.issues[].dimension",
            "$.issues[].evidence",
            "$.issues[].confidence",
            "$.next_actions",
            "$.next_actions[].type",
            "$.next_actions[].label",
            "$.retry_focus_turn_ids",
        },
    },
    "report.question_assessment": {
        "type": "object",
        "required_paths": {
            "$.dimension_results",
            "$.overall_status",
            "$.confidence",
            "$.strengths",
            "$.gaps",
            "$.recommended_framework",
            "$.review_status",
        },
    },
    "resume.parse": {
        "type": "object",
        "required_paths": {
            "$.displayName",
            "$.markdownText",
            "$.basics",
            "$.experiences",
            "$.projects",
            "$.education",
            "$.skills",
            "$.languages",
        },
    },
    "resume.tailor.gap_review": {
        "type": "object",
        "required_paths": {"$.matchSummary", "$.matchSummary.strengths", "$.matchSummary.gaps"},
    },
    "resume.tailor.bullet_suggestions": {
        "type": "object",
        "required_paths": {
            "$.suggestions",
            "$.suggestions[].originalBullet",
            "$.suggestions[].suggestedBullet",
            "$.suggestions[].reason",
        },
    },
}


def canonical_meta_json(meta: dict) -> bytes:
    meta_for_hash = {k: v for k, v in meta.items() if k != "template_hash"}
    return (
        json.dumps(
            meta_for_hash,
            sort_keys=True,
            ensure_ascii=False,
            separators=(",", ":"),
        )
        + "\n"
    ).encode("utf-8")


def expected_hash(body_bytes: bytes, meta: dict) -> str:
    return hashlib.sha256(body_bytes + canonical_meta_json(meta)).hexdigest()


def _filename_language(yaml_path: pathlib.Path) -> str:
    """Extract the language tag from a yaml filename per README §1."""
    # filename forms: v0.1.0.yaml or v0.1.0.<language>.yaml
    parts = yaml_path.name.split(".")
    # ['v0', '1', '0', 'yaml'] -> language 'multi'
    # ['v0', '1', '0', 'en', 'yaml'] -> language 'en'
    if len(parts) == 4:
        return "multi"
    if len(parts) == 5:
        return parts[3]
    return ""


def _read_yaml_with_order(path: pathlib.Path) -> tuple[dict, list[str]]:
    text = path.read_text(encoding="utf-8")
    parsed = yaml.safe_load(text)
    keys: list[str] = []
    for line in text.splitlines():
        if not line or line.startswith(" ") or line.startswith("#"):
            continue
        if ":" not in line:
            continue
        key = line.split(":", 1)[0].strip()
        if key:
            keys.append(key)
    return parsed, keys


def lint_prompt_yaml(yaml_path: pathlib.Path) -> list[str]:
    errors: list[str] = []
    parsed, key_order = _read_yaml_with_order(yaml_path)
    if not isinstance(parsed, dict):
        return [f"{yaml_path}: not a YAML mapping"]

    if key_order != REQUIRED_FIELD_ORDER:
        errors.append(
            f"{yaml_path}: field order {key_order} does not match required {REQUIRED_FIELD_ORDER}"
        )

    feature_key = parsed.get("feature_key")
    if feature_key != yaml_path.parent.name:
        errors.append(
            f"{yaml_path}: feature_key '{feature_key}' does not match parent dir '{yaml_path.parent.name}'"
        )
    if any(str(feature_key).startswith(prefix) for prefix in NON_CURRENT_FEATURE_KEY_PREFIXES):
        errors.append(f"{yaml_path}: feature_key '{feature_key}' is non-current")

    version = parsed.get("version")
    if not isinstance(version, str) or not SEMVER_RE.match(version):
        errors.append(f"{yaml_path}: version '{version}' is not a valid SemVer literal")

    language = parsed.get("language")
    if not isinstance(language, str) or not LANGUAGE_RE.match(language):
        errors.append(f"{yaml_path}: language '{language}' violates language rule")
    else:
        filename_lang = _filename_language(yaml_path)
        if filename_lang and language != filename_lang:
            errors.append(
                f"{yaml_path}: yaml language '{language}' does not match filename '{filename_lang}'"
            )

    status = parsed.get("status")
    if status not in STATUS_ENUM:
        errors.append(f"{yaml_path}: status '{status}' not in {sorted(STATUS_ENUM)}")

    md_path = yaml_path.with_suffix(".md")
    if not md_path.exists():
        errors.append(f"{yaml_path}: matching markdown body {md_path.name} missing")
        return errors

    body_bytes = md_path.read_bytes()
    body_text = body_bytes.decode("utf-8")

    if FORBIDDEN_BODY_TOKEN_RE.search(body_text):
        errors.append(f"{md_path}: body contains forbidden stub marker (TBD/placeholder)")
    if NON_CURRENT_MODULE_RE.search(body_text):
        errors.append(f"{md_path}: body contains non-current module name")

    actual_hash = expected_hash(body_bytes, parsed)
    if parsed.get("template_hash") != actual_hash:
        errors.append(
            f"{yaml_path}: template_hash drift "
            f"(yaml={parsed.get('template_hash')!r}, computed={actual_hash!r})"
        )

    return errors


def lint_prompts_directory(root: pathlib.Path) -> list[str]:
    if not root.exists():
        return [f"{root}: prompts directory missing"]
    errors: list[str] = []
    yaml_paths = sorted(p for p in root.rglob("*.yaml") if p.is_file())
    for yp in yaml_paths:
        errors.extend(lint_prompt_yaml(yp))
    errors.extend(lint_language_coordinates(root))
    return errors


def lint_seed_migration(prompts_root: pathlib.Path, migrations_root: pathlib.Path) -> list[str]:
    """Phase 4.6 enhancement: assert seed migration template_hash matches yaml hash.

    The check is enabled only when a seed migration named
    `*seed_baseline_prompt_rubric_versions*.up.sql` is present. Phase 4.4 lands
    that migration; until then this function is a no-op.
    """
    errors: list[str] = []
    if not migrations_root.exists():
        return errors

    seed_re = re.compile(
        r"INSERT\s+INTO\s+prompt_versions\s*\(([^)]+)\)\s*VALUES",
        re.IGNORECASE,
    )
    row_re = re.compile(
        r"\(\s*'[^']+'\s*,\s*'(?P<feature_key>[^']+)'\s*,\s*'(?P<version>[^']+)'\s*,"
        r"\s*'(?P<language>[^']+)'\s*,\s*'(?P<template_hash>[a-fA-F0-9]+)'",
    )

    yaml_index: dict[tuple[str, str, str], str] = {}
    for yp in prompts_root.rglob("*.yaml"):
        try:
            parsed = yaml.safe_load(yp.read_text(encoding="utf-8"))
        except Exception:
            continue
        if not isinstance(parsed, dict):
            continue
        key = (
            str(parsed.get("feature_key", "")),
            str(parsed.get("version", "")),
            str(parsed.get("language", "")),
        )
        yaml_index[key] = str(parsed.get("template_hash", ""))

    # Later module pruning migrations (e.g. product-scope v2.1 D-17 dropping
    # the jd_match feature keys) delete previously seeded rows; the net DB
    # state, not the raw seed insert, must match the prompts dir.
    non_current: set[str] = set()
    delete_re = re.compile(
        r"DELETE\s+FROM\s+(?:prompt|rubric)_versions\s+WHERE\s+feature_key\s+IN\s*\(([^)]+)\)",
        re.IGNORECASE,
    )
    for sql_path in sorted(migrations_root.glob("*drop*_module.up.sql")):
        text = sql_path.read_text(encoding="utf-8")
        for dm in delete_re.finditer(text):
            non_current.update(re.findall(r"'([^']+)'", dm.group(1)))

    for sql_path in sorted(migrations_root.glob("*seed_baseline_prompt_rubric*.up.sql")):
        text = sql_path.read_text(encoding="utf-8")
        if not seed_re.search(text):
            continue
        for m in row_re.finditer(text):
            key = (m.group("feature_key"), m.group("version"), m.group("language"))
            if m.group("feature_key") in non_current:
                continue
            sql_hash = m.group("template_hash")
            yaml_hash = yaml_index.get(key)
            if yaml_hash is None:
                errors.append(
                    f"{sql_path}: seed row {key} has no matching yaml under prompts dir"
                )
                continue
            if sql_hash != yaml_hash:
                errors.append(
                    f"{sql_path}: seed row {key} template_hash drift "
                    f"(sql={sql_hash}, yaml={yaml_hash})"
                )
    return errors


def _schema_path(prompts_root: pathlib.Path, feature_key: str, version: str) -> pathlib.Path:
    return prompts_root / feature_key / f"{version}.schema.json"


def _load_schema(path: pathlib.Path) -> tuple[dict | None, list[str]]:
    try:
        parsed = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return None, [f"{path}: parse output schema: {exc}"]
    if not isinstance(parsed, dict):
        return None, [f"{path}: output schema must be a JSON object"]
    return parsed, []


def _collect_prompt_metas(root: pathlib.Path) -> list[tuple[pathlib.Path, dict]]:
    metas: list[tuple[pathlib.Path, dict]] = []
    for yp in sorted(p for p in root.rglob("*.yaml") if p.is_file()):
        try:
            parsed = yaml.safe_load(yp.read_text(encoding="utf-8"))
        except Exception:
            continue
        if isinstance(parsed, dict):
            metas.append((yp, parsed))
    return metas


def lint_language_coordinates(root: pathlib.Path) -> list[str]:
    errors: list[str] = []
    metas = _collect_prompt_metas(root)
    by_feature_version: dict[tuple[str, str], list[tuple[pathlib.Path, str]]] = {}
    for yaml_path, meta in metas:
        feature_key = meta.get("feature_key")
        version = meta.get("version")
        language = meta.get("language")
        if not all(isinstance(v, str) and v for v in (feature_key, version, language)):
            continue
        by_feature_version.setdefault((feature_key, version), []).append((yaml_path, language))

        if language == "multi" and feature_key not in OUTPUT_SCHEMA_EXEMPT_FEATURE_KEYS:
            body_path = yaml_path.with_suffix(".md")
            if body_path.exists():
                body = body_path.read_text(encoding="utf-8")
                if "{{language}}" not in body:
                    errors.append(
                        f"{body_path}: multi prompt missing runtime language instruction "
                        "('{{language}}')"
                    )
            continue

        if language != "multi" and (
            feature_key,
            version,
            language,
        ) not in LANGUAGE_OVERRIDE_ALLOWLIST:
            errors.append(
                f"{yaml_path}: language override ({feature_key}, {version}, {language}) "
                "not allowlisted; baseline storage must use canonical multi"
            )

    for (feature_key, version), entries in sorted(by_feature_version.items()):
        languages = {language for _, language in entries}
        if "multi" not in languages:
            first_path = entries[0][0]
            errors.append(f"{first_path}: feature/version {feature_key} {version} missing multi prompt")

    return errors


def lint_output_schemas(prompts_root: pathlib.Path) -> list[str]:
    errors: list[str] = []
    metas = _collect_prompt_metas(prompts_root)
    feature_versions = sorted(
        {
            (str(meta.get("feature_key", "")), str(meta.get("version", "")))
            for _, meta in metas
            if meta.get("feature_key") and meta.get("version")
        }
    )

    for schema_path in sorted(prompts_root.rglob("*.schema.json")):
        version = schema_path.name[: -len(".schema.json")]
        if not SEMVER_RE.match(version):
            errors.append(f"{schema_path}: schema filename must be <version>.schema.json")

    for feature_key, version in feature_versions:
        if feature_key in OUTPUT_SCHEMA_EXEMPT_FEATURE_KEYS:
            continue
        contract = FEATURE_CONTRACTS.get(feature_key)
        if contract is None:
            errors.append(f"{feature_key}: missing prompt_lint FEATURE_CONTRACTS entry")
            continue

        schema_path = _schema_path(prompts_root, feature_key, version)
        if not schema_path.exists():
            errors.append(f"{schema_path}: output schema missing for chat feature_key {feature_key}")
            continue
        schema, schema_errors = _load_schema(schema_path)
        errors.extend(schema_errors)
        if schema is None:
            continue
        schema_errors = validate_schema_subset(schema_path, schema)
        schema_errors.extend(validate_schema_contract(schema_path, schema, contract))
        errors.extend(schema_errors)
        if schema_errors:
            continue

        expected_block = render_output_contract(schema)
        example = example_for_schema(schema)
        example_errors: list[str] = []
        validate_value_against_schema(example, schema, "$", example_errors)
        errors.extend(f"{schema_path}: rendered example invalid: {e}" for e in example_errors)

        for yaml_path, meta in metas:
            if meta.get("feature_key") != feature_key or meta.get("version") != version:
                continue
            body_path = yaml_path.with_suffix(".md")
            if not body_path.exists():
                continue
            body = body_path.read_text(encoding="utf-8")
            block = extract_output_contract_block(body)
            if block is None:
                errors.append(f"{body_path}: missing schema-rendered output contract block")
                continue
            if block != expected_block:
                errors.append(f"{body_path}: output contract block drift from {schema_path}")

    return errors


def validate_schema_subset(schema_path: pathlib.Path, schema: dict) -> list[str]:
    errors: list[str] = []

    def walk(node: dict, path: str) -> None:
        extra = sorted(set(node) - SCHEMA_ALLOWED_KEYS)
        if extra:
            errors.append(f"{schema_path}: {path} uses unsupported schema keys {extra}")
        desc = node.get("description")
        if not isinstance(desc, str) or not desc.strip():
            errors.append(f"{schema_path}: {path} missing non-empty description")
        schema_type = node.get("type")
        if schema_type not in SCHEMA_ALLOWED_TYPES:
            errors.append(f"{schema_path}: {path}.type {schema_type!r} is not allowed")
        required = node.get("required", [])
        if required is not None:
            if not isinstance(required, list) or any(not isinstance(k, str) for k in required):
                errors.append(f"{schema_path}: {path}.required must be a string list")
            props = node.get("properties")
            if required and not isinstance(props, dict):
                errors.append(f"{schema_path}: {path}.required present without properties")
            if isinstance(props, dict):
                missing = sorted(set(required) - set(props))
                if missing:
                    errors.append(f"{schema_path}: {path}.required keys missing from properties: {missing}")
        enum = node.get("enum")
        if enum is not None and (not isinstance(enum, list) or not enum):
            errors.append(f"{schema_path}: {path}.enum must be a non-empty list")
        props = node.get("properties")
        if props is not None:
            if not isinstance(props, dict):
                errors.append(f"{schema_path}: {path}.properties must be an object")
            else:
                for key, child in props.items():
                    if not isinstance(child, dict):
                        errors.append(f"{schema_path}: {path}.properties.{key} must be an object")
                        continue
                    walk(child, f"{path}.properties.{key}")
        items = node.get("items")
        if items is not None:
            if not isinstance(items, dict):
                errors.append(f"{schema_path}: {path}.items must be an object")
            else:
                walk(items, f"{path}.items")

    walk(schema, "$")
    return errors


def validate_schema_contract(schema_path: pathlib.Path, schema: dict, contract: dict) -> list[str]:
    errors: list[str] = []
    expected_type = contract["type"]
    if schema.get("type") != expected_type:
        errors.append(f"{schema_path}: top-level type {schema.get('type')!r}, want {expected_type!r}")
    actual_required = collect_required_paths(schema)
    expected_required = set(contract["required_paths"])
    missing = sorted(expected_required - actual_required)
    extra = sorted(actual_required - expected_required)
    if missing:
        errors.append(f"{schema_path}: required paths missing from schema: {missing}")
    if extra:
        errors.append(f"{schema_path}: required paths are not parser/struct-owned: {extra}")
    return errors


def collect_required_paths(schema: dict, path: str = "$") -> set[str]:
    out: set[str] = set()
    schema_type = schema.get("type")
    if schema_type == "array" and isinstance(schema.get("items"), dict):
        out.update(collect_required_paths(schema["items"], path + "[]"))
    props = schema.get("properties")
    required = schema.get("required") or []
    if isinstance(props, dict):
        for key in required:
            out.add(f"{path}.{key}")
        for key, child in props.items():
            if isinstance(child, dict):
                out.update(collect_required_paths(child, f"{path}.{key}"))
    return out


def schema_type_label(schema: dict) -> str:
    t = str(schema.get("type", "value"))
    enum = schema.get("enum")
    if isinstance(enum, list) and enum:
        return f"{t} enum({', '.join(str(v) for v in enum)})"
    return t


def ordered_schema_property_keys(schema: dict) -> list[str]:
    props = schema.get("properties") or {}
    required_keys = [k for k in schema.get("required") or [] if k in props]
    required_set = set(required_keys)
    return required_keys + [k for k in props if k not in required_set]


def render_output_contract(schema: dict) -> str:
    lines = [
        OUTPUT_CONTRACT_START,
        "Return strict JSON matching this schema-derived output contract.",
        "Produce a complete JSON value, not JSON Schema or an OpenAPI schema.",
        "",
        "Output shape:",
    ]

    def walk(node: dict, path: str, required: bool) -> None:
        marker = "required" if required else "optional"
        lines.append(f"- `{path}` ({marker}, {schema_type_label(node)}): {node['description'].strip()}")
        if node.get("type") == "array" and isinstance(node.get("items"), dict):
            walk(node["items"], path + "[]", True)
            return
        props = node.get("properties") or {}
        required_keys = set(node.get("required") or [])
        for key in ordered_schema_property_keys(node):
            child = props.get(key)
            if isinstance(child, dict):
                walk(child, f"{path}.{key}", key in required_keys)

    walk(schema, "$", True)
    example = example_for_schema(schema)
    example_text = json.dumps(example, ensure_ascii=False, indent=2)
    lines.extend(
        [
            "",
            "Example complete JSON output:",
            "```json",
            example_text,
            "```",
            OUTPUT_CONTRACT_END,
        ]
    )
    return "\n".join(lines)


ENUM_EXAMPLE_BY_PATH = {
    "$[].level": "senior",
    "$.interviewRounds[].type": "technical",
    "$.overall_status": "meets_bar",
    "$.questions[].severity": "medium",
    "$.requirements[].evidenceLevel": "explicit",
    "$.requirements[].kind": "must_have",
    "$.riskItems[].severity": "medium",
    "$.severity": "nudge",
    "$.suggestions[].source": "jd",
}

STRING_EXAMPLE_BY_PATH = {
    "$.displayName": "Candidate A - Backend engineer",
    "$.markdownText": "# Candidate A\n\n## Experience\n- Reduced p95 latency by 32% by redesigning cache invalidation.\n\n## Skills\n- Go",
    "$.basics.name": "Candidate A",
    "$.dimension_scores[].name": "System design",
    "$.dimension_scores[].reasoning": "Clear architecture tradeoffs, but limited quantified impact.",
    "$.dimension_results.system_design.score_level": "meets_bar",
    "$.dimension_results.system_design.status": "meets_bar",
    "$.experiences[].summary": "Owned high-throughput API reliability and platform migrations.",
    "$.highlights[].dimension": "System design",
    "$.highlights[].evidence": "Explained queue backpressure and deployment tradeoffs.",
    "$.issues[].dimension": "Risk handling",
    "$.issues[].evidence": "Rollback plan was mentioned but not made concrete.",
    "$.next_actions[].label": "Replay the system design follow-up",
    "$.projects[].name": "Interview Prep Platform",
    "$.projects[].summary": "Built evidence-backed interview practice workflows.",
    "$.questions[].aiAnalysis": "Good direction, but add numbers and rollback detail.",
    "$.questions[].interviewerReaction": "Asked for concrete failure metrics.",
    "$.questions[].myAnswerSummary": "Explained queue sizing and retry policy.",
    "$.questions[].questionText": "How did you handle backpressure in the migration?",
    "$.companyName": "Acme",
    "$.interviewRounds[].focus": "Probe distributed systems reliability and rollback decisions.",
    "$.interviewRounds[].name": "Technical system design",
    "$.requirements[].description": "The JD explicitly calls for owning high-availability backend systems.",
    "$.requirements[].label": "Design reliable distributed services",
    "$.riskItems[].label": "Thin rollback detail",
    "$.suggestions[].reason": "Adds scope, measurable impact, and target-JD language.",
    "$.suggestions[].suggestedBullet": "Improved API reliability by reducing incident rate 28% through retry-safe queue processing.",
}

STRING_EXAMPLE_BY_KEY = {
    "aiAnalysis": "The answer identified the main tradeoff but needs clearer evidence and metrics.",
    "answerSummary": "Candidate described implementation details but had not yet clarified the main tradeoff.",
    "branchDimension": "System design tradeoffs",
    "company": "Example Cloud",
    "companyTag": "Growth-stage SaaS",
    "comp": "$180k-$220k",
    "contact": "email and phone redacted",
    "cue": "Clarify the tradeoff before moving to implementation details.",
    "description": "This requirement is explicit in the JD and likely to be tested in system design.",
    "dimension": "System design",
    "dimensionHint": "System design",
    "evidence": "The candidate explained cache invalidation but did not quantify failure impact.",
    "evidenceLevel": "explicit",
    "focusDimension": "System design",
    "headline": "Backend engineer focused on distributed systems",
    "interviewerReaction": "The interviewer asked for more detail on rollback strategy.",
    "jobMatchId": "job-123",
    "label": "retry system design drill",
    "level": "senior",
    "location": "Remote US",
    "myAnswerSummary": "Described a queue-backed service migration and the operational safeguards used.",
    "name": "System design",
    "networkNote": "3 prior interview reports mention similar backend platform scope.",
    "overall_status": "meets_bar",
    "posted": "posted 2 days ago",
    "questionIntent": "Probe ownership, tradeoffs, and evidence quality.",
    "questionText": "Tell me about a time you improved reliability in a distributed system.",
    "reason": "Adds measurable impact and ties the bullet to the target JD.",
    "recommended_framework": "Use STAR with explicit constraints, tradeoffs, and measured outcome.",
    "review_status": "ready",
    "score_level": "meets_bar",
    "school": "Example University",
    "source": "jd",
    "sourceLabel": "internal jobs pool",
    "sourceUrl": "https://jobs.internal.example/job-123",
    "stage": "onsite",
    "start": "2021",
    "summary": "The candidate gave a structured answer with clear tradeoffs but should quantify impact.",
    "degree": "B.S. Computer Science",
    "end": "Present",
    "field": "Computer Science",
    "originalBullet": "Worked on API reliability.",
    "title": "Senior Backend Engineer",
    "technologies": "Go",
    "type": "retry_round",
    "whyLikelyAsked": "The JD emphasizes distributed systems and ownership of reliability.",
}

ARRAY_ITEM_EXAMPLE_BY_KEY = {
    "bullets": "Reduced p95 latency by 32% by redesigning cache invalidation.",
    "coreThemes": "Distributed systems reliability",
    "education": {
        "school": "Example University",
        "degree": "B.S. Computer Science",
        "field": "Computer Science",
        "start": "2014",
        "end": "2018",
    },
    "expectedSignals": "Names constraints, tradeoffs, measured impact, and rollback plan.",
    "experiences": {
        "company": "Example Cloud",
        "title": "Senior Backend Engineer",
        "start": "2021",
        "end": "Present",
        "summary": "Owned high-throughput API reliability and platform migrations.",
        "bullets": ["Reduced incident rate by introducing replayable job processing."],
    },
    "gaps": "Needs deeper rollback and failure-mode analysis.",
    "highlights": "Strong ownership of backend reliability work.",
    "interviewRounds": {
        "sequence": 1,
        "type": "technical",
        "name": "Technical system design",
        "durationMinutes": 60,
        "focus": "Probe distributed systems reliability and rollback decisions.",
    },
    "issues": {
        "dimension": "Risk handling",
        "evidence": "Rollback plan was mentioned but not made concrete.",
        "confidence": 0.73,
    },
    "languages": "English - professional",
    "next_actions": {"type": "retry_round", "label": "Replay the system design follow-up"},
    "projects": {
        "name": "Interview Prep Platform",
        "summary": "Built evidence-backed interview practice workflows.",
        "technologies": ["Go", "PostgreSQL", "React"],
        "bullets": ["Implemented structured AI output validation and retry-safe jobs."],
    },
    "questions": {
        "questionText": "How did you handle backpressure in the migration?",
        "myAnswerSummary": "Explained queue sizing and retry policy.",
        "interviewerReaction": "Asked for concrete failure metrics.",
        "aiAnalysis": "Good direction, but add numbers and rollback detail.",
    },
    "reasons": "Recent backend platform work maps directly to the JD.",
    "requirements": {
        "kind": "must_have",
        "label": "Design reliable distributed services",
        "description": "The JD explicitly calls for owning high-availability backend systems.",
        "evidenceLevel": "explicit",
    },
    "retry_focus_turn_ids": "turn-3",
    "riskItems": {"label": "Thin rollback detail", "severity": "medium"},
    "riskSignals": "The JD asks for on-call ownership without naming team support.",
    "risks": "Less evidence for frontend-heavy collaboration requirements.",
    "skills": "Go",
    "strengths": "Quantified backend reliability impact.",
    "strengths_to_amplify": {"topic": "Reliability ownership", "evidence": "Reduced incidents."},
    "suggestions": {
        "originalBullet": "Worked on API reliability.",
        "suggestedBullet": "Improved API reliability by reducing incident rate 28% through retry-safe queue processing.",
        "reason": "Adds scope, measurable impact, and target-JD language.",
    },
    "supporting_observations": "Used concrete operational examples from the session.",
}

INTEGER_EXAMPLE_BY_KEY = {
    "must": 4,
    "plus": 2,
    "score": 86,
    "similarInterviewers": 3,
    "timeBudgetSeconds": 180,
    "durationMinutes": 60,
    "sequence": 1,
    "total": 5,
    "totalPlus": 3,
}

NUMBER_EXAMPLE_BY_KEY = {
    "confidence": 0.82,
    "score": 4.2,
}


def _path_key(path: str) -> str:
    normalized = path.replace("[]", "")
    if "." not in normalized:
        return ""
    return normalized.rsplit(".", 1)[-1]


def _array_key(path: str) -> str:
    parent = path[:-2] if path.endswith("[]") else path
    return _path_key(parent)


def extract_output_contract_block(body: str) -> str | None:
    start = body.find(OUTPUT_CONTRACT_START)
    end = body.find(OUTPUT_CONTRACT_END)
    if start == -1 or end == -1 or end < start:
        return None
    end += len(OUTPUT_CONTRACT_END)
    return body[start:end]


def example_for_schema(schema: dict, path: str = "$") -> object:
    enum = schema.get("enum")
    if isinstance(enum, list) and enum:
        return ENUM_EXAMPLE_BY_PATH.get(path, enum[0])
    schema_type = schema.get("type")
    if schema_type == "object":
        props = schema.get("properties") or {}
        out: dict[str, object] = {}
        for key in ordered_schema_property_keys(schema):
            child = props.get(key)
            if isinstance(child, dict):
                out[key] = example_for_schema(child, f"{path}.{key}")
        if out:
            return out
        dynamic = ARRAY_ITEM_EXAMPLE_BY_KEY.get(_array_key(path))
        if isinstance(dynamic, dict):
            return dynamic
        if path == "$.dimension_results":
            return {
                "system_design": {
                    "score_level": "meets_bar",
                    "status": "meets_bar",
                    "confidence": 0.82,
                    "score": 4.2,
                }
            }
        return out
    if schema_type == "array":
        item = schema.get("items")
        if not isinstance(item, dict):
            return []
        first = example_for_schema(item, path + "[]")
        if path == "$.interviewRounds":
            second = {
                "sequence": 2,
                "type": "manager",
                "name": "Hiring manager ownership interview",
                "durationMinutes": 45,
                "focus": "Assess ownership scope, incident judgment, and cross-team collaboration.",
            }
            return [first, second]
        return [first]
    if schema_type == "integer":
        return INTEGER_EXAMPLE_BY_KEY.get(_path_key(path), 2)
    if schema_type == "number":
        return NUMBER_EXAMPLE_BY_KEY.get(_path_key(path), 0.82)
    if schema_type == "boolean":
        return True
    if schema_type == "null":
        return None
    path_example = STRING_EXAMPLE_BY_PATH.get(path)
    if path_example is not None:
        return path_example
    key = _path_key(path)
    if path.endswith("[]"):
        item_example = ARRAY_ITEM_EXAMPLE_BY_KEY.get(_array_key(path))
        if isinstance(item_example, str):
            return item_example
    return STRING_EXAMPLE_BY_KEY.get(key, f"example {key or 'value'}")


def validate_value_against_schema(value: object, schema: dict, path: str, errors: list[str]) -> None:
    schema_type = schema.get("type")
    if schema_type == "object":
        if not isinstance(value, dict):
            errors.append(f"{path}: expected object")
            return
        for key in schema.get("required") or []:
            if key not in value:
                errors.append(f"{path}: missing required field {key!r}")
        props = schema.get("properties") or {}
        for key, child in props.items():
            if key in value and isinstance(child, dict):
                validate_value_against_schema(value[key], child, f"{path}.{key}", errors)
    elif schema_type == "array":
        if not isinstance(value, list):
            errors.append(f"{path}: expected array")
            return
        item_schema = schema.get("items")
        if isinstance(item_schema, dict):
            for index, item in enumerate(value):
                validate_value_against_schema(item, item_schema, f"{path}[{index}]", errors)
    elif schema_type == "string":
        if not isinstance(value, str):
            errors.append(f"{path}: expected string")
    elif schema_type == "number":
        if not isinstance(value, (int, float)) or isinstance(value, bool):
            errors.append(f"{path}: expected number")
    elif schema_type == "integer":
        if not isinstance(value, int) or isinstance(value, bool):
            errors.append(f"{path}: expected integer")
    elif schema_type == "boolean":
        if not isinstance(value, bool):
            errors.append(f"{path}: expected boolean")
    elif schema_type == "null" and value is not None:
        errors.append(f"{path}: expected null")
    enum = schema.get("enum")
    if isinstance(enum, list) and enum and value not in enum:
        errors.append(f"{path}: value {value!r} not in enum {enum!r}")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--prompts-dir",
        default="config/prompts",
        help="Directory containing prompt yaml/md baseline files",
    )
    parser.add_argument(
        "--migrations-dir",
        default="migrations",
        help="Directory containing migration SQL files (for Phase 4.6 cross-check)",
    )
    args = parser.parse_args(argv)

    prompts_root = pathlib.Path(args.prompts_dir)
    migrations_root = pathlib.Path(args.migrations_dir)

    errors: list[str] = []
    errors.extend(lint_prompts_directory(prompts_root))
    errors.extend(lint_output_schemas(prompts_root))
    errors.extend(lint_seed_migration(prompts_root, migrations_root))

    if errors:
        for e in errors:
            print(f"prompt_lint: {e}", file=sys.stderr)
        return 1
    print(f"prompt_lint: {len(list(prompts_root.rglob('*.yaml')))} files clean")
    return 0


if __name__ == "__main__":
    sys.exit(main())
