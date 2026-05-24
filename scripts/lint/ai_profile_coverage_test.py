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


# DEFAULT_JUDGE is a valid active judge.default profile included by make_repo so
# the mandatory judge-coverage assertion has a passing baseline. Negative tests
# pass an override via the judge_body parameter.
DEFAULT_JUDGE = textwrap.dedent(
    """
    name: judge.default
    capability: judge
    status: active
    default:
      provider_ref: judge-deepseek
      model: deepseek-v4-pro
    timeout_ms: 1000
    version: 1.0.0
    """
).strip()


def make_repo(tmp_path: Path, profile_body: str, judge_body: str = DEFAULT_JUDGE) -> Path:
    repo = tmp_path / "repo"
    repo.mkdir()
    write(
        repo / "docs/spec/prompt-rubric-registry/spec.md",
        textwrap.dedent(
            """
            #### 3.1.1 11 个当前 baseline feature_key 字典

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
              - name: deepseek
                protocol: openai_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [chat]
                version: 1.0.0
              - name: judge-deepseek
                protocol: judge_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [judge]
                version: 1.0.0
              - name: judge-placeholder
                protocol: judge_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [judge]
                version: 1.0.0
            """
        ).strip(),
    )
    profiles = "profiles:\n  - " + textwrap.indent(profile_body, "    ").lstrip()
    if judge_body:
        profiles += "\n  - " + textwrap.indent(judge_body, "    ").lstrip()
    write(repo / "config/ai-profiles.yaml", profiles)
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
                provider_ref: deepseek
                model: deepseek-v4-flash
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
            capability: stt
            status: active
            default:
              provider_ref: deepseek
              model: stt-model
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
              provider_ref: deepseek
              model: deepseek-v4-flash
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
            capability: stt
            status: active
            default:
              provider_ref: deepseek
              model: stt-model
            timeout_ms: 1000
            version: 1.0.0
            """
        ).strip(),
    )
    (repo / "config/ai-providers.yaml").write_text(
        textwrap.dedent(
            """
            providers:
              - name: deepseek
                protocol: openai_compatible
                base_url_env: AI_PROVIDER_BASE_URL
                api_key_env: AI_PROVIDER_API_KEY
                capabilities: [chat, stt]
                version: 1.0.0
            """
        ).strip(),
        encoding="utf-8",
    )

    result = run(repo)
    assert result.returncode == 1
    assert "Product/UI capability mismatch" in result.stderr
    assert "practice.followup.default" in result.stderr


CHAT_FOLLOWUP = textwrap.dedent(
    """
    name: practice.followup.default
    capability: chat
    status: active
    default:
      provider_ref: deepseek
      model: deepseek-v4-flash
    timeout_ms: 1000
    version: 1.0.0
    """
).strip()


def test_passes_when_judge_default_active_and_non_placeholder(tmp_path: Path) -> None:
    repo = make_repo(tmp_path, CHAT_FOLLOWUP)
    result = run(repo)
    assert result.returncode == 0, result.stderr
    assert "OK" in result.stdout


def test_fails_when_judge_default_missing(tmp_path: Path) -> None:
    repo = make_repo(tmp_path, CHAT_FOLLOWUP, judge_body="")
    result = run(repo)
    assert result.returncode == 1
    assert "judge.default" in result.stderr


def test_fails_when_judge_default_unsupported(tmp_path: Path) -> None:
    judge = textwrap.dedent(
        """
        name: judge.default
        capability: judge
        status: unsupported
        unsupported_reason: reserved
        default:
          provider_ref: judge-deepseek
          model: deepseek-v4-pro
        timeout_ms: 1000
        version: 1.0.0
        """
    ).strip()
    repo = make_repo(tmp_path, CHAT_FOLLOWUP, judge_body=judge)
    result = run(repo)
    assert result.returncode == 1
    assert "judge.default" in result.stderr
    assert "active" in result.stderr


def test_fails_when_judge_default_uses_placeholder_provider(tmp_path: Path) -> None:
    judge = textwrap.dedent(
        """
        name: judge.default
        capability: judge
        status: active
        default:
          provider_ref: judge-placeholder
          model: judge-provider-required
        timeout_ms: 1000
        version: 1.0.0
        """
    ).strip()
    repo = make_repo(tmp_path, CHAT_FOLLOWUP, judge_body=judge)
    result = run(repo)
    assert result.returncode == 1
    assert "placeholder" in result.stderr


def test_fails_when_chat_profile_uses_placeholder_model(tmp_path: Path) -> None:
    chat = textwrap.dedent(
        """
        name: practice.followup.default
        capability: chat
        status: active
        default:
          provider_ref: deepseek
          model: chat-provider-required
        timeout_ms: 1000
        version: 1.0.0
        """
    ).strip()
    repo = make_repo(tmp_path, chat)
    result = run(repo)
    assert result.returncode == 1
    assert "placeholder" in result.stderr
