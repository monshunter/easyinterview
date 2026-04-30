#!/usr/bin/env python3
"""B4 migration lint gate.

This script starts with the file-contract checks needed by migrate-check and is
extended by later B4 phases for enum/check source drift and privacy red flags.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


MIGRATION_RE = re.compile(r"^([0-9]{6})_[a-z0-9]+(?:_[a-z0-9]+)*\.(up|down)\.sql$")


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


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    problems = validate_file_contract(repo_root / "migrations")
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1
    print("migration lint: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
