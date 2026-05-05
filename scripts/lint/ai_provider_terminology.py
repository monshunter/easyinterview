#!/usr/bin/env python3
"""Reject retired AI gateway terminology in active AI provider surfaces.

This gate is intentionally narrower than a whole-repo grep: historical
evidence under docs/work-journal, docs/reports, docs/bugs, and spec history
files may retain old wording, but active code, config, deploy assets, ADRs, and
generated AI convention artifacts must use provider-neutral terminology.
"""
from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path


TEXT_SUFFIXES = {
    ".env",
    ".example",
    ".go",
    ".json",
    ".md",
    ".sh",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}

TEXT_NAMES = {
    "Makefile",
}

EXCLUDED_NAMES = {
    "history.md",
}

SCAN_PATHS = [
    ".env.example",
    "backend",
    "config",
    "deploy/dev-stack",
    "docs/spec/ai-provider-and-model-routing",
    "docs/spec/engineering-roadmap/spec.md",
    "docs/spec/engineering-roadmap/decisions/ADR-Q4-cloud-deploy-target.md",
    "docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md",
    "docs/spec/engineering-roadmap/plans/001-decompose-subspecs/context.yaml",
    "docs/spec/local-dev-stack",
    "docs/spec/secrets-and-config",
    "docs/spec/shared-conventions-codified",
    "frontend/src/lib/conventions/ai.ts",
    "shared/conventions.yaml",
]


@dataclass(frozen=True)
class RetiredPattern:
    label: str
    pattern: re.Pattern[str]


RETIRED_PATTERNS = [
    RetiredPattern("AI_GATEWAY env key", re.compile(r"\bAI_GATEWAY_[A-Z0-9_]+\b")),
    RetiredPattern("gateway_route schema key", re.compile(r"\bgateway_route\b")),
    RetiredPattern("GatewayRoute API field", re.compile(r"\bGatewayRoute\b")),
    RetiredPattern("GatewayBaseURL API field", re.compile(r"\bGatewayBaseURL\b")),
    RetiredPattern("GatewayAPIKey API field", re.compile(r"\bGatewayAPIKey\b")),
    RetiredPattern(
        "ErrMissingGatewayConfig API error",
        re.compile(r"\bErrMissingGatewayConfig\b"),
    ),
    RetiredPattern("ai.gateway config path", re.compile(r"\bai\.gateway[A-Za-z0-9_]*\b")),
    RetiredPattern("gateway terminology", re.compile(r"\bgateway\b", re.IGNORECASE)),
]


def is_text_path(path: Path) -> bool:
    return path.name in TEXT_NAMES or path.suffix in TEXT_SUFFIXES


def iter_scan_files(repo: Path) -> list[Path]:
    files: list[Path] = []
    for rel in SCAN_PATHS:
        path = repo / rel
        if not path.exists():
            continue
        if path.is_file():
            if path.name not in EXCLUDED_NAMES and is_text_path(path):
                files.append(path)
            continue
        for child in sorted(path.rglob("*")):
            if not child.is_file():
                continue
            if child.name in EXCLUDED_NAMES:
                continue
            if is_text_path(child):
                files.append(child)
    return files


def scan_file(repo: Path, path: Path) -> list[str]:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []

    findings: list[str] = []
    rel = path.relative_to(repo)
    for lineno, line in enumerate(text.splitlines(), start=1):
        for retired in RETIRED_PATTERNS:
            if retired.pattern.search(line):
                findings.append(f"{rel}:{lineno}: {retired.label}: {line.strip()}")
                break
    return findings


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    args = parser.parse_args()

    repo = args.repo_root.resolve()
    findings: list[str] = []
    for path in iter_scan_files(repo):
        findings.extend(scan_file(repo, path))

    if findings:
        print("ai_provider_terminology: retired terminology found", file=sys.stderr)
        for finding in findings:
            print(f"  - {finding}", file=sys.stderr)
        print(
            "Fix: use AI provider endpoint/config/profile route terminology in active surfaces; "
            "keep historical wording only in history, work-journal, report, or bug records.",
            file=sys.stderr,
        )
        return 1

    print(f"ai_provider_terminology: OK ({len(iter_scan_files(repo))} active files scanned)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
