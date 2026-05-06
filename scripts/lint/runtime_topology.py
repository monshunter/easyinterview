#!/usr/bin/env python3
"""Reject retired standalone backend worker process terminology.

The P0 runtime topology is frontend + backend. Historical evidence and the
backend-runtime-topology owner docs may still mention retired worker terms, but
runtime code, active handoff plans, config, deploy assets, and generated
contracts must use backend internal runner / backend_async wording.
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
    ".py",
    ".sh",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}

TEXT_NAMES = {
    "Makefile",
}

SCAN_PATHS = [
    ".env.example",
    "Makefile",
    "backend",
    "frontend",
    "shared",
    "scripts",
    "config",
    "deploy/dev-stack",
    "docs/spec",
]

EXCLUDED_DIR_PARTS = {
    ".git",
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
    ".spec.ts",
    ".spec.tsx",
)


@dataclass(frozen=True)
class RetiredPattern:
    label: str
    pattern: re.Pattern[str]


RETIRED_PATTERNS = [
    RetiredPattern(
        "cmd/worker entrypoint",
        re.compile(r"(?:\./|backend/)?cmd/(?:\{api,)?worker(?=[,}'\"\/\s]|$)|\./cmd/worker\b"),
    ),
    RetiredPattern(
        "worker listen addr config",
        re.compile(
            r"\bWORKER_LISTEN_ADDR\b|\bworker\.listenAddr\b|\bapp/worker listen addr\b",
            re.IGNORECASE,
        ),
    ),
    RetiredPattern(
        "worker config bindings",
        re.compile(r"\bworker bindings?\b", re.IGNORECASE),
    ),
    RetiredPattern(
        "worker build target",
        re.compile(r"\bbuild-worker\b"),
    ),
    RetiredPattern(
        "worker producer enum",
        re.compile(
            r"(?:\bproducer\b.*(?:[`\"']?worker[`\"']?\b|/ worker\b|\bworker\s*/))|"
            r"(?:(?:[`\"']?worker[`\"']?\b|/ worker\b|\bworker\s*/).*\bproducer\b)"
        ),
    ),
    RetiredPattern(
        "backend async runner subject shorthand",
        re.compile(r"\bbackend-async-runtime\b"),
    ),
    RetiredPattern(
        "worker component probe",
        re.compile(r"worker 类组件"),
    ),
    RetiredPattern(
        "privacy worker wording",
        re.compile(r"\bprivacy workers?\b|\bC8 worker\b", re.IGNORECASE),
    ),
    RetiredPattern(
        "worker observability wording",
        re.compile(r"\bF1 worker span\b", re.IGNORECASE),
    ),
    RetiredPattern(
        "Asynq worker wording",
        re.compile(r"\bAsynq worker\b"),
    ),
]


def is_text_path(path: Path) -> bool:
    return path.name in TEXT_NAMES or path.suffix in TEXT_SUFFIXES


def is_test_path(path: Path) -> bool:
    return any(path.name.endswith(suffix) for suffix in TEST_FILE_SUFFIXES)


def is_excluded_path(repo: Path, path: Path) -> bool:
    rel = path.relative_to(repo)
    if any(part in EXCLUDED_DIR_PARTS for part in rel.parts):
        return True
    if is_test_path(path):
        return True
    if rel == Path("scripts/lint/runtime_topology.py"):
        return True
    if rel.parts[:3] == ("docs", "spec", "backend-runtime-topology"):
        return True
    if path.name == "history.md":
        return True
    return False


def iter_scan_files(repo: Path) -> list[Path]:
    files: list[Path] = []
    for rel in SCAN_PATHS:
        path = repo / rel
        if not path.exists():
            continue
        if path.is_file():
            if not is_excluded_path(repo, path) and is_text_path(path):
                files.append(path)
            continue
        for child in sorted(path.rglob("*")):
            if not child.is_file():
                continue
            if is_excluded_path(repo, child):
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
    files = iter_scan_files(repo)
    for path in files:
        findings.extend(scan_file(repo, path))

    if findings:
        print("runtime_topology: retired worker process terminology found", file=sys.stderr)
        for finding in findings:
            print(f"  - {finding}", file=sys.stderr)
        print(
            "Fix: use backend internal runner / backend_async wording; keep retired "
            "worker process terms only in history, tests, or backend-runtime-topology owner docs.",
            file=sys.stderr,
        )
        return 1

    print(f"runtime_topology: OK ({len(files)} active files scanned)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
