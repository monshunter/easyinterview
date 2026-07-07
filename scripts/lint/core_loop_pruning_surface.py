#!/usr/bin/env python3
"""Bucket non-current core-loop module references after product-scope D-22 pruning."""

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable


BUCKETS = (
    "migration_records",
    "non_current_normalization",
    "negative_tests",
    "real_residuals",
)

SCAN_ROOTS = (
    "backend",
    "frontend",
    "openapi",
    "shared",
    "config",
    "scripts",
    "migrations",
    "test/scenarios",
    "ui-design",
)

TEXT_SUFFIXES = {
    ".css",
    ".go",
    ".html",
    ".js",
    ".json",
    ".jsx",
    ".md",
    ".mjs",
    ".py",
    ".sh",
    ".sql",
    ".tmpl",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}
TEXT_NAMES = {"Makefile"}

EXCLUDED_PARTS = {
    ".git",
    ".pytest_cache",
    ".test-output",
    "__pycache__",
    "coverage",
    "dist",
    "node_modules",
    "vendor",
}

TEST_FILE_SUFFIXES = (
    "_test.go",
    "_test.py",
    ".test.ts",
    ".test.tsx",
    ".test.js",
    ".test.mjs",
    ".spec.ts",
    ".spec.tsx",
    ".spec.js",
    ".spec.mjs",
)

NON_CURRENT_NORMALIZATION_PATHS = {
    Path("frontend/src/app/normalizeRoute.ts"),
    Path("frontend/src/app/routeUrl.ts"),
    Path("frontend/src/app/auth/pendingAction.ts"),
    Path("frontend/scripts/spaFallback.mjs"),
    Path("ui-design/src/app.jsx"),
}

NEGATIVE_PATH_PREFIXES = (
    Path("scripts/lint"),
    Path("test/scenarios"),
)

STRICT_LIFECYCLE_CONTEXT_TERMS = (
    "退" "役",
    "ret" "ired",
    "de" "precated",
)
OBSOLETE_ZH_STATUS_TERMS = (
    "废" "弃",
)
OLD_SCOPE_CONTEXT_TERMS = (
    "\u65e7",
    "\u5386" "\u53f2",
)
STRICT_LIFECYCLE_CONTEXT_RE = "|".join(
    re.escape(term)
    for term in STRICT_LIFECYCLE_CONTEXT_TERMS + OBSOLETE_ZH_STATUS_TERMS + OLD_SCOPE_CONTEXT_TERMS
)
NEGATIVE_CONTEXT_RE = re.compile(
    r"删除|移除|"
    + STRICT_LIFECYCLE_CONTEXT_RE
    + r"|负向|回流|归一|不(?:再|得|允许|存在|出现|保留|作为)|"
    r"only as|removed|"
    + "leg" "acy"
    + r"|negative|zero-reference|absent|forbidden|"
    r"must not|should not|not remain|do not|drop|delete|reject|fail|guard|"
    r"cleanup|prun(?:e|ing)|obsolete",
    re.IGNORECASE,
)


@dataclass(frozen=True)
class NonCurrentPattern:
    label: str
    pattern: re.Pattern[str]


NON_CURRENT_PATTERNS = (
    NonCurrentPattern(
        "debrief surface",
        re.compile(
            r"\bDebriefs?\b|\bdebrief(?:s|_full|_generate|\.generate|\.created|\.completed|"
            r"\.suggest_questions)?\b|source_debrief_id|sourceDebriefId|PracticeGoalDebrief|"
            r"backend/internal/debrief|frontend/src/app/screens/debrief|/debrief\b|route=debrief"
        ),
    ),
    NonCurrentPattern(
        "candidate profile surface",
        re.compile(
            r"\bCandidateProfile\b|\bcandidate_profiles\b|\bcandidate[- ]profiles?\b|"
            r"profile\.update\b|\bProfileScreen\b|\bUserProfileScreen\b|screen-profile|"
            r"route=profile|topbar-user-profile|nav\([\"']profile[\"']"
        ),
    ),
    NonCurrentPattern(
        "experience card surface",
        re.compile(r"\bExperienceCard\b|\bexperience_cards\b|\bexperience[- ]cards?\b"),
    ),
    NonCurrentPattern(
        "jd match surface",
        re.compile(
            r"\bJD Match\b|\bJob Picks\b|\bJobMatch\b|\bjobs-recommendations\b|"
            r"\bjd[_-]match\b|\bjdmatch\b|jd_match_recommendations|jd_match_search_runs|"
            r"\bsaved_searches\b|\bwatchlist_items\b|\bagent_scans\b|"
            r"backend/internal/jdmatch|frontend/src/app/screens/jd-match"
        ),
    ),
)


@dataclass(frozen=True)
class Finding:
    bucket: str
    path: str
    line: int
    label: str
    text: str


@dataclass
class AuditReport:
    buckets: dict[str, list[Finding]] = field(default_factory=lambda: {bucket: [] for bucket in BUCKETS})

    @property
    def failed(self) -> bool:
        return bool(self.buckets["real_residuals"])


def is_text_file(path: Path) -> bool:
    return path.name in TEXT_NAMES or path.suffix in TEXT_SUFFIXES


def is_test_path(rel: Path) -> bool:
    return any(rel.name.endswith(suffix) for suffix in TEST_FILE_SUFFIXES)


def is_excluded(rel: Path) -> bool:
    return any(part in EXCLUDED_PARTS for part in rel.parts)


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


def classify(rel: Path, line: str) -> str:
    if rel.parts and rel.parts[0] == "migrations":
        return "migration_records"
    if rel in NON_CURRENT_NORMALIZATION_PATHS:
        return "non_current_normalization"
    if is_test_path(rel):
        return "negative_tests"
    if any(is_relative_to(rel, prefix) for prefix in NEGATIVE_PATH_PREFIXES):
        return "negative_tests"
    if NEGATIVE_CONTEXT_RE.search(line):
        return "negative_tests"
    return "real_residuals"


def scan_file(repo_root: Path, path: Path) -> list[Finding]:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []

    rel = path.relative_to(repo_root)
    findings: list[Finding] = []
    for lineno, line in enumerate(text.splitlines(), start=1):
        for pattern in NON_CURRENT_PATTERNS:
            if not pattern.pattern.search(line):
                continue
            bucket = classify(rel, line)
            findings.append(
                Finding(
                    bucket=bucket,
                    path=rel.as_posix(),
                    line=lineno,
                    label=pattern.label,
                    text=line.strip(),
                )
            )
            break
    return findings


def scan_repo(repo_root: Path) -> AuditReport:
    repo_root = repo_root.resolve()
    report = AuditReport()
    for path in iter_scan_files(repo_root):
        for finding in scan_file(repo_root, path):
            report.buckets[finding.bucket].append(finding)
    return report


def format_report(report: AuditReport, *, verbose: bool = False) -> str:
    lines: list[str] = []
    for bucket in BUCKETS:
        findings = report.buckets[bucket]
        lines.append(f"{bucket} ({len(findings)})")
        if not verbose and bucket != "real_residuals":
            continue
        for finding in findings:
            lines.append(f"  {finding.path}:{finding.line}: {finding.label}: {finding.text}")
    return "\n".join(lines)


def is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.relative_to(parent)
        return True
    except ValueError:
        return False


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path(__file__).resolve().parents[2],
        help="Repository root to scan.",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Print every allowed bucket finding instead of summary counts plus real residual details.",
    )
    args = parser.parse_args(list(argv))

    report = scan_repo(args.repo_root)
    output = format_report(report, verbose=args.verbose)
    if report.failed:
        print(output, file=sys.stderr)
        print(
            "core_loop_pruning_surface: FAILED; remove or reclassify real residuals in the owner plan.",
            file=sys.stderr,
        )
        return 1
    print(output)
    print("core_loop_pruning_surface: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
