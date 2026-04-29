#!/usr/bin/env python3
"""Local drift gate for shared/conventions.yaml generated outputs.

The B1 conventions generator is the only source of truth for enum, error-code,
pagination, ID, and AI vocabulary generated files. This wrapper renders the
current YAML into a temporary tree and compares every tracked output against that
canonical render.
"""
from __future__ import annotations

import argparse
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

GENERATED_OUTPUTS = (
    "backend/internal/shared/types/enums.go",
    "backend/internal/shared/types/http_dto.go",
    "backend/internal/shared/errors/codes.go",
    "backend/internal/shared/idx/generated.go",
    "backend/internal/shared/ai/vocabulary.go",
    "frontend/src/lib/conventions/enums.ts",
    "frontend/src/lib/conventions/errors.ts",
    "frontend/src/lib/conventions/ai.ts",
    "frontend/src/lib/conventions/pagination.ts",
    "frontend/src/lib/ids/generated.ts",
)


def read_generated_outputs(root: Path) -> dict[str, str]:
    outputs: dict[str, str] = {}
    for rel in GENERATED_OUTPUTS:
        path = root / rel
        if path.exists():
            outputs[rel] = path.read_text(encoding="utf-8")
    return outputs


def compare_generated_outputs(actual: dict[str, str], expected: dict[str, str]) -> list[str]:
    errs: list[str] = []
    for rel in GENERATED_OUTPUTS:
        asset = asset_label(rel)
        if rel not in expected:
            errs.append(f"{asset} generator did not render expected output: shared/conventions.yaml -> {rel}")
            continue
        if rel not in actual:
            errs.append(f"{asset} missing generated output: shared/conventions.yaml -> {rel}")
            continue
        if actual[rel] != expected[rel]:
            errs.append(
                f"{asset} drift: shared/conventions.yaml -> {rel} differs; "
                "run make codegen-conventions"
            )
    return errs


def asset_label(rel: str) -> str:
    if "/ai/" in rel or rel.endswith("/ai.ts"):
        return "AI vocabulary"
    if "/errors/" in rel or rel.endswith("/errors.ts"):
        return "error code"
    if "/types/enums.go" in rel or rel.endswith("/enums.ts"):
        return "enum"
    if "/idx/" in rel or rel.endswith("/ids/generated.ts"):
        return "id convention"
    if rel.endswith("http_dto.go") or rel.endswith("pagination.ts"):
        return "pagination/API structure"
    return "conventions"


def render_expected(repo_root: Path) -> dict[str, str]:
    with tempfile.TemporaryDirectory(prefix="conventions-drift-") as tmp_name:
        tmp = Path(tmp_name)
        result = subprocess.run(
            [
                "go",
                "run",
                "./cmd/codegen/conventions",
                "-yaml",
                str(repo_root / "shared" / "conventions.yaml"),
                "-repo-root",
                str(tmp),
            ],
            cwd=repo_root / "backend",
            text=True,
            capture_output=True,
        )
        if result.returncode != 0:
            stderr = result.stderr.strip()
            raise RuntimeError(f"codegen-conventions failed: {stderr}")
        return read_generated_outputs(tmp)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(ROOT), help="repository root")
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    try:
        expected = render_expected(repo_root)
    except RuntimeError as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        return 2

    actual = read_generated_outputs(repo_root)
    errs = compare_generated_outputs(actual, expected)
    if errs:
        for err in errs:
            print(f"FAIL: {err}", file=sys.stderr)
        return 1

    print(f"OK: {len(GENERATED_OUTPUTS)} conventions generated outputs match shared/conventions.yaml")
    return 0


if __name__ == "__main__":
    sys.exit(main())
