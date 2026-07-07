# 001 — Report Screen and Generating Handoff Test Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 0 范围与覆盖矩阵映射

本 test-plan 把 plan §3.5 Coverage Matrix + plan §4 Phase 1-5 的实施步骤映射到具体测试包、文件、测试函数与命令。覆盖矩阵中的 BDD 行（场景级）由 `bdd-plan.md` 承担，本文件聚焦 Vitest + Playwright + scenario 单元 / 集成 / parity / contract / privacy 测试。

| 覆盖矩阵行（关键摘要） | 测试包 / 文件 | 测试函数（关键） | 测试命令 |
|------------------------|---------------|----------|----------|
| GeneratingScreen happy path（轮询 → ready → nav report） | `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx` + `frontend/src/app/screens/generating/__tests__/useReportGenerationPoll.test.ts` | `TestGeneratingScreenHappyPath` / `TestUseReportGenerationPollExponentialBackoff` / `TestOnReadyCallback` | `pnpm --filter @easyinterview/frontend test src/app/screens/generating` |
| GeneratingScreen 失败/超时/visibility | 同上 | `TestOnFailedCallbackNavReportWithReportStatus` / `TestMaxAttemptsTriggersTimeoutState` / `TestVisibilityPausesPolling` | 同上 |
| ReportScreen 三态分发 | `frontend/src/app/screens/report/__tests__/ReportScreen.test.tsx` | `TestReportScreenDispatchesFailureState` / `TestReportScreenDispatchesMissingSession` / `TestReportScreenDispatchesDashboard` | `pnpm --filter @easyinterview/frontend test src/app/screens/report` |
| ReportFailureState + ReportMissingSessionState | `frontend/src/app/screens/report/__tests__/ReportFailureState.test.tsx` + `ReportMissingSessionState.test.tsx` | `TestReportFailureStateRendersErrorCodeMatrix` / `TestReportFailureStateRetryNavigatesGenerating` / `TestReportMissingSessionNavigatesWorkspace` | 同上 |
| useFeedbackReport 4 态 + cross-user 404 | `frontend/src/app/screens/report/__tests__/useFeedbackReport.test.ts` | `TestUseFeedbackReport4States` / `TestUseFeedbackReportCrossUser404` / `TestUseFeedbackReportNoIdempotencyHeader` | 同上 |
| ContextStrip 7 字段 + label 数据源 | `frontend/src/app/screens/report/__tests__/ReportContextStrip.test.tsx` + `useReportContextData.test.ts` | `TestContextStripDisplaysAll7Fields` / `TestReportContextDataLoadsTargetJobAndResumeVersion` / `TestReportContextDataFallsBackToIds` / `TestReportContextDataDoesNotReadRawBody` | 同上 |
| DetailSurface 5 tab 切换 | `frontend/src/app/screens/report/__tests__/DetailSurface.test.tsx` | `TestDetailSurfaceSwitches5Tabs` / `TestDetailSurfaceAriaTablist` / `TestDetailSurfaceDefaultQuestions` / `TestDetailSurfaceCanSwitchToReadiness` | 同上 |
| 5 Detail Tab 内容 | `frontend/src/app/screens/report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx` | `TestReadinessTabDial` / `TestDimensionsTabGrid` / `TestQuestionsTabListAndDetail` / `TestEvidenceTabRiskAndHighlight` / `TestNextTabPathAAndB` | 同上 |
| 4 档 readiness 文案矩阵 | `frontend/src/app/screens/report/__tests__/tabs/ReadinessTab.test.tsx` | `TestReadinessTier4LevelsZhAndEn` | 同上 |
| 维度状态三态映射 | `frontend/src/app/screens/report/__tests__/tabs/DimensionsTab.test.tsx` + `frontend/src/app/screens/report/__tests__/DimRow.test.tsx` | `TestDimensionStatus3StatesMapping` | 同上 |
| 复练 CTA 路径 A | `frontend/src/app/screens/report/__tests__/ReplayCta.test.tsx` | `TestReplayCtaPathA_AuthenticatedAutoStartPractice` / `TestReplayCtaPathA_UnauthenticatedUseRequestAuth` / `TestReplayCtaPathA_PayloadIntegrity` / `TestReplayCtaPathA_NoRawText` | 同上 |
| 复练 CTA 路径 B | `frontend/src/app/screens/report/__tests__/ReplayCta.test.tsx` | `TestNextRoundCta_AutoStartPractice` / `TestNextRoundCta_NextRoundIdInference` | 同上 |
| ReportFailure handoff | `frontend/src/app/screens/report/__tests__/ReportFailureHandoff.test.tsx` | `TestRetryCtaNavGenerating` / `TestBackToWorkspaceNavWorkspace` | 同上 |
| GeneratingScreen handoff 防抖 | `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx` | `TestReadyCallbackDebouncesNavReport` / `TestFailedCallbackNavReportWithStatus` | `pnpm --filter @easyinterview/frontend test src/app/screens/generating` |
| Privacy 红线 | `frontend/src/app/screens/report/__tests__/reportPrivacy.test.tsx` + `frontend/src/app/screens/generating/__tests__/generatingPrivacy.test.tsx` | `TestNoRawTextInRouteParams` / `TestNoRawTextInConsoleLog` / `TestNoRawTextInLocalStorage` | `pnpm --filter @easyinterview/frontend test` |
| mockTransport spy | `frontend/src/api/__tests__/mockTransport.spy.test.ts`（扩展） | `TestMockTransportSpyRecordsReportRequestsWithoutBody` | 同上 |
| i18n 完整性 | `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts` | `TestReportNamespaceZhEnSync` / `TestGeneratingNamespaceZhEnSync` / `TestErrorCodeI18nCoversAllAIErrors` | 同上 |
| pixel parity desktop + mobile | `frontend/tests/pixel-parity/generating.spec.ts` + `frontend/tests/pixel-parity/report.spec.ts` | `report main + 5 tab + failure + missing-session × desktop / mobile × dark / customAccent` | `pnpm --filter @easyinterview/frontend test:pixel-parity` |
| Non-current negative grep | `frontend/src/app/screens/report/__tests__/nonCurrentNegative.test.ts` + `frontend/src/app/screens/generating/__tests__/nonCurrentNegative.test.ts` + `scripts/lint/frontend_report_dashboard_non_current.py`（新增）+ `scripts/lint/frontend_report_dashboard_non_current_test.py`（pytest） | `test_frontend_report_dashboard_non_current_includes_terms` / `TestNonCurrentNegativeGrep` | `pnpm --filter @easyinterview/frontend test src/app/screens/{report,generating}/__tests__/nonCurrentNegative.test.ts && python3 scripts/lint/frontend_report_dashboard_non_current.py --repo-root . --phase all && python3 -m pytest scripts/lint/frontend_report_dashboard_non_current_test.py -q` |
| Cross-owner regression | scenario rerun `p0-044-047`（frontend-workspace-and-practice/002）+ `p0-052-055`（backend-review/001 如已 implement） | scenario verify.sh chain | scenario rerun commands |
| Phase 0 preflight: B2 errorCode / report-failed fixture / empty fixture / REPORT_NOT_FOUND | `frontend/src/app/screens/report/__tests__/preflight.test.ts` | `TestB2FeedbackReportSchemaHasErrorCode` / `TestReportFailedFixtureVariantExists`（断言 `scenarios.report-failed.response.body.status/errorCode`） / `TestListTargetJobReportsEmptyFixtureVariantExists`（断言 `scenarios.empty.response.body.items/pageInfo`） / `TestReportNotFoundErrorCodeRegistered` | `pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts` |
| PendingAction `replay_practice` 注册 + round-trip | `frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts` | `TestPendingActionEncodeDecodeReplayPractice` / `TestPendingActionReplayPracticeTypeAllowed` | `pnpm --filter @easyinterview/frontend test src/app/auth` |
| ReportFailureState 404 / `REPORT_NOT_FOUND` 文案分支 | `frontend/src/app/screens/report/__tests__/ReportFailureStateNotFound.test.tsx` | `TestReportFailureStateRendersNotFoundCopy` / `TestUseFeedbackReportEncodesNotFoundDistinctly` | `pnpm --filter @easyinterview/frontend test src/app/screens/report` |
| `listTargetJobReports` 0 调用反向断言 | `frontend/src/app/screens/{report,generating}/__tests__/nonCurrentNegative.test.ts` | `TestListTargetJobReportsNotInvokedInReportOrGenerating` | `pnpm --filter @easyinterview/frontend test src/app/screens/{report,generating}` |

## Phase 0: 跨 owner 前置 preflight

- **测试目标**：断言 backend-review/001 已交付 `FeedbackReport.errorCode` 字段、`report-failed` + `empty` fixture variants、`REPORT_NOT_FOUND` 错误码与 generated TS 常量；持续作为 OpenAPI / fixture / generated client drift guard。
- **测试文件**：
  - `frontend/src/app/screens/report/__tests__/preflight.test.ts`（新增）：4 个 assertion
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts`
- **预期 Red / Green 证据**：
  - Red：任一 OpenAPI / fixture / generated client contract 缺失时 preflight test fail，message 指向具体 drift source
  - Green：backend-review/001 contract 与 generated client 一致时 preflight test 通过

## Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

- **测试目标**：GeneratingScreen 渲染 5 阶段进度动画 + 实时观察流；reportId 缺失立即渲染 ErrorState；轮询 hook 指数退避 + max attempts + visibility/focus 暂停-恢复；ready callback nav report；failed callback nav report?reportStatus=failed；timeout state 显示 retry CTA；request init 不含 Idempotency-Key。
- **测试文件**：
  - `frontend/src/app/screens/generating/GeneratingScreen.tsx`（新增；源级复刻）
  - `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx`（新增）：≥ 15 用例
  - `frontend/src/app/screens/generating/hooks/useReportGenerationPoll.ts`（新增；轮询 hook）
  - `frontend/src/app/screens/generating/__tests__/useReportGenerationPoll.test.ts`（新增）：7 态 + 退避 + 暂停 + 防抖
  - `frontend/src/app/screens/generating/components/`（新增 6 个 component）
  - `frontend/src/app/i18n/locales/{zh,en}.ts`（扩展 `generating.*` ≥ 20 keys）
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/generating`
  - `pnpm --filter @easyinterview/frontend typecheck`
- **预期 Red / Green 证据**：
  - Red：GeneratingScreen 未实现前 ≥ 15 测试 fail
  - Green：Phase 1 完成后命令通过；BDD `E2E.P0.056` GeneratingScreen 部分通过

## Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

- **测试目标**：ReportScreen 三态分发（failure/missing/dashboard）；useFeedbackReport 4 态（loading/data/error/notFound）；404 cross-user → ReportFailureState；5xx + retry；ContextStrip 7 字段显示；`useReportContextData` 通过 `getTargetJob` / `getResumeVersion` 加载人类可读 label 并在失败时回退 ID；4 Summary Cards 渲染。
- **测试文件**：
  - `frontend/src/app/screens/report/ReportScreen.tsx`（新增；顶层分发）
  - `frontend/src/app/screens/report/__tests__/ReportScreen.test.tsx`（新增）：3 态 + loading/data/error/notFound 4 态
  - `frontend/src/app/screens/report/components/`（新增 6 个 component）
  - `frontend/src/app/screens/report/__tests__/ReportFailureState.test.tsx`（新增）：errorCode 文案映射 + retry CTA
  - `frontend/src/app/screens/report/__tests__/ReportFailureStateNotFound.test.tsx`（新增）：`TestReportFailureStateRendersNotFoundCopy`（cross-user 404 / `REPORT_NOT_FOUND` 使用 `failureState.notFound.*` i18n key 而非 AI_* enum 文案）
  - `frontend/src/app/screens/report/__tests__/ReportMissingSessionState.test.tsx`（新增）：缺会话兜底
  - `frontend/src/app/screens/report/__tests__/ReportContextStrip.test.tsx`（新增）：7 字段
  - `frontend/src/app/screens/report/__tests__/useReportContextData.test.ts`（新增）：getTargetJob/getResumeVersion 成功 + fallback + raw body negative
  - `frontend/src/app/screens/report/__tests__/SummaryCards.test.tsx`（新增）：4 张卡片
  - `frontend/src/app/screens/report/hooks/useFeedbackReport.ts`（新增）+ test
  - `frontend/src/app/i18n/locales/{zh,en}.ts`（扩展 `report.*` ≥ 40 keys）
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/report`
- **预期 Red / Green 证据**：
  - Red：Phase 1 完成后 ReportScreen 三态测试 fail
  - Green：Phase 2 完成后命令通过

## Phase 3: 5 detail tab 内容源级复刻

- **测试目标**：DetailSurface 5 tab 切换 + ARIA tablist + 默认 `questions`；readiness 通过显式切换覆盖；5 个 tab 内容源级复刻；4 档 readiness 文案矩阵；维度状态三态映射；不出现旧 5 档 readiness / rubric score_levels label。
- **测试文件**：
  - `frontend/src/app/screens/report/components/DetailSurface.tsx`（新增）+ test
  - `frontend/src/app/screens/report/components/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.tsx`（新增 5 个 tab component）
  - `frontend/src/app/screens/report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx`（新增 5 个 test 文件）
  - `frontend/src/app/screens/report/__tests__/DimRow.test.tsx`（新增）：状态三态映射
  - `frontend/src/app/screens/report/__tests__/ReadinessTier4Levels.test.tsx`（新增；4 档矩阵）
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/report/components`
- **预期 Red / Green 证据**：
  - Red：tab component 未实现前测试 fail
  - Green：Phase 3 完成后命令通过；BDD `E2E.P0.056` ReportDashboard 部分通过

## Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

- **测试目标**：复练 CTA 路径 A + B 已登录 / 未登录经 workspace auto-start 创建 fresh practice session；payload 字段完整 + 隐私（无 raw text）；ReportFailureState retry → nav generating；GeneratingScreen ready/failed/timeout nav 防抖；data 未 ready 时 CTA disabled。
- **测试文件**：
  - `frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts`（新增）：`TestPendingActionEncodeDecodeReplayPractice` / `TestPendingActionReplayPracticeTypeAllowed` / 负向断言 URL params / localStorage 不含 raw text
  - `frontend/src/app/screens/report/__tests__/ReplayCta.test.tsx`（新增）：路径 A + 路径 B auto-start 子用例
  - `frontend/src/app/screens/report/__tests__/ReportFailureHandoff.test.tsx`（新增）：retry / backToWorkspace 2 子用例
  - `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx`（扩展）：ready / failed / timeout 三态 nav + 防抖
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/{report,generating} src/app/auth`
- **预期 Red / Green 证据**：
  - Red：CTA / handoff / PendingAction registration 未实现前测试 fail
  - Green：Phase 4 完成后命令通过；BDD `E2E.P0.056` 整链 + `E2E.P0.057` + `E2E.P0.058` 通过

## Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + 非当前输入负向

- **测试目标**：全 frontend 测试 + typecheck + pixel parity + 非当前输入负向 + i18n 完整性 + 跨 owner regression；scenario 4 个目录加挂。
- **测试文件**：
  - `frontend/tests/pixel-parity/generating.spec.ts`（新增）+ `report.spec.ts`（新增）：Playwright desktop + mobile + theme 切换 + DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke；`toHaveScreenshot` 仅在稳定 baseline 已提交或本 phase 明确更新 baseline 时启用
  - `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts`（新增）：zh/en 同步 + errorCode i18n 覆盖
  - `frontend/src/app/screens/report/__tests__/nonCurrentNegative.test.ts`（新增）+ `frontend/src/app/screens/generating/__tests__/nonCurrentNegative.test.ts`（新增）：scoped grep negative + `TestListTargetJobReportsNotInvokedInReportOrGenerating`（mockTransport spy 反向断言）
  - `scripts/lint/frontend_report_dashboard_non_current.py`（新增）+ `scripts/lint/frontend_report_dashboard_non_current_test.py`（pytest 新增）：scoped non-current grep + allowlist
  - `test/scenarios/e2e/p0-{056,057,058,059}-*/`（新增 4 个 scenario 目录，每个含 README + data + scripts）
  - `test/scenarios/e2e/INDEX.md`（更新 P0 表追加 4 行）
  - `frontend/src/app/scenarios/p0-002-auth-pending-action-resume.test.tsx`（扩展 `replay_practice` resume path）
  - `frontend/src/app/App.test.tsx`（扩展 `generating-screen` / `report-dashboard` testid 命中）
  - `frontend/src/app/AppNormalize.test.tsx`（扩展 `generating` / `report` alias，如有）
  - `frontend/src/app/AppPendingAction.test.tsx`（扩展 `replay_practice` pendingAction 恢复）
  - `frontend/src/app/screens/{report,generating}/README.md`（新增 handoff doc）
- **测试命令**：
  - `pnpm --filter @easyinterview/frontend test`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend test:pixel-parity`
  - `pnpm --filter @easyinterview/frontend build`
  - `make codegen-check`
  - `make validate-fixtures`
  - `python3 scripts/lint/frontend_report_dashboard_non_current.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/frontend_report_dashboard_non_current_test.py -q`
  - scenario 4 个目录 setup → trigger → verify → cleanup
  - 跨 owner regression：scenario `p0-044-047` 重跑 + `cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1`（如 backend-review/001 已 implement）
- **预期 Red / Green 证据**：
  - Red：Phase 4 完成后 Playwright DOM/style/smoke parity 与 scenario 目录尚不存在；稳定 screenshot baseline 如未提交不得作为 clean-checkout 阻塞项
  - Green：Phase 5 完成后所有命令通过；BDD `E2E.P0.059` 通过

## 全局收口测试命令

```bash
pnpm --filter @easyinterview/frontend test
pnpm --filter @easyinterview/frontend typecheck
pnpm --filter @easyinterview/frontend test:pixel-parity
pnpm --filter @easyinterview/frontend build
make codegen-check
make validate-fixtures
python3 scripts/lint/frontend_report_dashboard_non_current.py --repo-root . --phase all
python3 -m pytest scripts/lint/frontend_report_dashboard_non_current_test.py -q
make docs-check
git diff --check
```

不引入硬编码覆盖率门槛（observational only）。
