#!/usr/bin/env python3
"""Relative-link validator for markdown trees.

Walks a directory recursively, parses inline markdown links `[text](target)`,
and verifies that each non-external `target` resolves to an existing path
relative to the source file. External links (http://, https://, mailto:, ftp://,
data:, javascript:) and pure in-page anchors (`#section`) are skipped.

Wired into `make docs-check` per [ci-pipeline-baseline plan §4.1.4](../../docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md).

Run: `python3 scripts/lint/check_md_links.py docs`
"""
from __future__ import annotations

import argparse
import fnmatch
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

INLINE_LINK_RE = re.compile(r"\[(?P<text>[^\]]*)\]\((?P<target>[^)\s]+)(?:\s+\"[^\"]*\")?\)")
HTML_COMMENT_RE = re.compile(r"<!--.*?-->", re.DOTALL)
INLINE_CODE_RE = re.compile(r"`[^`\n]*`")
EXTERNAL_SCHEMES = ("http://", "https://", "mailto:", "ftp://", "data:", "javascript:", "tel:", "//")
FENCE_RE = re.compile(r"^\s*(```|~~~)")


@dataclass(frozen=True)
class Finding:
    source: Path
    line: int
    target: str
    resolved: Path

    def format(self, root: Path) -> str:
        rel = self.source.relative_to(root) if self.source.is_relative_to(root) else self.source
        return f"{rel}:{self.line}: broken link -> {self.target} (resolved={self.resolved})"


def _is_external(target: str) -> bool:
    lowered = target.lower()
    return any(lowered.startswith(scheme) for scheme in EXTERNAL_SCHEMES)


def _strip_fragment(target: str) -> str:
    for cut in ("#", "?"):
        if cut in target:
            target = target.split(cut, 1)[0]
    return target


def _strip_html_comments(text: str) -> str:
    return HTML_COMMENT_RE.sub(lambda m: re.sub(r"[^\n]", " ", m.group(0)), text)


def scan_file(path: Path) -> list[Finding]:
    findings: list[Finding] = []
    try:
        raw = path.read_text(encoding="utf-8")
    except (UnicodeDecodeError, OSError):
        return findings
    text = _strip_html_comments(raw)
    in_fence = False
    for lineno, line in enumerate(text.splitlines(), start=1):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        sanitized = INLINE_CODE_RE.sub(lambda m: " " * len(m.group(0)), line)
        for match in INLINE_LINK_RE.finditer(sanitized):
            target = match.group("target").strip()
            if not target or target.startswith("#") or _is_external(target):
                continue
            stripped = _strip_fragment(target)
            if not stripped:
                continue
            resolved = (path.parent / stripped).resolve()
            if not resolved.exists():
                findings.append(
                    Finding(source=path, line=lineno, target=target, resolved=resolved)
                )
    return findings


def _matches_ignore(rel: Path, ignores: list[str]) -> bool:
    rel_str = rel.as_posix()
    name = rel.name
    for pat in ignores:
        if fnmatch.fnmatch(rel_str, pat):
            return True
        if fnmatch.fnmatch(name, pat):
            return True
        # `**/` prefix should also match a file at the root with no parent dirs.
        for prefix in ("**/", "*/"):
            if pat.startswith(prefix) and fnmatch.fnmatch(name, pat[len(prefix):]):
                return True
    return False


def iter_markdown_files(root: Path, ignores: list[str] | None = None) -> Iterable[Path]:
    ignores = ignores or []
    for path in sorted(root.rglob("*.md")):
        rel = path.relative_to(root)
        if any(part.startswith(".") for part in rel.parts):
            continue
        if ignores and _matches_ignore(rel, ignores):
            continue
        yield path


def scan_directory(root: Path, ignores: list[str] | None = None) -> list[Finding]:
    findings: list[Finding] = []
    for md in iter_markdown_files(root, ignores):
        findings.extend(scan_file(md))
    return findings


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", help="Directory to scan recursively for *.md files")
    parser.add_argument(
        "--ignore",
        action="append",
        default=[],
        help="Glob pattern of relative path or basename to exclude (repeatable, e.g. '**/TEMPLATES.md')",
    )
    args = parser.parse_args(argv)

    root = Path(args.root).resolve()
    if not root.is_dir():
        print(f"ERROR: not a directory: {root}", file=sys.stderr)
        return 2

    findings = scan_directory(root, ignores=args.ignore)
    if not findings:
        print(f"check_md_links: OK ({root})")
        return 0

    for finding in findings:
        print(finding.format(root), file=sys.stderr)
    print(f"check_md_links: {len(findings)} broken relative link(s) in {root}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
