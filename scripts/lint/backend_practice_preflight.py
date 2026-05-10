#!/usr/bin/env python3
"""Backend-practice Phase 0 preflight checks.

This gate verifies that F3 prompt/rubric baseline plan 001 is completed before
backend-practice starts AI-dependent implementation, then delegates the
runtime Resolve assertions to a focused Go test.
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path


STATUS_RE = re.compile(r"^> \*\*状态\*\*: ([a-z_]+)\s*$", re.MULTILINE)


def header_status(path: Path) -> str:
    match = STATUS_RE.search(path.read_text(encoding="utf-8"))
    if not match:
        raise ValueError(f"{path}: missing Header status")
    return match.group(1)


def validate_completed_headers(paths: list[Path]) -> list[str]:
    problems: list[str] = []
    for path in paths:
        try:
            status = header_status(path)
        except ValueError as exc:
            problems.append(str(exc))
            continue
        if status != "completed":
            problems.append(f"{path}: status {status} != completed")
    return problems


def run_go_preflight(repo_root: Path) -> int:
    return subprocess.run(
        [
            "go",
            "test",
            "./internal/ai/registry",
            "-run",
            "TestBackendPracticeF3Preflight",
            "-count=1",
        ],
        cwd=repo_root / "backend",
        check=False,
    ).returncode


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--skip-go", action="store_true", help="only validate document headers")
    args = parser.parse_args(argv)

    repo_root = Path(args.repo_root).resolve()
    f3_plan = repo_root / "docs/spec/prompt-rubric-registry/plans/001-baseline/plan.md"
    f3_checklist = repo_root / "docs/spec/prompt-rubric-registry/plans/001-baseline/checklist.md"
    problems = validate_completed_headers([f3_plan, f3_checklist])
    if problems:
        for problem in problems:
            print(f"ERROR: {problem}", file=sys.stderr)
        return 1

    if not args.skip_go:
        code = run_go_preflight(repo_root)
        if code != 0:
            return code

    print("backend_practice_preflight: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
