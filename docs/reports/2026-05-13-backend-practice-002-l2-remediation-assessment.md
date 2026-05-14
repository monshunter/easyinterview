# Backend Practice 002 L2 Remediation 交付复盘报告

> **日期**: 2026-05-13
> **审查人**: Codex

## 1 复盘范围与成功证据

- 修复范围：`backend-practice/plans/002-event-loop-and-completion` L2 review findings，覆盖 `answer_submitted` required payload 校验、`completePracticeSession` 状态机 guard、BDD 编号迁移与编号碰撞 gate。
- 计划资产：原 002 plan / checklist / test-plan / test-checklist / bdd-plan / bdd-checklist 已原地 bump 到 v1.1，并新增 L2 remediation checklist 与验证证据。
- Bug 资产：[BUG-0054](../bugs/BUG-0054.md) 记录 payload / status / BDD gate drift，状态 `resolved`。
- 验证证据：`cd backend && go test ./...`、focused backend practice/API/store/middleware/cmd tests、`python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`、`python3 scripts/lint/backend_practice_legacy.py --repo-root .`、`make lint-events`、`make codegen-events-check`、`make codegen-check`、`make validate-fixtures`、`python3 scripts/lint/conventions_drift.py --repo-root .`、`make docs-check`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- Completed 状态容易遮蔽语义缺口。
  - **证据**：002 checklist 已完成，但 L2 反查发现 `answer_submitted` 缺少 `payload.answerText` 校验，`completePracticeSession` 没有限制非法源状态。
  - **影响**：如果只复用历史 PASS，会把核心 API failure path 当成已闭环。
- BDD 编号没有可执行唯一性 gate。
  - **证据**：002 BDD 文档分配 `E2E.P0.034` / `035`，而 `test/scenarios/e2e/INDEX.md` 已由 backend-resume 占用这两个编号。
  - **影响**：不同 owner 的 BDD 证据会共享同一全局 ID，后续 scenario 索引和报告会变得不可追溯。
- D-35 replay 与状态 guard 的顺序需要明确测试。
  - **证据**：修复选择为先 replay 既有 report/job，再对无 report/job 的 session 做状态 guard；单测新增 `TestSQLRepositoryCompleteSessionReplaysExistingReportBeforeStatusGuard`。
  - **影响**：如果只加状态 guard，可能误伤已完成 session 的双 key replay；如果只保留 replay，又会放过 failed / queued 误创建 report。

## 3 根因归类

- Payload / status 缺口。
  - **类别**：spec-plan
  - **根因**：原计划覆盖主流程和幂等矩阵，但没有把 required payload field 与非法源状态列为独立 negative gate。
- BDD 编号碰撞。
  - **类别**：skill / spec-plan
  - **根因**：plan-code-review 过去检查 BDD 文件是否存在和测试是否通过，但没有把 `E2E.P0.xxx` 分配与全局 scenario index 交叉验证。
- 修复流程闭环。
  - **类别**：无新增仓库改动需要
  - **根因**：本次已经把缺口沉淀到 `backend_practice_legacy.py`、BUG-0054 和 002 v1.1 checklist，当前不需要再修改 AGENTS.md。

## 4 对流程资产的改进建议

- 在后续 backend feature plan 的 BDD 创建阶段就运行 scenario ID collision check。
  - **落点**：spec-plan / skill
  - **优先级**：high
  - 建议：`plan-review` 或 `plan-code-review` 命中 BDD 文件时，读取 `编号分配` 与 `test/scenarios/e2e/INDEX.md`，阻止复用非本 owner 已占用编号。
- 对 API event append 计划补 required payload field negative gate。
  - **落点**：spec-plan
  - **优先级**：medium
  - 建议：event kind 表格中每个 required payload 字段都应有 service-level validation test，不只依赖 OpenAPI schema 或 happy path fixture。
- 对 completion / lifecycle 计划固定 replay-before-guard 测试模式。
  - **落点**：spec-plan
  - **优先级**：medium
  - 建议：同时测试“既有结果 replay 不被状态 guard 阻断”和“无既有结果时非法源状态被拒绝”。

## 5 建议优先级与后续动作

- 最高优先级：后续进入 `backend-practice/003-mode-policies-and-provenance` 前，先在 plan-review 阶段确认 BDD 编号、required payload、状态转移 negative gates 已写入 checklist。
- 次优先级：若后续多个 owner 继续使用 Go HTTP scenario tests 承接 BDD，可把本次 `backend_practice_legacy.py` 的编号碰撞逻辑抽为通用 `scripts/lint` gate，而不是保留在 backend-practice 专属脚本内。
