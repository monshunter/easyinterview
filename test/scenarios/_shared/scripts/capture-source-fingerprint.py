#!/usr/bin/env python3
"""Capture a deterministic source/worktree fingerprint for screenshot evidence."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
from pathlib import Path


def git_output(repo_root: Path, *args: str) -> bytes:
    return subprocess.check_output(
        ["git", "-C", str(repo_root), *args],
        stderr=subprocess.DEVNULL,
    )


def normalize_source_paths(
    repo_root: Path,
    raw_paths: list[str],
) -> list[str]:
    normalized: set[str] = set()
    for raw_path in raw_paths:
        candidate = (repo_root / raw_path).resolve()
        try:
            relative = candidate.relative_to(repo_root)
        except ValueError as exc:
            raise SystemExit(f"source path escapes repository: {raw_path}") from exc
        if not candidate.exists() and not candidate.is_symlink():
            raise SystemExit(f"source path does not exist: {raw_path}")
        normalized.add(relative.as_posix())
    if not normalized:
        raise SystemExit("at least one source path is required")
    return sorted(normalized)


def source_files(repo_root: Path, source_paths: list[str]) -> list[Path]:
    files: set[Path] = set()
    for source_path in source_paths:
        candidate = repo_root / source_path
        if candidate.is_file() or candidate.is_symlink():
            files.add(candidate)
            continue
        files.update(path for path in candidate.rglob("*") if path.is_file() or path.is_symlink())
    return sorted(files, key=lambda path: path.relative_to(repo_root).as_posix())


def hash_sources(repo_root: Path, source_paths: list[str]) -> str:
    digest = hashlib.sha256()
    digest.update(b"easyinterview-source-fingerprint.v1\0")
    for path in source_files(repo_root, source_paths):
        relative = path.relative_to(repo_root).as_posix().encode("utf-8")
        if path.is_symlink():
            content = os.readlink(path).encode("utf-8")
            kind = b"symlink"
        else:
            content = path.read_bytes()
            kind = b"file"
        digest.update(len(relative).to_bytes(8, "big"))
        digest.update(relative)
        digest.update(kind)
        digest.update(len(content).to_bytes(8, "big"))
        digest.update(content)
    return digest.hexdigest()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    source_group = parser.add_mutually_exclusive_group(required=True)
    source_group.add_argument("--path", action="append", dest="source_paths")
    source_group.add_argument("--source-paths-from", type=Path)
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo_root = args.repo_root.resolve()
    if args.source_paths_from:
        source_manifest = json.loads(args.source_paths_from.read_text(encoding="utf-8"))
        raw_paths = source_manifest.get("source_paths")
        if not isinstance(raw_paths, list) or not all(isinstance(path, str) for path in raw_paths):
            raise SystemExit("source fingerprint manifest has invalid source_paths")
    else:
        raw_paths = args.source_paths
    source_paths = normalize_source_paths(repo_root, raw_paths)
    status = git_output(repo_root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
    payload = {
        "schema_version": "source-fingerprint.v1",
        "git_head": git_output(repo_root, "rev-parse", "HEAD").decode("ascii").strip(),
        "git_dirty": bool(status),
        "git_status_sha256": hashlib.sha256(status).hexdigest(),
        "source_sha256": hash_sources(repo_root, source_paths),
        "source_paths": source_paths,
    }
    output = args.output.resolve()
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(payload, ensure_ascii=True, indent=2) + "\n", encoding="utf-8")
    print(
        "SOURCE_FINGERPRINT_CAPTURED "
        f"head={payload['git_head']} dirty={str(payload['git_dirty']).lower()} "
        f"source_sha256={payload['source_sha256']}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
