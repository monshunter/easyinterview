# 004 — Derived Plans and Debrief Seeding

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

落地 backend-practice spec D-4 / D-14 / D-24 与 C-2 / C-3 预留的 derived practice plan 能力：

- `createPracticePlan` 支持 `goal IN ('retry_current_round','next_round') + sourceReportId`，写入 `practice_plans.source_report_id`，保持首题仍由 `practice.session.first_question` AI 生成。
- `createPracticePlan` 支持 `goal='debrief' + sourceDebriefId`，写入新列 `practice_plans.source_debrief_id`，并验证 source debrief 属于同一用户、同一 target job、状态已 completed 且至少有一条可用问题。
- `startPracticeSession` 在 `goal='debrief'` 时直接使用 debrief 已确认问题序列的首题作为第一 turn，不调用 F3 `practice.session.first_question` / A3 first-question AI；`mode` 仍只允许 `assisted` / `strict`，并继续控制 hint 策略。
- 所有 source id、question text、answer text 与 debrief notes 遵守 backend-practice D-11 隐私红线：API / event / audit / metric / log 只暴露 ID、状态、计数与错误码摘要。

本 plan 是 `backend-debrief/001` Phase 0.6 的阻塞依赖；完成后 `backend-debrief` 可恢复验证 `goal='debrief'` handoff。

## 2 背景

001 已把 baseline plan create/start 闭环并显式拒绝非 baseline goal，错误 detail 指向 future `004-derived-plans-debrief`。002/003 已完成 event loop、completion、hint mode policy 与 provenance；同时 003 验证了 `goal='debrief'` 在 mode policy 中只是一种 goal，不应变成第三个 `PracticeMode`。

当前阻塞事实：

- B1/B2/B4 已有 `PracticeGoal.debrief`，但 service 仍拒绝所有非 baseline goal。
- B4 baseline 已有 `source_report_id`，尚缺 `source_debrief_id` 与互斥 / goal CHECK。
- B2 `CreatePracticePlanRequest` / `PracticePlan` 尚无 `sourceReportId` / `sourceDebriefId` wire 字段。
- `startPracticeSession` 对所有 goal 都调用 `practice.session.first_question`；debrief flow 需要从 `debriefs.raw_questions` 取首题，证明复盘面试是对真实复盘问题的再练。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + migration + code-internal。
- **TDD 策略**: Code plan requires TDD。每个 checklist item 先补 Red：OpenAPI/generated/fixture contract、migration SQL contract、service validation、repository integration、handler envelope、HTTP scenario；再 Green 实现；最后 Refactor 与 checklist 同步。测试入口详见 [test-plan](./test-plan.md)。
- **BDD 策略**: Feature plan requires BDD。本 plan 改变用户可见 API 行为与复盘面试启动业务流，分配 `E2E.P0.070-073`，主 checklist 使用 `BDD-Gate:` 引用 [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md)。
- **替代验证 gate**: Contract / migration / privacy / legacy negative 使用 `make codegen-openapi`、`make lint-openapi`、`make validate-fixtures`、`migrations/lint.sh`、`make migrate-check`、focused Go tests、`make docs-check`、`git diff --check` 与 scoped grep 组合。

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` 新增 `report-derived` / `debrief-derived` / `source-missing` / `cross-user-source` scenarios | frontend-report-dashboard replay CTA 经 workspace/practice owner；frontend-debrief handoff 经 frontend-workspace-and-practice owner | `backend/internal/api/practice.Handler.CreatePracticePlan` + `backend/internal/practice.Service.CreatePracticePlan` + `backend/internal/store/practice.SQLRepository.CreatePlan` | `practice_plans.source_report_id` / `practice_plans.source_debrief_id` / `audit_events` / `idempotency_records` | none | `E2E.P0.070`, `E2E.P0.072` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` 补 derived source 字段场景 | workspace/practice state refresh | `Handler.GetPracticePlan` + `Service.GetPracticePlan` + `SQLRepository.GetPlan` | `practice_plans` read | none | `E2E.P0.070`, `E2E.P0.072` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` 新增 `debrief-derived-first-question` / `debrief-source-empty` scenarios | frontend-debrief "复盘面试" handoff 后的 practice route | `Handler.StartPracticeSession` + `Service.StartPracticeSession` + `SQLRepository.ReserveSessionStart` / `CommitSessionStart` | `practice_sessions` / `practice_turns` / `practice_session_events` / `outbox_events` / `audit_events` / `idempotency_records` | `practice.session.first_question` 仅 baseline / retry / next_round 调用；debrief 不调用 | `E2E.P0.071`, `E2E.P0.073` |

## 3.2 Coverage Matrix

| 行 | 类别 | source | plan_phase | verification | negative_scope |
|----|------|--------|------------|--------------|----------------|
| R1 | Cross-layer contract | D-24 source fields | Phase 0 | OpenAPI schema + generated Go/TS + fixtures + handler mapping tests | no hidden local-only source id fields |
| R2 | Migration | D-14 B4 source_debrief_id | Phase 0 | SQL contract + migration lint + migrate-check | no ALTER path; pre-launch baseline edit only |
| R3 | Primary | C-2 report-derived plan | Phase 1 | service/store/handler tests + `E2E.P0.070` | non-baseline goals no longer return owner=004 when valid source is present |
| R4 | Primary | C-3 debrief-derived plan | Phase 1 + Phase 2 | service/store/handler tests + `E2E.P0.071` | no `mode='debrief'`; only `goal='debrief'` |
| R5 | Failure / recovery | missing/cross-user/stale source | Phase 1 + Phase 3 | validation tests + cross-user 404/422 envelope tests + `E2E.P0.072` | no source existence leak across users |
| R6 | Boundary | debrief source empty or not completed | Phase 1 + Phase 2 | repository integration + start-session failure tests | no empty first turn; no draft debrief source |
| R7 | Privacy / observability | D-11 / D-14 | Phase 3 | audit/outbox/log/metric grep + HTTP scenario assertions | no debrief question text, answer text, notes, risk prose in audit/event/log/metric |
| R8 | Regression / legacy-negative | D-5/D-21 mode two-valued | Phase 3 | scoped grep + generated enum tests | `PracticeMode.debrief`, `mode=debrief`, `legacy debrief replay value` zero active references |

## 4 实施步骤

### Phase 0: Contract, Migration, And Plan Preflight

#### 0.1 B2 source fields

在 `openapi/openapi.yaml` 为 `CreatePracticePlanRequest` 与 `PracticePlan` 增加可空 `sourceReportId` / `sourceDebriefId`，更新 PracticePlans / PracticeSessions fixtures，运行 OpenAPI codegen 与 fixture validation。

#### 0.2 B4 source_debrief_id

在 `migrations/000001_create_baseline.up.sql` 为 `practice_plans` 增加 `source_debrief_id uuid`、FK 到 `debriefs(id) ON DELETE SET NULL`，并补 CHECK：

- `goal='baseline'` 时两个 source 均为空。
- `goal IN ('retry_current_round','next_round')` 时 `source_report_id IS NOT NULL` 且 `source_debrief_id IS NULL`。
- `goal='debrief'` 时 `source_debrief_id IS NOT NULL` 且 `source_report_id IS NULL`。

同步 SQL contract / migration lint / migrate-check。

#### 0.3 owner docs sync

将本 plan 链接到 backend-practice spec §7 与 plans INDEX；如果跨 owner 文档仍出现 `mode='debrief'` 作为 active contract，按当前 D-5/D-21 修订为 `goal='debrief' + mode IN ('assisted','strict')`。

### Phase 1: createPracticePlan Derived Source Validation

#### 1.1 DTO and service validation

扩展 domain DTO：`CreatePlanRequest` / `CreatePlanStoreInput` / `PlanRecord` 增加 source ids；service 按 goal 做互斥验证并返回 `422 VALIDATION_FAILED`：

- baseline 禁止任何 source id。
- retry / next_round 要求 `sourceReportId` 且禁止 `sourceDebriefId`。
- debrief 要求 `sourceDebriefId` 且禁止 `sourceReportId`。

#### 1.2 SQL source validation

repository 在插入时验证 source 同 user / same target job /可用状态：

- report source：`feedback_reports.user_id=user_id`、`target_job_id=targetJobId`、`status='ready'`。
- debrief source：`debriefs.user_id=user_id`、`target_job_id=targetJobId`、`status='completed'`、`jsonb_array_length(raw_questions) > 0`。

#### 1.3 handler and idempotency

handler 从 generated request 读取 source ids；idempotency fingerprint 必须包含 source ids，same key + different source 返回既有 mismatch；response replay 保留 source ids。

### Phase 2: debrief startPracticeSession First Turn Seeding

#### 2.1 reservation carries debrief question

`ReserveSessionStart` 在 selected plan 中读取 `source_debrief_id` 与 debrief `raw_questions[0]` 的 `questionText` / intent fallback；`SessionReservation` 增加 debrief first-turn 字段。

#### 2.2 service bypasses first_question AI for debrief

`StartPracticeSession` 在 `reservation.Goal == PracticeGoalDebrief` 时直接构造 first question，不调用 `ResolveActive(practice.session.first_question)` / `AIClient.Complete`；baseline / retry / next_round 保持现有 AI 首题路径。

#### 2.3 commit and event payload

`CommitSessionStart` 复用现有 started event / outbox / audit 结构，audit metadata 只包含 source ids 与计数摘要，不包含 debrief question text / answer summary / notes。

### Phase 3: Privacy, Isolation, And Legacy-Negative

#### 3.1 source isolation matrix

覆盖 missing source、cross-user source、wrong target job、draft debrief、empty raw_questions、source mismatch replay，确保错误 envelope 不泄露另一个用户的资源存在性。

#### 3.2 mode × goal regression

验证 debrief plan 在 `mode='assisted'` 与 `mode='strict'` 都可 start；hint 策略仍由 003 处理；`mode='debrief'` 在 OpenAPI / generated / fixtures / backend runtime 中不是合法值。

#### 3.3 privacy and audit redline

断言 audit_events / outbox_events / logs / metrics / idempotency response 不含 debrief question text、answer summary、interviewer reaction、notes、risk label prose。

### Phase 4: Gates And Handoff

#### 4.1 focused gates

运行 test-plan 中列出的 focused Go tests、OpenAPI/fixture/migration/docs gates。

#### 4.2 unblock backend-debrief

回到 `backend-debrief/001` checklist 0.6，记录本 plan 依赖与验证命令；随后执行 backend-debrief Phase 0.7 收口 gate。

## 5 验收标准

- `createPracticePlan` 对 baseline / retry_current_round / next_round / debrief 的 source id 规则与 DB CHECK 一致。
- `startPracticeSession(goal='debrief')` 返回第一条 debrief 问题作为 currentTurn，且不调用 first_question AI。
- `mode` 仍只有 `assisted` / `strict`，不会因为 debrief goal 引入第三值。
- `E2E.P0.070-073` 的 HTTP scenario 或等价场景证据全部通过。
- backend-debrief 0.6 被解除阻塞并记录依赖。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| derived source validation 与 DB CHECK 规则漂移 | service 表驱动测试 + SQL contract + migration lint 双向验证 |
| debrief 首题误走 first_question AI | fake registry / fake AIClient call-count 测试断言 debrief start 零调用 |
| source id 泄露跨用户资源存在性 | cross-user source 返回统一 validation/not-found envelope，BDD 反查错误 detail 不含 owner/source raw id 外的敏感上下文 |
| 旧 `mode='debrief'` 口径回流 | OpenAPI enum、generated enum、fixtures、docs/runtime scoped grep 作为 legacy-negative gate |
