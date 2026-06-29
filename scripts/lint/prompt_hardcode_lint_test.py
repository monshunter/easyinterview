#!/usr/bin/env python3
"""Tests for scripts/lint/prompt_hardcode_lint.{py,go}."""
from __future__ import annotations

import pathlib
import subprocess
import sys

THIS_DIR = pathlib.Path(__file__).resolve().parent
WRAPPER = THIS_DIR / "prompt_hardcode_lint.py"
REPO_ROOT = THIS_DIR.parents[1]


def _run(repo_root: pathlib.Path, roots: list[str]) -> subprocess.CompletedProcess:
    return subprocess.run(
        [
            sys.executable,
            str(WRAPPER),
            "--roots",
            ",".join(roots),
            "--repo-root",
            str(repo_root),
        ],
        capture_output=True,
        text=True,
        check=False,
    )


def _seed_go(tmp_path: pathlib.Path, rel_dir: str, filename: str, source: str) -> None:
    """Seed a Go source file under tmp_path/<rel_dir>/<filename>."""
    target = tmp_path / rel_dir
    target.mkdir(parents=True, exist_ok=True)
    (target / filename).write_text(source, encoding="utf-8")


def test_default_scan_passes():
    """Real backend code must not contain hardcoded prompt assignments."""
    result = _run(REPO_ROOT, [
        "backend/internal/practice",
        "backend/internal/report",
        "backend/internal/resume",
        "backend/internal/targetjob",
    ])
    assert result.returncode == 0, f"stdout={result.stdout!r} stderr={result.stderr!r}"


def test_negative_raw_string_prompt(tmp_path):
    """Plan §1.6 negative fixture: prompt := `multi-line raw string` must fail."""
    source = (
        "package practice\n\n"
        "func bad() {\n"
        "\tprompt := `You are an interviewer.\n"
        "Answer truthfully.`\n"
        "\t_ = prompt\n"
        "}\n"
    )
    _seed_go(tmp_path, "backend/internal/practice", "bad.go", source)
    result = _run(tmp_path, ["backend/internal/practice"])
    assert result.returncode == 1
    assert "prompt" in result.stderr


def test_negative_long_quoted_prompt(tmp_path):
    """A long single-line quoted assignment to Prompt = \"...\" must fail."""
    long_string = "You are an interviewer producing structured assessment output for the rubric."
    assert len(long_string) >= 70
    source = (
        "package report\n\n"
        "func bad() {\n"
        f"\tPrompt := \"{long_string}\"\n"
        "\t_ = Prompt\n"
        "}\n"
    )
    _seed_go(tmp_path, "backend/internal/report", "bad.go", source)
    result = _run(tmp_path, ["backend/internal/report"])
    assert result.returncode == 1
    assert "Prompt" in result.stderr


def test_prompt_version_short_string_passes(tmp_path):
    """PromptVersion = \"v0.1.0\" is a short version key, not a body."""
    source = (
        "package practice\n\n"
        "const PromptVersion = \"v0.1.0\"\n"
    )
    _seed_go(tmp_path, "backend/internal/practice", "ok.go", source)
    result = _run(tmp_path, ["backend/internal/practice"])
    assert result.returncode == 0, f"unexpected violation: {result.stderr!r}"


def test_test_file_allowlisted(tmp_path):
    """`*_test.go` is allowlisted even when it contains hardcoded prompts."""
    source = (
        "package practice\n\n"
        "func TestBad() {\n"
        "\tprompt := `multi line\nbody for fixture purposes only`\n"
        "\t_ = prompt\n"
        "}\n"
    )
    _seed_go(tmp_path, "backend/internal/practice", "bad_test.go", source)
    result = _run(tmp_path, ["backend/internal/practice"])
    assert result.returncode == 0, f"test file unexpectedly flagged: {result.stderr!r}"


def test_system_message_flagged(tmp_path):
    """systemMessage := `...` literal is flagged regardless of suffix rules."""
    source = (
        "package targetjob\n\n"
        "func bad() {\n"
        "\tsystemMessage := `You are a coach.\n"
        "Summarize the target role.`\n"
        "\t_ = systemMessage\n"
        "}\n"
    )
    _seed_go(tmp_path, "backend/internal/targetjob", "bad.go", source)
    result = _run(tmp_path, ["backend/internal/targetjob"])
    assert result.returncode == 1
    assert "systemMessage" in result.stderr


if __name__ == "__main__":
    import unittest

    unittest.main(verbosity=2)
