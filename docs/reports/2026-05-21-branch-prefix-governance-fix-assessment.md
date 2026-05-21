# Branch Prefix Governance Fix 交付复盘报告

> **日期**: 2026-05-21
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复当前本地分支使用 `codex/` 工具名前缀的问题，并把分支命名前缀治理固化到 `AGENTS.md`、入口 skills、BUG 知识库、模式库和工作日志。
- 当前分支已从 `codex/backend-profile-and-jobs-recommendations-bootstrap` 重命名为 `spec-design/backend-profile-and-jobs-recommendations-bootstrap`。
- 验证证据：
  - `python3 -m pytest .agent-skills/design/scripts/test_design_skill_contract.py .agent-skills/change-intake/scripts/test_change_intake_skill_contract.py .agent-skills/implement/scripts/test_execution_automation_docs_contract.py .agent-skills/implement/scripts/test_session_branch_helper.py`：30 passed。
  - `make docs-check`：Header / INDEX / Orphans / Warnings 均为 none，链接检查 OK。
  - `git diff --check`：PASS。
  - `git log -1 --format=%B | LC_ALL=C perl -ne 'if (/[^\\x00-\\x7F]/) { print; exit 1 }'`：commit message ASCII 校验通过。

## 2 会话中的主要阻点/痛点

- 分支前缀把工具身份暴露成协作域。
  - **证据**：当前分支原名为 `codex/backend-profile-and-jobs-recommendations-bootstrap`；本地历史分支还存在多个 `codex/*`。
  - **影响**：分支名无法表达工作性质，破坏用户预期的 `fix/spec/feat/design/spec-design` 等 domain signal。
- 既有规则只覆盖“不要在默认父分支写文件”，没有覆盖“feature branch 的第一段必须是语义前缀”。
  - **证据**：`AGENTS.md` 原 §7 未禁止工具名前缀；`/design` 有 `design/` / `spec-design/` 示例但没有负向规则；`/change-intake`、`/plan-review --fix`、`/plan-code-review --fix` 的 branch guard 也缺少命名前缀约束。
  - **影响**：Agent 可以满足“已创建 feature branch”的表层门禁，却仍创建不符合协作语义的分支。

## 3 根因归类

- 根因：入口流程的 branch guard 缺少语义命名检查。
  - **类别**：AGENTS.md + skill
- 根因：contract tests 未覆盖工具名前缀负向规则。
  - **类别**：skill
- 非根因：`/implement` 自动分支脚本。
  - **证据**：`detect_session_branch.py` 只匹配 `feat|fix|opt|docs`，`/implement` Step 4.5 约定为 `{type}/{subspec}-{plan}-{MMDD}`。

## 4 对流程资产的改进建议

- 已完成：`AGENTS.md` §7 明确允许语义前缀并禁止 `codex/`、`claude/`、`gemini/`、`agent/`。
  - **落点**：AGENTS.md
  - **优先级**：high
- 已完成：`/design`、`/change-intake`、`/plan-review --fix`、`/plan-code-review --fix` 的分支门禁继承禁止工具名前缀规则。
  - **落点**：skill
  - **优先级**：high
- 已完成：`BUG-0080` 与 `docs/bugs/PATTERNS.md` 模式 2 记录复发检查清单。
  - **落点**：BUG 知识库
  - **优先级**：medium
- 已完成：新增 contract test 覆盖 `AGENTS.md`、`/design`、`/change-intake` 的负向命名前缀规则。
  - **落点**：skill tests
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最高优先级：继续在当前 `spec-design/backend-profile-and-jobs-recommendations-bootstrap` 分支上执行 `backend-profile` 与 `backend-jobs-recommendations` 两个新 subject 的 L1 plan review，确认 operation matrix、BDD gate 和 frontend/backend contract preflight 完整后，再进入后端 implementation owner。
- 可延后：清理本地历史 `codex/*` 分支。它们是历史引用，是否删除取决于用户是否还需要查旧工作，不应在本次治理修复中自动删除。
