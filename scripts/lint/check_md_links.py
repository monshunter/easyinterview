#!/usr/bin/env python3
"""Relative-link validator for markdown trees.

Walks a directory recursively, parses inline markdown links `[text](target)`,
and verifies that each non-external `target` resolves to an existing path
relative to the source file. External links (http://, https://, mailto:, ftp://,
data:, javascript:) are skipped. Fragment checks are opt-in via
`--check-fragments` so legacy path-only calls keep their existing behavior.

Wired into `make docs-check` per [ci-pipeline-baseline plan §4.1.4](../../docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md).

Run: `python3 scripts/lint/check_md_links.py docs`
"""
from __future__ import annotations

import argparse
import fnmatch
import re
import sys
import unicodedata
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable
from urllib.parse import unquote

INLINE_LINK_RE = re.compile(r"\[(?P<text>[^\]]*)\]\((?P<target>[^)\s]+)(?:\s+\"[^\"]*\")?\)")
HTML_COMMENT_RE = re.compile(r"<!--.*?-->", re.DOTALL)
INLINE_CODE_RE = re.compile(r"`[^`\n]*`")
EXTERNAL_SCHEMES = ("http://", "https://", "mailto:", "ftp://", "data:", "javascript:", "tel:", "//")
FENCE_RE = re.compile(r"^\s*(```|~~~)")
HEADING_RE = re.compile(r"^\s{0,3}#{1,6}\s+(?P<text>.+?)\s*$")
TRAILING_HASH_RE = re.compile(r"\s+#+\s*$")
MARKDOWN_LINK_TEXT_RE = re.compile(r"\[([^\]]+)\]\([^)]+\)")


@dataclass(frozen=True)
class Finding:
    source: Path
    line: int
    target: str
    resolved: Path
    kind: str = "link"

    def format(self, root: Path) -> str:
        rel = self.source.relative_to(root) if self.source.is_relative_to(root) else self.source
        label = "broken fragment" if self.kind == "fragment" else "broken link"
        return f"{rel}:{self.line}: {label} -> {self.target} (resolved={self.resolved})"


def _is_external(target: str) -> bool:
    lowered = target.lower()
    return any(lowered.startswith(scheme) for scheme in EXTERNAL_SCHEMES)


def _strip_fragment(target: str) -> str:
    for cut in ("#", "?"):
        if cut in target:
            target = target.split(cut, 1)[0]
    return target


def _extract_fragment(target: str) -> str:
    if "#" not in target:
        return ""
    fragment = target.split("#", 1)[1]
    if "?" in fragment:
        fragment = fragment.split("?", 1)[0]
    return unquote(fragment.strip())


def _strip_html_comments(text: str) -> str:
    return HTML_COMMENT_RE.sub(lambda m: re.sub(r"[^\n]", " ", m.group(0)), text)


def github_heading_slug(text: str) -> str:
    """Return the GitHub-style slug used by this repo's docs/spec anchors."""
    text = text.strip().lower()
    text = MARKDOWN_LINK_TEXT_RE.sub(r"\1", text)
    text = INLINE_CODE_RE.sub(lambda m: m.group(0).strip("`"), text)
    text = "".join(
        ch
        for ch in text
        if ch in {"-", "_"} or ch.isspace() or unicodedata.category(ch)[0] in {"L", "N"}
    ).strip()
    return re.sub(r"\s", "-", text)


def heading_anchors(path: Path) -> set[str]:
    try:
        raw = path.read_text(encoding="utf-8")
    except (UnicodeDecodeError, OSError):
        return set()

    anchors: set[str] = set()
    seen: dict[str, int] = {}
    in_fence = False
    for line in raw.splitlines():
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = HEADING_RE.match(line)
        if match is None:
            continue
        heading = TRAILING_HASH_RE.sub("", match.group("text")).strip()
        slug = github_heading_slug(heading)
        suffix = seen.get(slug, 0)
        anchors.add(slug if suffix == 0 else f"{slug}-{suffix}")
        seen[slug] = suffix + 1
    return anchors


def _is_markdown_file(path: Path) -> bool:
    return path.is_file() and path.suffix.lower() in {".md", ".markdown"}


def _fragment_finding(path: Path, lineno: int, target: str, resolved: Path, fragment: str) -> Finding | None:
    if not fragment or not _is_markdown_file(resolved):
        return None
    if fragment in heading_anchors(resolved):
        return None
    return Finding(source=path, line=lineno, target=target, resolved=resolved, kind="fragment")


def scan_file(path: Path, check_fragments: bool = False) -> list[Finding]:
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
            if not target or _is_external(target):
                continue
            stripped = _strip_fragment(target)
            fragment = _extract_fragment(target)
            if not stripped:
                if check_fragments and fragment:
                    fragment_finding = _fragment_finding(path, lineno, target, path.resolve(), fragment)
                    if fragment_finding is not None:
                        findings.append(fragment_finding)
                continue
            resolved = (path.parent / stripped).resolve()
            if not resolved.exists():
                findings.append(
                    Finding(source=path, line=lineno, target=target, resolved=resolved)
                )
                continue
            if check_fragments and fragment:
                fragment_finding = _fragment_finding(path, lineno, target, resolved, fragment)
                if fragment_finding is not None:
                    findings.append(fragment_finding)
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


def scan_directory(
    root: Path, ignores: list[str] | None = None, check_fragments: bool = False
) -> list[Finding]:
    findings: list[Finding] = []
    for md in iter_markdown_files(root, ignores):
        findings.extend(scan_file(md, check_fragments=check_fragments))
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
    parser.add_argument(
        "--check-fragments",
        action="store_true",
        help="Validate local markdown #fragment targets against GitHub-style heading anchors",
    )
    args = parser.parse_args(argv)

    root = Path(args.root).resolve()
    if not root.is_dir():
        print(f"ERROR: not a directory: {root}", file=sys.stderr)
        return 2

    findings = scan_directory(root, ignores=args.ignore, check_fragments=args.check_fragments)
    if not findings:
        print(f"check_md_links: OK ({root})")
        return 0

    for finding in findings:
        print(finding.format(root), file=sys.stderr)
    print(f"check_md_links: {len(findings)} broken relative link(s) in {root}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
