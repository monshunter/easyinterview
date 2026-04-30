#!/usr/bin/env python3
"""B4 migration lint gate.

This script starts with the file-contract checks needed by migrate-check and is
extended by later B4 phases for enum/check source drift and privacy red flags.
"""

from __future__ import annotations

import argparse
import hashlib
import re
import sys
from pathlib import Path

import yaml


MIGRATION_RE = re.compile(r"^([0-9]{6})_[a-z0-9]+(?:_[a-z0-9]+)*\.(up|down)\.sql$")
CHECK_RE = re.compile(r"CHECK\s*\(\s*([a-z_]+)\s+IN\s*\((.*?)\)\s*\)", re.IGNORECASE | re.DOTALL)
CREATE_TABLE_RE = re.compile(r"CREATE\s+TABLE\s+([a-z_]+)\s*\((.*?)\);", re.IGNORECASE | re.DOTALL)
VALUE_RE = re.compile(r"'([^']+)'")
FORBIDDEN_SECRET_RE = re.compile(r"\b(raw_token|session_cookie|api_key|provider_token)\b", re.IGNORECASE)


def validate_file_contract(migrations_dir: Path) -> list[str]:
    problems: list[str] = []
    pairs: dict[int, set[str]] = {}

    if not migrations_dir.exists():
        return [f"migrations dir does not exist: {migrations_dir}"]

    for path in sorted(migrations_dir.iterdir()):
        if path.is_dir():
            if path.name == "backfill":
                continue
            problems.append(f"{path.name} is a directory; migrations must be flat")
            continue
        if path.suffix != ".sql":
            continue
        match = MIGRATION_RE.match(path.name)
        if not match:
            problems.append(f"invalid migration file name: {path.name}")
            continue
        version = int(match.group(1))
        direction = match.group(2)
        pairs.setdefault(version, set()).add(direction)

    versions = sorted(pairs)
    for offset, version in enumerate(versions, start=1):
        if version != offset:
            problems.append(f"expected version {offset:06d}, found {version:06d}")
        directions = pairs[version]
        if "up" not in directions:
            problems.append(f"missing up migration for {version:06d}")
        if "down" not in directions:
            problems.append(f"missing down migration for {version:06d}")

    return problems


def run_checks(repo_root: Path) -> list[str]:
    migrations_dir = repo_root / "migrations"
    problems = validate_file_contract(migrations_dir)
    problems.extend(validate_enum_sources(migrations_dir))
    problems.extend(validate_secret_red_lines(migrations_dir))
    return problems


def validate_enum_sources(migrations_dir: Path) -> list[str]:
    enum_sources = load_enum_sources(migrations_dir / "enum-sources.yaml")
    checks = extract_sql_checks(migrations_dir)
    problems: list[str] = []

    for key, values in sorted(checks.items()):
        source = enum_sources.get(key)
        if source is None:
            problems.append(f"{key} check list is not registered in migrations/enum-sources.yaml")
            continue
        declared_values = source.get("values") or []
        if declared_values != values:
            problems.append(f"{key} checksum drift: SQL values {values} != enum-sources values {declared_values}")
            continue
        want_checksum = checksum_values(values)
        if source.get("checksum") != want_checksum:
            problems.append(f"{key} checksum drift: expected {want_checksum}, got {source.get('checksum')}")

    for key in sorted(set(enum_sources) - set(checks)):
        problems.append(f"{key} registered in enum-sources.yaml but not present in SQL checks")

    return problems


def load_enum_sources(path: Path) -> dict[str, dict]:
    if not path.exists():
        return {}
    data = yaml.safe_load(path.read_text()) or {}
    out: dict[str, dict] = {}
    for item in data.get("checks", []):
        key = f"{item.get('table')}.{item.get('column')}"
        out[key] = item
    return out


def extract_sql_checks(migrations_dir: Path) -> dict[str, list[str]]:
    checks: dict[str, list[str]] = {}
    for path in sorted(migrations_dir.glob("*.up.sql")):
        sql = path.read_text()
        for table_match in CREATE_TABLE_RE.finditer(sql):
            table = table_match.group(1)
            body = table_match.group(2)
            for check_match in CHECK_RE.finditer(body):
                column = check_match.group(1)
                values = VALUE_RE.findall(check_match.group(2))
                checks[f"{table}.{column}"] = values
    return checks


def validate_secret_red_lines(migrations_dir: Path) -> list[str]:
    problems: list[str] = []
    for path in sorted(migrations_dir.glob("*.sql")):
        for lineno, line in enumerate(path.read_text().splitlines(), start=1):
            match = FORBIDDEN_SECRET_RE.search(line)
            if match:
                problems.append(f"{path.name}:{lineno}: forbidden plaintext secret field marker {match.group(1)}")
    return problems


def checksum_values(values: list[str]) -> str:
    digest = hashlib.sha256("|".join(values).encode()).hexdigest()[:16]
    return f"sha256:{digest}"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    problems = run_checks(repo_root)
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print("migration lint: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
