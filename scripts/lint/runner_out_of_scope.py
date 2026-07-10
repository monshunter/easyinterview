#!/usr/bin/env python3
"""Out-of-scope gate for backend-async-runner (spec D-12).

After plan 001 consolidates every in-process drainer / runner into the single
runner.Runtime kernel, the out-of-scope entry points below must not reappear.
Production entry-point checks exclude tests; the canonical targetjob contract
check also scans tests so a second runner model cannot survive as test support.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

# Forbidden production references. Each entry is a compiled regex matched line
# by line against non-test Go sources under backend/.
FORBIDDEN_PATTERNS = (
    (re.compile(r"\b(?:review|reviewdomain|domainreview)\.NewRunner\b"), "out-of-scope review.NewRunner (use runner.Runtime + review.GenerateHandler)"),
    (re.compile(r"\b(?:review|reviewdomain|domainreview)\.NewReaper\b"), "out-of-scope review.NewReaper (use runner.Runtime reaper)"),
    (re.compile(r"\bComputeReportFailureBackoff\b"), "out-of-scope ComputeReportFailureBackoff (use runner.BackoffPolicy)"),
    (re.compile(r"\bDefaultReportFailureBackoff\b"), "out-of-scope DefaultReportFailureBackoff (use runner.BackoffPolicy)"),
    (re.compile(r"\bBackgroundMailDispatcher\b"), "out-of-scope auth.BackgroundMailDispatcher (use EmailDispatchEnqueuer + EmailDispatchHandler)"),
    (re.compile(r"\bNewBackgroundMailDispatcher\b"), "out-of-scope auth.NewBackgroundMailDispatcher (use EmailDispatchEnqueuer)"),
    (re.compile(r"\btargetjob\.NewDrainer\("), "out-of-scope targetjob.NewDrainer instantiation (register handlers on runner.Runtime)"),
    (re.compile(r"\b(?:jdmatchRuntime|resumeRuntime|targetJobRuntime)\.Drainer\b"), "out-of-scope per-domain Drainer field (use runner.Runtime)"),
    (re.compile(r"\breportRuntime\.(?:Runner|Reaper)\b"), "out-of-scope reportRuntime Runner/Reaper field (use runner.Runtime)"),
)

# Production Go roots that must stay clean. Tests are excluded by suffix.
SCAN_ROOTS = (
    ("backend", "cmd"),
    ("backend", "internal"),
)

EXCLUDED_PARTS = {".git", "node_modules", "dist", "coverage", ".test-output", "__pycache__"}

REMOVED_TARGETJOB_PATHS = (
    ("backend", "internal", "targetjob", "drainer.go"),
    ("backend", "internal", "runner", "adapter_targetjob.go"),
    ("backend", "internal", "runner", "adapter_targetjob_test.go"),
)

TARGETJOB_LOCAL_PATTERNS = (
    (re.compile(r"\btype\s+(?:ClaimedJob|JobOutcome|JobHandler|JobHandlerFunc|AsyncJobStore|DrainerOptions|Drainer)\b"), "duplicate targetjob async job type"),
    (re.compile(r"\bfunc\s+NewDrainer\b"), "targetjob.NewDrainer"),
    (re.compile(r"\bClaimNextAsyncJob\b"), "duplicate targetjob ClaimNextAsyncJob API"),
    (re.compile(r"\bFinalizeAsyncJob\b"), "duplicate targetjob FinalizeAsyncJob API"),
)

QUALIFIED_TARGETJOB_PATTERN = re.compile(
    r"\btargetjob\.(?:ClaimedJob|JobOutcome|JobHandler|JobHandlerFunc|AsyncJobStore|DrainerOptions|Drainer|NewDrainer)\b"
)
ADAPTER_PATTERN = re.compile(r"\b(?:runner\.)?FromTargetjobHandler\b")


def is_production_go(path: Path) -> bool:
    if path.suffix != ".go":
        return False
    if path.name.endswith("_test.go"):
        return False
    return True


def iter_scan_files(repo_root: Path) -> list[Path]:
    files: list[Path] = []
    for prefix in SCAN_ROOTS:
        root = repo_root.joinpath(*prefix)
        if not root.exists():
            continue
        for path in root.rglob("*.go"):
            if path.is_symlink() or not path.is_file():
                continue
            if any(part in EXCLUDED_PARTS for part in path.parts):
                continue
            if is_production_go(path):
                files.append(path)
    return sorted(set(files))


def scan_paths(paths: list[Path]) -> list[str]:
    problems: list[str] = []
    for path in paths:
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            for pattern, message in FORBIDDEN_PATTERNS:
                if pattern.search(line):
                    problems.append(f"{path}:{lineno}: {message}")
    return problems


def scan_removed_targetjob_contract(repo_root: Path) -> list[str]:
    problems: list[str] = []
    for parts in REMOVED_TARGETJOB_PATHS:
        path = repo_root.joinpath(*parts)
        if path.is_file():
            problems.append(f"{path}: removed runner contract file still exists")

    api_root = repo_root / "backend" / "cmd" / "api"
    if api_root.exists():
        for path in sorted(api_root.glob("*_drainer_scenario_test.go")):
            problems.append(f"{path}: drainer_scenario_test.go name must use runner kernel terminology")

    backend_root = repo_root / "backend"
    if not backend_root.exists():
        return problems
    targetjob_root = backend_root / "internal" / "targetjob"
    for path in sorted(backend_root.rglob("*.go")):
        if path.is_symlink() or not path.is_file() or any(part in EXCLUDED_PARTS for part in path.parts):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except UnicodeDecodeError:
            continue
        for lineno, line in enumerate(lines, start=1):
            if QUALIFIED_TARGETJOB_PATTERN.search(line):
                problems.append(f"{path}:{lineno}: targetjob async job contract must use runner.ClaimedJob / runner.JobOutcome")
            if ADAPTER_PATTERN.search(line):
                problems.append(f"{path}:{lineno}: FromTargetjobHandler adapter must not exist")
            if path == targetjob_root or targetjob_root in path.parents:
                for pattern, message in TARGETJOB_LOCAL_PATTERNS:
                    if pattern.search(line):
                        problems.append(f"{path}:{lineno}: {message}")
    return problems


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(ROOT), help="repository root")
    parser.add_argument("--phase", default="all", choices=["all"], help="reserved for future phased gates")
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    problems = scan_paths(iter_scan_files(repo_root))
    problems.extend(scan_removed_targetjob_contract(repo_root))
    if problems:
        for problem in problems:
            print(f"FAIL: {problem}", file=sys.stderr)
        return 1
    print("OK: backend-async-runner has one canonical runtime, handler contract, and async job SQL owner")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
