#!/usr/bin/env python3
"""Reject out-of-scope standalone backend worker process terminology.

The P0 runtime topology is frontend + backend. Prior evidence and the
backend-runtime-topology owner docs may still mention out-of-scope worker terms, but
runtime code, active handoff plans, config, deploy assets, and generated
contracts must use backend internal runner / backend_async wording.
"""
from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml


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
class OutOfScopePattern:
    label: str
    pattern: re.Pattern[str]


OUT_OF_SCOPE_PATTERNS = [
    OutOfScopePattern(
        "cmd/worker entrypoint",
        re.compile(r"(?:\./|backend/)?cmd/(?:\{api,)?worker(?=[,}'\"\/\s]|$)|\./cmd/worker\b"),
    ),
    OutOfScopePattern(
        "worker listen addr config",
        re.compile(
            r"\bWORKER_LISTEN_ADDR\b|\bworker\.listenAddr\b|\bapp/worker listen addr\b",
            re.IGNORECASE,
        ),
    ),
    OutOfScopePattern(
        "worker config bindings",
        re.compile(r"\bworker bindings?\b", re.IGNORECASE),
    ),
    OutOfScopePattern(
        "worker build target",
        re.compile(r"\bbuild-worker\b"),
    ),
    OutOfScopePattern(
        "worker producer enum",
        re.compile(
            r"(?:\bproducer\b.*(?:[`\"']?worker[`\"']?\b|/ worker\b|\bworker\s*/))|"
            r"(?:(?:[`\"']?worker[`\"']?\b|/ worker\b|\bworker\s*/).*\bproducer\b)"
        ),
    ),
    OutOfScopePattern(
        "backend async runner subject shorthand",
        re.compile(r"\bbackend-async-runtime\b"),
    ),
    OutOfScopePattern(
        "worker component probe",
        re.compile(r"worker 类组件"),
    ),
    OutOfScopePattern(
        "privacy worker wording",
        re.compile(r"\bprivacy workers?\b|\bC8 worker\b", re.IGNORECASE),
    ),
    OutOfScopePattern(
        "worker observability wording",
        re.compile(r"\bF1 worker span\b", re.IGNORECASE),
    ),
    OutOfScopePattern(
        "Asynq worker wording",
        re.compile(r"\bAsynq worker\b"),
    ),
]

STRUCTURED_PRODUCER_ROOTS = (
    Path("shared"),
)
STRUCTURED_WORKER_LINE_RE = re.compile(r"(?:-\s*)?[`\"']?worker[`\"']?,?\s*$")
PRODUCER_CONTEXT_RE = re.compile(r"\bproducer\b", re.IGNORECASE)
OWNER_CURRENT_CUE_RE = re.compile(
    r"\bcurrent\b|\bhandoff\b|\bbuild\b|\brun\b|\bruntime\b|\bentry\b|"
    r"\bverification command\b|\bcommand=|当前|构建|运行|执行|入口|验证[:：]",
    re.IGNORECASE,
)
STRICT_LIFECYCLE_CONTEXT_TERMS = (
    "退" "役",
    "ret" "ired",
    "de" "precated",
)
OLD_SCOPE_CONTEXT_TERMS = (
    "\u65e7",
    "\u5386" "\u53f2",
)
STRICT_LIFECYCLE_CONTEXT_RE = "|".join(
    re.escape(term) for term in STRICT_LIFECYCLE_CONTEXT_TERMS + OLD_SCOPE_CONTEXT_TERMS
)
OWNER_NEGATIVE_CONTEXT_RE = re.compile(
    r"删除|移除|取消|不(?:再|得|保留|存在|构建|要求|作为|新增)|无独立|"
    r"负向|回流|拦截|"
    + STRICT_LIFECYCLE_CONTEXT_RE
    + r"|removed|must not|not remain|absent|"
    r"zero-reference|negative|not-retained|assertions?|rejects?|fails?|failed|omits?",
    re.IGNORECASE,
)


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
    if rel.parts[:3] == ("docs", "spec", "backend-runtime-topology") and not is_owner_plan_handoff_path(rel):
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
    if is_owner_plan_handoff_path(rel):
        return scan_owner_current_handoff_file(rel, text)
    for lineno, line in enumerate(text.splitlines(), start=1):
        for pattern in OUT_OF_SCOPE_PATTERNS:
            if pattern.pattern.search(line):
                findings.append(f"{rel}:{lineno}: {pattern.label}: {line.strip()}")
                break
    findings.extend(scan_structured_producer_values(repo, path, text, findings))
    return findings


def scan_structured_producer_values(repo: Path, path: Path, text: str, existing: list[str]) -> list[str]:
    rel = path.relative_to(repo)
    if path.suffix not in {".json", ".yaml", ".yml"}:
        return []
    if not any(is_relative_to(rel, root) for root in STRUCTURED_PRODUCER_ROOTS):
        return []
    try:
        data = yaml.safe_load(text)
    except yaml.YAMLError:
        return []
    if not contains_out_of_scope_producer_value(data):
        return []
    if any(finding.startswith(f"{rel}:") and ": worker producer enum:" in finding for finding in existing):
        return []

    lineno = find_structured_worker_lineno(text)
    finding = f"{rel}:{lineno}: worker producer enum: structured producer value contains out-of-scope worker"
    return [finding]


def contains_out_of_scope_producer_value(value: Any) -> bool:
    if isinstance(value, dict):
        string_keys = {str(key): item for key, item in value.items()}
        if string_keys.get("name") == "producer" and contains_worker_scalar(string_keys.get("values")):
            return True
        if "producer" in string_keys and contains_worker_scalar(string_keys["producer"]):
            return True
        return any(contains_out_of_scope_producer_value(item) for item in value.values())
    if isinstance(value, list):
        return any(contains_out_of_scope_producer_value(item) for item in value)
    return False


def contains_worker_scalar(value: Any) -> bool:
    if isinstance(value, str):
        return value.strip() == "worker"
    if isinstance(value, dict):
        return any(contains_worker_scalar(item) for item in value.values())
    if isinstance(value, list):
        return any(contains_worker_scalar(item) for item in value)
    return False


def find_structured_worker_lineno(text: str) -> int:
    lines = text.splitlines()
    for index, line in enumerate(lines):
        if not STRUCTURED_WORKER_LINE_RE.search(line.strip()):
            continue
        window = "\n".join(lines[max(0, index - 12) : index + 1])
        if PRODUCER_CONTEXT_RE.search(window):
            return index + 1
    for index, line in enumerate(lines):
        if PRODUCER_CONTEXT_RE.search(line):
            return index + 1
    return 1


def is_relative_to(path: Path, base: Path) -> bool:
    try:
        path.relative_to(base)
    except ValueError:
        return False
    return True


def is_owner_plan_handoff_path(rel: Path) -> bool:
    return (
        rel.parts[:4] == ("docs", "spec", "backend-runtime-topology", "plans")
        and rel.name in {"plan.md", "checklist.md"}
    )


def scan_owner_current_handoff_file(rel: Path, text: str) -> list[str]:
    findings: list[str] = []
    for lineno, line in enumerate(text.splitlines(), start=1):
        if OWNER_NEGATIVE_CONTEXT_RE.search(line):
            continue
        if not OWNER_CURRENT_CUE_RE.search(line):
            continue
        for pattern in OUT_OF_SCOPE_PATTERNS:
            if pattern.pattern.search(line):
                findings.append(f"{rel}:{lineno}: owner current handoff: {pattern.label}: {line.strip()}")
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
        print("runtime_topology: out-of-scope worker process terminology found", file=sys.stderr)
        for finding in findings:
            print(f"  - {finding}", file=sys.stderr)
        print(
            "Fix: use backend internal runner / backend_async wording; keep out-of-scope "
            "worker process terms only in history, tests, or backend-runtime-topology owner docs.",
            file=sys.stderr,
        )
        return 1

    print(f"runtime_topology: OK ({len(files)} active files scanned)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
