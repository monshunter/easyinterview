# 001 — Report Screen and Generating Handoff

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [frontend-report-dashboard spec](../../spec.md) v1.0 §2.1 / §6 C-1~C-15 / §7 锁定的第一个 plan 范围落地，承接 [frontend-workspace-and-practice/002-practice-text-event-loop](../../../frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)（completed, 2026-05-14）已交付的 `completePracticeSession` 真实调用 + 13 字段 generating route params（`buildPracticeHandoffParams` 输出：`{planId, targetJobId, jdId, resumeVersionId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}`）+ `ReportWithJob` 缓存 + `nav("generating", ...)` handoff，闭合 P0 用户路径中"报告生成过渡态 → 证据化报告 dashboard → 复练 / 下一轮 handoff"段：

- `generating` 路由从 `PlaceholderScreen` 切换为正式 `GeneratingScreen`，源级复刻 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399)：页头 + 进度条 + 5 阶段列表（done/active/pending 状态圆圈，每阶段 700-1200ms 动画节奏）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA "<12s target" + 「通知我」UI-only 按钮）。
- 通过 generated client + fixture-backed transport 消费 `getFeedbackReport(reportId)`；指数退避轮询（初始 1.5s × 1.5 上限 8s，max attempts 30）+ visibility 暂停 / 恢复；status='ready' → nav `report?sessionId=&reportId=&...passThrough`；status='failed' → nav `report?reportStatus=failed&errorCode=&...`；max attempts 达到 → ErrorState 「报告生成超时」+ retry / 返回 workspace。
- `report` 路由从 `PlaceholderScreen` 切换为正式 `ReportScreen`，源级复刻 `ui-design/src/screen-report.jsx::ReportScreen` (lines 1-516) 三态：`ReportDashboard`（正常报告）/ `ReportFailureState`（reportStatus='failed'）/ `ReportMissingSessionState`（缺 sessionId）。
- ReportDashboard 完整源级复刻：返回按钮 + Header + `ReportContextStrip`（sessionId / targetJob / round / resume / modality / practiceMode / hints）+ 4 个 Summary Cards（准备度 / 维度 / 题目 / 下一步）+ `ReportDetailSurface` 5 个 tab（readiness / dimensions / questions / evidence / next）+ 维度卡片行 + 优先级 + 复练重点 + 题目回顾概览（5 题）+ 风险 & 亮点 + 复练 CTA（路径 A 复练当前轮 / 路径 B 进入下一轮）。
- 复练 CTA 行为：路径 A `goReplay()` → `nav("workspace", { sourceSessionId, replayItems:retryFocusTurnIds, evidenceGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'retry_current_round', autoStartPractice:'1' })`；路径 B `goNextRound()` → `nav("workspace", { nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'next_round', autoStartPractice:'1' })`；未登录走 `useRequestAuth({type:'replay_practice', route:'workspace', params:{...sameParams, autoStartPractice:'1'}})`；workspace owner 创建 fresh session 后再进入 `practice`，不能复用报告来源 session。
- i18n 双语：新增 `report.*` + `generating.*` 命名空间（≥ 60 keys）；不复用 `workspace.*` 或 `practice.*`。
- 旧口径负向 grep + Playwright pixel parity (desktop 1440×900 + mobile 390×844) + i18n 完整性断言。

完成后用户在 frontend-workspace-and-practice plan 002 已经送出 `nav("generating", ...)` 后能进入真实 GeneratingScreen 轮询 → 真实 ReportScreen 渲染 → 复练 CTA 交给 workspace auto-start → fresh practice session，完整闭环 P0 报告路径。Backend 真实数据由 [backend-review plan 001](../../../backend-review/plans/001-report-generation-baseline/plan.md) Phase 5 提供；本 plan 在 backend-review Phase 5 完成前用 fixture-backed transport 测试，Phase 5 切真 API regression。

## 2 背景

frontend-workspace-and-practice spec v1.3 + plan [002-practice-text-event-loop](../../../frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)（completed, 2026-05-14）已交付 `completePracticeSession` 真实调用 + `useCompletePracticeSession` hook + `buildPracticeHandoffParams` 13 字段构造器（`{planId, targetJobId, jdId, resumeVersionId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}`）+ `nav("generating", ...)` 跳转 + `ReportWithJob` 缓存防双触发。当前 `frontend/src/app/App.tsx::renderRouteScreen` 中 `generating` 与 `report` 路由仍渲染 `PlaceholderScreen`（D1 占位），等待本 plan 替换。

> **关于 `useRequestAuth` payload type 与 UI 真理源的 deliberate divergence**：UI 真理源 `ui-design/src/screen-report.jsx:114` 在复练 / 下一轮 CTA 触发 auth 时使用通用 `requestAuth({type:'create_session', ...})`。生产前端按 [spec D-5](../../spec.md#31-已锁定决策) + product-scope §4.1 "复练优先" + backend-practice D-4 plan goal 四值的要求需要区分 `replay_practice`（路径 A retry_current_round）与 `start_practice`（workspace 立即面试）/ `next_round_practice`（未来路径 B 单独区分）三类 PendingAction，使 pendingAction 恢复后能正确派发到 `goReplay` vs `goNextRound` vs workspace 立即面试不同的 nav payload。此分叉只影响 auth/routing 契约，不影响视觉源级复刻 gate；plan 001 引入 `replay_practice` 一种新 type（路径 A + 路径 B 共用），未来如需要进一步细分由后续 plan 演化。

backend-review spec v1.0 + plan [001-report-generation-baseline](../../../backend-review/plans/001-report-generation-baseline/plan.md) 是本 plan 的 backend handoff source；backend-review Phase 0 会扩 `openapi/fixtures/Reports/getFeedbackReport.json` 的 `report-failed` variant + `openapi/fixtures/Reports/listTargetJobReports.json` 的 `empty` variant + B1 `REPORT_NOT_FOUND` 错误码 + B2 `getFeedbackReport` 404 response schema。本 plan 仅消费这些 fixtures + generated client，不直接修改 OpenAPI / B1 / B2 truth source。

UI 真理源：[`ui-design/src/screen-report.jsx`](../../../../../ui-design/src/screen-report.jsx)（`ReportScreen` 主控制器三态分发 + `ReportDashboard` 主组件 + 5 个 detail tab 内容 + `ReportStatButton` / `DimRow` / `StatCard` primitives）+ [`ui-design/src/screens-p0-complete.jsx`](../../../../../ui-design/src/screens-p0-complete.jsx) lines 269-399（`ReportGeneratingScreen`）+ [`ui-design/src/data.jsx`](../../../../../ui-design/src/data.jsx) lines 175-242（`report` sample 数据结构）+ [`docs/ui-design/report-dashboard.md`](../../../../ui-design/report-dashboard.md)（dashboard-only 形态 + 准备度 4 档 + 维度状态 + 复练规则）。

001 plan 不依赖 frontend-workspace-and-practice 任何待落地 plan（003 voice / 004 deferred handoff）；004 计划本来预留为本 spec 同款 generating handoff plan，本 spec v1.0 创建后 004 编号回收并由 frontend-report-dashboard 完整承接（与 frontend-workspace-and-practice spec.md §7 关联计划 "不再预留本 subspec 内的 report 或 company_intel plan" 一致）。

InterviewContext reducer 已经在 frontend-workspace-and-practice plan 001 + 002 落地完整 13 字段写入（与 `buildPracticeHandoffParams` 输出一致）；本 plan 仅 read，不新增 reducer action（与 spec D-8 一致）。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）+ cross-layer contract（消费 backend-review 已落地的 wire 契约 + frontend-workspace-and-practice 已落地的 handoff）。
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest + @testing-library/react + jsdom）；每个 Phase 在新增组件 / hook / utils 前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言（method、path、query、body schema、header 反查）、URL/state 隐私反查与负向旧 testid / 旧 enum / 旧 reportLayout 字面量断言；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 5 扩展为 `report.spec.ts` + `generating.spec.ts`（desktop 1440×900 + mobile 390×844）。新增组件位于 `frontend/src/app/screens/report/` 与 `frontend/src/app/screens/generating/`；测试文件与组件 colocate（`*.test.tsx` / `*.test.ts`）。`test-plan.md` + `test-checklist.md` 拓展测试形态与 phase 映射。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 [bdd-plan.md](./bdd-plan.md) 定义 4 个场景 `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059`，[bdd-checklist.md](./bdd-checklist.md) 跟踪每个场景资产创建与执行；主 [checklist.md](./checklist.md) 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖 + pixel parity + contract drift）。

## 3.5 Coverage Matrix

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path · GeneratingScreen happy path | 进入 generating 携带 reportId + 13 字段；渲染 5 阶段进度动画 + 实时观察流；轮询 `getFeedbackReport(reportId)` 多次（指数退避节奏）；status='ready' 自动 nav report | `screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 | 1 | E2E.P0.056 + Vitest `generating/GeneratingScreen.test.tsx` + `hooks/useReportGenerationPoll.test.ts` |
| Primary path · ReportDashboard 渲染（ready） | 携带 sessionId + reportId 进入 report；`getFeedbackReport` 返回完整 FeedbackReport；渲染 Header + ContextStrip + 4 Summary Cards + 5 Detail Tabs（默认 `questions`）+ 维度卡片 + 优先级 + 复练重点 + 题目回顾 + 风险亮点；readiness tab 通过显式切换覆盖 | `screen-report.jsx::ReportDashboard` lines 80-257 | 2+3 | E2E.P0.056 + Vitest `report/ReportScreen.test.tsx` + `report/ReportDashboard.test.tsx` |
| Primary path · 5 detail tab 内容 | 切换 5 个 tab；每个 tab 内容源级复刻：readiness（拨号盘 + JD 对齐 + 证据密度 + 下一档门槛）/ dimensions（二级网格）/ questions（侧栏 + 当前题分析）/ evidence（风险 + 亮点）/ next（路径 A vs 路径 B 对比） | `screen-report.jsx::ReportDetailSurface` lines 311-516 | 3 | E2E.P0.056 子断言 + Vitest `report/DetailTabs.test.tsx` |
| Primary path · 复练 CTA 路径 A | 用户点击「复练当前轮」CTA → nav workspace auto-start with retry_current_round payload + retry_focus_turn_ids；workspace owner 创建 fresh session 后进入 practice；未登录走 useRequestAuth 到 workspace | `screen-report.jsx::goReplay` line 116 | 4 | E2E.P0.057 + Vitest `report/ReplayCta.test.tsx` |
| Primary path · 复练 CTA 路径 B | 用户点击「进入下一轮」CTA → nav workspace auto-start with next_round payload；workspace owner 创建 fresh session 后进入 practice；未登录走 useRequestAuth 到 workspace | `screen-report.jsx::goNextRound` line 117 | 4 | E2E.P0.057 + Vitest `report/ReplayCta.test.tsx` |
| Alternate path · GeneratingScreen 失败处理 | fixture report-failed（status='failed' + errorCode）→ 自动 nav report?reportStatus=failed&errorCode=... | spec §2.1 generating 失败分支 | 1 | Vitest `useReportGenerationPoll.test.ts` 子用例 + E2E.P0.058 |
| Alternate path · GeneratingScreen 超时 | fixture 永久 generating → max attempts 达到 ErrorState「报告生成超时」+ retry 重启轮询 | spec D-3 | 1 | Vitest `GeneratingScreen.test.tsx` 子用例 |
| Alternate path · GeneratingScreen visibility 暂停 | tab 隐藏 → 暂停轮询；恢复显示 → 恢复轮询 | spec D-3 | 1 | Vitest fake visibility event + `useReportGenerationPoll.test.ts` |
| Failure / recovery · ReportFailureState | `report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`；不调 `getFeedbackReport`；渲染失败卡片 + errorCode 文案映射 + CTA「重新生成」/「返回 workspace」 | `screen-report.jsx::ReportFailureState` lines 61-77 | 2 + 5 | E2E.P0.058 + Vitest `ReportFailureState.test.tsx` |
| Failure / recovery · ReportMissingSessionState | `report?reportId=R`（缺 sessionId）→ 不调 `getFeedbackReport`；渲染缺会话卡片 + CTA「返回 workspace」 | `screen-report.jsx::ReportMissingSessionState` lines 46-59 | 2 | E2E.P0.058 + Vitest `ReportMissingSessionState.test.tsx` |
| Failure / recovery · `getFeedbackReport` 5xx 错误 | network 5xx / timeout → InlineError + retry 按钮 + retry 复用 reportId；3 次失败显示「返回 workspace」fallback | spec §3.2 待确认事项 | 5 | Vitest `report/ReportErrors.test.tsx` + E2E.P0.058 子断言 |
| Failure / recovery · `getFeedbackReport` 404 cross-user | 用户 B 访问 R_A → 404 `REPORT_NOT_FOUND` → 渲染 ReportFailureState 但使用独立的 `report.failureState.notFound.{title,desc}` / `report.failureState.errorCode.REPORT_NOT_FOUND` i18n key（与 `failureState.errorCode.AI_*` 通用映射区分），文案为 zh "未找到该报告" / en "Report not found"；不暴露 reportX 存在性；不与 AI errorCode 文案混淆 | backend-review D-15 + B1 `REPORT_NOT_FOUND` | 5 | Vitest `report/CrossUserNotFound.test.tsx` + `report/ReportFailureStateNotFound.test.tsx` + E2E.P0.058 子断言 |
| UI stale-contract negative · `listTargetJobReports` 0 调用 | `listTargetJobReports` 在 `frontend/src/app/screens/{report,generating}/` 范围零调用（三层断言：scoped legacy grep + Vitest mockTransport spy + scenario verify.sh）；dashboard-only D-7 边界 | spec §2.2 + frontend-workspace-and-practice 边界 | 全 phase | Vitest `mockTransport.spy.test.ts` 反向断言 + `scripts/lint/frontend_report_dashboard_legacy.py` literal grep + scenario verify.sh |
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
| UX · loading state | generating mount 阶段 1 个 viewport 不闪烁；report mount 阶段 skeleton placeholder（数据未 ready） | spec §4 + ui-design 隐式 | 1+2 | Vitest fake timer + `reportLoading.test.tsx` |
| UX · empty state | report 中 highlights / issues / nextActions / questionAssessments 任一为空数组 → 该区块显示 EmptyHint 占位（非 dashboard fail） | spec D-4 ready status 字段填充 | 3 | Vitest `report/EmptyHint.test.tsx` |
| UX · error state | `getFeedbackReport` 5xx / network → inline error + retry；3 次失败显示返回 workspace fallback | spec §3.2 | 5 | Vitest `ReportErrors.test.tsx` |
| UX · i18n zh/en | 全文案通过 typed locale helper；新增 `report.*` + `generating.*` namespace ≥ 60 keys；切换立即重绘 | spec D-7 + 001 D1 typed locale helper | 1-5 | Vitest `reportI18n.test.tsx` + namespace 同步断言 |
| UX · dark + customAccent + 主题切换 | generating + report 在 8 主题 × dark 组合下 computed background / color 出现可见变化 | D2 `data-theme / data-mode / data-custom-accent` + `ui-design/src/primitives.jsx` | 5 | Playwright `tests/pixel-parity/{generating,report}.spec.ts` 主题循环 |
| UX · responsive layout (mobile 390×844) | generating 居中进度态不溢出视口；report 三列折叠为单列 + Detail Surface 切 collapsible Accordion + 复练 CTA sticky bottom；BindingPill 不溢出 | spec §4 mobile | 1+5 | Playwright mobile project + Vitest jsdom 视口模拟 |
| UI source structure parity · GeneratingScreen | 页头 + 进度条 + 5 阶段列表 + 实时观察流 + 底部提示 + 「通知我」按钮；testid `generating-{header,progress,phase-${idx},live-stream,evidence-${idx},sla-hint,notify-cta}` | `screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 | 1 | Vitest + testid 命中 + 控件类型断言 |
| UI source structure parity · ReportScreen 顶层三态 | ReportScreen 主组件按 params 分发：`ReportFailureState` / `ReportMissingSessionState` / `ReportDashboard`；testid 顶层 `report-screen` + 子树根 `report-failure-state` / `report-missing-session` / `report-dashboard` | `screen-report.jsx::ReportScreen` lines 1-44 | 2 | Vitest + 三态测试用例 |
| UI source structure parity · ReportDashboard | 返回按钮 + Header（titles + 双 CTA）+ ContextStrip + 4 Summary Cards + Detail Surface + 维度卡片行 + 优先级 + 复练重点 + 题目回顾 + 风险亮点；testid `report-{return,header-title,header-subtitle,replay-cta,next-cta,context-strip,summary-{readiness,dimensions,questions,next},detail-tab-${tab},detail-panel-${tab},dim-row-${idx},top-priority,next-practice-${idx},perq-${idx},issue-${idx},highlight-${idx}}` | `screen-report.jsx::ReportDashboard` lines 80-257 | 2+3 | Vitest + testid 命中 |
| UI source structure parity · ReportFailureState | 失败卡片 + errorCode 文案 + CTA「重新生成」+ CTA「返回 workspace」；testid `report-failure-{title,desc,error-code,retry-cta,back-to-workspace}` | `screen-report.jsx::ReportFailureState` lines 61-77 | 2 | Vitest + testid |
| UI source structure parity · ReportMissingSessionState | 缺会话卡片 + CTA「返回 workspace」；testid `report-missing-session-{title,desc,cta}` | `screen-report.jsx::ReportMissingSessionState` lines 46-59 | 2 | Vitest + testid |
| UI source structure parity · ReportContextStrip | sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条；testid `report-context-{session,job,round,resume,modality,practice-mode,hints}` | `screen-report.jsx::ReportContextStrip` lines 145 | 2 | Vitest + testid |
| UI source structure parity · 5 detail tabs | readiness（拨号盘 `report-readiness-dial`+ 二级详情 `report-readiness-jd-align / -evidence-density / -next-threshold`）/ dimensions（二级网格 `report-dimensions-grid` + 各 `report-dim-card-${idx}`）/ questions（侧栏 `report-questions-list` + 主体 `report-questions-detail-{topic,good,missing,frame,evidence,follow-up}`）/ evidence（`report-evidence-risk-${idx}` + `report-evidence-highlight-${idx}`）/ next（`report-next-path-{a,b}` + `report-next-cta-{a,b}`） | `screen-report.jsx` lines 311-516 | 3 | Vitest + testid 命中（每 tab 独立测试） |
| UI source structure parity · DimRow (维度行) | `DimRow` name + score bar + state tag + confidence；testid `report-dim-row-{name,score,state,confidence}` | `screen-report.jsx::DimRow` lines 565-577 | 3 | Vitest |
| UI source structure parity · 准备度 tier 4 档文案与色调 | not_ready / needs_practice / basically_ready / well_prepared 4 档 zh/en 文案 + 色调；不引入 5 档 readiness 旧 numeric score 字面量 | spec D-10 + `ui-design/src/data.jsx` readinessLabel | 2+3 | Vitest + 4 档矩阵测试 |
| UI source structure parity · 维度卡片状态映射 | strong / meets_bar / needs_work 三态文案与色调；不引入 weak / developing / proficient / acceptable rubric 内部 score_levels label（rubric label 不暴露到 UI） | spec D-11 + B1 DimensionStatus | 3 | Vitest + 三态矩阵 |
| UI visual geometry parity · desktop | 1440×900 generating + report 主屏 + 5 detail tab + dashboard skeleton + failure state + missing session state bounding box stays in viewport, no overlap；`generating` TopBar 隐藏不占位；`report` 默认 App chrome / TopBar 可见且不进入一级导航；底部 CTA sticky 不被遮挡 | n/a | 5 | Playwright `tests/pixel-parity/{generating,report}.spec.ts` desktop project |
| UI visual geometry parity · mobile | 390×844 generating 居中 + report 三列折叠 + Detail Surface Accordion + CTA sticky | n/a | 5 | Playwright mobile project |
| UI visual geometry parity · dark / customAccent / theme | 8 主题 × dark + customAccent oklch 切换可见变化 | n/a | 5 | Playwright |
| UI visual geometry parity · clean-checkout gate | DOM anchor + computed style + bounding box + responsive geometry + non-empty screenshot smoke；仅当稳定 baseline 已提交或本 phase 明确更新 baseline 时才追加 `toHaveScreenshot` | n/a | 5 | Playwright + frontend clean-checkout smoke |
| UI stale-contract negative · 旧 reportLayout | 旧 `reportLayout='timeline'` / `reportLayout='document'` 等旧字面量在 report / generating 新代码中 0 命中（不计 negative tests / docs） | spec D-12 + product-scope D-7 | 全 phase | Vitest + scenario verify negative grep |
| UI stale-contract negative · 旧 readiness 5 档 | 旧 readiness numeric (`readinessScore`) / 旧 5 档（not_ready / needs_practice / basically_ready / well_prepared / fully_prepared 之类的 5 档值）/ 旧 `readiness_score` 字段在 report / generating 新代码中 0 命中 | spec D-10 + D-12 | 全 phase | grep negative |
| UI stale-contract negative · 独立错题 / Drill / Growth | 旧 `mistakes` route / `mistake_queue` testid / `drill_builder` / `growth_center` / 报告时间线 / 多形态 report / 独立 `report` 一级导航 entry / `practiceModeCard` 在 report / generating 模块 0 命中 | spec D-12 + product-scope D-6 | 全 phase | grep negative |
| UI stale-contract negative · 不直接 import prototype | `frontend/src/app/screens/{report,generating}/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getReportSampleDimensions` 等 prototype helper | spec §4 | 全 phase | Vitest + tsc grep |
| UI stale-contract negative · 不调 Practice operation | report / generating 模块不调 `getPracticeSession` / `appendSessionEvent` / `completePracticeSession` / `startPracticeSession` / `createPracticePlan` / `getPracticePlan`（归 frontend-workspace-and-practice 与 backend-practice） | spec §2.2 + frontend-workspace-and-practice owner | 全 phase | Vitest spy + tsc |
| UI stale-contract negative · 不调 createPracticeVoiceTurn / getCompanyIntel / getDebrief | report / generating 模块不调 voice / company intel / debrief operation；不渲染 voice surface / company intel screen / debrief screen 任何 DOM | spec §2.2 | 全 phase | Vitest spy + tsc + grep |
| Regression / legacy-negative · 工作区 + 后端契约 | `E2E.P0.044-047`（frontend-workspace-and-practice 002 practice 文本闭环）+ `E2E.P0.052-055`（backend-review 001 report 生成）全部作为真实 regression gate 重跑；fixture-backed PASS 只能证明前端 mock 合同，不替代真实闭环 | n/a | 5 | scenario rerun + `cd backend && go test ./cmd/api -run 'TestE2EP0052\|TestE2EP0053\|TestE2EP0054\|TestE2EP0055' -count=1` |
| Regression / legacy-negative · 不直接调用 LLM | report / generating 模块不出现 AI provider key / provider registry / prompt registry / AIClient / LLM endpoint / bypass generated client 的 ad hoc fetch | n/a | 全 phase | Vitest + grep negative |
| BDD 主路径 + 关键分支 + 失败恢复 + 旧口径负向 | 见 [bdd-plan.md](./bdd-plan.md) 4 场景矩阵 | n/a | 1-5 | E2E.P0.056/057/058/059 |

### 高风险类别 N/A 说明

- **隐私 / 安全 · audio buffer**：本 plan 不实现 voice surface / STT / TTS；audio buffer 不进入 report 屏。N/A 原因记录在此。
- **Privacy · LLM prompt raw text**：B2 在 backend-review 服务端 redact prompt 与 response；前端不直接调用 LLM；`provenance` 字段（`promptVersion / modelId`）只是版本/标识，不含 prompt body；前端只渲染版本号到 AI 透明度区域（report dashboard ContextStrip / questions tab evidence），因此 prompt-response 明文不在前端泄漏面。N/A 原因记录在此。
- **Out-of-scope · 报告时间线 / 多形态 report / 报告列表 UI**：D-12 retired；plan 001 不实现；归 plan 002 future（且需先修订 product-scope D-7）。

## 3.6 Frontend / Backend Operation Matrix

本 plan 走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated client + fixture-backed transport 完成 P0 UI/BDD；同时当前仓库 backend-review/001 正在并行设计，Phase 5 完成后必须跑对应真实 handler regression。fixture-backed PASS 不等于真实 backend 闭环；尤其 `getFeedbackReport` 真实 200 + status 转换 + cross-user 404 由 backend-review/001 Phase 5 接管。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | 当前已有 `default` / `report-generating` / `prototype-baseline`；本 plan 依赖 backend-review/001 Phase 0 新增 `report-failed`（status='failed', errorCode='AI_PROVIDER_TIMEOUT' 等） | `useReportGenerationPoll` GeneratingScreen + ReportScreen 单次拉取；OpenAPI path param `reportId` only | backend-review/001 Phase 5 真实 handler；read-only；user-scoped；返回 status 决定 placeholder vs ready vs failed | `feedback_reports` + `question_assessments` read | none in frontend | frontend `E2E.P0.056 / E2E.P0.058`；backend `E2E.P0.053 / E2E.P0.055` |
| `getTargetJob` | 当前已有 `openapi/fixtures/TargetJobs/getTargetJob.json` `default` | `useReportContextData` / `ReportContextStrip` 只读 target job title/companyName；失败时显示 targetJobId fallback，不阻塞报告正文 | backend-targetjob 既有 handler；read-only；user-scoped | `target_jobs` read | none | `E2E.P0.056` ContextStrip 子断言 + Vitest `ReportContextStrip.test.tsx` |
| `getResumeVersion` | 当前已有 `openapi/fixtures/Resumes/getResumeVersion.json` `default` | `useReportContextData` / `ReportContextStrip` 只读 resume version displayName；失败时显示 resumeVersionId fallback；不得读取 raw resume text | backend-resume 既有 handler；read-only；user-scoped | `resume_versions` read | none | `E2E.P0.056` ContextStrip 子断言 + privacy negative |
| `listTargetJobReports` | N/A（本 plan 不消费） | 不消费（dashboard-only D-7） | backend-review/001 Phase 5 真实 handler；read-only；user-scoped；cursor 分页 | `feedback_reports` cursor read | none | 负向断言（在 generating / report 模块零调用） |
| `completePracticeSession` | N/A（本 plan 不消费） | 由 frontend-workspace-and-practice plan 002 消费；本 plan 不调用 | backend-practice/002 已落地 | — | — | 负向断言 |
| `appendSessionEvent` / `getPracticeSession` / 其他 Practice operation | N/A | 由 frontend-workspace-and-practice plan 002 消费；本 plan 不调用 | — | — | — | 负向断言 |
| `createPracticeVoiceTurn` | N/A | 本 plan **不消费**；voice surface deferred | missing operation；blocked | — | — | 负向断言（grep `createPracticeVoiceTurn` 在 report / generating 模块 0 命中） |
| `getCompanyIntel` | N/A | 本 plan **不消费**；company-intel owner 承接 | external owner | — | — | 负向断言 |

## 3.7 InterviewContext × PracticeDisplayContext View-Model Mapping

正式前端不得从 `ui-design/src/data.jsx` 或未声明 fixture 字段补齐 `InterviewContext` 之外的数据；ContextStrip 所需的人类可读 job/resume label 只能来自 §3.6 operation matrix 中声明的 `getTargetJob` / `getResumeVersion`，并在失败时回退 ID 显示。本 plan 在 frontend-workspace-and-practice plan 001 + 002 已落地的 `InterviewContext` reducer 基础上**仅 read**，不新增 reducer action。具体 mapping：

| 字段 | Source | Rule |
|------|--------|------|
| `reportId` | route param 或 InterviewContext.reportId（由 frontend-workspace-and-practice plan 002 buildPracticeHandoffParams 写入） | 必填；缺失 → 渲染 ErrorState「报告 ID 缺失」 |
| `sessionId` | route param 或 InterviewContext.sessionId | 必填；缺失 → 渲染 ReportMissingSessionState |
| `planId / targetJobId / jdId / resumeVersionId / roundId` | route param 或 InterviewContext | 本 plan 只 read，用于 ContextStrip 显示 + 复练 CTA payload 构造；不新增 handoff key |
| `targetJob.title / targetJob.companyName` | generated `getTargetJob(targetJobId)`；失败时回退 `targetJobId` | ContextStrip 人类可读 target job label；不得从 `ui-design/src/data.jsx` 或硬编码 fixture 复制 |
| `resumeVersion.displayName` | generated `getResumeVersion(resumeVersionId)`；失败时回退 `resumeVersionId` | ContextStrip 人类可读 resume label；不得读取 `ResumeAsset.originalText` / `parsedTextSnapshot` / raw resume body |
| `mode / modality` | route param（默认 `text/text`） | 本 plan UI 不切换；用于 ContextStrip 显示「文本面试 / 语音面试」 |
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

**目标**：阻塞性 preflight assert — 在 Phase 1 编码前必须把以下 4 个 backend-review/001 Phase 0 deliverable 落实到当前仓库；任一缺失则本 plan Phase 1 红灯进入 `blocked` 状态等待 backend-review/001 Phase 0 推进，并通知 backend-review owner。preflight 失败不属于本 plan 编码工作范围，但本 plan 必须在 Phase 1 启动前显式断言；不得在缺位状态下 silently fallback 到 frontend-only stub。

#### 0.1 `FeedbackReport.errorCode` 字段存在断言

新增 `frontend/src/app/screens/report/__tests__/preflight.test.ts`：

- 读取 `openapi/openapi.yaml` 中 `FeedbackReport` schema；断言 `errorCode` 字段存在且类型 `oneOf: [ApiErrorCode, null]`；缺失则测试 fail + 在 message 中提示 "blocked on backend-review/001 Phase 0.2 + 0.4 (errorCode field in B2 schema)"。
- 兼容 generated TS client：断言 `frontend/src/api/generated/` 中 `FeedbackReport` interface 含可选 `errorCode` 属性。

#### 0.2 `report-failed` fixture variant 存在断言

在 `preflight.test.ts` 中追加：

- 读取 `openapi/fixtures/Reports/getFeedbackReport.json`，断言 `scenarios.report-failed.response.body.status === 'failed'` + `scenarios.report-failed.response.body.errorCode` 非 null（建议值为 `AI_PROVIDER_TIMEOUT`）；缺失则 fail + 提示 "blocked on backend-review/001 Phase 0.4 (report-failed fixture variant)"。
- 同时断言 `scenarios.report-generating` 与 `scenarios.default` 已经存在（这两个已在仓库内）。

#### 0.3 `listTargetJobReports.empty` fixture variant 存在断言

在 `preflight.test.ts` 中追加：

- 读取 `openapi/fixtures/Reports/listTargetJobReports.json`，断言 `scenarios.empty.response.body.items === []`、`scenarios.empty.response.body.pageInfo.hasMore === false`、`scenarios.empty.response.body.pageInfo.nextCursor === null`；缺失则 fail + 提示 "blocked on backend-review/001 Phase 0.4 (empty fixture variant)"。
- 虽然本 plan 不消费 `listTargetJobReports`，empty fixture 的存在性是 backend-review/001 Phase 0 deliverable 的健康指示器；缺失提示 backend-review/001 Phase 0 整体未完成，本 plan 不应进入 Phase 1。

#### 0.4 `REPORT_NOT_FOUND` 错误码存在断言

在 `preflight.test.ts` 中追加：

- 读取 `shared/conventions.yaml#errors`，断言 `REPORT_NOT_FOUND` 行存在 + `httpStatus: 404` + `retryable: false`；缺失则 fail + 提示 "blocked on backend-review/001 Phase 0.1 (REPORT_NOT_FOUND in B1)"。
- 同时断言 `frontend/src/api/generated/` 中存在 generated TS 等价常量（如 `ApiErrorCode.REPORT_NOT_FOUND` 或 `errors.REPORT_NOT_FOUND`）；缺失则 fail + 提示 "blocked on backend-review/001 Phase 0.1 (generated TS error constant)"。

#### 0.5 Phase 0 收口 gate

- `pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts` 全绿（即 4 项断言全部通过）
- 如有任一断言 fail：本 plan 状态保持 `active` 但 Phase 1 不启动；通过 [bug-report](../../../../bugs/PATTERNS.md) 或 retrospective 联动通知 backend-review/001 owner
- 通过后，删除任何临时 stub 文件（如有）并进入 Phase 1

### Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

#### 1.1 新增 `frontend/src/app/screens/generating/GeneratingScreen.tsx`

按 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399) 源级复刻渲染：页头（标题 + 副文案 "Reading every turn. Evidence first."）+ 进度条（百分比 + phase indicator）+ 5 阶段列表（每个 done/active/pending 状态圆圈 + 标签）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA "<12s target" + 「通知我」UI-only 按钮）。本 phase 接入轮询 hook；reportId 缺失时不发请求、直接渲染 ErrorState。

#### 1.2 新增 `frontend/src/app/screens/generating/hooks/useReportGenerationPoll.ts`

通过 generated client 轮询 `getFeedbackReport(reportId)`；React state 跟踪 `idle / polling / ready / failed / timeout / error / paused` 七态；指数退避（初始 1.5s × 1.5 上限 8s，max attempts 30）；visibility / focus 事件暂停-恢复轮询；网络 5xx / network error retry 复用当前 attempt count；status='ready' → 调 `onReady(report)` callback；status='failed' → 调 `onFailed(errorCode)` callback；max attempts 达到 → state='timeout' 触发 ErrorState；request init 不含 `Idempotency-Key` header。

#### 1.3 新增 `frontend/src/app/screens/generating/components/`

包含：`HeaderHero.tsx`（标题 + 副文案）/ `ProgressBar.tsx`（百分比 + phase indicator）/ `PhaseList.tsx`（5 阶段列表 + 状态圆圈）/ `LiveEvidenceStream.tsx`（fade-in 流）/ `SlaHint.tsx`（底部提示 + 通知我按钮）/ `GeneratingErrorState.tsx`（reportId 缺失或 timeout 兜底）。每个组件从 `ui-design/src/screens-p0-complete.jsx` 同名片段复刻 DOM。

#### 1.4 路由壳替换

在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `generating` → `<GeneratingScreen route={route} />`（替换 D1 `PlaceholderScreen`）；保持 `generating` 在 `NO_CHROME_ROUTES` 中隐藏 TopBar；`report` 仍渲染 `PlaceholderScreen`，待 Phase 2 替换。

#### 1.5 i18n locale 扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `generating.*` 命名空间（≥ 20 keys：header.title / header.subtitle / phase.1 / phase.2 / phase.3 / phase.4 / phase.5 / progress.phaseN / evidence.streamLabel / sla.target / sla.notifyCta / errors.missingReportId / errors.timeout / errors.retry / errors.backToWorkspace 等）；`messages.ts` 类型聚合补齐。

#### 1.6 Vitest 红灯 → 绿灯

新增 `generating/__tests__/GeneratingScreen.test.tsx`：测 i18n zh/en 切换重绘、≥ 10 个 `generating-*` testid 存在、reportId 缺失 → 不发请求 + 渲染 ErrorState、timeout 状态渲染 retry CTA、负向断言不出现 `mistakesQueue` / 旧 `report-timeline` testid。

新增 `generating/__tests__/useReportGenerationPoll.test.ts`：7 态、指数退避节奏（fake timer）、max attempts、visibility / focus 暂停-恢复、status='ready' callback、status='failed' callback、unmount 取消、不含 `Idempotency-Key`。

#### 1.7 BDD-Gate Phase 1

- BDD-Gate: 验证 `E2E.P0.056` GeneratingScreen 部分通过（mount → 进度动画 → 轮询 → status='ready' nav report）

### Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

#### 2.1 新增 `frontend/src/app/screens/report/ReportScreen.tsx`

按 `ui-design/src/screen-report.jsx::ReportScreen` (lines 1-44) 源级复刻：顶层根据 `params.reportStatus === 'failed'` 或 `!sessionId` 分别渲染 `ReportFailureState` / `ReportMissingSessionState` / `ReportDashboard`。本 phase 实现 `ReportDashboard` 静态壳（含 Header + ContextStrip + 4 Summary Cards + Detail Surface 骨架占位）；5 个 detail tab 内容由 Phase 3 完成。

#### 2.2 新增 `frontend/src/app/screens/report/components/`

包含：`ReportHeader.tsx`（标题 + 副标题 + 双 CTA `复练当前轮` + `进入下一轮`）/ `ReportContextStrip.tsx`（sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条）/ `SummaryCards.tsx`（4 张 ReportStatButton：准备度 / 维度 / 题目 / 下一步）/ `ReportFailureState.tsx`（lines 61-77 源级复刻；CTA「重新生成」+「返回 workspace」）/ `ReportMissingSessionState.tsx`（lines 46-59 源级复刻；CTA「返回 workspace」）/ `ReportErrorBoundary.tsx`（getFeedbackReport 5xx 兜底）。

#### 2.3 新增 report 数据 hooks

- `useFeedbackReport.ts`：通过 generated client 单次拉 `getFeedbackReport(reportId)`；React state 跟踪 `loading / data / error / notFound`；404 → state='notFound' 渲染 ReportFailureState（cross-user 隔离）；5xx → state='error' + retry 按钮；request init 不含 `Idempotency-Key`。
- `useReportContextData.ts`：通过 generated `getTargetJob(targetJobId)` + `getResumeVersion(resumeVersionId)` 只读 ContextStrip label；成功时显示 target title/companyName + resume version displayName；任一失败时回退 targetJobId / resumeVersionId，不阻塞 ReportDashboard；不得读取 raw resume/JD/body 字段。

#### 2.4 路由壳替换

在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `report` → `<ReportScreen route={route} />`（替换 D1 `PlaceholderScreen`）；保持 `report` 不在 `NO_CHROME_ROUTES` 中，默认 App chrome / TopBar 可见，同时不把 `report` 加入一级导航。

#### 2.5 i18n locale 扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `report.*` 命名空间（≥ 40 keys：header.title / header.subtitle / cta.replay / cta.nextRound / context.session / context.job / context.round / context.resume / context.modality / context.practiceMode.assisted / context.practiceMode.strict / context.hints / summary.readiness / summary.dimensions / summary.questions / summary.next / readiness.tier.notReady / readiness.tier.needsPractice / readiness.tier.basicallyReady / readiness.tier.wellPrepared / failureState.title / failureState.desc / failureState.errorCode.* / failureState.retry / failureState.backToWorkspace / missingSession.title / missingSession.desc / missingSession.cta / loading / errors.network / errors.retry 等）；`messages.ts` 类型聚合补齐。

#### 2.6 Vitest 红灯 → 绿灯

新增 `report/__tests__/ReportScreen.test.tsx`：测三态切换（reportStatus='failed' / 缺 sessionId / 正常）、loading / data / error / notFound 四态、≥ 10 个 `report-*` testid 存在。

新增 `report/__tests__/ReportFailureState.test.tsx` / `ReportMissingSessionState.test.tsx`：分别测各自渲染 + CTA 行为。

新增 `report/__tests__/useFeedbackReport.test.ts`：测 4 态 + 404 / 5xx 处理 + 不含 `Idempotency-Key`。

新增 `report/__tests__/useReportContextData.test.ts`：测 `getTargetJob` / `getResumeVersion` 成功渲染人类可读 label、单 operation 失败回退对应 ID、不读取 raw resume/JD/body 字段、request init 不含写操作 header。

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

新增 `report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx`：每个测对应 testid + 数据驱动 + 边界（空 dimensions / 空 questions / 空 issues / 空 highlights 各自 EmptyHint）+ 4 档 readiness 文案矩阵 + 三档维度状态矩阵 + 不出现旧 5 档 readiness / rubric score_levels label。

#### 3.5 BDD-Gate Phase 3

- BDD-Gate: 验证 `E2E.P0.056` ReportDashboard 渲染部分通过（含 5 detail tab 切换 + 维度卡片 + 题目回顾 + 风险亮点）

### Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

#### 4.0 注册 `replay_practice` PendingAction type

在 `frontend/src/app/auth/pendingAction.ts` 中（当前 `PendingAction.type: string` 仅在测试中使用 `start_practice` 一个值）扩展：

- 把 `replay_practice` 加入允许的 type allowlist（如果当前实现使用 string union 或运行时 validator）。
- `encodePendingAction` / `decodePendingActionRoute` 必须支持 `replay_practice` round-trip（params 含 sourceSessionId / replayItems / evidenceGaps / planId / targetJobId / jdId / resumeVersionId / roundId / mode / modality / practiceMode / practiceGoal / autoStartPractice 等键）。
- 路由恢复（`AppPendingAction` 或同款）必须支持 `replay_practice` 恢复到 `workspace` 并触发 `autoStartPractice=1`（与 `start_practice` 在 `workspace` 路由 mount 时自动触发同款机制）。

新增 `frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts`（或扩展现有 `pendingAction.test.ts`）：

- `TestPendingActionEncodeDecodeReplayPractice` 测 encode → decode round-trip 字段对等
- `TestPendingActionReplayPracticeTypeAllowed` 测 type allowlist / discriminated union 包含 `replay_practice`
- 负向断言：encode 后 URL params / localStorage 不含 raw text（与 spec D-13 一致）

#### 4.1 实现复练 CTA `goReplay()` 路径 A

在 `ReportHeader.tsx` 与 `tabs/NextTab.tsx` 的 `report-next-cta-a` 按钮上绑定 `goReplay()`：组装 payload `{ sourceSessionId:sessionId, replayItems:retryFocusTurnIds, evidenceGaps:focusGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:InterviewContext.practiceMode, practiceGoal:'retry_current_round', autoStartPractice:'1' }`；未登录 → `useRequestAuth({type:'replay_practice', route:'workspace', params:{...sameParams}})`；登录后 pendingAction 回到 workspace auto-start；已登录 → 直接 `nav("workspace", payload)`，由 workspace owner 创建 fresh session 后进入 practice。

#### 4.2 实现复练 CTA `goNextRound()` 路径 B

同上，但 payload 为 `{ nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:InterviewContext.practiceMode, practiceGoal:'next_round', autoStartPractice:'1' }`；nextRoundId 来源：默认从 InterviewContext.roundId 推断（roundId + 1）或 fixed mapping；如未来需要从 backend 拿真实下一轮 metadata，先回 backend-targetjob owner 修订（本 plan 不引入新 backend 调用）。

#### 4.3 完整 ReportFailureState handoff

`ReportFailureState.tsx` CTA「重新生成」点击 → `nav("generating", { sessionId, reportId, ...passThroughContext })` 重新进入 generating 屏触发轮询（不直接调用 backend；轮询 hook 自然命中既有 failed report）；CTA「返回 workspace」点击 → `nav("workspace", { targetJobId, jdId, planId, resumeVersionId })`。

#### 4.4 完整 GeneratingScreen handoff

`useReportGenerationPoll` 的 `onReady(report)` callback → `nav("report", { sessionId, reportId, ...passThrough })`；`onFailed(errorCode)` callback → `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode, ...passThrough })`；timeout state → 不自动 nav；用户点 retry 重启轮询；nav 调用必须防抖（handoffNavigatedRef）。

#### 4.5 Vitest 红灯 → 绿灯

新增 `report/__tests__/ReplayCta.test.tsx`：测路径 A 已登录经 workspace auto-start 创建 fresh session + 未登录 useRequestAuth 恢复到 workspace auto-start；payload 字段完整；负向断言 raw text 不在 payload。

在 `report/__tests__/ReplayCta.test.tsx` 覆盖路径 B 同上；断言 nextRoundId 推断逻辑与 fresh session start。

新增 `report/__tests__/ReportFailureHandoff.test.tsx`：测「重新生成」nav generating + 「返回 workspace」nav workspace。

扩展 `generating/__tests__/GeneratingScreen.test.tsx`：测 ready / failed / timeout 三态分别 nav + 防抖（多次 ready callback 只 nav 一次）。

#### 4.6 BDD-Gate Phase 4

- BDD-Gate: 验证 `E2E.P0.057` 通过（复练 CTA 路径 A + 路径 B 经 workspace auto-start 进入 fresh practice session）
- BDD-Gate: 验证 `E2E.P0.058` 通过（GeneratingScreen 轮询命中 `status='failed'` → nav failed report + ReportFailureState + ReportMissingSessionState + 跨用户 + 隐私 route params）
- BDD-Gate: 验证 `E2E.P0.056` 整链完整通过（含 GeneratingScreen mount → 进度动画 → 轮询 ready → nav report → ReportDashboard 渲染 → 5 detail tab 切换 → CTA wire 完整）；Phase 1 + Phase 3 仅做局部断言，Phase 4 复练 CTA wire 完成后才算完整通过

### Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + 旧口径负向

#### 5.1 完整状态机集成回归

`pnpm vitest run`（全 frontend 测试）+ `pnpm typecheck` 全绿；扩展现有 `App.test.tsx` 添加 `generating-screen` 与 `report-dashboard` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` / `report` route alias 处理；扩展 `pendingActionReplayPractice.test.ts` 添加 `replay_practice` pendingAction 恢复到 workspace auto-start 的 round-trip；扩展 `scenarios/p0-002-auth-pending-action-resume.test.tsx` 添加 `replay_practice` resume path 验证。

#### 5.2 Playwright pixel parity 加挂

新增 `frontend/tests/pixel-parity/generating.spec.ts` + `frontend/tests/pixel-parity/report.spec.ts`：desktop 1440×900 + mobile 390×844 两 viewport；测 generating + report 主屏 + 5 detail tab + 三态（dashboard/failure/missing-session）+ 8 主题 × dark 切换；clean-checkout 必须断言 DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke，只有在稳定 baseline 已提交或本 phase 明确更新 baseline 时才追加 `toHaveScreenshot`。

#### 5.3 scenario 加挂

在 `test/scenarios/e2e/` 派生 4 个 scenario 目录 `p0-056-generating-to-report-happy-path/` / `p0-057-replay-cta-paths-a-and-b/` / `p0-058-report-failure-and-missing-session/` / `p0-059-report-pixel-parity-i18n-and-legacy-negative/`，每个含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x 可执行）；trigger 跑对应 Vitest 套件；verify 反查 testid / nav payload / 负向 grep。

#### 5.4 旧口径负向 grep

scoped grep 在 `frontend/src/app/screens/{report,generating}/` 范围：
- `reportLayout` / `report_layout`
- 旧 5 档 readiness（如 `fully_prepared` 旧字面量）/ `readinessScore` numeric 字段
- `mistakes_queue` / `mistakesQueue` / `mistake-queue` testid
- `drill_builder` / `drillBuilder` / `drill-builder` testid
- `growth_center` / `growthCenter` / `growth-center` testid
- `report_timeline` / `reportTimeline`
- `report_form` / `reportForm`
- 旧独立 `mistakes` route entry
- 旧 prototype data import：`ui-design/src/data.jsx` / `window.EI_DATA` / `getReportSampleDimensions` 等
- `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief` 调用
- `listTargetJobReports` 调用（dashboard-only D-7；plan 001 不消费列表）

负向 grep 在 `scripts/lint/frontend_report_dashboard_legacy.py`（新增）+ `frontend/src/app/screens/{report,generating}/__tests__/legacyNegative.test.ts` 实现。

#### 5.5 i18n 完整性断言

新增 `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts`：断言 `report.*` 与 `generating.*` 命名空间 zh / en 同步无缺漏（key 集合相等）；新增 key ≥ 60；切换 locale 时所有 testid 文案重绘。

#### 5.6 跨 owner regression

在 Phase 5 收口阶段重跑：
- frontend-workspace-and-practice/002 BDD `E2E.P0.044-047`（保证未被破坏）
- backend-review/001 BDD `E2E.P0.052-055`（如已 implement；如未 implement 则跳过 backend regression，仅跑 fixture-backed scenario）
- backend-practice/002 BDD `E2E.P0.038-043`（必要时通过 cmd/api 重跑）

#### 5.7 BDD-Gate Phase 5

- BDD-Gate: 验证 `E2E.P0.059` 通过（Playwright pixel parity + i18n + 旧口径负向）

#### 5.8 文档收口

更新 `docs/spec/frontend-report-dashboard/plans/INDEX.md`：001 状态保持 `active`（plan-review / sync-doc-index 推进到 completed 由后续动作完成）。

新增 `frontend/src/app/screens/report/README.md` + `frontend/src/app/screens/generating/README.md`：简明 handoff 段落，记录 001 新增 component / hook / nav 边界 / handoff 给 backend-review 与 frontend-workspace-and-practice 的边界。

## 5 验收标准

- Phase 1 ~ Phase 5 checklist 全部勾选
- 关联 BDD 场景 `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059` 均由对应 Vitest + Playwright + scenario 执行通过
- `pnpm --filter @easyinterview/frontend test` / `pnpm --filter @easyinterview/frontend typecheck` / `pnpm --filter @easyinterview/frontend test:pixel-parity` / `pnpm --filter @easyinterview/frontend build` 全绿
- `make codegen-check` 通过（不修改 OpenAPI / generated client，但 build 时反查 drift）
- `make validate-fixtures` 通过（依赖 backend-review/001 Phase 0 新增 fixture variants）
- `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过
- 001 范围内代码与文档中无 §3.5 / §D-12 列出的 legacy 术语 / 旧 reportLayout / 5 档 readiness / 独立 mistakes route / drill_builder / growth_center / 报告时间线 / 多形态 report 字面量出现
- frontend-workspace-and-practice/002 BDD regression（`E2E.P0.044-047`）通过；backend-review/001 BDD regression（`E2E.P0.052-055`）如已 implement 则通过

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| backend-review/001 Phase 0 fixture 扩展（`report-failed` / `empty`）未及时 land 导致本 plan Phase 1 测试 fail | Phase 1 测试用本地 mock 模拟 fixture variant（不进 generated client），等 backend-review Phase 0 land 后切换；如 backend-review/001 严重 delay，本 plan Phase 1 可以独立 land + 待 backend ready 后串联 scenario |
| backend-review/001 Phase 5 真实 handler 未 land 导致本 plan Phase 5 跨 owner regression 跑不通 | Phase 5 跨 owner regression 用 fixture-backed transport 跑 frontend BDD；backend-review Phase 5 land 后再跑真实 backend regression；retrospective 记录 mock vs real 切换证据 |
| GeneratingScreen 5 阶段动画时长（700-1200ms）与真实 backend 生成时间（P95 < 12s）节奏不匹配，导致用户感知卡顿 | Phase 1 动画仅作为视觉反馈，不阻塞 nav；status='ready' 在任何阶段都立刻 nav；最坏情况下 5 阶段 ~5s + 轮询命中 ~3s = 8s 内完成；超过 12s 进入 timeout state |
| `useReportGenerationPoll` 在 visibility 切换时 race condition 导致重复 nav | Phase 1 nav 防抖 ref + 单次 fire-and-forget；onReady / onFailed callback 只触发一次；test fake clock 验证 |
| ReportScreen 5 detail tab 在 mobile (390×844) viewport 折叠为 Accordion 时与 desktop 切换 keyboard a11y 不一致 | Phase 5 Playwright mobile project 单独 verify ARIA tablist → ARIA accordion 转换；Vitest 模拟 viewport 切换测试 |
| 复练 CTA 路径 A 的 `replayItems:retryFocusTurnIds` 在 `getFeedbackReport` 响应中是 turn UUID 列表，但 backend-practice 创建 `goal='retry_current_round'` plan 需要 `source_report_id` 而非 turn id 列表 | 复练 CTA payload 仅作为 frontend nav route params；frontend-workspace-and-practice plan 002 或 future plan 在执行 `createPracticePlan` 时把 `sourceReportId` 设为当前 reportId；`replayItems` 作为本地 hint 用于 UI 显示，不传给 backend；本 plan handoff doc 标记此约束 |
| 路径 B `nextRoundId` 推断逻辑（默认 roundId + 1）在 backend 没有真实 round metadata 时不准确 | plan 001 默认走本地推断；如产品确认需要真实下一轮 metadata，先回 backend-targetjob owner 修订；不在本 plan 引入新 backend 调用 |
| i18n `errors.errorCode.*` 文案需要覆盖 backend-review D-8 所有 B1 `AI_*` enum；新增 B1 enum 时本 plan 文案漏更新 | Phase 2 i18n 测试断言 `errors.errorCode` 覆盖 B1 `AI_*` 当前全部 enum（用 generated B1 常量做 source of truth）；新增 enum 时 lint fail |
| Playwright pixel parity baseline 与 ui-design 原型微调时容易 drift | clean-checkout PASS 以 DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke 为硬 gate；`toHaveScreenshot` 仅在稳定 baseline 已提交或本 phase 明确更新 baseline 时启用；ui-design 微调时同步记录 baseline 更新原因 |
| 复练 CTA 在未登录场景的 useRequestAuth handoff 未正确恢复到可创建新 session 的 owner | Phase 4 测试覆盖 `replay_practice` pendingAction 恢复到 workspace auto-start + fresh session；扩展 `scenarios/p0-002-auth-pending-action-resume.test.tsx` 覆盖该路径 |
| **Polling timeout vs backend lease/retry 节奏不一致**（Open Question） | `useReportGenerationPoll` max attempts 30 ≈ 3.67 min；backend-review/001 D-13 lease_timeout 5 min + retry `min(2^attempt_count * 30s, 30min)` 可达 30 min × 5 attempts。当 backend 处于 retry 等待时 frontend 会先 timeout。当前默认：用户 retry 重启轮询（attempts 归零），不自动续轮；如 backend 在 generating → generating 持续 > 3.67 min 用户会反复看到 timeout 卡片。**待 design 决策**：是否把 max attempts 提到 ≥ 5 min（对齐 backend lease_timeout）？是否在 timeout 自动 retry 一次再显示卡片？是否新增 backend `?wait=true` long-poll API？plan 001 默认保留 3.67 min + manual retry；如需调整在 design 确认后修订本 risk row |
| **`nextRoundId` 推断算法字符串格式 与 backend round metadata 缺位**（Open Question） | spec D-5 + plan §4.2 推断 `nextRoundId = roundId + 1` 仅对 numeric / `round-${N}` 字符串成立；如果 backend `practice_plans.round_id` 是 uuid 推断无意义。**待 design 决策**：（A）约定 `roundId` 必须是可递增字符串（如 `round-${N}`）+ 推断 `round-${parseInt(N)+1}`；（B）path B CTA 在 backend-targetjob 提供真实 metadata 前 disabled / 不渲染；（C）path B 走临时 fallback：复用当前 roundId（"continue current round" 隐含语义）。plan 001 默认采用选项 A + 在 ContextStrip 显示 zh "第 2 轮" / en "Round 2" 文案；如 backend 实际格式不符合，design 需先回 backend-targetjob owner 修订 |
