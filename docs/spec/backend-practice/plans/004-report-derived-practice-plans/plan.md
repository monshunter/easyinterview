# 004 — Report-derived Practice Plans

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本计划只承接 report-derived practice plan 当前契约：

- `createPracticePlan` 支持 closed derived request `goal IN ('retry_current_round','next_round') + sourceReportId`，写入 `practice_plans.source_report_id`；请求不接收 focus、settings 或 identity 副本。
- `getPracticePlan` 返回当前 plan 的 `sourceReportId`，同用户重放保持一致。
- `startPracticeSession` 对 report-derived plan 走统一 `practice.session.chat` opening message 路径，不从报告注入预设消息。
- backend-practice 只构造 JSON 编码的 `semanticFocus` runtime payload；F3/002 拥有 immutable `practice.session.chat/v0.2.0` prompt/schema+rubric pair、hash/version parity 与最终激活。
- `goal='debrief'`、`sourceDebriefId`、`source_debrief_id`、`PracticeGoalDebrief` 和 debrief-derived message seeding 是禁止输入 / 禁止契约字段，只能出现在负向断言中。

## 2 背景

当前 `backend-practice/spec.md` 把 `PracticeGoal` 定义为 `baseline / retry_current_round / next_round`，并把 `sourceReportId` 作为唯一派生计划 source 字段。


## 3 质量门禁分类

- **Plan 类型**: feature-behavior + OpenAPI consumer + backend service/store + prompt context + scenario。

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | report-derived plan fixtures | Report retry/next actions | practice plan owner | source report + plan/idempotency/audit | none | 当前无真实 E2E owner；root `make test` |
| `getPracticePlan` | current plan fixture | Workspace/practice refresh | practice plan read owner | plan read | none | 当前无真实 E2E owner；root `make test` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` current plan goals only | Interview session start | `Handler.StartPracticeSession` + `Service.StartPracticeSession` + `SQLRepository.ReserveSessionStart` / `CommitSessionStart` | `practice_sessions` / `practice_messages` / `practice_session_events` / `outbox_events` / `idempotency_records` | `practice.session.chat` opening for baseline / retry_current_round / next_round | Covered by active backend-practice start-session gates; no report-seeded message bypass |

## 3.2 Coverage Matrix

| 行 | 类别 | source | verification | negative_scope |
|----|------|--------|--------------|----------------|
| R3 | Cross-layer contract | B2 `sourceReportId`, B4 `source_report_id`, generated Go/TS | OpenAPI inventory + generated artifact search + fixture validation owner gates | no `sourceDebriefId` / `source_debrief_id` positive fields |
| R4 | Regression / negative | prohibited source fields / goals | negative grep across runtime/generated/fixtures/scenario docs | no `PracticeGoalDebrief`, `goal='debrief'`, debrief start scenario, or seeded-message bypass |
| R5 | Regression / naming | current report-local dimension focus | generated/API/store/prompt/scenario exact-set gate | no positive `focusCompetencyCodes`, `focus_competency_codes`, `retryFocusCompetencyCodes`, `retry_focus_competency_codes`, `retryFocusTurnIds`, `retry_focus_turn_ids`, `retry_round` or code-only prompt focus |

## 4 实施步骤

### Phase 1: Report-derived Plan Contract

#### 1.1 Source report request / response contract

Keep `sourceReportId` as the only derived-plan source field in `CreatePracticePlanRequest` and `PracticePlan`.

#### 1.2 Report source validation

Keep service/store validation for `retry_current_round` and `next_round`: `sourceReportId` is required, must belong to the same user and target job, and must not leak cross-user source existence.

#### 1.3 Start-session behavior

Keep report-derived starts on the regular AI conversation-opening path. Do not add a report-seeded message bypass or raw transcript seed.

### Phase 2: Source Boundary Reconciliation

#### 2.1 Prohibited source removal

Ensure current docs, context, fixtures, generated clients, runtime code, and scenarios do not list `sourceDebriefId`, `source_debrief_id`, `PracticeGoalDebrief`, or `goal='debrief'` as current positive contract.



### Phase 3: Server-owned dimension focus and frozen identity

- OpenAPI exposes closed derived request `{goal, sourceReportId}` and deletes request `focusCompetencyCodes` plus copied persona/difficulty/language/time-budget/target/resume/round fields.
- In one transaction, retry validates owned/ready/current-shape source report and exact target/resume/round, then projects `retryFocusDimensionCodes` into `practice_plans.focus_dimension_codes`. Empty focus is valid and means generic same-round retry; non-empty focus must be unique and every code must resolve to an issue-backed `needs_work` dimension. next selects the frozen canonical successor and persists empty focus.
- Derived request only contains goal + sourceReportId. Server reuses source persona/difficulty/language, derives target/resume/current-or-successor round and duration, and rejects missing/cross-user/non-ready/invalid-context/unsupported-or-duplicate-non-empty-focus/idempotency cases without insert/leak; empty focus remains valid.
- Retry start/send with non-empty focus resolves each stored report-local code back to immutable source report label + issue summaries and injects that structured semantic focus as untrusted prompt context. Empty focus starts a generic same-round retry without fabricated dimension guidance. Code-only prompt focus, non-empty missing/unsupported cross-ref, raw transcript/anchors and next/baseline report focus are forbidden.
- Runtime builder and focused tests use `semanticFocus` / `{{semantic_focus_json}}`. Before release, backend-practice validates the F3-owned v0.2 exact candidate rather than requiring `ResolveActive`; F3/002 alone activates the prompt/rubric pair after its marker gate. v0.1 remains immutable rollback material and must not be rewritten by this plan.
- Existing sourceReportId response/start behavior remains; no report-seeded transcript/message bypass.

The current-name negative gate requires zero positive runtime/generated/OpenAPI/fixture/scenario hits for `focusCompetencyCodes`, `focus_competency_codes`, `retryFocusCompetencyCodes`, `retry_focus_competency_codes`, `retryFocusTurnIds`, `retry_focus_turn_ids`, `retry_round` and code-only report focus. Historical migrations, history rows, immutable `practice.session.chat/v0.1.0` rollback assets and explicitly named negative fixtures may retain literals, but the final active v0.2 runtime cannot consume them. This owner alone emits `REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS` after the exact set is clean and F3 emits `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS`.


## 5 验收标准

- Current plan/test/BDD/context describe only report-derived `retry_current_round` / `next_round` behavior as positive scope.
- `sourceDebriefId` / `source_debrief_id` / `PracticeGoalDebrief` / `goal='debrief'` do not appear in active runtime, generated artifacts, OpenAPI fixtures, or this plan's positive gates.
- Empty retry focus creates a generic same-round plan; every non-empty focus code is issue-backed; the F3 v0.2 active pair consumes structured semantic focus, and final active runtime has zero positive old competency/turn/action identifier consumers outside explicit rollback/history/negative allowlists.
- `validate_context.py`, `make docs-check`, and `git diff --check` pass.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 范围外 source 字段看起来可用 | 只在负向断言中枚举禁止字段，正向 contract 只列 `sourceReportId` / `source_report_id` |
| 禁止 source 字段回流到 generated artifacts | Negative grep 覆盖 OpenAPI、generated Go/TS、fixtures、backend runtime 和 scenario docs |
| backend runtime 与 F3 active version 提前耦合 | 发布前 Phase 3 只断言 exact v0.2 coordinate；F3 marker 后 final gate 断言 `ResolveActive` v0.2，8-status rollback 与 000019 仍由 F3/002 独占 |

## 7 修订记录

| 日期 | 版本 | 变更 | 原因 |
|------|------|------|------|
| 2026-07-12 | 1.7 | Hand off semantic-focus prompt assets and activation to F3/002 through immutable practice v0.2 exact-candidate and final marker gates. | Coordinated scheme A activation. |
| 2026-07-12 | 1.6 | Make derived-plan legacy negative marker owner-specific and include snake_case focus/turn identifiers in the exact set. | Cross-owner evidence reconcile. |
| 2026-07-12 | 1.5 | Allow generic empty-focus retry, require issue-backed non-empty focus, close server-derived request semantics and lock legacy-name negatives. | Owner reconcile. |
| 2026-07-12 | 1.4 | Reopen for server-owned report-local dimension focus and frozen-context derived-plan validation. | Grounded report plan A. |
| 2026-07-12 | 1.3 | Align report-derived start with `practice.session.chat` and `practice_messages`. | Practice is now one continuous conversation without turn/question structures. |
| 2026-07-06 | 1.2 | Rename owner path to `004-report-derived-practice-plans`; current contract remains report-derived retry / next-round only. | Product-scope pruning requires current owner docs to use current owner language. |
| 2026-07-06 | 1.1 | Reconcile completed plan after product-scope D-22: current positive scope is report-derived retry / next-round only; out-of-scope source fields move to negative assertions. | Completed plan/context was still a discovery source and could reintroduce deleted work. |
| 2026-05-16 | 1.0 | Initial implementation of derived practice plans. | Initial contract delivery. |
