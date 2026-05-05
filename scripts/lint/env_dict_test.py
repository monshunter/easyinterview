"""Unit tests for env_dict drift checker."""
from __future__ import annotations

import subprocess
import sys
import textwrap
from pathlib import Path

THIS_DIR = Path(__file__).resolve().parent
SCRIPT = THIS_DIR / "env_dict.py"


def make_repo(
    tmp_path: Path,
    env_example: str,
    spec_table: str,
    code_files: dict[str, str] | None = None,
    provider_registry: str | None = None,
) -> Path:
    repo = tmp_path / "repo"
    repo.mkdir()
    (repo / ".env.example").write_text(env_example)
    spec_dir = repo / "docs" / "spec" / "secrets-and-config"
    spec_dir.mkdir(parents=True)
    (spec_dir / "spec.md").write_text(
        "#### 3.1.1 P0 必备 env key 字典\n\n" + spec_table + "\n\n#### 3.1.2 next\n",
        encoding="utf-8",
    )
    code_root = repo / "backend" / "internal" / "platform" / "config"
    code_root.mkdir(parents=True)
    for name, body in (code_files or {}).items():
        (code_root / name).write_text(body)
    if provider_registry is not None:
        config_dir = repo / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        (config_dir / "ai-providers.yaml").write_text(provider_registry, encoding="utf-8")
    return repo


def run(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        capture_output=True,
        text=True,
    )


def test_passes_when_three_sources_aligned(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        env_example="APP_ENV=dev\nDATABASE_URL=postgres://x\n",
        spec_table=textwrap.dedent("""
            | Key | a | b |
            |-----|---|---|
            | `APP_ENV` | x | y |
            | `DATABASE_URL` | x | y |
        """).strip(),
    )
    result = run(repo)
    assert result.returncode == 0, result.stderr
    assert "OK" in result.stdout


def test_fails_when_env_example_missing_key(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        env_example="APP_ENV=dev\n",
        spec_table="| Key | a |\n|-----|---|\n| `APP_ENV` | x |\n| `DATABASE_URL` | y |\n",
    )
    result = run(repo)
    assert result.returncode == 1
    assert "missing from .env.example" in result.stderr
    assert "DATABASE_URL" in result.stderr


def test_fails_when_code_reads_undocumented_key(tmp_path: Path) -> None:
    code = 'package config\nimport "os"\nfunc x() { os.Getenv("MYSTERY_KEY") }\n'
    repo = make_repo(
        tmp_path,
        env_example="APP_ENV=dev\n",
        spec_table="| Key | a |\n|-----|---|\n| `APP_ENV` | x |\n",
        code_files={"loader.go": code},
    )
    result = run(repo)
    assert result.returncode == 1
    assert "MYSTERY_KEY" in result.stderr


def test_fails_when_binding_literal_declares_undocumented_key(tmp_path: Path) -> None:
    code = textwrap.dedent("""
        package config

        var defaultEnvBindings = map[string]string{
            "DATABASE_URL": "database.url",
        }
    """).strip()
    repo = make_repo(
        tmp_path,
        env_example="APP_ENV=dev\n",
        spec_table="| Key | a |\n|-----|---|\n| `APP_ENV` | x |\n",
        code_files={"bindings.go": code},
    )
    result = run(repo)
    assert result.returncode == 1
    assert "DATABASE_URL" in result.stderr


def test_fails_when_provider_registry_env_ref_is_missing_from_dictionary(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        env_example="APP_ENV=dev\n",
        spec_table="| Key | a |\n|-----|---|\n| `APP_ENV` | x |\n",
        provider_registry=textwrap.dedent("""
            providers:
              - name: default-openai-compatible
                protocol: openai_compatible
                base_url_env: CUSTOM_PROVIDER_BASE_URL
                api_key_env: CUSTOM_PROVIDER_API_KEY
                capabilities: [chat]
                version: 1.0.0
        """).strip(),
    )
    result = run(repo)
    assert result.returncode == 1
    assert "CUSTOM_PROVIDER_BASE_URL" in result.stderr
    assert "CUSTOM_PROVIDER_API_KEY" in result.stderr


def test_fails_when_spec_missing_311_section(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    repo.mkdir()
    (repo / ".env.example").write_text("APP_ENV=dev\n")
    spec_dir = repo / "docs" / "spec" / "secrets-and-config"
    spec_dir.mkdir(parents=True)
    (spec_dir / "spec.md").write_text("# wrong\n")
    result = run(repo)
    assert result.returncode == 2
    assert "missing section header" in result.stderr or "3.1.1" in result.stderr
