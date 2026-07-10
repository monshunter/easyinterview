# 001 — Report Screen and Generating Handoff

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 0.0 D-19 CTA Convergence Reopen (v1.3)

2026-06-12 [product-scope v2.1 D-19](../../../product-scope/spec.md#31-已锁定决策) 报告 CTA 单点收敛。本 plan Phase 6 承接 [spec v1.3 §3.1 D-19 / C-16 / C-17](../../spec.md#31-已锁定决策)：next（复练计划）tab 删除 `report-next-cta-a` / `report-next-cta-b` 两个重复 CTA 按钮，改以 footer 文案引导 Header CTA；题目回顾 `加入本轮复练`（`report-questions-add-to-replay`）从 `goReplay` nav 改为 per-question 本地标记 toggle。v1.5 同步 B2/backend-review 术语，`report-generating` fixture 表达 generating 状态元数据。v1.6 对齐当前 route fallback 命名，不再引用已移除的旧组件名。v1.7 将 ContextStrip modality 文案收敛为「文本面试 / 电话模式」。v1.8 将 preflight/risk wording 收敛为 fixture-only workaround 口径。v1.9 将 loading / empty / skeleton 口径收敛为 skeleton state、empty state 与 layout-space 术语。

**Phase 6 质量门禁**：纯前端 UI 收敛，无 OpenAPI/backend/契约变更；BDD 不适用（既有 `E2E.P0.056-059` 场景维持回归）。替代验证 gate 为 focused Vitest（NextTab 无 CTA、QuestionsTab 本地标记不 nav）+ topbar/report pixel parity + removed CTA testid 负向断言 + `pnpm typecheck/test/build`。删除属 TDD：先以 Red（断言 `report-next-cta-a/b` 缺失、点 add-to-replay 不触发 navigate）表达目标态，再改实现转绿。

### Phase 6: D-19 report CTA single-point convergence

#### 6.1 next tab 删除重复 CTA 按钮

`NextTab.tsx` 删除 `report-next-cta-a` / `report-next-cta-b` 按钮与 `onReplay` / `onNextRound` props；路径 A/B 卡片保留说明与复练清单，新增 footer 文案「开练入口在页面顶部：复练当前轮 / 进入下一轮」（对照 `ui-design/src/screen-report.jsx` PATH A/B footer）；`DetailSurface.tsx` 去掉 NextTab 的 `onReplay`/`onNextRound` 传参。

#### 6.2 题目回顾 `加入本轮复练` 改本地标记

`QuestionsTab.tsx` 的 `report-questions-add-to-replay` 从 `onAddToReplay={goReplay}`（nav workspace）改为 per-question 本地 toggle state（`markedForReplay` map，对照原型 `replayQueued` + `toggleQueued`）：点击只 toggle 当前题目标记，文案 `加入本轮复练` ↔ `已加入本轮复练`，不 `nav`/不调 API/不改 URL；`DetailSurface.tsx` 去掉 `onAddToReplay` 传参；新增 i18n `report.questions.detail.addedToReplay`。

#### 6.3 Phase 6 operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `N/A`（UI-only：CTA 收敛 + 本地标记） | N/A | `NextTab.tsx` / `QuestionsTab.tsx` / `DetailSurface.tsx`；Header CTA 维持既有 `useReplayCtaHandlers` | 无 API 变更 | 无（本地 React state） | 无 | focused Vitest + report/topbar pixel parity + removed CTA testid 负向；E2E.P0.056-059 回归 |

#### 6.4 Phase 6 回归与负向 gate

`report-next-cta-a` / `report-next-cta-b` testid 在 report 模块源码与渲染 DOM 0 命中（负向断言测试除外）；`report-questions-add-to-replay` 点击不触发 `navigate` / `useRequestAuth`；`pnpm --filter @easyinterview/frontend typecheck/test/build` 通过；report + topbar pixel parity 通过；`frontend_report_dashboard_out_of_scope` lint 通过。

## 0.1 Generating hook test runtime isolation (v1.10)

### Phase 7: remove unrelated provider effects from hook unit tests

`useReportGenerationPoll.test.tsx` only exercises the poller against an injected generated client. Mount `AppRuntimeContext` directly, following the established frontend test pattern, and delete the fake client's unused runtime-config/auth methods. This removes asynchronous `AppRuntimeProvider` state updates and their React `act(...)` warnings without changing production hook behavior, OpenAPI, UI, BDD or scenario contracts.

**Phase 7 gate**: the focused Vitest command must pass 11 tests with no `not wrapped in act` output; full frontend Vitest, typecheck/build, owner context and docs/pruning gates must remain green. BDD is not applicable because this is test-harness isolation only.

## 0.2 P0.059 browser evidence reconciliation (v1.11)

### Phase 8: align pixel-parity claims with executable evidence

P0.059 owns deterministic report and generating browser smoke at desktop and mobile viewports. The executable contract is DOM/state visibility, viewport geometry, overflow bounds and an in-memory non-empty screenshot for every covered state. It does not own theme matrices or persistent image-comparison files; those claims are removed instead of being kept as conditional or future behavior.

**Phase 8 gate**: a focused preflight must reject expanded browser claims in the active spec and six plan artifacts, and require both Playwright sources to call `page.screenshot()` with a non-empty buffer assertion. `E2E.P0.059` must execute that preflight, build the frontend, run both Playwright specs and pass setup/trigger/verify/cleanup. No OpenAPI, backend, persistence, AI dependency or product UI behavior changes.

## 0.3 P0.057 direct-start contract reconciliation (v1.12)

### Phase 9: remove the obsolete workspace side-effect handoff

The current UI source, frontend-workspace-and-practice D-9 and production code all start replay/next-round practice from the report owner through generated `createPracticePlan` / `startPracticeSession`, then navigate directly to `practice`. Signed-out users authenticate and return to `report` to retry the CTA. Phase 9 removes the contradictory workspace mount side-effect contract from the spec, plan/test/BDD documents and P0.057 assets.

**Phase 9 gate**: owner preflight rejects the obsolete workspace side-effect terms and requires the direct-start generated-client calls, `practice` navigation, signed-out `report` return and P0.057 preflight wiring. P0.057 setup/trigger/verify/cleanup, owner context and global docs/pruning gates must pass. No API schema or product UI changes.

## 0.4 P0.056 focused-runner evidence reconciliation (v1.13)

### Phase 10: bind scenario claims to the five focused owner test files

P0.056 composes preflight, polling-hook, GeneratingScreen, ReportScreen and DetailSurface Vitest files plus static verification. It is not a single browser or live-backend journey, and it does not prove fixed poll counts, exact fixture sequencing, theme switching or transcript persistence. Phase 10 compresses the scenario and BDD claims to the assertions those five files actually execute and requires verify to observe every test-file marker.

**Phase 10 gate**: owner preflight rejects the expanded claims and pre-flatten Resume vocabulary, requires all five trigger paths and verify markers, and keeps the real-mode configuration gate distinct from fixture-backed UI evidence. P0.056 setup/trigger/verify/cleanup and owner/global gates must pass. No runtime behavior changes.

## 0.5 P0.058 failure-contract evidence reconciliation (v1.14)

### Phase 11: separate hook, component and route-state failure evidence

P0.058 composes owner preflight plus five focused files for ReportFailureState, ReportMissingSessionState, useFeedbackReport, ReportScreen and useReportGenerationPoll. The poll hook covers failed/404/timeout state, but the scenario does not mount GeneratingScreen or prove a multi-timeout UI fallback. Report components separately cover error-copy, CTA handlers, missing inputs and not-found rendering.

**Phase 11 gate**: preflight rejects GeneratingScreen UI and broad URL/storage/telemetry claims not executed by this runner, requires all six trigger/verify markers, and distinguishes real-mode configuration from deterministic failure clients. P0.058 setup/trigger/verify/cleanup and owner/global gates must pass. No runtime behavior changes.

## 0.6 Active visual contract reconciliation (v1.15)

### Phase 12: bind the active spec to P0.059 browser evidence

The active spec and all six plan artifacts must describe the browser contract actually executed by P0.059: seven deterministic generating/report states across desktop and mobile, state-specific DOM or viewport assertions, report TopBar visibility, mobile overflow bounds and one non-empty in-memory screenshot per state. Theme ownership remains with frontend-shell, while component structure and detail-tab behavior remain covered by the existing source-parity Vitest gates.

**Phase 12 gate**: extend the owner preflight to read the active spec, reject unsupported visual and responsive claims across all seven owner artifacts, and preserve the current Playwright source/runner checks. Reconcile the spec, owner plan, test/BDD artifacts and P0.059 scenario wording; then pass focused Vitest, both Playwright specs, P0.059 setup/trigger/verify/cleanup, owner/product contexts and global docs/pruning gates. No runtime behavior, API, persistence or AI dependency changes.

## 0.7 Unconsumed report error helper removal (v1.16)

### Phase 13: remove the isolated AI-error predicate branch

Delete `isAiErrorCode` and its private `FAILURE_AI_ERROR_KEYS` array. Repository inventory proves neither has a consumer; `ReportFailureState` continues to use `failureErrorCodeKey` and its existing `FAILURE_LABEL_BY_CODE` mapping. BDD is not applicable because the removed branch has no executable caller. Alternative gates are a scoped source-surface red/green test, focused failure-state tests, typecheck and owner/global checks.

## 1 目标

把 [frontend-report-dashboard spec](../../spec.md) v1.0 §2.1 / §6 C-1~C-15 / §7 锁定的第一个 plan 范围落地，承接 [frontend-workspace-and-practice/002-practice-text-event-loop](../../../frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)（completed, 2026-05-14）已交付的 `completePracticeSession` 真实调用 + 13 字段 generating route params（`buildPracticeHandoffParams` 输出：`{planId, targetJobId, jdId, resumeId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}`）+ `ReportWithJob` 缓存 + `nav("generating", ...)` handoff，闭合 P0 用户路径中"报告生成过渡态 → 证据化报告 dashboard → 复练 / 下一轮 handoff"段：

- `generating` 路由从 route fallback shell 切换为正式 `GeneratingScreen`，源级复刻 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399)：页头 + 进度条 + 5 阶段列表（done/active/pending 状态圆圈，每阶段 700-1200ms 动画节奏）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA "<12s target" + 「通知我」UI-only 按钮）。
- 通过 generated client + fixture-backed transport 消费 `getFeedbackReport(reportId)`；指数退避轮询（初始 1.5s × 1.5 上限 8s，max attempts 30）+ visibility 暂停 / 恢复；status='ready' → nav `report?sessionId=&reportId=&...passThrough`；status='failed' → nav `report?reportStatus=failed&errorCode=&...`；max attempts 达到 → ErrorState 「报告生成超时」+ retry / 返回 workspace。
- `report` 路由从 route fallback shell 切换为正式 `ReportScreen`，源级复刻 `ui-design/src/screen-report.jsx::ReportScreen` (lines 1-516) 三态：`ReportDashboard`（正常报告）/ `ReportFailureState`（reportStatus='failed'）/ `ReportMissingSessionState`（缺 sessionId）。
- ReportDashboard 完整源级复刻：返回按钮 + Header + `ReportContextStrip`（sessionId / targetJob / round / resume / modality / practiceMode / hints）+ 4 个 Summary Cards（准备度 / 维度 / 题目 / 下一步）+ `ReportDetailSurface` 5 个 tab（readiness / dimensions / questions / evidence / next）+ 维度卡片行 + 优先级 + 复练重点 + 题目回顾概览（5 题）+ 风险 & 亮点 + 复练 CTA（路径 A 复练当前轮 / 路径 B 进入下一轮）。
- 复练 CTA 行为：路径 A `goReplay()` 构造 retry_current_round + source report/session + replayItems/evidenceGaps payload；路径 B `goNextRound()` 构造 next_round + source report/session + nextRoundId payload。已登录时共享 `startPracticeFromParams` 通过 generated client 创建派生 plan、启动 fresh session 并直接 `nav("practice", ...)`；未登录走 `useRequestAuth({type:'replay_practice', route:'report', params:route.params})`，鉴权后回报告重试；不能复用报告来源 session。
- i18n 双语：新增 `report.*` + `generating.*` 命名空间（≥ 60 keys）；不复用 `workspace.*` 或 `practice.*`。
- stale-contract negative grep + Playwright pixel parity (desktop 1440×900 + mobile 390×844) + i18n 完整性断言。

完成后用户在 frontend-workspace-and-practice plan 002 已经送出 `nav("generating", ...)` 后能进入真实 GeneratingScreen 轮询 → 真实 ReportScreen 渲染 → Header CTA 创建 fresh plan/session 并直接进入 practice，完整闭环 P0 报告路径。Backend 真实数据由 [backend-review plan 001](../../../backend-review/plans/001-report-generation-baseline/plan.md) 提供；2026-05-23 L2 remediation 后，P0.056-P0.059 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker。fixture-backed UI variants 继续用于 DOM / 状态分支确定性测试，但不能单独代表真实 backend 闭环。

## 2 背景

frontend-workspace-and-practice spec v1.3 + plan [002-practice-text-event-loop](../../../frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)（completed, 2026-05-14）已交付 `completePracticeSession` 真实调用 + `useCompletePracticeSession` hook + `buildPracticeHandoffParams` 13 字段构造器（`{planId, targetJobId, jdId, resumeId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}`）+ `nav("generating", ...)` 跳转 + `ReportWithJob` 缓存防双触发。本 plan 接管 `frontend/src/app/App.tsx::renderRouteScreen` 中 `generating` 与 `report` 的 route fallback shell。

> **关于 `useRequestAuth` payload type 与 UI 真理源的必要适配**：UI 真理源 `ui-design/src/screen-report.jsx` 用通用 `create_session` action 直接进入 practice。生产前端保留专用 `replay_practice` type 以表达鉴权意图，但 pendingAction 回到 `report`，用户重试 Header CTA 后再由 generated client 创建 plan/session；该差异只影响鉴权恢复，不改变“从报告直接开始对应 session”的交互。

backend-review spec v1.0 + plan [001-report-generation-baseline](../../../backend-review/plans/001-report-generation-baseline/plan.md) 是本 plan 的 backend handoff source；backend-review Phase 0 会扩 `openapi/fixtures/Reports/getFeedbackReport.json` 的 `report-failed` variant + `openapi/fixtures/Reports/listTargetJobReports.json` 的 `empty` variant + B1 `REPORT_NOT_FOUND` 错误码 + B2 `getFeedbackReport` 404 response schema。本 plan 仅消费这些 fixtures + generated client，不直接修改 OpenAPI / B1 / B2 truth source。

UI 真理源：[`ui-design/src/screen-report.jsx`](../../../../../ui-design/src/screen-report.jsx)（`ReportScreen` 主控制器三态分发 + `ReportDashboard` 主组件 + 5 个 detail tab 内容 + `ReportStatButton` / `DimRow` / `StatCard` primitives）+ [`ui-design/src/screens-p0-complete.jsx`](../../../../../ui-design/src/screens-p0-complete.jsx) lines 269-399（`ReportGeneratingScreen`）+ [`ui-design/src/data.jsx`](../../../../../ui-design/src/data.jsx) lines 175-242（`report` sample 数据结构）+ [`docs/ui-design/report-dashboard.md`](../../../../ui-design/report-dashboard.md)（dashboard-only 形态 + 准备度 4 档 + 维度状态 + 复练规则）。

001 plan 不依赖 frontend-workspace-and-practice 任何待落地 plan（003 voice / 004 deferred handoff）；004 计划本来预留为本 spec 同款 generating handoff plan，本 spec v1.0 创建后 004 编号回收并由 frontend-report-dashboard 完整承接（与 frontend-workspace-and-practice spec.md §7 关联计划 "不再预留本 subspec 内的 report 或 company_intel plan" 一致）。

InterviewContext reducer 已经在 frontend-workspace-and-practice plan 001 + 002 落地完整 13 字段写入（与 `buildPracticeHandoffParams` 输出一致）；本 plan 仅 read，不新增 reducer action（与 spec D-8 一致）。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）+ cross-layer contract（消费 backend-review 已落地的 wire 契约 + frontend-workspace-and-practice 已落地的 handoff）。
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest + @testing-library/react + jsdom）；每个 Phase 在新增组件 / hook / utils 前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言（method、path、query、body schema、header 反查）、URL/state 隐私反查与 stale-contract testid / enum / reportLayout 字面量负向断言；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 5 扩展为 `report.spec.ts` + `generating.spec.ts`（desktop 1440×900 + mobile 390×844）。新增组件位于 `frontend/src/app/screens/report/` 与 `frontend/src/app/screens/generating/`；测试文件与组件 colocate（`*.test.tsx` / `*.test.ts`）。`test-plan.md` + `test-checklist.md` 拓展测试形态与 phase 映射。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 [bdd-plan.md](./bdd-plan.md) 定义 4 个场景 `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059`，[bdd-checklist.md](./bdd-checklist.md) 跟踪每个场景资产创建与执行；主 [checklist.md](./checklist.md) 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖 + pixel parity + contract drift）。

## 3.5 Coverage Matrix

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path · GeneratingScreen happy path | 进入 generating 携带 reportId + 13 字段；渲染 5 阶段进度动画 + 实时观察流；轮询 `getFeedbackReport(reportId)` 多次（指数退避节奏）；status='ready' 自动 nav report | `screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 | 1 | E2E.P0.056 + Vitest `generating/GeneratingScreen.test.tsx` + `hooks/useReportGenerationPoll.test.ts` |
| Primary path · ReportDashboard 渲染（ready） | 携带 sessionId + reportId 进入 report；`getFeedbackReport` 返回完整 FeedbackReport；渲染 Header + ContextStrip + 4 Summary Cards + 5 Detail Tabs（默认 `questions`）+ 维度卡片 + 优先级 + 复练重点 + 题目回顾 + 风险亮点；readiness tab 通过显式切换覆盖 | `screen-report.jsx::ReportDashboard` lines 80-257 | 2+3 | E2E.P0.056 + Vitest `report/ReportScreen.test.tsx` + `report/ReportDashboard.test.tsx` |
| Primary path · 5 detail tab 内容 | 切换 5 个 tab；每个 tab 内容源级复刻：readiness（拨号盘 + JD 对齐 + 证据密度 + 下一档门槛）/ dimensions（二级网格）/ questions（侧栏 + 当前题分析）/ evidence（风险 + 亮点）/ next（路径 A vs 路径 B 对比） | `screen-report.jsx::ReportDetailSurface` lines 311-516 | 3 | E2E.P0.056 子断言 + Vitest `report/DetailTabs.test.tsx` |
| Primary path · 复练 CTA 路径 A | 已登录用户点击「复练当前轮」→ generated client 创建 retry_current_round 派生 plan、启动 fresh session并直接进入 practice；未登录鉴权后回 report 重试 | `screen-report.jsx::goReplay` | 4+9 | E2E.P0.057 + Vitest `report/ReplayCta.test.tsx` |
| Primary path · 复练 CTA 路径 B | 已登录用户点击「进入下一轮」→ generated client 创建 next_round 派生 plan、启动 fresh session并直接进入 practice；未登录鉴权后回 report 重试 | `screen-report.jsx::goNextRound` | 4+9 | E2E.P0.057 + Vitest `report/ReplayCta.test.tsx` |
| Alternate path · GeneratingScreen 失败处理 | fixture report-failed（status='failed' + errorCode）→ 自动 nav report?reportStatus=failed&errorCode=... | spec §2.1 generating 失败分支 | 1 | Vitest `useReportGenerationPoll.test.ts` 子用例 + E2E.P0.058 |
| Alternate path · GeneratingScreen 超时 | fixture 永久 generating → max attempts 达到 ErrorState「报告生成超时」+ retry 重启轮询 | spec D-3 | 1 | Vitest `GeneratingScreen.test.tsx` 子用例 |
| Alternate path · GeneratingScreen visibility 暂停 | tab 隐藏 → 暂停轮询；恢复显示 → 恢复轮询 | spec D-3 | 1 | Vitest fake visibility event + `useReportGenerationPoll.test.ts` |
| Failure / recovery · ReportFailureState | `report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`；不调 `getFeedbackReport`；渲染失败卡片 + errorCode 文案映射 + CTA「重新生成」/「返回 workspace」 | `screen-report.jsx::ReportFailureState` lines 61-77 | 2 + 5 | E2E.P0.058 + Vitest `ReportFailureState.test.tsx` |
| Failure / recovery · ReportMissingSessionState | `report?reportId=R`（缺 sessionId）→ 不调 `getFeedbackReport`；渲染缺会话卡片 + CTA「返回 workspace」 | `screen-report.jsx::ReportMissingSessionState` lines 46-59 | 2 | E2E.P0.058 + Vitest `ReportMissingSessionState.test.tsx` |
| Failure / recovery · `getFeedbackReport` 5xx 错误 | network 5xx / timeout → InlineError + retry 按钮 + retry 复用 reportId；3 次失败显示「返回 workspace」fallback | spec §3.2 待确认事项 | 5 | Vitest `report/ReportErrors.test.tsx` + E2E.P0.058 子断言 |
| Failure / recovery · `getFeedbackReport` 404 cross-user | 用户 B 访问 R_A → 404 `REPORT_NOT_FOUND` → 渲染 ReportFailureState 但使用独立的 `report.failureState.notFound.{title,desc}` / `report.failureState.errorCode.REPORT_NOT_FOUND` i18n key（与 `failureState.errorCode.AI_*` 通用映射区分），文案为 zh "未找到该报告" / en "Report not found"；不暴露 reportX 存在性；不与 AI errorCode 文案混淆 | backend-review D-15 + B1 `REPORT_NOT_FOUND` | 5 | Vitest `report/CrossUserNotFound.test.tsx` + `report/ReportFailureStateNotFound.test.tsx` + E2E.P0.058 子断言 |
| UI stale-contract negative · `listTargetJobReports` 0 调用 | `listTargetJobReports` 在 `frontend/src/app/screens/{report,generating}/` 范围零调用（三层断言：scoped out-of-scope grep + Vitest mockTransport spy + scenario verify.sh）；dashboard-only D-7 边界 | spec §2.2 + frontend-workspace-and-practice 边界 | 全 phase | Vitest `mockTransport.spy.test.ts` 反向断言 + `scripts/lint/frontend_report_dashboard_out_of_scope.py` literal grep + scenario verify.sh |
| Boundary · GeneratingScreen mount 时 reportId 缺失 | `generating?sessionId=S`（缺 reportId）→ 不发请求；渲染 ErrorState「报告 ID 缺失」+ 返回 workspace | spec D-3 | 1 | Vitest `GeneratingScreen.test.tsx` 子用例 |
| Boundary · ReportDashboard mount 时 reportId 缺失 | `report?sessionId=S`（缺 reportId）→ 不发请求；渲染缺 reportId ErrorState | spec D-3 | 2 | Vitest `ReportScreen.test.tsx` 子用例 |
| Boundary · 复练 CTA 在数据未 ready 时禁用 | report status='generating' 时（理论上不应进入但兜底）CTA disabled；不发 nav | spec D-5 | 4 | Vitest `ReplayCta.test.tsx` 子用例 |
| Cross-layer contract · getFeedbackReport schema | response 含 status / preparednessLevel / highlights / issues / nextActions / questionAssessments / provenance / retryFocusTurnIds / errorCode（按 status 不同字段填充策略）；不暴露 runtime 字段 | OpenAPI `FeedbackReport` + spec D-4 | 1+2+3 | Vitest `mockTransport.spy.test.ts` 子用例 + Vitest schema parity test |
| Cross-layer contract · GenerationProvenance wire 6 字段 | 任何 provenance 渲染只读 6 wire keys（promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion）；runtime 字段不出现在 UI 或 props | spec D-13 + backend-review D-9 | 2+3 | Vitest `report/AiTransparency.test.tsx`（如适用）+ 负向 grep |
| Cross-layer contract · 不调 `Idempotency-Key` 在 getFeedbackReport | hook 内部断言 request init 不含 `Idempotency-Key` header；read 路径无副作用 | OpenAPI `getFeedbackReport` (无 idempotency 要求) | 1 | Vitest `useReportGenerationPoll.test.ts` 反向断言 |
| Cross-layer contract · `Prefer: example=<scenario>` fixture variant 切换 | hook 通过 fixture-backed transport 切换 variant；scenario verify.sh 通过 `EI_FIXTURE_SCENARIO_*` 环境变量驱动 | `frontend/src/api/mockTransport.ts` createFixtureBackedFetch | 全 phase | Vitest `mockTransport.spy.test.ts` + scenario setup.sh |
| Privacy / security · route params | nav generating + report 携带 13 个 handoff params（7 个稳定 owner IDs + 6 个 display knobs）；不传 raw `answerText` / `questionText` / `hint` / `promptHash` / `modelId raw value` | spec D-13 + frontend-workspace-and-practice plan 002 privacy 红线 | 全 phase | Vitest `reportPrivacy.test.tsx` 反向 grep + E2E.P0.058 子断言 |
| Privacy / security · localStorage / sessionStorage | 不缓存 report 内容到 localStorage；React state only；refresh 时重新拉取 | spec §4 隐私 | 全 phase | Vitest 反向断言 |
| Privacy / security · console.log / telemetry | 不打印 report 内容 / hint / answer / question；mockTransport spy 仅记录 status / latency / 4xx code / scenario name | spec §4 + frontend-workspace-and-practice plan 002 observability 红线 | 全 phase | Vitest `mockTransport.spy.test.ts` |
| Observability · mockTransport spy | 仅记录 status / latency / 4xx code / scenario name；不带 body；不带 report 内容字段 | frontend-workspace-and-practice 同款 spy | 全 phase | Vitest `mockTransport.spy.test.ts` |
| UX · loading state | generating mount 阶段 1 个 viewport 不闪烁；report mount 阶段 skeleton state（数据未 ready） | spec §4 + ui-design 隐式 | 1+2 | Vitest fake timer + `reportLoading.test.tsx` |
| UX · empty state | report 中 highlights / issues / nextActions / questionAssessments 任一为空数组 → 该区块显示 EmptyHint empty state（非 dashboard fail） | spec D-4 ready status 字段填充 | 3 | Vitest `report/EmptyHint.test.tsx` |
| UX · error state | `getFeedbackReport` 5xx / network → inline error + retry；3 次失败显示返回 workspace fallback | spec §3.2 | 5 | Vitest `ReportErrors.test.tsx` |
| UX · i18n zh/en | 全文案通过 typed locale helper；新增 `report.*` + `generating.*` namespace ≥ 60 keys；切换立即重绘 | spec D-7 + 001 D1 typed locale helper | 1-5 | Vitest `reportI18n.test.tsx` + namespace 同步断言 |
| UX · responsive layout (mobile 390×844) | generating 与 report 主屏的 `documentElement.scrollWidth <= 390`；每个 mobile 状态取得非空内存截图 | spec §4 mobile | 5+8+12 | Playwright mobile project |
| UI source structure parity · GeneratingScreen | 页头 + 进度条 + 5 阶段列表 + 实时观察流 + 底部提示 + 「通知我」按钮；testid `generating-{header,progress,phase-${idx},live-stream,evidence-${idx},sla-hint,notify-cta}` | `screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 | 1 | Vitest + testid 命中 + 控件类型断言 |
| UI source structure parity · ReportScreen 顶层三态 | ReportScreen 主组件按 params 分发：`ReportFailureState` / `ReportMissingSessionState` / `ReportDashboard`；testid 顶层 `report-screen` + 子树根 `report-failure-state` / `report-missing-session` / `report-dashboard` | `screen-report.jsx::ReportScreen` lines 1-44 | 2 | Vitest + 三态测试用例 |
| UI source structure parity · ReportDashboard | 返回按钮 + Header（titles + 双 CTA）+ ContextStrip + 4 Summary Cards + Detail Surface + 维度卡片行 + 优先级 + 复练重点 + 题目回顾 + 风险亮点；testid `report-{return,header-title,header-subtitle,replay-cta,next-cta,context-strip,summary-{readiness,dimensions,questions,next},detail-tab-${tab},detail-panel-${tab},dim-row-${idx},top-priority,next-practice-${idx},perq-${idx},issue-${idx},highlight-${idx}}` | `screen-report.jsx::ReportDashboard` lines 80-257 | 2+3 | Vitest + testid 命中 |
| UI source structure parity · ReportFailureState | 失败卡片 + errorCode 文案 + CTA「重新生成」+ CTA「返回 workspace」；testid `report-failure-{title,desc,error-code,retry-cta,back-to-workspace}` | `screen-report.jsx::ReportFailureState` lines 61-77 | 2 | Vitest + testid |
| UI source structure parity · ReportMissingSessionState | 缺会话卡片 + CTA「返回 workspace」；testid `report-missing-session-{title,desc,cta}` | `screen-report.jsx::ReportMissingSessionState` lines 46-59 | 2 | Vitest + testid |
| UI source structure parity · ReportContextStrip | sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条；testid `report-context-{session,job,round,resume,modality,practice-mode,hints}` | `screen-report.jsx::ReportContextStrip` lines 145 | 2 | Vitest + testid |
| UI source structure parity · 5 detail tabs | readiness（拨号盘 `report-readiness-dial`+ 二级详情 `report-readiness-jd-align / -evidence-density / -next-threshold`）/ dimensions（二级网格 `report-dimensions-grid` + 各 `report-dim-card-${idx}`）/ questions（侧栏 `report-questions-list` + 主体 `report-questions-detail-{topic,good,missing,frame,evidence,follow-up}`）/ evidence（`report-evidence-risk-${idx}` + `report-evidence-highlight-${idx}`）/ next（`report-next-path-{a,b}` + `report-next-cta-{a,b}`） | `screen-report.jsx` lines 311-516 | 3 | Vitest + testid 命中（每 tab 独立测试） |
| UI source structure parity · DimRow (维度行) | `DimRow` name + score bar + state tag + confidence；testid `report-dim-row-{name,score,state,confidence}` | `screen-report.jsx::DimRow` lines 565-577 | 3 | Vitest |
| UI source structure parity · 准备度 tier 4 档文案与色调 | not_ready / needs_practice / basically_ready / well_prepared 4 档 zh/en 文案 + 色调；不引入 5 档 readiness numeric score 字面量 | spec D-10 + `ui-design/src/data.jsx` readinessLabel | 2+3 | Vitest + 4 档矩阵测试 |
| UI source structure parity · 维度卡片状态映射 | strong / meets_bar / needs_work 三态文案与色调；不引入 weak / developing / proficient / acceptable rubric 内部 score_levels label（rubric label 不暴露到 UI） | spec D-11 + B1 DimensionStatus | 3 | Vitest + 三态矩阵 |
| UI visual geometry parity · desktop/state | generating 主屏断言 5-phase 关键 DOM 与 root 起始坐标，缺 reportId 断言 error/back CTA；report dashboard 断言 header/context/summary/detail + TopBar，缺 sessionId 与 failed 断言各自 CTA；七个状态均取得非空内存截图 | n/a | 5+8+12 | Playwright `tests/pixel-parity/{generating,report}.spec.ts` |
| UI visual geometry parity · mobile | 390×844 generating/report 主屏断言 document width 不超过 390px，并取得非空内存截图 | n/a | 5+8+12 | Playwright mobile project |
| UI visual geometry parity · clean-checkout gate | generating/report 既有 desktop/mobile 状态的 DOM 锚点、可见性、bounding box、viewport overflow 与逐状态非空内存截图 | n/a | 5+8 | Playwright + Phase 8 owner preflight |
| UI stale-contract negative · reportLayout | `reportLayout='timeline'` / `reportLayout='document'` 等字面量在 report / generating 新代码中 0 命中（不计 negative tests / docs） | spec D-12 + product-scope D-7 | 全 phase | Vitest + scenario verify negative grep |
| UI stale-contract negative · readiness 5 档 | readiness numeric (`readinessScore`) / 5 档（not_ready / needs_practice / basically_ready / well_prepared / fully_prepared 之类的 5 档值）/ `readiness_score` 字段在 report / generating 新代码中 0 命中 | spec D-10 + D-12 | 全 phase | grep negative |
| UI stale-contract negative · 独立错题 / Drill / Growth | `mistakes` route / `mistake_queue` testid / `drill_builder` / `growth_center` / 报告时间线 / 多形态 report / 独立 `report` 一级导航 entry / `practiceModeCard` 在 report / generating 模块 0 命中 | spec D-12 + product-scope D-6 | 全 phase | grep negative |
| UI stale-contract negative · 不直接 import prototype | `frontend/src/app/screens/{report,generating}/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getReportSampleDimensions` 等 prototype helper | spec §4 | 全 phase | Vitest + tsc grep |
| UI stale-contract negative · 不调面试中 operation | report / generating 模块不调 `getPracticeSession` / `appendSessionEvent` / `completePracticeSession` / voice operation；Header CTA 仅允许经共享 `startPracticeFromParams` 调用 generated `createPracticePlan` / `startPracticeSession` | spec §2.2 + frontend-workspace-and-practice D-9 | 全 phase | Vitest spy + owner preflight |
| UI stale-contract negative · 不调范围外 operation | report / generating 模块不调 voice / company intel operation；不渲染 voice surface / company intel screen 任何 DOM | spec §2.2 | 全 phase | Vitest spy + tsc + grep |
| Regression / out-of-scope-negative · 工作区 + 后端契约 | `E2E.P0.044-047`（frontend-workspace-and-practice 002 practice 文本闭环）+ `E2E.P0.052-055`（backend-review 001 report 生成）全部作为真实 regression gate 重跑；fixture-backed PASS 只能证明前端 mock 合同，不替代真实闭环 | n/a | 5 | scenario rerun + `cd backend && go test ./cmd/api -run 'TestE2EP0052\|TestE2EP0053\|TestE2EP0054\|TestE2EP0055' -count=1` |
| Regression / out-of-scope-negative · 不直接调用 LLM | report / generating 模块不出现 AI provider key / provider registry / prompt registry / AIClient / LLM endpoint / bypass generated client 的 ad hoc fetch | n/a | 全 phase | Vitest + grep negative |
| BDD 主路径 + 关键分支 + 失败恢复 + stale-contract negative | 见 [bdd-plan.md](./bdd-plan.md) 4 场景矩阵 | n/a | 1-5 | E2E.P0.056/057/058/059 |

### 高风险类别 N/A 说明

- **隐私 / 安全 · audio buffer**：本 plan 不实现 voice surface / STT / TTS；audio buffer 不进入 report 屏。N/A 原因记录在此。
- **Privacy · LLM prompt raw text**：B2 在 backend-review 服务端 redact prompt 与 response；前端不直接调用 LLM；`provenance` 字段（`promptVersion / modelId`）只是版本/标识，不含 prompt body；前端只渲染版本号到 AI 透明度区域（report dashboard ContextStrip / questions tab evidence），因此 prompt-response 明文不在前端泄漏面。N/A 原因记录在此。
- **Out-of-scope · 报告时间线 / 多形态 report / 报告列表 UI**：D-12 范围外边界；plan 001 不实现；归 plan 002 future（且需先修订 product-scope D-7）。

## 3.6 Frontend / Backend Operation Matrix

本 plan 最初走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated client + fixture-backed transport 完成 P0 UI/BDD。2026-05-23 复查时 backend-review/001 真实 handler 已落地；P0.056-P0.059 trigger 必须先跑 `frontendOwners.realApiMode.test.ts`，fixture-backed PASS 不得单独代表真实 backend 闭环。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `default` / `report-generating` / `prototype-baseline` / `report-failed` | `useReportGenerationPoll` GeneratingScreen + ReportScreen 单次拉取；OpenAPI path param `reportId` only | backend-review real handler；read-only；user-scoped；返回 status 决定 queued/generating vs ready vs failed | `feedback_reports` + `question_assessments` read | none in frontend | frontend `E2E.P0.056 / E2E.P0.058` + real-mode gate；backend `E2E.P0.053 / E2E.P0.055` |
| `getTargetJob` | 当前已有 `openapi/fixtures/TargetJobs/getTargetJob.json` `default` | `useReportContextData` / `ReportContextStrip` 只读 target job title/companyName；失败时显示 targetJobId fallback，不阻塞报告正文 | backend-targetjob 既有 handler；read-only；user-scoped | `target_jobs` read | none | `E2E.P0.056` ContextStrip 子断言 + Vitest `ReportContextStrip.test.tsx` |
| `getResume` | 当前已有 `openapi/fixtures/Resumes/getResume.json` `default` | `useReportContextData` / `ReportContextStrip` 只读 resume displayName；失败时显示 resumeId fallback；不得读取 raw resume text | backend-resume 既有 handler；read-only；user-scoped | `resumes` read | none | `E2E.P0.056` ContextStrip 子断言 + privacy negative |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json` (`default`, `empty`) | 不消费 UI（dashboard-only D-7）；real-mode gate 覆盖 generated client routing | backend-review real handler；read-only；user-scoped；cursor 分页 | `feedback_reports` cursor read | none | real-mode gate + 负向断言（在 generating / report 模块零调用） |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` | `useReplayCtaHandlers` → `startPracticeFromParams` 创建 retry_current_round / next_round 派生 plan；携带 `Idempotency-Key` | backend-practice real handler | `practice_plans` write | backend-only first question prep | E2E.P0.057 + `ReplayCta.test.tsx` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | `useReplayCtaHandlers` → `startPracticeFromParams` 启动 fresh session；携带 `Idempotency-Key`，成功后直接 nav practice | backend-practice real handler | `practice_sessions` + first turn write | backend-only first question prep | E2E.P0.057 + `ReplayCta.test.tsx` |
| `completePracticeSession` | N/A（本 plan 不消费） | 由 frontend-workspace-and-practice plan 002 消费；本 plan 不调用 | backend-practice/002 已落地 | — | — | 负向断言 |
| `appendSessionEvent` / `getPracticeSession` / voice operations | N/A | 由 frontend-workspace-and-practice plan 002 消费；本 plan 不调用 | — | — | — | 负向断言 |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` | 本 plan **不消费**；voice surface deferred to practice-voice owner | practice-voice/backend-practice real handler | voice session events | STT/LLM/TTS backend-only | report/generating 负向断言 + real-mode gate coverage |
| `getCompanyIntel` | N/A | 本 plan **不消费**；company-intel owner 承接 | external owner | — | — | 负向断言 |

## 3.7 InterviewContext × PracticeDisplayContext View-Model Mapping

正式前端不得从 `ui-design/src/data.jsx` 或未声明 fixture 字段补齐 `InterviewContext` 之外的数据；ContextStrip 所需的人类可读 job/resume label 只能来自 §3.6 operation matrix 中声明的 `getTargetJob` / `getResume`，并在失败时回退 ID 显示。本 plan 在 frontend-workspace-and-practice plan 001 + 002 已落地的 `InterviewContext` reducer 基础上**仅 read**，不新增 reducer action。具体 mapping：

| 字段 | Source | Rule |
|------|--------|------|
| `reportId` | route param 或 InterviewContext.reportId（由 frontend-workspace-and-practice plan 002 buildPracticeHandoffParams 写入） | 必填；缺失 → 渲染 ErrorState「报告 ID 缺失」 |
| `sessionId` | route param 或 InterviewContext.sessionId | 必填；缺失 → 渲染 ReportMissingSessionState |
| `planId / targetJobId / jdId / resumeId / roundId` | route param 或 InterviewContext | 本 plan 只 read，用于 ContextStrip 显示 + 复练 CTA payload 构造；不新增 handoff key |
| `targetJob.title / targetJob.companyName` | generated `getTargetJob(targetJobId)`；失败时回退 `targetJobId` | ContextStrip 人类可读 target job label；不得从 `ui-design/src/data.jsx` 或硬编码 fixture 复制 |
| `resume.displayName` | generated `getResume(resumeId)`；失败时回退 `resumeId` | ContextStrip 人类可读 resume label；不得读取 raw resume text / parsed snapshot / raw resume body |
| `mode / modality` | route param（默认 `text/text`） | 本 plan UI 不切换；用于 ContextStrip 显示「文本面试 / 电话模式」 |
| `practiceMode` | route param（默认 `strict`） | 二值；用于 ContextStrip 显示「带提示 / 严格模拟」+ 复练 CTA payload `practiceMode:lastPracticeMode` 传递 |
| `practiceGoal` | route param（默认 `baseline`） | 用于 ContextStrip 显示「首轮 / 复练当前轮 / 进入下一轮 / 真实复盘」；复练 CTA 派生新 goal |
| `hintUsed / hintCount` | route param（默认 `'false' / '0'`） | 用于 ContextStrip 显示「已使用 N 次提示」 |
| `reportStatus`（URL search param） | `report?reportStatus=failed`（仅在失败路径由 GeneratingScreen 设置） | 进入 report 屏决定渲染 `ReportFailureState` vs `ReportDashboard` |
| `errorCode`（URL search param） | `report?errorCode=AI_PROVIDER_TIMEOUT`（仅在失败路径） | ReportFailureState 文案映射 |
| `feedbackReport` (response) | `getFeedbackReport(reportId)` | 临时状态；ReportDashboard 渲染整个 dashboard 内容；不持久化到 localStorage |
| `retryFocusTurnIds` | `feedbackReport.retryFocusTurnIds` | 复练 CTA 路径 A payload 字段 `replayItems` |
| `nextRoundId / roundName`（路径 B） | InterviewContext.roundId 推进 / 或 backend-targetjob future 提供下一轮元数据 | plan 001 默认从 InterviewContext 取 `nextRoundId = roundId + 1` 或本地推断，无 backend 单独调用；如未来需要从 backend 拿真实下一轮 metadata，先回 backend-targetjob owner 修订（**Open Question**：当前 plan 不规约 `roundId` 字符串格式与递增算法，由 design 决策；见 §6 风险与应对 row 8） |
| `roundName`（generating + report ContextStrip 显示） | 由 ContextStrip 在本 spec owner route 内通过 `InterviewContext.roundId` + i18n typed locale helper 本地推导显示 | **不**包含在 `buildPracticeHandoffParams` 输出的 13 字段内；ContextStrip 在本 plan 内基于 `roundId` 推导 zh/en label（如 "第 1 轮 / Round 1"）；与路径 B `nextRoundId` 派生 `nextRound` 显示同款 |

新增 InterviewContext reducer action（在 frontend-workspace-and-practice 已有 reducer 基础上）：**无**。spec D-8 已锁定本 plan 不新增 action；reportId / sessionId 来源已在 buildPracticeHandoffParams（plan 002）写入；本 plan 通过 route params + URL search params 读取。

> 备注：与 frontend-workspace-and-practice plan 002 `INCREMENT_HINT_COUNT` 同款扩展机制不在本 plan 范围。

## 4 实施步骤

### Phase 0: 跨 owner 前置 preflight

**目标**：阻塞性 preflight assert — 本 plan 用于确认 `backend-review/001` 交付的 4 个 cross-owner contract 已落实；本节测试作为 contract guard，任何失败都表示 OpenAPI / fixture / generated client drift，不得用 frontend-only fixture workaround 掩盖 drift。

#### 0.1 `FeedbackReport.errorCode` 字段存在断言

新增 `frontend/src/app/screens/report/__tests__/preflight.test.ts`：

- 读取 `openapi/openapi.yaml` 中 `FeedbackReport` schema；断言 `errorCode` 字段存在且类型 `oneOf: [ApiErrorCode, null]`；缺失则测试 fail 并指向 `backend-review/001` 的 schema/error-code contract。
- 兼容 generated TS client：断言 `frontend/src/api/generated/` 中 `FeedbackReport` interface 含可选 `errorCode` 属性。

#### 0.2 `report-failed` fixture variant 存在断言

在 `preflight.test.ts` 中追加：

- 读取 `openapi/fixtures/Reports/getFeedbackReport.json`，断言 `scenarios.report-failed.response.body.status === 'failed'` + `scenarios.report-failed.response.body.errorCode` 非 null（建议值为 `AI_PROVIDER_TIMEOUT`）；缺失则 fail 并指向 report-failed fixture contract。
- 同时断言 `scenarios.report-generating` 与 `scenarios.default` 已经存在（这两个已在仓库内）。

#### 0.3 `listTargetJobReports.empty` fixture variant 存在断言

在 `preflight.test.ts` 中追加：

- 读取 `openapi/fixtures/Reports/listTargetJobReports.json`，断言 `scenarios.empty.response.body.items === []`、`scenarios.empty.response.body.pageInfo.hasMore === false`、`scenarios.empty.response.body.pageInfo.nextCursor === null`；缺失则 fail 并指向 empty fixture contract。
- 虽然本 plan 不消费 `listTargetJobReports`，empty fixture 的存在性仍是 backend-review/001 contract 健康指示器；缺失表示 fixture/generated-client drift。

#### 0.4 `REPORT_NOT_FOUND` 错误码存在断言

在 `preflight.test.ts` 中追加：

- 读取 `shared/conventions.yaml#errors`，断言 `REPORT_NOT_FOUND` 行存在 + `httpStatus: 404` + `retryable: false`；缺失则 fail 并指向 shared error contract。
- 同时断言 `frontend/src/api/generated/` 中存在 generated TS 等价常量（如 `ApiErrorCode.REPORT_NOT_FOUND` 或 `errors.REPORT_NOT_FOUND`）；缺失则 fail 并指向 generated TS error constant drift。

#### 0.5 Phase 0 收口 gate

- `pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts` 全绿（即 4 项断言全部通过）
- 如有任一断言 fail：本 plan 状态保持 `active` 但 Phase 1 不启动；通过 [bug-report](../../../../bugs/PATTERNS.md) 或 retrospective 联动通知 backend-review/001 owner
- 通过后，删除任何临时 fixture-only workaround 文件（如有）并进入 Phase 1

### Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

#### 1.1 新增 `frontend/src/app/screens/generating/GeneratingScreen.tsx`

按 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399) 源级复刻渲染：页头（标题 + 副文案 "Reading every turn. Evidence first."）+ 进度条（百分比 + phase indicator）+ 5 阶段列表（每个 done/active/pending 状态圆圈 + 标签）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA "<12s target" + 「通知我」UI-only 按钮）。本 phase 接入轮询 hook；reportId 缺失时不发请求、直接渲染 ErrorState。

#### 1.2 新增 `frontend/src/app/screens/generating/hooks/useReportGenerationPoll.ts`

通过 generated client 轮询 `getFeedbackReport(reportId)`；React state 跟踪 `idle / polling / ready / failed / timeout / error / paused` 七态；指数退避（初始 1.5s × 1.5 上限 8s，max attempts 30）；visibility / focus 事件暂停-恢复轮询；网络 5xx / network error retry 复用当前 attempt count；status='ready' → 调 `onReady(report)` callback；status='failed' → 调 `onFailed(errorCode)` callback；max attempts 达到 → state='timeout' 触发 ErrorState；request init 不含 `Idempotency-Key` header。

#### 1.3 新增 `frontend/src/app/screens/generating/components/`

包含：`HeaderHero.tsx`（标题 + 副文案）/ `ProgressBar.tsx`（百分比 + phase indicator）/ `PhaseList.tsx`（5 阶段列表 + 状态圆圈）/ `LiveEvidenceStream.tsx`（fade-in 流）/ `SlaHint.tsx`（底部提示 + 通知我按钮）/ `GeneratingErrorState.tsx`（reportId 缺失或 timeout 兜底）。每个组件从 `ui-design/src/screens-p0-complete.jsx` 同名片段复刻 DOM。

#### 1.4 路由壳替换

在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `generating` → `<GeneratingScreen route={route} />`（替换 route fallback shell）；保持 `generating` 在 `NO_CHROME_ROUTES` 中隐藏 TopBar；`report` 仍由 route fallback shell 承接，待 Phase 2 替换。

#### 1.5 i18n locale 扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `generating.*` 命名空间（≥ 20 keys：header.title / header.subtitle / phase.1 / phase.2 / phase.3 / phase.4 / phase.5 / progress.phaseN / evidence.streamLabel / sla.target / sla.notifyCta / errors.missingReportId / errors.timeout / errors.retry / errors.backToWorkspace 等）；`messages.ts` 类型聚合补齐。

#### 1.6 Vitest 红灯 → 绿灯

新增 `generating/__tests__/GeneratingScreen.test.tsx`：测 i18n zh/en 切换重绘、≥ 10 个 `generating-*` testid 存在、reportId 缺失 → 不发请求 + 渲染 ErrorState、timeout 状态渲染 retry CTA、负向断言不出现 `mistakesQueue` / `report-timeline` testid。

新增 `generating/__tests__/useReportGenerationPoll.test.ts`：7 态、指数退避节奏（fake timer）、max attempts、visibility / focus 暂停-恢复、status='ready' callback、status='failed' callback、unmount 取消、不含 `Idempotency-Key`。

#### 1.7 BDD-Gate Phase 1

- BDD-Gate: 验证 `E2E.P0.056` GeneratingScreen 部分通过（mount → 进度动画 → 轮询 → status='ready' nav report）

### Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

#### 2.1 新增 `frontend/src/app/screens/report/ReportScreen.tsx`

按 `ui-design/src/screen-report.jsx::ReportScreen` (lines 1-44) 源级复刻：顶层根据 `params.reportStatus === 'failed'` 或 `!sessionId` 分别渲染 `ReportFailureState` / `ReportMissingSessionState` / `ReportDashboard`。本 phase 实现 `ReportDashboard` 静态壳（含 Header + ContextStrip + 4 Summary Cards + Detail Surface 骨架态）；5 个 detail tab 内容由 Phase 3 完成。

#### 2.2 新增 `frontend/src/app/screens/report/components/`

包含：`ReportHeader.tsx`（标题 + 副标题 + 双 CTA `复练当前轮` + `进入下一轮`）/ `ReportContextStrip.tsx`（sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条）/ `SummaryCards.tsx`（4 张 ReportStatButton：准备度 / 维度 / 题目 / 下一步）/ `ReportFailureState.tsx`（lines 61-77 源级复刻；CTA「重新生成」+「返回 workspace」）/ `ReportMissingSessionState.tsx`（lines 46-59 源级复刻；CTA「返回 workspace」）/ `ReportErrorBoundary.tsx`（getFeedbackReport 5xx 兜底）。

#### 2.3 新增 report 数据 hooks

- `useFeedbackReport.ts`：通过 generated client 单次拉 `getFeedbackReport(reportId)`；React state 跟踪 `loading / data / error / notFound`；404 → state='notFound' 渲染 ReportFailureState（cross-user 隔离）；5xx → state='error' + retry 按钮；request init 不含 `Idempotency-Key`。
- `useReportContextData.ts`：通过 generated `getTargetJob(targetJobId)` + `getResume(resumeId)` 只读 ContextStrip label；成功时显示 target title/companyName + resume displayName；任一失败时回退 targetJobId / resumeId，不阻塞 ReportDashboard；不得读取 raw resume/JD/body 字段。

#### 2.4 路由壳替换

在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `report` → `<ReportScreen route={route} />`（替换 route fallback shell）；保持 `report` 不在 `NO_CHROME_ROUTES` 中，默认 App chrome / TopBar 可见，同时不把 `report` 加入一级导航。

#### 2.5 i18n locale 扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `report.*` 命名空间（≥ 40 keys：header.title / header.subtitle / cta.replay / cta.nextRound / context.session / context.job / context.round / context.resume / context.modality / context.practiceMode.assisted / context.practiceMode.strict / context.hints / summary.readiness / summary.dimensions / summary.questions / summary.next / readiness.tier.notReady / readiness.tier.needsPractice / readiness.tier.basicallyReady / readiness.tier.wellPrepared / failureState.title / failureState.desc / failureState.errorCode.* / failureState.retry / failureState.backToWorkspace / missingSession.title / missingSession.desc / missingSession.cta / loading / errors.network / errors.retry 等）；`messages.ts` 类型聚合补齐。

#### 2.6 Vitest 红灯 → 绿灯

新增 `report/__tests__/ReportScreen.test.tsx`：测三态切换（reportStatus='failed' / 缺 sessionId / 正常）、loading / data / error / notFound 四态、≥ 10 个 `report-*` testid 存在。

新增 `report/__tests__/ReportFailureState.test.tsx` / `ReportMissingSessionState.test.tsx`：分别测各自渲染 + CTA 行为。

新增 `report/__tests__/useFeedbackReport.test.ts`：测 4 态 + 404 / 5xx 处理 + 不含 `Idempotency-Key`。

新增 `report/__tests__/useReportContextData.test.ts`：测 `getTargetJob` / `getResume` 成功渲染人类可读 label、单 operation 失败回退对应 ID、不读取 raw resume/JD/body 字段、request init 不含写操作 header。

#### 2.7 BDD-Gate Phase 2

无独立 BDD（Phase 2 静态壳由 Phase 4 后串联 `E2E.P0.056` + `E2E.P0.058` 覆盖；本 Phase 仅 Vitest）。

### Phase 3: 5 detail tab 内容源级复刻

#### 3.1 新增 `frontend/src/app/screens/report/components/DetailSurface.tsx`

按 `screen-report.jsx::ReportDetailSurface` (lines 162-174 + 311-516) 源级复刻：5 个 tab 触发按钮 + panel 切换；testid `report-detail-tab-{readiness,dimensions,questions,evidence,next}` + `report-detail-panel-{key}`；ARIA tablist + tab + tabpanel 角色。

#### 3.2 实现 5 个 tab 内容组件

每个 tab 独立组件文件，源级复刻 `screen-report.jsx` 对应 lines：

- `tabs/ReadinessTab.tsx` (lines 335-357)：拨号盘（4 档色环）+ 二级详情（JD 对齐、证据密度、下一档门槛）。
- `tabs/DimensionsTab.tsx` (lines 360-382)：二级维度卡片网格（使用 `DimRow` primitive，状态三态映射）。
- `tabs/QuestionsTab.tsx` (lines 385-442)：题目列表侧栏（5 题）+ 当前题分析（有效点、缺口、建议框架、证据片段、下次追问）；选中 turn 切换右侧 detail。
- `tabs/EvidenceTab.tsx` (lines 445-467)：风险证据卡片列表 + 可复用亮点证据卡片列表。
- `tabs/NextTab.tsx` (lines 470-514)：路径 A vs 路径 B 对比展示 + 各自行动 CTA（在 Phase 4 wire 复练 CTA 行为）。

#### 3.3 新增维度卡片行 + 优先级 + 复练重点 + 题目回顾 + 风险亮点

在 `ReportDashboard.tsx` 主体补齐 lines 176-253 的源级复刻：维度卡片行（horizontal scroll on mobile）+ 优先级（topPriority 单行）+ 复练重点（nextPractice 3 条 list）+ 题目回顾概览（5 题 quick state 卡片）+ 风险 issues + 亮点 highlights。

#### 3.4 Vitest 红灯 → 绿灯

新增 `report/__tests__/DetailSurface.test.tsx`：测 5 个 tab 切换 + ARIA tablist + 默认 `questions` 激活；readiness tab 通过显式点击覆盖。

新增 `report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx`：每个测对应 testid + 数据驱动 + 边界（空 dimensions / 空 questions / 空 issues / 空 highlights 各自 EmptyHint）+ 4 档 readiness 文案矩阵 + 三档维度状态矩阵 + 不出现 5 档 readiness / rubric score_levels label。

#### 3.5 BDD-Gate Phase 3

- BDD-Gate: 验证 `E2E.P0.056` ReportDashboard 渲染部分通过（含 5 detail tab 切换 + 维度卡片 + 题目回顾 + 风险亮点）

### Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

#### 4.0 注册 `replay_practice` PendingAction type

在 `frontend/src/app/auth/pendingAction.ts` 中（当前 `PendingAction.type: string` 仅在测试中使用 `start_practice` 一个值）扩展：

- 把 `replay_practice` 加入允许的 type allowlist（如果当前实现使用 string union 或运行时 validator）。
- `encodePendingAction` / `decodePendingActionRoute` 必须支持 `replay_practice` round-trip（params 保留原 report route 的 sessionId / reportId / targetJobId / resumeId / roundId / display knobs）。
- 路由恢复（`AppPendingAction` 或同款）把 `replay_practice` 返回 `report`；不在恢复 route mount 时执行 session 创建副作用，用户鉴权后重试 Header CTA。

新增 `frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts`（或扩展现有 `pendingAction.test.ts`）：

- `TestPendingActionEncodeDecodeReplayPractice` 测 encode → decode round-trip 字段对等
- `TestPendingActionReplayPracticeTypeAllowed` 测 type allowlist / discriminated union 包含 `replay_practice`
- 负向断言：encode 后 URL params / localStorage 不含 raw text（与 spec D-13 一致）

#### 4.1 实现复练 CTA `goReplay()` 路径 A

在唯一 Header `report-replay-cta` 上绑定 `goReplay()`：组装 `{ sourceSessionId, sourceReportId, replayItems, evidenceGaps, planId, targetJobId, jdId, resumeId, roundId, mode:'text', modality:'text', practiceMode, practiceGoal:'retry_current_round' }`。已登录时调用共享 `startPracticeFromParams`，通过 generated `createPracticePlan` / `startPracticeSession` 创建 fresh session 后直接 nav practice；未登录通过 `replay_practice` pendingAction 返回 report 重试。

#### 4.2 实现复练 CTA `goNextRound()` 路径 B

唯一 Header `report-next-cta` 组装 `{ sourceSessionId, sourceReportId, nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeId, mode:'text', modality:'text', practiceMode, practiceGoal:'next_round' }`，并走同一 direct-start helper；nextRoundId 由 canonical interview ladder 推断。若未来由 backend 提供真实 round metadata，先回 backend-targetjob owner 修订。

#### 4.3 完整 ReportFailureState handoff

`ReportFailureState.tsx` CTA「重新生成」点击 → `nav("generating", { sessionId, reportId, ...passThroughContext })` 重新进入 generating 屏触发轮询（不直接调用 backend；轮询 hook 自然命中既有 failed report）；CTA「返回 workspace」点击 → `nav("workspace", { targetJobId, jdId, planId, resumeId })`。

#### 4.4 完整 GeneratingScreen handoff

`useReportGenerationPoll` 的 `onReady(report)` callback → `nav("report", { sessionId, reportId, ...passThrough })`；`onFailed(errorCode)` callback → `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode, ...passThrough })`；timeout state → 不自动 nav；用户点 retry 重启轮询；nav 调用必须防抖（handoffNavigatedRef）。

#### 4.5 Vitest 红灯 → 绿灯

新增 `report/__tests__/ReplayCta.test.tsx`：测路径 A 已登录创建/启动 fresh session 并直接进入 practice，未登录进入 auth_login 且 pendingAction 回 report；payload 字段完整；负向断言 raw text 不在 payload。

在 `report/__tests__/ReplayCta.test.tsx` 覆盖路径 B 同上；断言 nextRoundId 推断逻辑与 fresh session start。

新增 `report/__tests__/ReportFailureHandoff.test.tsx`：测「重新生成」nav generating + 「返回 workspace」nav workspace。

扩展 `generating/__tests__/GeneratingScreen.test.tsx`：测 ready / failed / timeout 三态分别 nav + 防抖（多次 ready callback 只 nav 一次）。

#### 4.6 BDD-Gate Phase 4

- BDD-Gate: 验证 `E2E.P0.057` 通过（复练 CTA 路径 A + 路径 B 通过 generated client 创建/启动 fresh session 并直接进入 practice；未登录鉴权后回 report）
- BDD-Gate: 验证 `E2E.P0.058` 通过（GeneratingScreen 轮询命中 `status='failed'` → nav failed report + ReportFailureState + ReportMissingSessionState + 跨用户 + 隐私 route params）
- BDD-Gate: 验证 `E2E.P0.056` 整链完整通过（含 GeneratingScreen mount → 进度动画 → 轮询 ready → nav report → ReportDashboard 渲染 → 5 detail tab 切换 → CTA wire 完整）；Phase 1 + Phase 3 仅做局部断言，Phase 4 复练 CTA wire 完成后才算完整通过

### Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + stale-contract negative

#### 5.1 完整状态机集成回归

`pnpm vitest run`（全 frontend 测试）+ `pnpm typecheck` 全绿；扩展现有 `App.test.tsx` 添加 `generating-screen` 与 `report-dashboard` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` / `report` route alias 处理；扩展 `pendingActionReplayPractice.test.ts` 添加 `replay_practice` pendingAction 回 report 的 round-trip；扩展 auth pending-action 场景覆盖 resume path。

#### 5.2 Playwright pixel parity 加挂

新增 `frontend/tests/pixel-parity/generating.spec.ts` + `frontend/tests/pixel-parity/report.spec.ts`：desktop 1440×900 + mobile 390×844 两 viewport；generating 覆盖 desktop 主屏、缺 reportId 错误态和 mobile overflow，report 覆盖 desktop dashboard、缺 sessionId、failed state 和 mobile overflow；每个状态断言关键 DOM/可见性或 viewport geometry，并执行非空内存截图。

#### 5.3 scenario 加挂

在 `test/scenarios/e2e/` 派生 4 个 scenario 目录 `p0-056-generating-to-report-happy-path/` / `p0-057-replay-cta-paths-a-and-b/` / `p0-058-report-failure-and-missing-session/` / `p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/`，每个含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x 可执行）；trigger 跑对应 Vitest 套件；verify 反查 testid / nav payload / 负向 grep。

#### 5.4 stale-contract negative grep

scoped grep 在 `frontend/src/app/screens/{report,generating}/` 范围：
- `reportLayout` / `report_layout`
- 5 档 readiness（如 `fully_prepared` 字面量）/ `readinessScore` numeric 字段
- `mistakes_queue` / `mistakesQueue` / `mistake-queue` testid
- `drill_builder` / `drillBuilder` / `drill-builder` testid
- `growth_center` / `growthCenter` / `growth-center` testid
- `report_timeline` / `reportTimeline`
- `report_form` / `reportForm`
- 独立 `mistakes` route entry
- prototype data import：`ui-design/src/data.jsx` / `window.EI_DATA` / `getReportSampleDimensions` 等
- `createPracticeVoiceTurn` / `getCompanyIntel` 调用
- `listTargetJobReports` 调用（dashboard-only D-7；plan 001 不消费列表）

负向 grep 在 `scripts/lint/frontend_report_dashboard_out_of_scope.py`（新增）+ `frontend/src/app/screens/{report,generating}/__tests__/outOfScopeNegative.test.ts` 实现。

#### 5.5 i18n 完整性断言

新增 `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts`：断言 `report.*` 与 `generating.*` 命名空间 zh / en 同步无缺漏（key 集合相等）；新增 key ≥ 60；切换 locale 时所有 testid 文案重绘。

#### 5.6 跨 owner regression

在 Phase 5 收口阶段重跑：
- frontend-workspace-and-practice/002 BDD `E2E.P0.044-047`（保证未被破坏）
- backend-review/001 BDD `E2E.P0.052-055`（如已 implement；如未 implement 则跳过 backend regression，仅跑 fixture-backed scenario）
- backend-practice/002 BDD `E2E.P0.038-043`（必要时通过 cmd/api 重跑）

#### 5.7 BDD-Gate Phase 5

- BDD-Gate: 验证 `E2E.P0.059` 通过（Playwright pixel parity + i18n + stale-contract negative）

#### 5.8 文档收口

更新 `docs/spec/frontend-report-dashboard/plans/INDEX.md`：001 状态保持 `active`（plan-review / sync-doc-index 推进到 completed 由后续动作完成）。

新增 `frontend/src/app/screens/report/README.md` + `frontend/src/app/screens/generating/README.md`：简明 handoff 段落，记录 001 新增 component / hook / nav 边界 / handoff 给 backend-review 与 frontend-workspace-and-practice 的边界。

## 5 验收标准

- Phase 1 ~ Phase 5 checklist 全部勾选
- 关联 BDD 场景 `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059` 均由对应 Vitest + Playwright + scenario 执行通过
- `pnpm --filter @easyinterview/frontend test` / `pnpm --filter @easyinterview/frontend typecheck` / `pnpm --filter @easyinterview/frontend test:pixel-parity` / `pnpm --filter @easyinterview/frontend build` 全绿
- `make codegen-check` 通过（不修改 OpenAPI / generated client，但 build 时反查 drift）
- `make validate-fixtures` 通过（覆盖 backend-review/001 已交付 fixture variants）
- `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all` 通过
- 001 范围内代码与文档中无 §3.5 / §D-12 列出的 stale-contract 术语 / reportLayout / 5 档 readiness / 独立 mistakes route / drill_builder / growth_center / 报告时间线 / 多形态 report 字面量出现
- frontend-workspace-and-practice/002 BDD regression（`E2E.P0.044-047`）通过；backend-review/001 BDD regression（`E2E.P0.052-055`）如已 implement 则通过

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| backend-review/001 已交付 fixture variants（`report-failed` / `empty`）与 OpenAPI/generated client 漂移，导致 preflight contract guard 失败 | 停止 report owner 变更，先回 `openapi/openapi.yaml`、`openapi/fixtures/Reports/*.json`、generated TS 与 `shared/conventions.yaml` 同步；不得用 frontend-only fixture workaround 掩盖 drift |
| backend-review/001 真实 handler 与 fixture-backed frontend BDD 表现不一致，导致 real-backend gate 或跨 owner regression 失败 | 以真实 handler 结果为准反修 fixture / viewmodel mapping / scenario expectation；保留 fixture-backed BDD 作为 UI variant 覆盖，但不能把 mock green 宣称为真实 backend 闭环 |
| GeneratingScreen 5 阶段动画时长（700-1200ms）与真实 backend 生成时间（P95 < 12s）节奏不匹配，导致用户感知卡顿 | Phase 1 动画仅作为视觉反馈，不阻塞 nav；status='ready' 在任何阶段都立刻 nav；最坏情况下 5 阶段 ~5s + 轮询命中 ~3s = 8s 内完成；超过 12s 进入 timeout state |
| `useReportGenerationPoll` 在 visibility 切换时 race condition 导致重复 nav | Phase 1 nav 防抖 ref + 单次 fire-and-forget；onReady / onFailed callback 只触发一次；test fake clock 验证 |
| Mobile width smoke 被误读为 detail-tab 组件级响应式证明 | P0.059 只声明 document width 与显式 DOM；detail tab 结构、切换和 keyboard a11y 由 Phase 3 Vitest 持续覆盖；扩展浏览器 AC 前先增加对应断言 |
| 复练 CTA 路径 A 的 `replayItems:retryFocusTurnIds` 在 `getFeedbackReport` 响应中是 turn UUID 列表，但 backend-practice 创建 `goal='retry_current_round'` plan 需要 `source_report_id` 而非 turn id 列表 | 复练 CTA payload 仅作为 frontend nav route params；frontend-workspace-and-practice plan 002 或 future plan 在执行 `createPracticePlan` 时把 `sourceReportId` 设为当前 reportId；`replayItems` 作为本地 hint 用于 UI 显示，不传给 backend；本 plan handoff doc 标记此约束 |
| 路径 B `nextRoundId` 推断逻辑（默认 roundId + 1）在 backend 没有真实 round metadata 时不准确 | plan 001 默认走本地推断；如产品确认需要真实下一轮 metadata，先回 backend-targetjob owner 修订；不在本 plan 引入新 backend 调用 |
| i18n `errors.errorCode.*` 文案需要覆盖 backend-review D-8 所有 B1 `AI_*` enum；新增 B1 enum 时本 plan 文案漏更新 | Phase 2 i18n 测试断言 `errors.errorCode` 覆盖 B1 `AI_*` 当前全部 enum（用 generated B1 常量做 source of truth）；新增 enum 时 lint fail |
| Playwright 注释、owner 文档与真实浏览器断言容易 drift | Phase 8/12 preflight 同时读取 active spec、六份 plan artifact、两份 Playwright 源码与 P0.059 scenario claims；只允许七个已执行状态的显式证据，并要求每个状态执行非空内存截图 |
| 复练 CTA 在未登录场景的 useRequestAuth handoff 丢失原报告上下文 | Phase 4 测试覆盖 `replay_practice` encode/decode 后回到同一 report params；鉴权恢复不自动创建 session，用户重试 CTA 后再走 direct-start helper |
| **Polling timeout vs backend lease/retry 节奏不一致**（Open Question） | `useReportGenerationPoll` max attempts 30 ≈ 3.67 min；backend-review/001 D-13 lease_timeout 5 min + retry `min(2^attempt_count * 30s, 30min)` 可达 30 min × 5 attempts。当 backend 处于 retry 等待时 frontend 会先 timeout。当前默认：用户 retry 重启轮询（attempts 归零），不自动续轮；如 backend 在 generating → generating 持续 > 3.67 min 用户会反复看到 timeout 卡片。**待 design 决策**：是否把 max attempts 提到 ≥ 5 min（对齐 backend lease_timeout）？是否在 timeout 自动 retry 一次再显示卡片？是否新增 backend `?wait=true` long-poll API？plan 001 默认保留 3.67 min + manual retry；如需调整在 design 确认后修订本 risk row |
| **`nextRoundId` 推断算法字符串格式 与 backend round metadata 缺位**（Open Question） | spec D-5 + plan §4.2 推断 `nextRoundId = roundId + 1` 仅对 numeric / `round-${N}` 字符串成立；如果 backend `practice_plans.round_id` 是 uuid 推断无意义。**待 design 决策**：（A）约定 `roundId` 必须是可递增字符串（如 `round-${N}`）+ 推断 `round-${parseInt(N)+1}`；（B）path B CTA 在 backend-targetjob 提供真实 metadata 前 disabled / 不渲染；（C）path B 走临时 fallback：复用当前 roundId（"continue current round" 隐含语义）。plan 001 默认采用选项 A + 在 ContextStrip 显示 zh "第 2 轮" / en "Round 2" 文案；如 backend 实际格式不符合，design 需先回 backend-targetjob owner 修订 |
