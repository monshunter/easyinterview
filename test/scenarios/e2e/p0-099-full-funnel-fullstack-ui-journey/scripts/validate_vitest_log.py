#!/usr/bin/env python3
"""Fail closed unless the P0.099 Vitest log proves the exact focused run."""

from __future__ import annotations

import re
import sys
from pathlib import Path


TARGETS = {
    "src/app/screens/generating/__tests__/GeneratingScreen.test.tsx",
    "src/app/screens/report/__tests__/ConversationReport.test.tsx",
    "src/app/screens/report/__tests__/reportContract.test.ts",
}
ANSI_ESCAPE = re.compile(r"\x1b\[[0-?]*[ -/]*[@-~]")
RESULT_LINE = re.compile(
    r"(?m)^\s*[✓✔]\s+(?P<path>\S+\.(?:test|spec)\.[cm]?[jt]sx?)\s+\((?P<count>[1-9][0-9]*)\s+tests?\)"
)


def fail(message: str) -> None:
    raise ValueError(message)


def validate(path: Path) -> int:
    try:
        text = ANSI_ESCAPE.sub("", path.read_text(encoding="utf-8"))
    except OSError as exc:
        fail(f"Vitest log cannot be read: {exc}")

    if re.search(r"(?m)^\s*RUN\s+v\d", text) is None:
        fail("Vitest runner segment is missing")
    for marker in (
        "No test files found",
        "no tests to run",
        "Test Files  0 passed (0)",
        "Tests  0 passed (0)",
    ):
        if marker.lower() in text.lower():
            fail(f"Vitest zero-run marker found: {marker}")
    if re.search(r"(?i)(?<![0-9])0 tests?\b", text):
        fail("Vitest zero-test result row found")
    if re.search(r"(?mi)^\s*(?:Test Files|Tests)\s+.*\bfailed\b", text):
        fail("Vitest failure summary found")

    results = list(RESULT_LINE.finditer(text))
    executed = [match.group("path") for match in results]
    if len(executed) != 3 or set(executed) != TARGETS:
        fail(f"Vitest executed files={sorted(executed)}, expected exactly {sorted(TARGETS)}")
    if re.search(r"(?m)^\s*Test Files\s+3 passed \(3\)\s*$", text) is None:
        fail("Vitest Test Files summary must be exactly 3 passed (3)")

    summaries = re.findall(r"(?m)^\s*Tests\s+([0-9]+) passed \(([0-9]+)\)\s*$", text)
    if len(summaries) != 1:
        fail("Vitest Tests summary is missing or ambiguous")
    passed, total = (int(value) for value in summaries[0])
    if passed <= 0 or passed != total:
        fail("Vitest Tests summary must contain a positive all-passing count")
    executed_tests = sum(int(match.group("count")) for match in results)
    if passed != executed_tests:
        fail(f"Vitest Tests summary={passed}, executed result rows total={executed_tests}")
    return passed


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: validate_vitest_log.py <trigger-log>", file=sys.stderr)
        return 2
    try:
        tests = validate(Path(sys.argv[1]))
    except ValueError as exc:
        print(f"P0.099 Vitest evidence invalid: {exc}", file=sys.stderr)
        return 1
    print(f"P0_099_VITEST_PASS files=3 tests={tests}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
