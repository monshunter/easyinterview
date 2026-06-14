# Debrief Suggest SessionId Context 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-debrief/001-debrief-record-and-analysis` 的 `suggestDebriefQuestions` `sessionId?` contract 后续 sweep，闭环 fixture → generated handler → domain service → store context →真实 F3 prompt marker → P0.063 scenario gate。
- 修复内容：store 按 `(user_id, target_job_id, session_id)` 读取 completed practice session derived summary；service 替换 `{{mock_report_summary}}` / `{{job_summary}}` / `{{role_title}}` / `{{resume_highlights}}`；handler 增加 sessionId 映射测试；P0.063 wrapper 增加 sessionId PASS marker 与 no-op test 负向 gate。
- 成功证据：
  - Red：store session tests 先失败；service real prompt marker test 先失败。
  - Green：focused store/service/API tests 通过。
  - BDD：`test/scenarios/e2e/p0-063-debrief-suggest-questions/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` PASS，日志含 `E2E.P0.063 sessionId backend contract PASS`。
  - Regression：`cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1` PASS。
  - Contract/docs：`make validate-fixtures`、`make docs-check`、`sync-doc-index --check`、`git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- optional context key 没有成组审查。
  - **证据**：BUG-0121 已闭环同一 operation 的 `resumeId`，但 `sessionId` 仍只停在 fixture/handler/service 字段转发层，store 与 prompt 未闭环。
  - **影响**：如果只检查用户明确指出的单个字段，同一个 request body 中的相邻 optional context 仍可能 false-green。

- service tests 使用了内部 marker，未覆盖真实 F3 prompt marker。
  - **证据**：`TestServiceSuggestQuestions_ResumeContextInPrompt` 使用 `{{resumeSummary}}`，而真实 prompt `config/prompts/debrief.suggest_questions/v0.1.0.md` 使用 `{{resume_highlights}}` / `{{mock_report_summary}}`。
  - **影响**：prompt payload 看似有字段，真实 registry prompt 仍会带 unreplaced marker。

## 3 根因归类

- 根因：completed owner plan 的 Phase 3 checklist 只写了“可选 session 摘要”，缺少 sessionId-specific handler/store/prompt/scenario gate。
  - **类别**：spec-plan。

- 根因：P0.063 wrapper 在 BUG-0121 后强化了 `resumeId`，但没有要求同 fixture 中所有 optional context ids 都有独立 PASS marker。
  - **类别**：spec-plan / scenario gate。

- 根因：prompt marker compatibility 没有绑定真实 prompt file。
  - **类别**：spec-plan；现有 Bug Pattern 5 已覆盖“production caller prompt payload 关键上下文非空”，无需新增模式条目。

## 4 对流程资产的改进建议

- 在 backend-debrief/001 后续 L2 review 中，把 `suggestDebriefQuestions` 的 optional request context keys 当作一个集合审查：`targetJobId`、`sessionId`、`resumeId` 每个都需要 handler/service/store/prompt/scenario 证据。
  - **落点**：spec-plan
  - **优先级**：high

- 对所有 AI prompt context tests，至少保留一条使用真实 `config/prompts/<feature>/v*.md` marker 或等价真实 marker 字符串的测试。
  - **落点**：spec-plan / test checklist
  - **优先级**：medium

- Scenario verify 中保留 per-context PASS marker，而不是只写 operation-level PASS。
  - **落点**：scenario gate
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：后续若继续 review Debrief owner，优先 sweep `createDebrief` / `getDebrief` / `debrief_generate` 的 prompt marker 与 optional context payload，确认没有同类“字段存在但真实 prompt/SQL 未消费”的漂移。
- 可延后：把真实 prompt marker 覆盖沉淀为通用 lint 或 test helper；当前已通过 BUG-0122 和 P0.063 wrapper 固化本次范围。
