#!/usr/bin/env python3
"""Fixture-backed operation registry for mock runtime tooling.

The registry is derived from `openapi/openapi.yaml` plus
`openapi/fixtures/<tag>/<operationId>.json`; it is intentionally not a
hand-maintained operation inventory.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Mapping

from scripts.lint import validate_fixtures


class FixtureRegistryError(Exception):
    """Raised when fixture registry construction or lookup fails."""


@dataclass(frozen=True)
class FixtureEntry:
    tag: str
    operation_id: str
    method: str
    path: str
    fixture_path: Path
    default_status: int
    request_schema_ref: str | None
    response_schema_ref: str | None
    scenarios: tuple[str, ...]


class FixtureRegistry:
    def __init__(self, entries: Mapping[str, FixtureEntry]) -> None:
        self._entries = dict(entries)

    def lookup(self, operation_id: str) -> FixtureEntry:
        try:
            return self._entries[operation_id]
        except KeyError as exc:
            raise FixtureRegistryError(f"unknown operationId: {operation_id}") from exc

    def entries(self) -> tuple[FixtureEntry, ...]:
        return tuple(self._entries[operation_id] for operation_id in sorted(self._entries))


def build_fixture_registry(repo_root: Path) -> FixtureRegistry:
    repo_root = repo_root.resolve()
    spec = validate_fixtures.load_openapi(repo_root / "openapi" / "openapi.yaml")
    op_index = validate_fixtures.build_operation_index(spec)
    fixtures_root = repo_root / "openapi" / "fixtures"

    entries: dict[str, FixtureEntry] = {}
    errors: list[str] = []
    for tag, operation_id, fixture_path, data in validate_fixtures.walk_fixtures(fixtures_root):
        op_meta = op_index.get(operation_id)
        if op_meta is None:
            errors.append(f"{fixture_path.relative_to(repo_root)}: operationId not in openapi.yaml")
            continue
        expected_tag = op_meta.get("tag")
        if expected_tag != tag:
            errors.append(
                f"{fixture_path.relative_to(repo_root)}: tag {tag!r} != openapi tag {expected_tag!r}"
            )
        scenarios = data.get("scenarios")
        if not isinstance(scenarios, dict) or "default" not in scenarios:
            errors.append(f"{fixture_path.relative_to(repo_root)}: scenarios.default missing")
            continue
        response = scenarios["default"].get("response") or {}
        status = response.get("status")
        if not isinstance(status, int):
            errors.append(f"{fixture_path.relative_to(repo_root)}: default response.status must be int")
            continue
        operation = op_meta["operation"]
        response_schema, _status_key = validate_fixtures._select_response_schema(operation, status)
        request_schema = validate_fixtures._request_schema(operation)
        entries[operation_id] = FixtureEntry(
            tag=tag,
            operation_id=operation_id,
            method=op_meta["method"],
            path=op_meta["path"],
            fixture_path=fixture_path.resolve(),
            default_status=status,
            request_schema_ref=_schema_ref(request_schema),
            response_schema_ref=_schema_ref(response_schema),
            scenarios=tuple(scenarios.keys()),
        )

    missing = set(op_index) - set(entries)
    for operation_id in sorted(missing):
        errors.append(f"missing fixture for operationId {operation_id}")

    if errors:
        raise FixtureRegistryError("; ".join(errors))
    return FixtureRegistry(entries)


def _schema_ref(schema: dict | None) -> str | None:
    if not isinstance(schema, dict):
        return None
    ref = schema.get("$ref")
    return ref if isinstance(ref, str) else None
