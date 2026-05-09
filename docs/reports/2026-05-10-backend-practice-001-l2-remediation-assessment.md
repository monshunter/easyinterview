# Backend Practice 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-practice/001-plan-and-session-orchestration` L2 code review remediation，覆盖首题 parser、`startPracticeSession` success replay snapshot、`000003` down migration ownership。
- Plan 原地修订：`plan.md` / `checklist.md` 升至 1.1，追加并完成 `0.11` / `1.10` / `2.10` remediation checklist，最终恢复 `completed`。
- 验证证据：backend practice 相关 Go 包全绿；`make validate-fixtures`、migration/preflight/legacy lint、`make codegen-check`、`make lint-events` 全绿；`E2E.P0.022` ~ `E2E.P0.026` scenario setup / trigger / verify / cleanup 全绿；`sync-doc-index --check` zero drift。
- Bug 记录：新增 [BUG-0033](../bugs/BUG-0033.md)，记录本次 L2 finding 与修复证据。

## 2 会话中的主要阻点/痛点

- Completed plan 需要重新进入 TDD 修复路径。
  - **证据**：原 checklist 全部已勾选，必须先按 create-doc/tdd 规则将原 plan 调回 `active` 并追加 remediation items，才能执行红绿流程。
  - **影响**：增加一次文档生命周期操作，但避免了绕过 checklist 的直接修复。
- Scenario 脚本入口第一次执行路径假设错误。
  - **证据**：最初执行根级 `./setup.sh` 报 `no such file or directory`，随后按 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 通过。
  - **影响**：轻微验证返工；仓库 README 已写明场景契约，无需修改仓库流程资产。
- 原 gate 对语义 drift 覆盖不足。
  - **证据**：历史 gate 覆盖了“有首题”“无重复副作用”“migration up contract”，但没有覆盖 F3 prompt JSON keys、replay stored snapshot、migration down ownership。
  - **影响**：L2 review 才发现真实 provider 输出与 replay/migration rollback 语义缺口。

## 3 根因归类

- Parser 与 prompt truth source 没有同测。
  - **类别**：spec-plan
- Idempotency replay gate 只验证副作用数量，没有制造 mutable state drift。
  - **类别**：spec-plan
- Migration baseline rebase 后缺少 down-path ownership negative test。
  - **类别**：spec-plan
- Scenario 脚本路径误执行属于本次操作失误。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续涉及 AI prompt JSON 输出的 backend plan，应在 checklist 中明确“prompt markdown schema keys -> parser test” gate。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 idempotency success replay 应默认要求 stored response snapshot drift test，而不是只断言副作用不重复。
  - **落点**：spec-plan
  - **优先级**：high
- migration rebase / integrator 模式中，down migration 必须有 owner zero-drop negative test。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在 `backend-practice/002-event-loop-and-completion` 开始前，把 AI output parser schema gate 与 idempotency replay snapshot drift gate 写入该 plan 的 test/coverage matrix。
- 可延后：若未来再次出现 migration ownership rebase，再把 down-path zero-drop 检查提升为共享 migration lint rule；本轮已有 focused contract test 覆盖当前缺口。
