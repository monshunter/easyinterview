# Backend Debrief 001 Debrief Record And Analysis 交付复盘报告

> **日期**: 2026-05-16
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付范围为 `backend-debrief/001-debrief-record-and-analysis`：落地 `createDebrief` / `getDebrief` / `suggestDebriefQuestions` API，`debrief_generate` worker，debrief service/store/handler，AI task run 与 audit/outbox 写入，idempotency mismatch 错误码修正，E2E.P0.060-064 场景资产，以及 Phase 0 B1/B2/B3/B4/F3/backend-practice 前置 addendum 验证。

成功证据已写入 plan checklist 与 test-checklist：`cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race`、`cd backend && go test ./... -count=1`、`python3 -m pytest scripts/lint -q`、`python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations`、`python3 scripts/lint/rubric_lint.py --rubrics-dir config/rubrics`、`make codegen-check`、`make validate-fixtures`、`make lint-events`、`make codegen-events-check`、`./migrations/lint.sh`、`set -a; . deploy/dev-stack/.env; set +a; make migrate-check`、`make docs-check`、`git diff --check` 均通过。5 个 E2E.P0.060-064 场景四段脚本均顺序执行通过。

## 2 会话中的主要阻点/痛点

- Phase 0 跨 owner 前置面较大，B1/B2/B3/B4/F3/backend-practice 都必须先闭环，否则 backend plan 无法安全进入实现。
  - **证据**：plan Phase 0.1-0.7 分别记录 shared conventions、OpenAPI fixtures、events baseline、migration enum、prompt-rubric seed、backend-practice handoff 验证与全局 gates。
  - **影响**：实现前置成本高，但避免了 handler/store 先落地后再反向修契约的返工。

- `debrief.generate` prompt baseline 与当前 backend-debrief 输出语义曾不一致，必须在 close-out 前修正。
  - **证据**：实现阶段按 plan/schema 解析 `questions` + `riskItems`；原 `config/prompts/debrief.generate/v0.1.0*.md` 保留旧 `timeline` / `lessons` / `follow_up_actions` 风格输出口径。本轮已改为 `questions` / `riskItems` schema，并同步 prompt hash、rubric 描述与 seed migration；详见 [BUG-0065](../bugs/BUG-0065.md)。
  - **影响**：如果只依赖 mock AI 单测，会让真实 AI 运行时进入 `AI_OUTPUT_INVALID`；本轮通过 prompt/rubric/migration lint 把该风险闭环。

- idempotency fingerprint mismatch 仍带有旧 practice 专用错误码。
  - **证据**：`TestCreateDebrief_IdempotencyMismatch_DifferentBody` red 阶段暴露 mismatch 返回 `PRACTICE_SESSION_CONFLICT`；本轮改为通用 `IDEMPOTENCY_KEY_MISMATCH` 并补中间件与既有 practice/resume/upload 回归测试。
  - **影响**：说明共享中间件的错误码不能只由首个调用方语义命名，后续多 owner 复用时需要负向断言。

- legacy negative gate 需要从文字 grep 升级为结构化 lint。
  - **证据**：plan 原始 gate 使用 `grep -rn`，但 docs 内需要保留 retired terms 作为负向口径说明；本轮新增 `scripts/lint/backend_debrief_legacy.py`，只扫描 runtime/contract/fixture/scenario 目标面并配套 pytest。
  - **影响**：减少误报，同时把旧 Mistakes/Growth/Drill/Voice 口径的零残留要求变成可复用 gate。

## 3 根因归类

- 跨 owner 契约依赖集中在同一个 backend plan 内收口。
  - **类别**：spec-plan

- F3 prompt baseline 的输出 schema 原本没有被 backend-debrief plan 的 worker parser gate 反向锁住。
  - **类别**：spec-plan

- idempotency middleware 的错误码历史命名没有共享契约级回归断言。
  - **类别**：README

- legacy negative gate 需要区分运行时真理源和文档负向说明。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 backend-debrief 后续 owner plan 或 F3 lint 中新增“prompt 输出 schema 与 worker parser 期望一致”的显式 gate，避免未来 prompt 文案修订再次回到旧 shape。
  - **落点**：spec-plan
  - **优先级**：high

- 在 shared idempotency middleware 文档或测试规范中增加“错误码必须为调用方无关通用码”的回归要求，避免新 owner 复用时泄漏首个业务域语义。
  - **落点**：README
  - **优先级**：medium

- 对需要保留 retired term 负向说明的计划，优先使用结构化 lint 脚本限定扫描面，不把 docs 说明文字直接纳入零残留 grep。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是把“prompt 输出 schema 必须反查 worker parser”的 gate 固化到后续 F3/backend-debrief 计划或 lint 中，防止同类 drift 复发。随后可把 idempotency 通用错误码要求沉淀到共享中间件 README 或测试约定。legacy gate 脚本已在本轮落地，后续同类 owner 可以复用本模式。
