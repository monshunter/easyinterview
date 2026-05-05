"""Unit tests for ai_profile_coverage lint."""
from __future__ import annotations

import subprocess
import sys
import textwrap
from pathlib import Path

THIS_DIR = Path(__file__).resolve().parent
SCRIPT = THIS_DIR / "ai_profile_coverage.py"


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def make_repo(tmp_path: Path, profile_body: str) -> Path:
    repo = tmp_path / "repo"
    repo.mkdir()
    write(
        repo / "docs/spec/prompt-rubric-registry/spec.md",
        textwrap.dedent(
            """
            #### 3.1.1 12 个当前 baseline feature_key 字典

            | feature_key | 用途 | 关联业务域 | 关联 Model Profile（默认） |
            |-------------|------|-----------|--------------------------|
            | `practice.session.follow_up` | 追问生成 | C5 | `practice.followup.default` |
            """
        ).strip(),
    )
    write(
        repo / "docs/spec/ai-provider-and-model-routing/spec.md",
        textwrap.dedent(
            """
            ### 4.5 Product/UI AI Capability Catalog

            | 产品 / UI 场景 | 主要输入 | Capability family | 默认 profile 命名 |
            |----------------|----------|-------------------|-------------------|
            | 模拟面试追问 | transcript | `chat` | `practice.followup.default` |
            """
        ).strip(),
    )
    write(
        repo / "config/ai-providers.yaml",
        textwrap.dedent(
            """
            providers:
              - name: unit-test-stub
                protocol: stub
                capabilities: [chat]
                version: 1.0.0
              - name: default-openai-compatible
                protocol: openai_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [chat]
                version: 1.0.0
            """
        ).strip(),
    )
    write(
        repo / "config/ai-profiles.yaml",
        "profiles:\n  - " + textwrap.indent(profile_body, "    ").lstrip(),
    )
    write(
        repo / "deploy/dev-stack/.env.example",
        textwrap.dedent(
            """
            AI_PROVIDER_REGISTRY_PATH=config/ai-providers.yaml
            AI_PROVIDER_BASE_URL=
            AI_PROVIDER_API_KEY=
            AI_MODEL_PROFILE_PATH=config/ai-profiles.yaml
            """
        ).strip(),
    )
    return repo


def run(repo: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        capture_output=True,
        text=True,
    )


def test_passes_when_docs_and_catalog_align(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        textwrap.dedent(
            """
            name: practice.followup.default
            capability: chat
            status: active
            default:
              provider_ref: default-openai-compatible
              model: default-chat
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    result = run(repo)
    assert result.returncode == 0, result.stderr
    assert "OK" in result.stdout


def test_fails_when_referenced_profile_is_missing(tmp_path: Path) -> None:
    repo = make_repo(tmp_path, "")
    (repo / "config/ai-profiles.yaml").write_text("profiles: []\n", encoding="utf-8")
    result = run(repo)
    assert result.returncode == 1
    assert "missing profiles" in result.stderr
    assert "practice.followup.default" in result.stderr


def test_fails_when_provider_does_not_support_capability(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        textwrap.dedent(
            """
            name: practice.followup.default
            capability: rerank
            status: active
            default:
              provider_ref: default-openai-compatible
              model: rerank-model
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    result = run(repo)
    assert result.returncode == 1
    assert "capability not declared by provider" in result.stderr


def test_fails_when_active_profile_uses_stub_provider(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        textwrap.dedent(
            """
            name: practice.followup.default
            capability: chat
            status: active
            default:
              provider_ref: unit-test-stub
              model: stub-chat-1
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    result = run(repo)
    assert result.returncode == 1
    assert "active profile must not use stub provider" in result.stderr


def test_fails_when_dev_stack_env_uses_legacy_profile_directory(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        textwrap.dedent(
            """
            name: practice.followup.default
            capability: chat
            status: active
            default:
              provider_ref: default-openai-compatible
              model: default-chat
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    write(
        repo / "deploy/dev-stack/.env.example",
        textwrap.dedent(
            """
            AI_PROVIDER_BASE_URL=
            AI_PROVIDER_API_KEY=
            AI_MODEL_PROFILE_PATH=config/ai-profiles/
            """
        ).strip(),
    )

    result = run(repo)
    assert result.returncode == 1
    assert "deploy/dev-stack/.env.example" in result.stderr
    assert "AI_MODEL_PROFILE_PATH must point to config/ai-profiles.yaml" in result.stderr
    assert "AI_PROVIDER_REGISTRY_PATH" in result.stderr


def test_fails_when_product_ui_capability_disagrees_with_catalog(tmp_path: Path) -> None:
    repo = make_repo(
        tmp_path,
        textwrap.dedent(
            """
            name: practice.followup.default
            capability: embed
            status: active
            default:
              provider_ref: default-openai-compatible
              model: default-embed
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    (repo / "config/ai-providers.yaml").write_text(
        textwrap.dedent(
            """
            providers:
              - name: default-openai-compatible
                protocol: openai_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [chat, embed]
                version: 1.0.0
            """
        ).strip(),
        encoding="utf-8",
    )

    result = run(repo)
    assert result.returncode == 1
    assert "Product/UI capability mismatch" in result.stderr
    assert "practice.followup.default" in result.stderr
