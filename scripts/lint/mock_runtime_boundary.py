#!/usr/bin/env python3
"""Lint mock runtime boundaries.

Checks that frontend runtime code does not import prototype data directly and
that fixture response bodies do not leak prototype-only display fields.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any, Iterable

import yaml


FRONTEND_EXTENSIONS = {".ts", ".tsx", ".js", ".jsx"}
UI_DATA_IMPORT_RE = re.compile(
    r"""(?:import\s+(?:[^;]*?\s+from\s+)?|import\(\s*|require\(\s*)["'][^"']*ui-design/src/data\.jsx["']""",
    re.MULTILINE,
)
PROTOTYPE_ONLY_RESPONSE_FIELDS = {
    "statusTone",
    "readinessLabel",
    "qIdx",
    "t",
}
OWNER_SPEC_HINT = "docs/spec/mock-contract-suite/spec.md"
RETIRED_CONTRACT_TOKENS = (
    "/mistakes",
    "/growth",
    "/drill",
    "/voice",
    "Mistakes",
    "Growth",
    "Drill",
    "Voice",
    "single_drill",
    "gateway_route",
    "ai.gateway",
    "default.provider",
    "task_type",
)
RETIRED_TOKEN_SCAN_ROOTS = (
    "openapi/fixtures",
    "frontend/src/api",
    "backend/internal/api/mockruntime",
    "openapi/templates/ts/client.tmpl",
)
RETIRED_TOKEN_EXTENSIONS = {".go", ".ts", ".json", ".tmpl"}


def lint(repo_root: Path) -> list[str]:
    repo_root = repo_root.resolve()
    errors: list[str] = []
    errors.extend(_lint_fixture_tag_directories(repo_root))
    errors.extend(_lint_frontend_imports(repo_root))
    errors.extend(_lint_fixture_response_fields(repo_root))
    errors.extend(_lint_retired_contract_tokens(repo_root))
    return errors


def _lint_fixture_tag_directories(repo_root: Path) -> list[str]:
    fixtures_root = repo_root / "openapi" / "fixtures"
    openapi_path = repo_root / "openapi" / "openapi.yaml"
    if not fixtures_root.is_dir() or not openapi_path.is_file():
        return []

    spec = yaml.safe_load(openapi_path.read_text(encoding="utf-8"))
    expected_tags = tuple(
        tag["name"]
        for tag in (spec.get("tags") or [])
        if isinstance(tag, dict) and isinstance(tag.get("name"), str)
    )
    expected = set(expected_tags)
    actual = {
        path.name
        for path in fixtures_root.iterdir()
        if path.is_dir()
    }

    errors: list[str] = []
    for tag in sorted(actual - expected):
        errors.append(
            f"openapi/fixtures/{tag}: unexpected fixture tag directory {tag!r}; "
            f"owner spec: {OWNER_SPEC_HINT}"
        )
    for tag in expected_tags:
        if tag not in actual:
            errors.append(
                f"openapi/fixtures/{tag}: missing fixture tag directory {tag!r}; "
                f"owner spec: {OWNER_SPEC_HINT}"
            )
    return errors


def _lint_frontend_imports(repo_root: Path) -> list[str]:
    root = repo_root / "frontend" / "src"
    if not root.is_dir():
        return []
    errors: list[str] = []
    for path in sorted(p for p in root.rglob("*") if p.suffix in FRONTEND_EXTENSIONS):
        text = path.read_text(encoding="utf-8")
        if UI_DATA_IMPORT_RE.search(text):
            errors.append(
                f"{path.relative_to(repo_root)}: frontend runtime must not import ui-design/src/data.jsx"
            )
    return errors


def _lint_fixture_response_fields(repo_root: Path) -> list[str]:
    fixtures_root = repo_root / "openapi" / "fixtures"
    if not fixtures_root.is_dir():
        return []
    errors: list[str] = []
    for fixture_path in sorted(fixtures_root.glob("*/*.json")):
        data = json.loads(fixture_path.read_text(encoding="utf-8"))
        operation_id = data.get("operationId", fixture_path.stem)
        scenarios = data.get("scenarios") or {}
        if not isinstance(scenarios, dict):
            continue
        for scenario_name, scenario in scenarios.items():
            body = ((scenario or {}).get("response") or {}).get("body")
            for key_path, key in _walk_keys(body):
                if key in PROTOTYPE_ONLY_RESPONSE_FIELDS:
                    errors.append(
                        f"{fixture_path.relative_to(repo_root)}:{operation_id}.{scenario_name}.response.body"
                        f"{key_path}: prototype-only display field {key!r} is forbidden"
                    )
    return errors


def _lint_retired_contract_tokens(repo_root: Path) -> list[str]:
    errors: list[str] = []
    for path in _retired_scan_files(repo_root):
        text = path.read_text(encoding="utf-8")
        for token in RETIRED_CONTRACT_TOKENS:
            if token in text:
                errors.append(
                    f"{path.relative_to(repo_root)}: retired mock/API token {token!r} is forbidden; "
                    f"owner spec: {OWNER_SPEC_HINT}"
                )
    return errors


def _retired_scan_files(repo_root: Path) -> Iterable[Path]:
    for rel in RETIRED_TOKEN_SCAN_ROOTS:
        path = repo_root / rel
        if path.is_file():
            if path.suffix in RETIRED_TOKEN_EXTENSIONS:
                yield path
            continue
        if not path.is_dir():
            continue
        for child in sorted(p for p in path.rglob("*") if p.is_file()):
            if child.suffix in RETIRED_TOKEN_EXTENSIONS:
                yield child


def _walk_keys(value: Any, prefix: str = "") -> Iterable[tuple[str, str]]:
    if isinstance(value, dict):
        for key, child in value.items():
            key_path = f"{prefix}.{key}" if prefix else f".{key}"
            yield key_path, key
            yield from _walk_keys(child, key_path)
    elif isinstance(value, list):
        for index, child in enumerate(value):
            yield from _walk_keys(child, f"{prefix}[{index}]")


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path(__file__).resolve().parents[2],
        help="Repository root containing frontend/ and openapi/.",
    )
    args = parser.parse_args(list(argv))

    errors = lint(args.repo_root)
    if errors:
        for error in errors:
            print(f"mock-runtime-boundary: {error}", file=sys.stderr)
        print(f"mock-runtime-boundary: FAILED with {len(errors)} error(s)", file=sys.stderr)
        return 1
    print("mock-runtime-boundary: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
