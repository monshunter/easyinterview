# 001 — Report Screen and Generating Handoff Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 owner 前置 preflight

- [x] Phase 0 本计划定义的测试项全部通过：
  - `TestB2FeedbackReportSchemaHasErrorCode`（preflight：`openapi/openapi.yaml` 中 `FeedbackReport` schema 含 `errorCode: oneOf [ApiErrorCode, null]`；generated TS 等价字段存在）
  - `TestReportFailedFixtureVariantExists`（preflight：`openapi/fixtures/Reports/getFeedbackReport.json` `scenarios.report-failed.response.body.status === 'failed'` + `response.body.errorCode` 非 null）
  - `TestListTargetJobReportsEmptyFixtureVariantExists`（preflight：`openapi/fixtures/Reports/listTargetJobReports.json` `scenarios.empty.response.body.items=[]` + `response.body.pageInfo.hasMore=false`）
  - `TestReportNotFoundErrorCodeRegistered`（preflight：`shared/conventions.yaml#errors` 含 `REPORT_NOT_FOUND` + generated TS 等价常量）
  - 任一 fail → Phase 1 不启动；通过 bug-report / retrospective 通知 backend-review/001 owner

## Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

- [x] Phase 1 本计划定义的测试项全部通过：
  - `TestGeneratingScreenHappyPath`（mount → 5 阶段渲染 → 进度条 + 实时观察流；≥ 10 testid 命中）
  - `TestGeneratingScreenMissingReportIdRendersErrorState`（reportId 缺失立即渲染 ErrorState；不发请求）
  - `TestGeneratingScreenTimeoutStateShowsRetryCta`（max attempts → timeout state + retry CTA + backToWorkspace CTA）
  - `TestUseReportGenerationPoll7States`（idle / polling / ready / failed / timeout / error / paused）
  - `TestUseReportGenerationPollExponentialBackoff`（fake timer 验证 1.5s → 2.25s → 3.375s → ... 上限 8s；max attempts 30）
  - `TestUseReportGenerationPollVisibilityPauseResume`（visibility hidden → paused；visible → polling 继续 + 不重发当前请求）
  - `TestUseReportGenerationPollFocusEvents`（同 visibility 但 focus 事件）
  - `TestOnReadyCallbackNavReport`（status='ready' → onReady(report) → nav report 带 sessionId + reportId + passThrough）
  - `TestOnFailedCallbackNavReportWithReportStatus`（status='failed' → onFailed(errorCode) → nav report?reportStatus=failed&errorCode=...）
  - `TestReadyCallbackDebouncesNavReport`（多次 ready callback 只 nav 一次；fake timer 验证 handoffNavigatedRef）
  - `TestUseReportGenerationPollNoIdempotencyHeader`（反向断言 request init 不含 `Idempotency-Key` header）
  - `TestUseReportGenerationPollCrossUser404`（404 → 触发 onFailed callback with REPORT_NOT_FOUND）
  - `TestUseReportGenerationPollUnmountCancels`（unmount inflight 请求取消）
  - i18n zh/en 切换重绘验证（generating.* 命名空间 ≥ 20 keys）
  - 负向断言：不出现 `mistakesQueue` / 旧 `report-timeline` testid / 旧 `reportLayout` 字面量

## Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

- [x] Phase 2 本计划定义的测试项全部通过：
  - `TestReportScreenDispatchesFailureState`（params.reportStatus='failed' → ReportFailureState）
  - `TestReportScreenDispatchesMissingSession`（缺 sessionId → ReportMissingSessionState）
  - `TestReportScreenDispatchesDashboard`（正常 sessionId + 非 failed → ReportDashboard）
  - `TestReportScreenLoadingState`（useFeedbackReport state=loading → skeleton）
  - `TestReportScreenErrorState`（useFeedbackReport state=error → InlineError + retry）
  - `TestReportScreenNotFoundState`（useFeedbackReport state=notFound → ReportFailureState 渲染 cross-user 兜底）
  - `TestReportFailureStateRendersErrorCodeMatrix`（B1 `AI_*` enum 所有值映射到对应文案：AI_PROVIDER_TIMEOUT / AI_PROVIDER_SECRET_MISSING / AI_PROVIDER_CONFIG_INVALID / AI_OUTPUT_INVALID / UNKNOWN）
  - `TestReportFailureStateRendersNotFoundCopy`（404 / `REPORT_NOT_FOUND` 路径使用 `failureState.notFound.*` i18n key 而非 AI_* enum 文案；zh "未找到该报告" / en "Report not found"）
  - `TestUseFeedbackReportEncodesNotFoundDistinctly`（hook state='notFound' 渲染时 errorCode 显式为 'REPORT_NOT_FOUND' 而非 fallback 到 UNKNOWN）
  - `TestReportFailureStateRetryNavigatesGenerating`（CTA「重新生成」→ nav generating 带 sessionId + reportId）
  - `TestReportFailureStateBackToWorkspaceCta`（CTA「返回 workspace」→ nav workspace 带 targetJobId）
  - `TestReportMissingSessionNavigatesWorkspace`（CTA → nav workspace）
  - `TestReportMissingSessionNoApiCall`（缺 sessionId 不调用 `getFeedbackReport`）
  - `TestContextStripDisplaysAll7Fields`（sessionId / targetJob / round / resume / modality / practiceMode / hints；modality.text/voice + practiceMode.assisted/strict 文案）
  - `TestReportContextDataLoadsTargetJobAndResumeVersion`（通过 generated `getTargetJob` / `getResumeVersion` 渲染 target title/companyName + resume displayName）
  - `TestReportContextDataFallsBackToIds`（任一 ContextStrip label operation 失败时回退 targetJobId / resumeVersionId，不阻塞 dashboard）
  - `TestReportContextDataDoesNotReadRawBody`（不得读取 raw resume/JD/body 字段）
  - `TestUseFeedbackReport4States`（loading / data / error / notFound）
  - `TestUseFeedbackReportCrossUser404`（404 REPORT_NOT_FOUND → state='notFound'）
  - `TestUseFeedbackReportNoIdempotencyHeader`（反向断言）
  - `TestUseFeedbackReportUnmountCancels`（unmount inflight 请求取消）
  - i18n zh/en 切换重绘验证（report.* 命名空间 ≥ 40 keys）

## Phase 3: 5 detail tab 内容源级复刻

- [x] Phase 3 本计划定义的测试项全部通过：
  - `TestDetailSurfaceSwitches5Tabs`（点击 tab 触发按钮 → panel 切换；testid `report-detail-tab-{key}` + `report-detail-panel-{key}` 命中）
  - `TestDetailSurfaceAriaTablist`（ARIA tablist / tab / tabpanel role 正确）
  - `TestDetailSurfaceDefaultQuestions`（mount 时默认 `questions` panel 激活）
  - `TestDetailSurfaceCanSwitchToReadiness`（显式点击 readiness 后 readiness panel 激活）
  - `TestReadinessTabDial`（拨号盘 + 4 档色环 + 二级详情 JD 对齐 + 证据密度 + 下一档门槛）
  - `TestReadinessTier4LevelsZhAndEn`（4 档 readiness 文案 × zh/en 矩阵）
  - `TestDimensionsTabGrid`（二级维度卡片网格 + DimRow primitive 使用）
  - `TestDimensionStatus3StatesMapping`（strong / meets_bar / needs_work 三态文案与色调；不出现旧 acceptable / weak / developing / proficient）
  - `TestQuestionsTabListAndDetail`（题目列表侧栏 5 题 + 选中切换右侧 detail；当前题分析 5 字段）
  - `TestEvidenceTabRiskAndHighlight`（risk + highlight 两列；空数组 EmptyHint）
  - `TestNextTabPathAAndB`（路径 A vs 路径 B 对比 + CTA testid `report-next-cta-{a,b}`）
  - `TestDimRowStateMapping`（DimRow 三态 + 分数条 + 置信度）
  - `TestDashboardDimensionsCardRow`（维度卡片行 horizontal scroll + 优先级 + 复练重点 3 条）
  - `TestDashboardQuestionRecap`（5 题 quick state 卡片）
  - `TestDashboardIssuesAndHighlights`（issues + highlights 列表 + 空数组 EmptyHint）
  - 负向断言：不出现旧 5 档 readiness numeric / rubric score_levels label / 旧 `reportLayout` / `report_timeline`

## Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

- [x] Phase 4 本计划定义的测试项全部通过：
  - `TestPendingActionEncodeDecodeReplayPractice`（`replay_practice` PendingAction encode → decode round-trip 字段对等；params 含 sourceSessionId / replayItems / evidenceGaps / planId / targetJobId / jdId / resumeVersionId / roundId / mode / modality / practiceMode / practiceGoal / autoReplay）
  - `TestPendingActionReplayPracticeTypeAllowed`（type allowlist / discriminated union 包含 `replay_practice`；负向断言 URL params / localStorage 不含 raw text）
  - `TestReplayCtaPathA_AuthenticatedNavPractice`（已登录 → nav practice with retry_current_round payload + retryFocusTurnIds）
  - `TestReplayCtaPathA_UnauthenticatedUseRequestAuth`（未登录 → useRequestAuth({type:'replay_practice', route:'report', params:{...}})）
  - `TestReplayCtaPathA_PayloadIntegrity`（payload 字段完整：sourceSessionId / replayItems / evidenceGaps / planId / targetJobId / jdId / resumeVersionId / roundId / mode / modality / practiceMode / practiceGoal:'retry_current_round'）
  - `TestReplayCtaPathA_NoRawText`（负向断言 payload 不含 `answerText` / `questionText` / `hint` / `promptHash` / `modelId raw`）
  - `TestReplayCtaPathA_NoBackendCalls`（CTA 触发不调用 `getFeedbackReport` / `appendSessionEvent`）
  - `TestNextRoundCta_NavPractice`（路径 B → nav practice with next_round payload）
  - `TestNextRoundCta_NextRoundIdInference`（nextRoundId 推断逻辑测试）
  - `TestNextRoundCta_PayloadIntegrity`（payload 字段完整）
  - `TestRetryCtaNavGenerating`（ReportFailureState「重新生成」→ nav generating）
  - `TestBackToWorkspaceNavWorkspace`（ReportFailureState「返回 workspace」→ nav workspace）
  - `TestReadyCallbackDebouncesNavReport`（GeneratingScreen onReady 防抖；nav 调用次数 = 1）
  - `TestFailedCallbackNavReportWithStatus`（GeneratingScreen onFailed → nav report?reportStatus=failed）
  - `TestTimeoutStateNoAutoNav`（timeout state 不自动 nav；用户点 retry 重启轮询）
  - `TestCtaDisabledWhenDataNotReady`（report status='generating' 兜底时 CTA disabled）

## Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + 旧口径负向

- [x] Phase 5 本计划定义的测试项全部通过：
  - `pnpm --filter @easyinterview/frontend test` 全绿（覆盖 Phase 1-4 全部测试）
  - `pnpm --filter @easyinterview/frontend typecheck` 全绿
  - 扩展 `App.test.tsx` 添加 `generating-screen` / `report-dashboard` testid 命中
  - 扩展 `AppNormalize.test.tsx`（如有 `generating` / `report` route alias）
  - 扩展 `AppPendingAction.test.tsx` 添加 `replay_practice` pendingAction 在 report 屏自动恢复测试
  - 扩展 `scenarios/p0-002-auth-pending-action-resume.test.tsx` 添加 `replay_practice` resume path 验证
  - Playwright `tests/pixel-parity/generating.spec.ts` 全绿（desktop 1440×900 + mobile 390×844 + 8 主题 × dark；DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke）
  - Playwright `tests/pixel-parity/report.spec.ts` 全绿（同上 + 5 detail tab 切换 + 三态；`report` 默认 App chrome / TopBar 可见、不进入一级导航）
  - `TestReportNamespaceZhEnSync`（report.* zh/en 同步无缺漏）
  - `TestGeneratingNamespaceZhEnSync`（generating.* zh/en 同步无缺漏）
  - `TestErrorCodeI18nCoversAllAIErrors`（report.failureState.errorCode.* 覆盖 B1 `AI_*` enum 全集）
  - `TestReportFailureStateErrorCodeCoversReportNotFound`（显式断言 `report.failureState.errorCode.REPORT_NOT_FOUND` 与 `report.failureState.notFound.*` key 存在且 zh / en 同步；不归入 AI_* 通用映射）
  - `TestI18nKeyCountAtLeast60`（`report.*` + `generating.*` 合计 ≥ 60 keys）
  - `frontend/src/app/screens/report/__tests__/legacyNegative.test.ts` 全绿（不 import ui-design / window.EI_DATA / Voice 组件 / 不调 Practice operation；旧字面量 0 命中；含 `TestListTargetJobReportsNotInvokedInReportOrGenerating` mockTransport spy 反向断言）
  - `frontend/src/app/screens/generating/__tests__/legacyNegative.test.ts` 全绿（同上）
  - `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过
  - `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q` 通过
  - 4 个 P0 scenario 目录 setup → trigger → verify → cleanup 通过：p0-056 / p0-057 / p0-058 / p0-059
  - 跨 owner regression：scenario p0-044-047 重跑通过；backend `E2E.P0.052-055` 如已 implement 则通过
  - `frontend/src/app/screens/report/README.md` + `frontend/src/app/screens/generating/README.md` 创建完成
  - `make codegen-check` / `make validate-fixtures` / `make docs-check` / `git diff --check` 全绿
  - `pnpm build` 通过

## 全局收口

- [x] `pnpm --filter @easyinterview/frontend test`
- [x] `pnpm --filter @easyinterview/frontend typecheck`
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity`
- [x] `pnpm --filter @easyinterview/frontend build`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q`
- [x] 4 个 P0 scenario 执行通过
- [x] `make docs-check`
- [x] `git diff --check`

## 2026-05-16 L2 review regression evidence

- `ReportScreen.test.tsx` / `ReportMissingSessionState.test.tsx`: 覆盖 `report?sessionId=S` 缺 `reportId` 不发 `getFeedbackReport` 并渲染缺 reportId 错误态。
- `useReportGenerationPoll.test.tsx`: 覆盖 visibility/focus 恢复后沿用已调度 retry，不立即重复请求。
- `bootstrapRoute.test.ts`: 覆盖 `#route=generating/report` hash bootstrap，保证 pixel parity 从真实 route 启动。
- `ReportScreen.test.tsx`: 覆盖报告 header 标题包含 target job label，而非只显示 round label。
- `tests/pixel-parity/generating.spec.ts` + `tests/pixel-parity/report.spec.ts`: 覆盖 desktop / mobile DOM anchor、computed style、bounding box、no-overflow、主题切换与三态渲染。
- `frontend_report_dashboard_legacy.py` + pytest: 覆盖 prototype short CSS tokens 和旧口径 literal 的 scoped negative gate。
