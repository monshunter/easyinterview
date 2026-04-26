#!/usr/bin/env python3
"""error_codes.py — local enforcement of error-code casing and boundary.

Implements the local executable lint required by
[shared-conventions-codified spec §6 C-5](../../docs/spec/shared-conventions-codified/spec.md#6-验收标准).
A5 `ci-pipeline-baseline` wires this into CI; this script is also invoked by
`make lint` as the local gate.

Enforces:
  - Every Go `Code*` constant in backend/internal/shared/errors/codes.go has
    a value matching `^[A-Z][A-Z0-9_]+$` (UPPER_SNAKE_CASE).
  - Every entry in `ERROR_CODES = { ... }` inside
    frontend/src/lib/conventions/errors.ts is UPPER_SNAKE_CASE and its key
    matches its value.
  - No file under frontend/src/ outside lib/conventions/errors.ts declares
    its own `ERROR_CODES = {` literal.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GO_CODES = ROOT / "backend" / "internal" / "shared" / "errors" / "codes.go"
TS_CODES = ROOT / "frontend" / "src" / "lib" / "conventions" / "errors.ts"
TS_SRC_ROOT = ROOT / "frontend" / "src"

ERROR_CODE_RE = re.compile(r"^[A-Z][A-Z0-9_]+$")
GO_CONST_RE = re.compile(r'\bCode([A-Za-z0-9]+)\s*=\s*"([^"]+)"')
TS_OBJECT_RE = re.compile(r"\bERROR_CODES\s*=\s*\{(?P<body>.*?)\}\s*as\s+const", re.DOTALL)
TS_ENTRY_RE = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*'([^']+)'\s*$")
TS_BOUNDARY_RE = re.compile(r"\bERROR_CODES\s*=\s*\{")


def check_go_codes() -> list[str]:
    errs: list[str] = []
    if not GO_CODES.exists():
        return [f"missing {GO_CODES.relative_to(ROOT)}"]
    src = GO_CODES.read_text(encoding="utf-8")
    matches = GO_CONST_RE.findall(src)
    if not matches:
        errs.append(f"{GO_CODES.relative_to(ROOT)}: no Code* constants found")
        return errs
    for name, value in matches:
        if not ERROR_CODE_RE.match(value):
            errs.append(
                f"{GO_CODES.relative_to(ROOT)}: Code{name} = {value!r} is not UPPER_SNAKE_CASE"
            )
    return errs


def check_ts_codes() -> list[str]:
    errs: list[str] = []
    if not TS_CODES.exists():
        return [f"missing {TS_CODES.relative_to(ROOT)}"]
    src = TS_CODES.read_text(encoding="utf-8")
    matches, parse_errs = parse_ts_error_entries(src)
    errs.extend(parse_errs)
    if not matches and not parse_errs:
        errs.append(f"{TS_CODES.relative_to(ROOT)}: no ERROR_CODES entries found")
        return errs
    for key, value in matches:
        if not ERROR_CODE_RE.match(key):
            errs.append(
                f"{TS_CODES.relative_to(ROOT)}: ERROR_CODES key {key!r} is not UPPER_SNAKE_CASE"
            )
        if not ERROR_CODE_RE.match(value):
            errs.append(
                f"{TS_CODES.relative_to(ROOT)}: ERROR_CODES.{key} = {value!r} is not UPPER_SNAKE_CASE"
            )
        if key != value:
            errs.append(
                f"{TS_CODES.relative_to(ROOT)}: ERROR_CODES key {key!r} != value {value!r}"
            )
    return errs


def parse_ts_error_entries(src: str) -> tuple[list[tuple[str, str]], list[str]]:
    match = TS_OBJECT_RE.search(src)
    if not match:
        return [], [f"{TS_CODES.relative_to(ROOT)}: ERROR_CODES object not found"]

    entries: list[tuple[str, str]] = []
    errs: list[str] = []
    for lineno, raw in enumerate(match.group("body").splitlines(), start=1):
        line = raw.strip()
        if not line or line.startswith("//"):
            continue
        if line.endswith(","):
            line = line[:-1].rstrip()
        entry = TS_ENTRY_RE.match(line)
        if not entry:
            errs.append(
                f"{TS_CODES.relative_to(ROOT)}: unparseable ERROR_CODES entry near object line {lineno}: {raw.strip()!r}"
            )
            continue
        entries.append((entry.group(1), entry.group(2)))
    return entries, errs


def check_ts_boundary() -> list[str]:
    errs: list[str] = []
    if not TS_SRC_ROOT.exists():
        return errs
    for ts_file in TS_SRC_ROOT.rglob("*.ts"):
        if ts_file.resolve() == TS_CODES.resolve():
            continue
        try:
            text = ts_file.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        if TS_BOUNDARY_RE.search(text):
            errs.append(
                f"{ts_file.relative_to(ROOT)}: ERROR_CODES literal declared outside "
                f"frontend/src/lib/conventions/errors.ts"
            )
    return errs


def main() -> int:
    all_errs: list[str] = []
    all_errs.extend(check_go_codes())
    all_errs.extend(check_ts_codes())
    all_errs.extend(check_ts_boundary())
    if all_errs:
        for e in all_errs:
            print(f"FAIL: {e}", file=sys.stderr)
        return 1
    go_count = len(GO_CONST_RE.findall(GO_CODES.read_text(encoding="utf-8")))
    ts_entries, _ = parse_ts_error_entries(TS_CODES.read_text(encoding="utf-8"))
    ts_count = len(ts_entries)
    print(
        f"OK: {go_count} Go constants, {ts_count} TS entries; boundary clean"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
