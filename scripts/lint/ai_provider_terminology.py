#!/usr/bin/env python3
"""Reject out-of-scope AI gateway terminology in active AI provider surfaces.

This gate is intentionally narrower than a whole-repo grep: evidence under
docs/work-journal, docs/reports, docs/bugs, and spec history files may retain
out-of-scope wording, but active code, config, deploy assets, ADRs, and
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
class OutOfScopePattern:
    label: str
    pattern: re.Pattern[str]


OUT_OF_SCOPE_PATTERNS = [
    OutOfScopePattern("AI_GATEWAY env key", re.compile(r"\bAI_GATEWAY_[A-Z0-9_]+\b")),
    OutOfScopePattern("gateway_route schema key", re.compile(r"\bgateway_route\b")),
    OutOfScopePattern("GatewayRoute API field", re.compile(r"\bGatewayRoute\b")),
    OutOfScopePattern("GatewayBaseURL API field", re.compile(r"\bGatewayBaseURL\b")),
    OutOfScopePattern("GatewayAPIKey API field", re.compile(r"\bGatewayAPIKey\b")),
    OutOfScopePattern(
        "ErrMissingGatewayConfig API error",
        re.compile(r"\bErrMissingGatewayConfig\b"),
    ),
    OutOfScopePattern("ai.gateway config path", re.compile(r"\bai\.gateway[A-Za-z0-9_]*\b")),
    OutOfScopePattern("gateway terminology", re.compile(r"\bgateway\b", re.IGNORECASE)),
]

OPENAI_GO_IMPORT = re.compile(r'"github\.com/openai/openai-go/v3(?:/[^"\s]*)?"')
OPENAI_GO_ALLOWED_PREFIXES = (
    Path("backend/internal/ai/aiclient/providers/openai_compatible"),
    Path("backend/internal/ai/aiclient/providers/judge_compatible"),
    Path("backend/internal/ai/aiclient/providers/internal/openaisdk"),
)


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
        if (
            path.suffix == ".go"
            and OPENAI_GO_IMPORT.search(line)
            and not any(rel == prefix or prefix in rel.parents for prefix in OPENAI_GO_ALLOWED_PREFIXES)
        ):
            findings.append(
                f"{rel}:{lineno}: OpenAI SDK import boundary: {line.strip()}"
            )
            continue
        scan_line = line.replace("host-gateway", "")
        for pattern in OUT_OF_SCOPE_PATTERNS:
            if pattern.pattern.search(scan_line):
                findings.append(f"{rel}:{lineno}: {pattern.label}: {line.strip()}")
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
        print("ai_provider_terminology: out-of-scope terminology found", file=sys.stderr)
        for finding in findings:
            print(f"  - {finding}", file=sys.stderr)
        print(
            "Fix: use AI provider registry/profile/provider-ref terminology in active surfaces; "
            "keep out-of-scope wording only in history, work-journal, report, or bug records.",
            file=sys.stderr,
        )
        return 1

    print(f"ai_provider_terminology: OK ({len(iter_scan_files(repo))} active files scanned)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
