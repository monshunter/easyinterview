#!/usr/bin/env python3
"""Generate lightweight Harness indexes from repository facts."""

from __future__ import annotations

import argparse
import json
import re
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

_LINK_PATTERN = re.compile(r"\[[^\]]+\]\(([^)]+)\)")
_HEADER_PATTERN = re.compile(r"^> \*\*(版本|状态|更新日期)\*\*: (.+)$", re.MULTILINE)
_CODE_PATTERN = re.compile(r"`([^`\n]+)`")
_TOKEN_PATTERN = re.compile(r"[A-Za-z0-9][A-Za-z0-9_.:/-]*")
_GENERIC_TOKENS = {
    "change",
    "checklist",
    "context",
    "context.yaml",
    "docs",
    "index",
    "md",
    "plan",
    "plan.md",
    "spec",
    "spec.md",
}


class StaleIndexError(ValueError):
    """Raised when a cache was generated for another repository commit."""


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


def _canonical_documents(spec_root: Path) -> list[tuple[str, str, Path]]:
    documents: list[tuple[str, str, Path]] = []
    for subject_dir in sorted(path for path in spec_root.iterdir() if path.is_dir()):
        spec_path = subject_dir / "spec.md"
        if spec_path.is_file():
            documents.append((subject_dir.name, "spec", spec_path))
        for kind, directory_name in (("change", "changes"), ("decision", "decisions")):
            directory = subject_dir / directory_name
            if directory.is_dir():
                documents.extend((subject_dir.name, kind, path) for path in sorted(directory.glob("*.md")))
    return documents


def _document_record(repo_root: Path, subject: str, kind: str, path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    title = next((line[2:].strip() for line in text.splitlines() if line.startswith("# ")), path.stem)
    headers = {key: value.strip() for key, value in _HEADER_PATTERN.findall(text)}
    links = sorted(set(match.strip() for match in _LINK_PATTERN.findall(text) if not match.startswith(("http://", "https://", "#"))))
    identifiers = sorted(
        {
            value.strip()
            for value in _CODE_PATTERN.findall(text)
            if 2 < len(value.strip()) <= 120 and " " not in value.strip()
        }
    )
    return {
        "path": path.relative_to(repo_root).as_posix(),
        "subject": subject,
        "kind": kind,
        "title": title,
        "version": headers.get("版本"),
        "status": headers.get("状态"),
        "updated": headers.get("更新日期"),
        "links": links,
        "identifiers": identifiers,
        "bytes": len(text.encode("utf-8")),
    }


def build_index(repo_root: Path) -> dict[str, Any]:
    """Build a disposable index from canonical checked-in documents."""

    root = repo_root.resolve()
    spec_root = root / "docs" / "spec"
    documents = [
        _document_record(root, subject, kind, path)
        for subject, kind, path in _canonical_documents(spec_root)
    ]
    return {
        "schema": "easyinterview.harness-index/v1",
        "git_commit": _git_commit(root),
        "documents": documents,
    }


def write_index_cache(index: dict[str, Any], cache_path: Path) -> None:
    cache_path.parent.mkdir(parents=True, exist_ok=True)
    cache_path.write_text(json.dumps(index, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def load_index_cache(cache_path: Path, *, expected_commit: str) -> dict[str, Any]:
    index = json.loads(cache_path.read_text(encoding="utf-8"))
    actual_commit = index.get("git_commit")
    if actual_commit != expected_commit:
        raise StaleIndexError(
            f"harness index is stale: cache={actual_commit!r}, expected={expected_commit!r}; rebuild it"
        )
    return index


def _query_tokens(value: str) -> set[str]:
    return {token.lower() for token in _TOKEN_PATTERN.findall(value)}


def route_query(index: dict[str, Any], query: str, *, limit: int = 3) -> dict[str, Any]:
    """Route a query using exact evidence first and expose uncertainty."""

    normalized = query.strip().lower()
    query_tokens = _query_tokens(query)
    meaningful_tokens = query_tokens - _GENERIC_TOKENS
    ranked: list[dict[str, Any]] = []

    for document in index.get("documents", []):
        score = 0
        reasons: list[str] = []
        subject = str(document.get("subject", "")).lower()
        path = str(document.get("path", ""))
        title_tokens = _query_tokens(str(document.get("title", ""))) - _GENERIC_TOKENS

        if normalized == path.lower() or normalized.endswith(path.lower()):
            score += 120
            reasons.append("exact path")
        if normalized == subject or subject in meaningful_tokens:
            score += 100
            reasons.append("exact subject")

        for identifier_value in document.get("identifiers", []):
            identifier = str(identifier_value).lower()
            if identifier in _GENERIC_TOKENS:
                continue
            if identifier in query_tokens or (len(identifier) >= 4 and identifier in normalized):
                score += 80
                reasons.append(f"exact identifier: {identifier_value}")

        shared_title_tokens = meaningful_tokens & title_tokens
        if shared_title_tokens:
            score += 8 * len(shared_title_tokens)
            reasons.append("title tokens: " + ", ".join(sorted(shared_title_tokens)))

        if score:
            ranked.append(
                {
                    "path": path,
                    "subject": document.get("subject"),
                    "kind": document.get("kind"),
                    "title": document.get("title"),
                    "score": score,
                    "reasons": reasons,
                }
            )

    ranked.sort(key=lambda candidate: (-candidate["score"], candidate["path"]))
    candidates = ranked[: max(1, min(limit, 3))]
    top_score = candidates[0]["score"] if candidates else 0
    second_score = candidates[1]["score"] if len(candidates) > 1 else 0
    top_has_exact = bool(candidates) and any(
        reason.startswith(("exact path", "exact subject", "exact identifier"))
        for reason in candidates[0]["reasons"]
    )
    confidence = "high" if top_has_exact and top_score - second_score >= 20 else "low"
    return {
        "query": query,
        "confidence": confidence,
        "requires_user_choice": confidence == "low",
        "candidates": candidates,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    parser.add_argument("command", choices=("index", "locate", "audit-legacy", "baseline"))
    parser.add_argument("--cache", type=Path)
    parser.add_argument("--query")
    parser.add_argument("--limit", type=int, default=3)
    args = parser.parse_args()

    if args.command == "index":
        result = build_index(args.repo_root)
        if args.cache:
            write_index_cache(result, args.cache)
    elif args.command == "locate":
        if not args.query:
            parser.error("locate requires --query")
        result = route_query(build_index(args.repo_root), args.query, limit=args.limit)
    elif args.command == "audit-legacy":
        result = audit_legacy_structure(args.repo_root)
    else:
        result = collect_legacy_baseline(args.repo_root)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
