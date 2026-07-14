"""Doc contract checks for execution automation closure."""

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
AGENTS_PATH = REPO_ROOT / "AGENTS.md"
PLAN_CONTEXT_CONTRACT_PATH = (
    REPO_ROOT / ".agent-skills" / "implement" / "shared" / "references" / "plan-context-contract.md"
)


def _agents_text() -> str:
    return AGENTS_PATH.read_text(encoding="utf-8")


def _plan_context_contract_text() -> str:
    return PLAN_CONTEXT_CONTRACT_PATH.read_text(encoding="utf-8")


def test_agents_declares_git_branch_strategy_section():
    text = _agents_text()

    assert "## 7 Git 分支策略" in text
    assert "- 默认父分支: main" in text
    assert "分支名前缀必须表达工作类型或业务域，而不是 Agent / CLI 工具身份" in text
    assert "禁止使用 `codex/`、`claude/`、`gemini/`、`agent/` 等工具名前缀" in text
    assert "docs/spec 设计派生或新 subject bootstrap 默认使用 `design/{subject}` 或 `spec-design/{subject}`" in text
    assert "- `/implement` 自动从父分支创建 feature branch" in text
    assert "- 创建 feature branch 前必须先更新父分支到最新远端状态；更新必须采用 fast-forward-only 语义，失败时停止并报告，不得从过期父分支派生新分支" in text
    assert "- phase commit 只在当前 feature branch 上提交并记录工作日志，默认不自动 merge / ff-merge 回父分支" in text
    assert "- 多方并行协作期间，父分支整合由用户在明确的 merge / rebase / PR review 阶段决定；Agent 不得在 phase 边界自动切回父分支合并" in text


def test_agents_work_journal_row_mentions_dual_mode():
    text = _agents_text()

    assert "| `/work-journal` | 记录工作日志并提交 | 显式调用 + `/tdd` phase-commit | 用户决定提交时机 / phase 边界自动提交时 |" in text


def test_agents_declares_continue_resume_must_reenter_implement():
    text = _agents_text()

    assert "继续已有计划或恢复当前 plan 执行 → 必须调用 `/implement`" in text


def test_agents_consolidates_skill_protocol_and_list_into_single_section():
    text = _agents_text()

    assert "### 3.1 Skills 使用协议与列表（强制）" in text
    assert "### 3.2 Skills 列表" not in text


def test_agents_code_ownership_section_uses_current_harness_areas():
    text = _agents_text()

    assert "| `governance` | 治理文档 | `AGENTS.md`、`CLAUDE.md`、`GEMINI.md` |" in text
    assert "| `docs` | 项目文档 | `docs/` |" in text
    assert "| `skills` | Skills 与共享脚本 | `.agent-skills/` |" in text
    assert "| `scenarios` | 场景测试框架 | `test/scenarios/` |" in text
    assert "当前仓库尚不存在的目录，不得提前写入 ownership 映射" in text
    assert "| `workspace` |" not in text
    assert "| `practice` |" not in text
    assert "| `frontend` |" not in text
    assert "| `config` |" not in text


def test_agents_completed_plan_policy_defaults_to_in_place_revision():
    text = _agents_text()

    assert "命中 `completed` plan 时，不得新建同主题 sibling follow-up / bugfix plan；应在原 spec/plan/checklist 上原地修订" in text
    assert "必须让原计划成为当前 owner，先更新 spec/plan/checklist，再继续 `/implement` 或其他明确 owner skill" in text
    assert "禁止重开已完成计划" not in text


def test_agents_declares_tdd_bdd_quality_gate_rules():
    text = _agents_text()

    assert "Code plan requires TDD" in text
    assert "Feature plan requires BDD" in text
    assert "前端 / 后端 / 工具脚本 / 迁移 / codegen" in text
    assert "不适用原因 + 替代验证 gate" in text


def test_plan_context_contract_mentions_branch_metadata_usage():
    text = _plan_context_contract_text()

    assert "| `metadata.baseBranch` | string | No | Base branch used by `/implement` Step 4.5 for fast-forward refresh, feature branch creation, and explicit integration decisions |" in text
    assert "| `metadata.branch` | string | No | Feature branch name stem used by `/implement` Step 4.5 before the date/collision suffix is appended |" in text
    assert "1. `context.yaml` `metadata.baseBranch`" in text
    assert "2. `AGENTS.md` project-level Git branch strategy" in text
    assert "3. Git default branch auto-detection" in text
    assert "Before creating a new feature branch, `/implement` must update the resolved base" in text
    assert "branch to the latest upstream state with fast-forward-only semantics." in text


def test_plan_context_contract_limits_validator_scope_to_manifest_and_paths():
    text = _plan_context_contract_text()

    assert "It does not parse or lint the" in text
    assert "consumers read those documents directly" in text


def test_plan_context_contract_declares_conditional_test_bdd_rules_are_document_owned():
    text = _plan_context_contract_text()

    assert "Conditional Test/BDD Document Rules" in text
    assert "Code plan requires TDD" in text
    assert "Feature plan requires BDD" in text
    assert "domain Behavior IDs backed by code-level behavior tests" in text
    assert "An `E2E.*` ID requires a real running frontend/backend flow" in text
    assert "must declare `BDD-N/A`" in text
    assert "`/design`, `/create-doc`, `/plan-review`, and `/implement` enforce these document-level rules" in text
