#!/usr/bin/env python3
"""scripts/lint/prompt_lint.py - F3 prompt registry linter.

Validates `config/prompts/<feature_key>/<version>[.<language>].{yaml,md}`
against the schema and canonical hash algorithm fixed by
`config/prompts/README.md` and `docs/spec/prompt-rubric-registry/spec.md` v2.1.

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
STATUS_ENUM = {"draft", "active", "deprecated"}
SEMVER_RE = re.compile(r"^v\d+\.\d+\.\d+(-[A-Za-z0-9\.-]+)?(\+[A-Za-z0-9\.-]+)?$")
LANGUAGE_RE = re.compile(r"^multi$|^[a-z]{2,3}$")

# Forbidden tokens reserved for the lint gate. They live here, not in the
# README, so `grep -rE "\bTBD\b|placeholder" config/prompts/` stays clean
# while the lint script still rejects them inside Markdown bodies.
FORBIDDEN_BODY_TOKEN_RE = re.compile(r"\bTBD\b|\bplaceholder\b", re.IGNORECASE)
RETIRED_MODULE_RE = re.compile(r"\bmistakes\b|\bgrowth\b|\bdrill\b|mistake\.extract")


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
    if RETIRED_MODULE_RE.search(body_text):
        errors.append(f"{md_path}: body contains retired-module name")

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

    for sql_path in sorted(migrations_root.glob("*seed_baseline_prompt_rubric*.up.sql")):
        text = sql_path.read_text(encoding="utf-8")
        if not seed_re.search(text):
            continue
        for m in row_re.finditer(text):
            key = (m.group("feature_key"), m.group("version"), m.group("language"))
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
    errors.extend(lint_seed_migration(prompts_root, migrations_root))

    if errors:
        for e in errors:
            print(f"prompt_lint: {e}", file=sys.stderr)
        return 1
    print(f"prompt_lint: {len(list(prompts_root.rglob('*.yaml')))} files clean")
    return 0


if __name__ == "__main__":
    sys.exit(main())
