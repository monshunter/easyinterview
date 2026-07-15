#!/usr/bin/env python3
"""Reject the removed UI Demo and active contracts that depend on it."""

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable


SCAN_ROOTS = (
    ".agent-skills",
    "AGENTS.md",
    "CLAUDE.md",
    "GEMINI.md",
    "Makefile",
    "docs",
    "frontend",
    "openapi",
    "scripts",
    "test",
)

TEXT_SUFFIXES = {
    ".css",
    ".html",
    ".js",
    ".json",
    ".jsx",
    ".md",
    ".mjs",
    ".py",
    ".sh",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}
TEXT_NAMES = {"Makefile"}

EXCLUDED_PARTS = {
    ".git",
    ".playwright-output",
    ".pytest_cache",
    ".test-output",
    "__pycache__",
    "coverage",
    "dist",
    "node_modules",
    "vendor",
}

HISTORICAL_PREFIXES = (
    Path("docs/bugs"),
    Path("docs/reports"),
    Path("docs/work-journal"),
)

SELF_PATHS = {
    Path("scripts/lint/ui_demo_pruning.py"),
    Path("scripts/lint/ui_demo_pruning_test.py"),
}

NEGATIVE_CONTEXT_RE = re.compile(
    r"不(?:再|得|允许|存在|保留|依赖|要求|使用)|未(?:引入|保留|使用|依赖)|删除|移除|禁止|零残留|"
    r"remove|delete|drop|reject|forbid|must not|no longer|negative|prun(?:e|ing)|absent",
    re.IGNORECASE,
)


@dataclass(frozen=True)
class ForbiddenPattern:
    label: str
    pattern: re.Pattern[str]


FORBIDDEN_PATTERNS = (
    ForbiddenPattern(
        "ui demo path",
        re.compile(
            r"ui-design/(?:src(?:/|$)|index\.html|canvas\.html|run\.sh|"
            r"ui-design-contract\.test\.mjs)|(?<![./\w-])ui-design/"
        ),
    ),
    ForbiddenPattern(
        "pixel parity contract",
        re.compile(
            r"test:(?:pixel-parity|responsive-browser)|"
            r"serve-(?:pixel-parity|responsive-browser)|"
            r"responsive browser verification|golden preview|pixel parity|"
            r"source-level parity|source-level replication|源码复刻|源级复刻|"
            r"UI truth source|UI 真理源",
            re.IGNORECASE,
        ),
    ),
)


@dataclass(frozen=True)
class Finding:
    path: str
    line: int
    label: str
    text: str


@dataclass
class AuditReport:
    demo_directory_exists: bool = False
    findings: list[Finding] = field(default_factory=list)

    @property
    def failed(self) -> bool:
        return self.demo_directory_exists or bool(self.findings)


def is_relative_to(path: Path, prefix: Path) -> bool:
    try:
        path.relative_to(prefix)
    except ValueError:
        return False
    return True


def is_historical(rel: Path) -> bool:
    return rel.name == "history.md" or any(
        is_relative_to(rel, prefix) for prefix in HISTORICAL_PREFIXES
    )


def is_text_file(path: Path) -> bool:
    return path.name in TEXT_NAMES or path.suffix in TEXT_SUFFIXES


def is_excluded(rel: Path) -> bool:
    return (
        rel in SELF_PATHS
        or is_historical(rel)
        or any(part in EXCLUDED_PARTS for part in rel.parts)
    )


def iter_scan_files(repo_root: Path) -> Iterable[Path]:
    for root in SCAN_ROOTS:
        path = repo_root / root
        if not path.exists():
            continue
        if path.is_file():
            rel = path.relative_to(repo_root)
            if not is_excluded(rel) and is_text_file(path):
                yield path
            continue
        for child in sorted(path.rglob("*")):
            if not child.is_file():
                continue
            rel = child.relative_to(repo_root)
            if is_excluded(rel) or not is_text_file(child):
                continue
            yield child


def scan_file(repo_root: Path, path: Path) -> list[Finding]:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []

    rel = path.relative_to(repo_root)
    lines = text.splitlines()
    findings: list[Finding] = []
    for lineno, line in enumerate(lines, start=1):
        context = "\n".join(lines[max(0, lineno - 3) : min(len(lines), lineno + 2)])
        if NEGATIVE_CONTEXT_RE.search(context):
            continue
        for forbidden in FORBIDDEN_PATTERNS:
            if forbidden.pattern.search(line):
                findings.append(
                    Finding(
                        path=rel.as_posix(),
                        line=lineno,
                        label=forbidden.label,
                        text=line.strip(),
                    )
                )
                break
    return findings


def scan_repo(repo_root: Path) -> AuditReport:
    repo_root = repo_root.resolve()
    report = AuditReport(demo_directory_exists=(repo_root / "ui-design").is_dir())
    for path in iter_scan_files(repo_root):
        report.findings.extend(scan_file(repo_root, path))
    return report


def format_report(report: AuditReport) -> str:
    lines = [
        f"ui_demo_directory: {'present' if report.demo_directory_exists else 'absent'}",
        f"active_residuals ({len(report.findings)})",
    ]
    lines.extend(
        f"  {finding.path}:{finding.line} [{finding.label}] {finding.text}"
        for finding in report.findings
    )
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Reject the removed UI Demo and active contracts that depend on it."
    )
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    args = parser.parse_args()

    report = scan_repo(args.repo_root)
    output = format_report(report)
    stream = sys.stderr if report.failed else sys.stdout
    print(output, file=stream)
    return 1 if report.failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
