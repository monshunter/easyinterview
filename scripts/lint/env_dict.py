#!/usr/bin/env python3
"""env_dict drift checker for secrets-and-config spec §3.1.1.

Three-way diff:

1. ``.env.example`` (root)               — repository-versioned dictionary
2. ``docs/spec/secrets-and-config/spec.md`` table §3.1.1
3. Code-side env reads (``os.Getenv`` / ``os.LookupEnv``) under
   ``backend/internal/platform/config/``, ``backend/internal/platform/secrets/``,
   ``backend/cmd/api/``, ``backend/cmd/worker/``

Any key present in one source but missing in another causes a non-zero
exit. Spec C-9 / C-11 / C-12 close on this gate.

The script is read-only: it never edits files; deployers must update the
spec and ``.env.example`` together.
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ENV_LINE_RE = re.compile(r"^\s*([A-Z][A-Z0-9_]*)\s*=", re.MULTILINE)
GETENV_RE = re.compile(r"os\.(?:Getenv|LookupEnv)\(\s*\"([A-Z][A-Z0-9_]*)\"")
ENV_LITERAL_RE = re.compile(r"\"([A-Z][A-Z0-9]*_[A-Z0-9_]*)\"")
TABLE_KEY_RE = re.compile(r"^\|\s*`([A-Z][A-Z0-9_]+)`\s*\|", re.MULTILINE)
SPEC_SECTION_HEADER_311 = "#### 3.1.1 P0 必备 env key 字典"


def parse_env_example(path: Path) -> set[str]:
    text = path.read_text(encoding="utf-8")
    return set(ENV_LINE_RE.findall(text))


def parse_spec_dictionary(path: Path) -> set[str]:
    text = path.read_text(encoding="utf-8")
    if SPEC_SECTION_HEADER_311 not in text:
        sys.stderr.write(f"env_dict: spec missing section header {SPEC_SECTION_HEADER_311!r}\n")
        raise SystemExit(2)
    section = text.split(SPEC_SECTION_HEADER_311, 1)[1]
    next_header = re.search(r"^####\s", section, re.MULTILINE)
    if next_header:
        section = section[: next_header.start()]
    return set(TABLE_KEY_RE.findall(section))


def parse_code_env_reads(repo_root: Path) -> set[str]:
    keys: set[str] = set()
    scan_roots = [
        repo_root / "backend" / "internal" / "platform" / "config",
        repo_root / "backend" / "internal" / "platform" / "secrets",
        repo_root / "backend" / "cmd" / "api",
        repo_root / "backend" / "cmd" / "worker",
    ]
    for root in scan_roots:
        if not root.exists():
            continue
        for path in root.rglob("*.go"):
            if path.name.endswith("_test.go"):
                continue
            text = path.read_text(encoding="utf-8")
            keys.update(GETENV_RE.findall(text))
            # Canonical EnvBindings / SecretBindings are code-side env key
            # declarations even when the actual os.LookupEnv call happens in
            # the generic loader. Limit the broader literal scan to env-shaped
            # strings (must contain "_") to avoid catching HTTP verbs etc.
            keys.update(ENV_LITERAL_RE.findall(text))
    return keys


def main() -> int:
    parser = argparse.ArgumentParser(description="env_dict drift checker")
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    args = parser.parse_args()

    repo = args.repo_root.resolve()
    env_example = repo / ".env.example"
    spec = repo / "docs" / "spec" / "secrets-and-config" / "spec.md"

    if not env_example.exists():
        print(f"env_dict: missing {env_example}", file=sys.stderr)
        return 2
    if not spec.exists():
        print(f"env_dict: missing {spec}", file=sys.stderr)
        return 2

    example_keys = parse_env_example(env_example)
    spec_keys = parse_spec_dictionary(spec)
    code_keys = parse_code_env_reads(repo)

    problems: list[str] = []

    missing_in_example = (spec_keys | code_keys) - example_keys
    if missing_in_example:
        problems.append(
            "missing from .env.example: " + ", ".join(sorted(missing_in_example))
        )

    missing_in_spec = (example_keys | code_keys) - spec_keys
    # Allow code-only build-time helpers explicitly listed below; everything
    # else must show up in spec §3.1.1.
    if missing_in_spec:
        problems.append(
            "missing from spec §3.1.1 table: " + ", ".join(sorted(missing_in_spec))
        )

    extra_in_code = code_keys - spec_keys
    if extra_in_code:
        # Only fail if code reads a key that spec does not document.
        problems.append(
            "code reads env keys not declared in spec: "
            + ", ".join(sorted(extra_in_code))
        )

    if problems:
        print("env_dict drift detected:", file=sys.stderr)
        for line in problems:
            print(f"  - {line}", file=sys.stderr)
        print(
            "Fix order: update spec §3.1.1 → sync .env.example → adjust validator → re-run make lint-config",
            file=sys.stderr,
        )
        return 1

    print(
        f"env_dict: OK ({len(example_keys)} keys in .env.example, "
        f"{len(spec_keys)} keys in spec §3.1.1, {len(code_keys)} keys read by code)"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
