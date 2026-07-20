#!/usr/bin/env python3
"""Reject duplicated D-* and C-* table IDs within each subject spec."""

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path


CONTRACT_ID_RE = re.compile(r"^\s*\|\s*(?P<id>[DC]-\d+)\s*\|")
FENCE_RE = re.compile(r"^\s*(```|~~~)")
# Existing duplicates outside this remediation remain explicit debt instead of
# forcing unrelated subject rewrites. The gate rejects every new duplicate.
LEGACY_DUPLICATES = frozenset(
    {
        ("ai-provider-and-model-routing/spec.md", "C-14"),
        ("frontend-home-job-picks-and-parse/spec.md", "D-15"),
        ("frontend-home-job-picks-and-parse/spec.md", "D-16"),
    }
)


@dataclass(frozen=True)
class Finding:
    source: Path
    contract_id: str
    first_line: int
    duplicate_line: int

    def format(self, root: Path) -> str:
        source = (
            self.source.relative_to(root)
            if self.source.is_relative_to(root)
            else self.source
        )
        return (
            f"{source}:{self.duplicate_line}: duplicate contract ID "
            f"{self.contract_id} (first defined at line {self.first_line})"
        )


def scan_spec(source: Path) -> list[Finding]:
    seen: dict[str, int] = {}
    findings: list[Finding] = []
    in_fence = False

    for line_number, line in enumerate(
        source.read_text(encoding="utf-8").splitlines(), start=1
    ):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue

        match = CONTRACT_ID_RE.match(line)
        if match is None:
            continue
        contract_id = match.group("id")
        first_line = seen.setdefault(contract_id, line_number)
        if first_line != line_number:
            findings.append(
                Finding(
                    source=source,
                    contract_id=contract_id,
                    first_line=first_line,
                    duplicate_line=line_number,
                )
            )

    return findings


def scan_directory(root: Path) -> list[Finding]:
    findings: list[Finding] = []
    for source in sorted(root.glob("*/spec.md")):
        for finding in scan_spec(source):
            relative_source = source.relative_to(root).as_posix()
            if (relative_source, finding.contract_id) in LEGACY_DUPLICATES:
                continue
            findings.append(finding)
    return findings


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Check subject spec D/C contract IDs for file-local uniqueness."
    )
    parser.add_argument("root", type=Path, help="Path to docs/spec")
    args = parser.parse_args(argv)

    root = args.root.resolve()
    if not root.is_dir():
        print(f"check_spec_contract_ids: not a directory: {root}", file=sys.stderr)
        return 2

    findings = scan_directory(root)
    if findings:
        for finding in findings:
            print(finding.format(root), file=sys.stderr)
        print(
            f"check_spec_contract_ids: {len(findings)} duplicate ID(s) in {root}",
            file=sys.stderr,
        )
        return 1

    print(f"check_spec_contract_ids: OK ({root})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
