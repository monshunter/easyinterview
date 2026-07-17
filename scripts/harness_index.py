#!/usr/bin/env python3
"""Generate lightweight Harness indexes from repository facts."""

from __future__ import annotations

import subprocess
from pathlib import Path
from typing import Any


REQUIRED_QUALITY_BOUNDARIES = {
    "executable-tests",
    "real-e2e",
    "contract-owner",
    "backend-persistence",
    "current-evidence",
    "high-risk-confirmation",
    "security-privacy",
    "failure-recovery",
}

_LEGACY_NAMES = {
    "context.yaml": "context-manifest",
    "plan.md": "plan-wrapper",
    "checklist.md": "checklist-wrapper",
    "bdd-plan.md": "bdd-wrapper",
    "bdd-checklist.md": "bdd-wrapper",
    "test-plan.md": "plan-wrapper",
    "test-checklist.md": "checklist-wrapper",
    "history.md": "history-wrapper",
    "INDEX.md": "layer-index",
}


def _git_commit(repo_root: Path) -> str:
    result = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=repo_root,
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout.strip()


def collect_legacy_baseline(repo_root: Path) -> dict[str, Any]:
    """Return the checked-in legacy document cost before migration."""

    root = repo_root.resolve()
    spec_root = root / "docs" / "spec"
    markdown_files = tuple(spec_root.rglob("*.md"))

    return {
        "repo_root": str(root),
        "git_commit": _git_commit(root),
        "subjects": sum(1 for path in spec_root.iterdir() if (path / "spec.md").is_file()),
        "context_files": sum(1 for _ in spec_root.rglob("context.yaml")),
        "plan_files": sum(1 for _ in spec_root.rglob("plan.md")),
        "checklist_files": sum(1 for _ in spec_root.rglob("checklist.md")),
        "bdd_files": sum(1 for path in markdown_files if path.name in {"bdd-plan.md", "bdd-checklist.md"}),
        "layer_indexes": sum(1 for _ in spec_root.rglob("INDEX.md")),
        "spec_bytes": sum(path.stat().st_size for path in spec_root.rglob("*") if path.is_file()),
    }


def audit_legacy_structure(repo_root: Path) -> list[dict[str, str]]:
    """List checked-in document wrappers forbidden by the target structure."""

    root = repo_root.resolve()
    spec_root = root / "docs" / "spec"
    violations: list[dict[str, str]] = []
    for path in sorted(spec_root.rglob("*")):
        kind = _LEGACY_NAMES.get(path.name)
        if path.is_file() and kind:
            violations.append({"kind": kind, "path": path.relative_to(root).as_posix()})
    return violations
