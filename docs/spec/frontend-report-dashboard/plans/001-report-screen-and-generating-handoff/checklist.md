# 001 — Report Screen and Generating Handoff Checklist

> **版本**: 1.18
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## 2026-07-10 Status Metadata Wording

- [x] `report-generating` fixture / OpenAPI / owner docs use queued/generating status metadata terminology rather than empty-report wording; verification: B2 codegen, frontend typecheck, focused report tests, context validation, doc sync and old wording grep.
- [x] Route fallback naming uses `RouteShellScreen` / route fallback shell terminology rather than the removed old component name; verification: scoped docs/code grep, frontend typecheck, focused route-shell tests and pixel parity.

## Phase 0: 跨 owner 前置 preflight

- [x] 0.1 新增 `frontend/src/app/screens/report/__tests__/preflight.test.ts`：断言 `openapi/openapi.yaml` 中 `FeedbackReport` schema 含 `errorCode` 字段（`oneOf: [ApiErrorCode, null]`）+ generated TS `FeedbackReport` interface 含可选 `errorCode` 属性；缺失 fail message 指向 backend-review/001 schema/error-code contract drift
- [x] 0.2 在 `preflight.test.ts` 追加：断言 `openapi/fixtures/Reports/getFeedbackReport.json` 的 `scenarios.report-failed.response.body.status === 'failed'` + `scenarios.report-failed.response.body.errorCode` 非 null；同时断言 `scenarios.default` + `scenarios.report-generating` 已存在；缺失 fail message 指向 report-failed fixture contract drift
- [x] 0.3 在 `preflight.test.ts` 追加：断言 `openapi/fixtures/Reports/listTargetJobReports.json` 的 `scenarios.empty.response.body.items === []` + `scenarios.empty.response.body.pageInfo.hasMore === false` + `scenarios.empty.response.body.pageInfo.nextCursor === null`；缺失 fail message 指向 empty fixture contract drift
- [x] 0.4 在 `preflight.test.ts` 追加：断言 `shared/conventions.yaml#errors` 含 `REPORT_NOT_FOUND` + `httpStatus: 404` + `retryable: false` + generated TS 等价常量；缺失 fail message 指向 shared/generated error contract drift
- [x] 0.5 收口 gate：`pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts` 全绿；如任一断言 fail，Phase 1 不启动，通过 bug-report / retrospective 通知 backend-review/001 owner

## Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

- [x] 1.1 新增 `frontend/src/app/screens/generating/GeneratingScreen.tsx` 按 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 源级复刻：页头（标题 + 副文案）+ 进度条（百分比 + phase indicator）+ 5 阶段列表（done/active/pending 圆圈 + 标签）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA + 「通知我」UI-only 按钮）；reportId 缺失立即渲染 `GeneratingErrorState`，不发请求
- [x] 1.2 新增 `frontend/src/app/screens/generating/hooks/useReportGenerationPoll.ts`：7 态（idle/polling/ready/failed/timeout/error/paused）+ 指数退避（初始 1.5s × 1.5 上限 8s）+ max attempts 30 + visibility/focus 暂停-恢复 + onReady/onFailed callback + request init 不含 `Idempotency-Key` header
- [x] 1.3 新增 `frontend/src/app/screens/generating/components/`：`HeaderHero.tsx` / `ProgressBar.tsx` / `PhaseList.tsx` / `LiveEvidenceStream.tsx` / `SlaHint.tsx` / `GeneratingErrorState.tsx`；每个从 `ui-design/src/screens-p0-complete.jsx` 同名片段复刻 DOM；testid `generating-{header,progress,phase-${idx},live-stream,evidence-${idx},sla-hint,notify-cta,error-{title,desc,retry,back-to-workspace}}`
- [x] 1.4 路由壳替换：在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `generating` → `<GeneratingScreen route={route} />`（替换 route fallback shell）；保持 `generating` 在 `NO_CHROME_ROUTES` 中隐藏 TopBar；`report` 仍由 route fallback shell 等待 Phase 2
- [x] 1.5 扩展 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `generating.*` 命名空间（≥ 20 keys：header.title / header.subtitle / phase.1 / phase.2 / phase.3 / phase.4 / phase.5 / progress.phaseN / evidence.streamLabel / sla.target / sla.notifyCta / errors.missingReportId / errors.timeout / errors.retry / errors.backToWorkspace 等）；`messages.ts` 类型聚合补齐；zh/en 同步无缺漏
- [x] 1.6 实现 `generating/__tests__/GeneratingScreen.test.tsx`：i18n zh/en 切换重绘 + ≥ 10 个 `generating-*` testid 存在 + reportId 缺失 → 不发请求 + 渲染 ErrorState + timeout 状态渲染 retry CTA + 负向断言不出现 `mistakesQueue` / `report-timeline` testid
- [x] 1.7 实现 `generating/__tests__/useReportGenerationPoll.test.ts`：7 态 + 指数退避节奏（fake timer 验证 1.5s / 2.25s / 3.375s / ... 上限 8s）+ max attempts 30 后 → state='timeout' + visibility/focus 暂停-恢复 + status='ready' callback + status='failed' callback + unmount 取消 inflight + 不含 `Idempotency-Key` header + cross-user 404 → state='notFound' callback
- [x] 1.8 兼容契约测试：扩展 `App.test.tsx` 添加 `generating-screen` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` route alias 处理（如必要）
- [x] 1.9 BDD-Gate: 验证 `E2E.P0.056` GeneratingScreen 部分通过（mount → 进度动画 → 轮询 → status='ready' nav report；Vitest scenario `frontend/src/app/scenarios/p0-056-generating-to-report-happy-path.test.tsx` 或对应 Playwright runner）

## Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

- [x] 2.1 新增 `frontend/src/app/screens/report/ReportScreen.tsx` 按 `ui-design/src/screen-report.jsx::ReportScreen` lines 1-44 源级复刻三态分发：`params.reportStatus === 'failed'` → `ReportFailureState`；缺 sessionId → `ReportMissingSessionState`；其他 → `ReportDashboard`
- [x] 2.2 新增 `frontend/src/app/screens/report/components/`：`ReportHeader.tsx`（标题 + 副标题 + 双 CTA `复练当前轮` + `进入下一轮`）/ `ReportContextStrip.tsx`（sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条）/ `SummaryCards.tsx`（4 张 ReportStatButton：准备度 / 维度 / 题目 / 下一步）/ `ReportFailureState.tsx`（lines 61-77 复刻，含 errorCode 文案映射 + CTA「重新生成」+ CTA「返回 workspace」）/ `ReportMissingSessionState.tsx`（lines 46-59 复刻，含 CTA「返回 workspace」）/ `ReportDashboard.tsx`（顶层组件，调用 `useFeedbackReport` + 渲染 Header + ContextStrip + 4 Summary Cards + Detail Surface 骨架态）
- [x] 2.3 新增 report 数据 hooks：`useFeedbackReport.ts` 单次拉取 `getFeedbackReport(reportId)`，4 态（loading/data/error/notFound），404 → state='notFound'，5xx → retry，request init 不含 `Idempotency-Key`；`useReportContextData.ts` 通过 generated `getTargetJob(targetJobId)` + `getResume(resumeId)` 只读 ContextStrip label，失败时回退 ID，不读取 raw resume/JD/body 字段
- [x] 2.4 路由壳替换：`App.tsx::renderRouteScreen` 中绑定 `report` → `<ReportScreen route={route} />`（替换 route fallback shell）；保持 `report` 不在 `NO_CHROME_ROUTES` 中，默认 App chrome / TopBar 可见，同时不加入一级导航
- [x] 2.5 扩展 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `report.*` 命名空间（≥ 40 keys：header.title / header.subtitle / cta.replay / cta.nextRound / context.session / context.job / context.round / context.resume / context.modality.text / context.modality.voice / context.practiceMode.assisted / context.practiceMode.strict / context.hints / summary.readiness / summary.dimensions / summary.questions / summary.next / readiness.tier.notReady / readiness.tier.needsPractice / readiness.tier.basicallyReady / readiness.tier.wellPrepared / failureState.title / failureState.desc / failureState.errorCode.AI_PROVIDER_TIMEOUT / failureState.errorCode.AI_PROVIDER_SECRET_MISSING / failureState.errorCode.AI_PROVIDER_CONFIG_INVALID / failureState.errorCode.AI_OUTPUT_INVALID / failureState.errorCode.REPORT_NOT_FOUND / failureState.errorCode.UNKNOWN / failureState.notFound.title / failureState.notFound.desc / failureState.retry / failureState.backToWorkspace / missingSession.title / missingSession.desc / missingSession.cta / loading / errors.network / errors.retry / errors.backToWorkspace 等；其中 `failureState.errorCode.REPORT_NOT_FOUND` 与 `failureState.notFound.*` 用于 cross-user 404 通用失败态，不混入 AI_* 通用文案）；`messages.ts` 类型聚合补齐
- [x] 2.6 实现 `report/__tests__/ReportScreen.test.tsx`：三态切换（reportStatus='failed' / 缺 sessionId / 正常）+ loading / data / error / notFound 四态 + ≥ 10 个 `report-*` testid 存在
- [x] 2.7 实现 `report/__tests__/ReportFailureState.test.tsx` / `ReportMissingSessionState.test.tsx`：分别测各自渲染 + errorCode 文案映射（B1 `AI_*` enum 全覆盖）+ CTA 行为（点击「重新生成」/「返回 workspace」分别触发 nav）+ `TestReportFailureStateRendersNotFoundCopy`：404 / `REPORT_NOT_FOUND` 路径使用 `failureState.notFound.*` i18n key 与 AI_* enum 文案显式区分
- [x] 2.8 实现 `report/__tests__/useFeedbackReport.test.ts`：4 态 + 404 cross-user → ReportFailureState + 5xx + retry + 不含 `Idempotency-Key`
- [x] 2.9 实现 `report/__tests__/useReportContextData.test.ts`：`getTargetJob` + `getResume` 成功显示人类可读 label；单 operation 失败回退对应 ID；不读取 raw resume/JD/body 字段；request init 不含写操作 header
- [x] 2.10 实现 `report/__tests__/ReportContextStrip.test.tsx`：所有 7 字段显示 + modality.text/voice + practiceMode.assisted/strict + hints count

## Phase 3: 5 detail tab 内容源级复刻

- [x] 3.1 新增 `frontend/src/app/screens/report/components/DetailSurface.tsx`：按 `screen-report.jsx::ReportDetailSurface` lines 162-174 + 311-516 源级复刻；5 个 tab 触发按钮 + panel 切换；testid `report-detail-tab-{readiness,dimensions,questions,evidence,next}` + `report-detail-panel-{key}`；ARIA tablist + tab + tabpanel 角色；默认激活 `questions`，readiness 通过显式切换覆盖
- [x] 3.2 实现 `report/components/tabs/ReadinessTab.tsx`（lines 335-357 复刻）：拨号盘（4 档色环）+ 二级详情（JD 对齐、证据密度、下一档门槛）；testid `report-readiness-dial / report-readiness-jd-align / report-readiness-evidence-density / report-readiness-next-threshold`
- [x] 3.3 实现 `report/components/tabs/DimensionsTab.tsx`（lines 360-382 复刻）：二级维度卡片网格（使用 `DimRow` primitive）；状态三态映射 strong / meets_bar / needs_work；testid `report-dimensions-grid / report-dim-card-{idx,name,state,score,confidence}`
- [x] 3.4 实现 `report/components/tabs/QuestionsTab.tsx`（lines 385-442 复刻）：题目列表侧栏（5 题，可选中切换）+ 当前题分析（有效点、缺口、建议框架、证据片段、下次追问）；testid `report-questions-list / report-questions-detail-{topic,good,missing,frame,evidence,follow-up}`
- [x] 3.5 实现 `report/components/tabs/EvidenceTab.tsx`（lines 445-467 复刻）：风险证据卡片列表 + 可复用亮点证据卡片列表；testid `report-evidence-risk-${idx}` + `report-evidence-highlight-${idx}`；空数组显示 EmptyHint
- [x] 3.6 实现 `report/components/tabs/NextTab.tsx`（lines 470-514 复刻）：路径 A vs 路径 B 对比展示 + 各自行动 CTA（CTA wire 在 Phase 4）；testid `report-next-path-{a,b}` + `report-next-cta-{a,b}` + `report-next-desc-{a,b}`
- [x] 3.7 在 `ReportDashboard.tsx` 主体补齐 lines 176-253 源级复刻：维度卡片行（horizontal scroll on mobile，使用 `DimRow`）+ 优先级（topPriority 单行）+ 复练重点（nextPractice 3 条 list）+ 题目回顾概览（5 题 quick state 卡片）+ 风险 issues + 亮点 highlights；testid `report-dim-row-${idx}` / `report-top-priority` / `report-next-practice-${idx}` / `report-perq-${idx}` / `report-issue-${idx}` / `report-highlight-${idx}`
- [x] 3.8 实现 `report/__tests__/DetailSurface.test.tsx`：5 个 tab 切换 + ARIA tablist + 默认 `questions` 激活 + 显式切换 readiness + 切换后其他 panel 不渲染（或 display:none）
- [x] 3.9 实现 `report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx`：每个测对应 testid + 数据驱动 + 边界（空 dimensions / 空 questions / 空 issues / 空 highlights 各自 EmptyHint）+ 4 档 readiness 文案矩阵 + 三档维度状态矩阵 + 负向断言不出现 5 档 readiness numeric / rubric score_levels label
- [x] 3.10 BDD-Gate: 验证 `E2E.P0.056` ReportDashboard 渲染部分通过（含 5 detail tab 切换 + 维度卡片 + 题目回顾 + 风险亮点）

## Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

- [x] 4.0 在 `frontend/src/app/auth/pendingAction.ts` 注册 `replay_practice` PendingAction type；`encodePendingAction` / `decodePendingActionRoute` 支持原 report route params round-trip；新增 `pendingActionReplayPractice.test.ts` 覆盖 type、回 report 与 URL/localStorage 无 raw text
- [x] 4.1 实现复练 CTA `goReplay()` 路径 A：唯一 Header CTA 组装 retry_current_round + source report/session + replayItems/evidenceGaps payload；已登录经共享 `startPracticeFromParams` 调 generated `createPracticePlan` / `startPracticeSession` 并直接进入 practice；未登录 pendingAction 回 report 重试
- [x] 4.2 实现复练 CTA `goNextRound()` 路径 B：唯一 Header CTA 组装 next_round + source report/session + nextRoundId payload；已登录创建 fresh plan/session 并直接进入 practice；未登录 pendingAction 回 report 重试
- [x] 4.3 完整 ReportFailureState handoff：`ReportFailureState.tsx` CTA「重新生成」点击 → `nav("generating", { sessionId, reportId, ...passThroughContext })`；CTA「返回 workspace」点击 → `nav("workspace", { targetJobId, jdId, planId, resumeId })`
- [x] 4.4 完整 GeneratingScreen handoff：`useReportGenerationPoll` 的 `onReady(report)` callback → `nav("report", { sessionId, reportId, ...passThrough })`；`onFailed(errorCode)` callback → `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode, ...passThrough })`；timeout state → 不自动 nav，用户点 retry 重启轮询；nav 调用必须防抖（handoffNavigatedRef）
- [x] 4.5 复练 CTA 数据未 ready 时禁用：report status='generating' 兜底（虽然不应进入但仍兜底）→ CTA disabled；不发 nav
- [x] 4.6 实现 `report/__tests__/ReplayCta.test.tsx`：路径 A 已登录创建/启动 fresh session 后直接进入 practice + 未登录 useRequestAuth 回 report；payload 字段完整；负向断言 raw text 不在 payload；CTA 点击后不重复调用 `getFeedbackReport`
- [x] 4.7 在 `report/__tests__/ReplayCta.test.tsx` 覆盖路径 B 同上；nextRoundId 推断逻辑测试
- [x] 4.8 实现 `report/__tests__/ReportFailureHandoff.test.tsx`：「重新生成」nav generating + 「返回 workspace」nav workspace；errorCode 不在 generating route params 中暴露 raw provider error
- [x] 4.9 扩展 `generating/__tests__/GeneratingScreen.test.tsx`：ready / failed / timeout 三态分别 nav + 防抖（多次 ready callback 只 nav 一次；fake timer 验证 1 次 nav）
- [x] 4.10 BDD-Gate: 验证 `E2E.P0.057` 通过（复练 CTA 路径 A + B 通过 generated client 创建/启动 fresh session 并直接进入 practice；未登录回 report）
- [x] 4.11 BDD-Gate: 验证 `E2E.P0.058` 通过（GeneratingScreen 轮询命中 `status='failed'` → nav failed report + ReportFailureState + ReportMissingSessionState + 跨用户 + 隐私 route params）
- [x] 4.12 BDD-Gate: 验证 `E2E.P0.056` 整链完整通过（含 GeneratingScreen mount → 进度动画 → 轮询 ready → nav report → ReportDashboard 渲染 → 5 detail tab 切换 → CTA wire 完整；Phase 1 + Phase 3 仅作局部断言，Phase 4 复练 CTA wire 完成后整链通过）

## Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + stale-contract negative

- [x] 5.1 `pnpm vitest run`（全 frontend 测试）+ `pnpm typecheck` 全绿；扩展 `App.test.tsx` 添加 `generating-screen` 与 `report-dashboard` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` / `report` route alias 处理；扩展 `pendingActionReplayPractice.test.ts` 添加 `replay_practice` pendingAction 回 report 的 round-trip；扩展 auth pending-action 场景覆盖 resume path
- [x] 5.2 新增 `frontend/tests/pixel-parity/generating.spec.ts`：desktop 1440×900 主屏、缺 reportId 错误态与 mobile 390×844 overflow；断言关键 DOM、主屏 bounding box 与每个状态的非空内存截图
- [x] 5.3 新增 `frontend/tests/pixel-parity/report.spec.ts`：desktop ReportDashboard、ReportMissingSessionState、ReportFailureState 与 mobile 390×844 overflow；断言关键 DOM、`report` 默认 App chrome / TopBar 可见与每个状态的非空内存截图
- [x] 5.4 派生 4 个 scenario 目录 `test/scenarios/e2e/p0-056-generating-to-report-happy-path/` / `p0-057-replay-cta-paths-a-and-b/` / `p0-058-report-failure-and-missing-session/` / `p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/`，每个含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x 可执行）；trigger 跑对应 Vitest 套件 / Playwright spec；verify 反查 testid / nav payload / 负向 grep / i18n 完整性
- [x] 5.5 更新 `test/scenarios/e2e/INDEX.md` 在 P0 表追加 4 行（E2E.P0.056 / 057 / 058 / 059）；状态 Ready，automated
- [x] 5.6 新增 `scripts/lint/frontend_report_dashboard_out_of_scope.py`：scoped grep 在 `frontend/src/app/screens/{report,generating}/`：`reportLayout` / `report_layout` / 5 档 readiness（`fully_prepared` 等字面量） / `readinessScore` / `readiness_score` numeric / `mistakes_queue` / `mistakesQueue` / `drill_builder` / `drillBuilder` / `growth_center` / `growthCenter` / `report_timeline` / `reportTimeline` / `report_form` / `reportForm` / 独立 `mistakes` route entry / `createPracticeVoiceTurn` / `getCompanyIntel` / `listTargetJobReports` 在实现代码中零出现；本 plan / BDD / test docs / spec §D-12 prohibition / preflight.test.ts 不属于实现范围；新增 `scripts/lint/frontend_report_dashboard_out_of_scope_test.py` 覆盖 `test_frontend_report_dashboard_out_of_scope_includes_terms` / `test_frontend_report_dashboard_out_of_scope_allows_negative_docs`
- [x] 5.7 新增 `frontend/src/app/screens/report/__tests__/outOfScopeNegative.test.ts` + `frontend/src/app/screens/generating/__tests__/outOfScopeNegative.test.ts`：grep negative 同上集合 + 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / prototype helper + 不 import `ui-design/src/screen-practice` practice DOM + 不调 Practice operation + `TestListTargetJobReportsNotInvokedInReportOrGenerating`（mockTransport spy 断言 `listTargetJobReports` 调用次数 = 0）
- [x] 5.8 新增 `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts`：断言 `report.*` 与 `generating.*` 命名空间 zh / en 同步无缺漏（key 集合相等）+ 新增 key ≥ 60 + 切换 locale 时所有 testid 文案重绘 + `report.failureState.errorCode.*` 覆盖 B1 `AI_*` enum 当前全部值（用 generated B1 常量做 source of truth）+ `TestReportFailureStateErrorCodeCoversReportNotFound` 显式断言 `report.failureState.errorCode.REPORT_NOT_FOUND` 与 `report.failureState.notFound.*` key 存在且 zh / en 同步
- [x] 5.9 跨 owner regression：在 Phase 5 收口阶段重跑：frontend-workspace-and-practice/002 BDD `E2E.P0.044-047`（保证未被破坏）；backend-review/001 BDD `E2E.P0.052-055` 在真实 handler 落地后作为 real-backend regression；backend-practice/002 BDD `E2E.P0.038-043` 必要时通过 cmd/api 重跑
- [x] 5.10 BDD-Gate: 验证 `E2E.P0.059` 通过（Playwright pixel parity + i18n + stale-contract negative）
- [x] 5.11 新增 `frontend/src/app/screens/report/README.md` + `frontend/src/app/screens/generating/README.md`：简明 handoff 段落，记录 001 新增 component / hook / nav 边界 / handoff 给 backend-review 与 frontend-workspace-and-practice 的边界；引用 D-1 ~ D-14 决策
- [x] 5.12 收口 gate：`pnpm vitest run` 全绿 + `pnpm typecheck` 全绿 + `pnpm test:pixel-parity` 全绿 + `pnpm build` 全绿 + `make codegen-check` 通过 + `make validate-fixtures` 通过（覆盖 backend-review/001 已交付 fixture variants）+ `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all` 通过 + `python3 -m pytest scripts/lint/frontend_report_dashboard_out_of_scope_test.py -q` 通过 + `make docs-check` 通过 + `git diff --check` 通过
- [x] 5.13 更新 `docs/spec/frontend-report-dashboard/plans/INDEX.md`：001 状态推进到 `completed`，并通过 sync-doc-index / docs-check 校验 Header 与 INDEX 一致

## 收口证据

- 2026-05-16 merge gate: `git fetch origin main` + `git merge --no-edit main`，结果为 `Already up to date.`，无冲突、无 merge commit。
- 2026-05-16 L2 code-review fixes: 补齐 `report?sessionId=S` 缺 `reportId` 的无请求错误态；修正 `useReportGenerationPoll` visibility/focus 恢复时重复请求；新增 hash bootstrap 使 Playwright parity route 能真实进入 `generating` / `report`；修复 report title target job 缺失；替换 prototype CSS short tokens；修复 390px mobile report overflow。
- `pnpm --filter @easyinterview/frontend test`：171 个 test files / 985 tests passed（首次出现 `HomeRecentMocks` 异步加载抖动，单测与全量重跑均通过）。
- `pnpm --filter @easyinterview/frontend typecheck`
- `pnpm --filter @easyinterview/frontend build`
- `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts`：14 passed。
- `make codegen-check`
- `make validate-fixtures`
- `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/frontend_report_dashboard_out_of_scope_test.py -q`：3 passed。
- `test/scenarios/e2e/p0-056-generating-to-report-happy-path/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-058-report-failure-and-missing-session/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/scripts/{setup,trigger,verify,cleanup}.sh`
- 2026-05-16 cross-owner regression: `E2E.P0.044` / `E2E.P0.045` / `E2E.P0.046` / `E2E.P0.047` setup → trigger → verify → cleanup 全部通过。
- 2026-05-23 L2 real-backend generated-client gate: P0.056-P0.059 trigger 前置 `frontendOwners.realApiMode.test.ts`；verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker；focused Vitest `frontendOwners.realApiMode.test.ts` PASS。
- `make docs-check`
- `git diff --check`

## Phase 6: D-19 report CTA single-point convergence

- [x] 6.1 next tab 删除重复 CTA 按钮；验证: focused Vitest 断言 `report-detail-panel-next` 渲染路径 A/B 说明 + 复练清单 + footer 引导文案，`report-next-cta-a` / `report-next-cta-b` 在 NextTab 渲染 DOM 0 命中；NextTab 不再接收 `onReplay`/`onNextRound` props；DetailSurface 去传参
  <!-- verified: 2026-06-13 command="pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/DetailSurface.test.tsx" evidence="Red: next-tab CTA 测试改为断言 report-next-cta-a/b 不存在 + path-a/b-footer 存在 + Header report-replay-cta/report-next-cta 仍在，先失败；Green: NextTab 删两 CTA 按钮与 onReplay/onNextRound props，新增 report-next-path-{a,b}-footer，DetailSurface 去 replayHandlers prop；10/10 通过" -->
- [x] 6.2 题目回顾 `加入本轮复练` 改本地标记；验证: focused Vitest 断言点击 `report-questions-add-to-replay` 不触发 `navigate`/`useRequestAuth`，仅 toggle 当前题目本地标记（文案 `加入本轮复练` ↔ `已加入本轮复练`），切换不同题目各自独立标记；新增 i18n `report.questions.detail.addedToReplay`（zh/en）
  <!-- verified: 2026-06-13 command="pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/DetailSurface.test.tsx src/app/i18n" evidence="QuestionsTab 改 per-question markedForReplay 本地 state + toggleActiveMarked（对照原型 replayQueued/toggleQueued），data-marked toggle、不 nav；断言 data-marked false->true->false、route 不变、report-dashboard 仍在；i18n addedToReplay zh/en 新增；62 测试通过" -->
- [x] 6.3 Phase 6 回归与负向 gate；验证: `report-next-cta-a`/`report-next-cta-b` 源码与渲染 0 命中（负向断言除外）；`pnpm --filter @easyinterview/frontend typecheck/test/build` 通过；report + topbar pixel parity 通过；`frontend_report_dashboard_out_of_scope` lint 通过；`make docs-check` + `sync-doc-index --check` 零漂移
  <!-- verified: 2026-06-13 command="pnpm --filter @easyinterview/frontend typecheck; pnpm --filter @easyinterview/frontend test; pnpm exec playwright test tests/pixel-parity/report.spec.ts; python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all; make docs-check" evidence="report-next-cta-a/b 仅存于负向断言；typecheck OK；vitest 1077/1077；report pixel parity 8 passed；out-of-scope lint OK；docs-check OK；sync-doc-index 零漂移" -->

## Phase 7: Generating hook test runtime isolation

- [x] 7.1 `useReportGenerationPoll` 单元测试直接注入 `AppRuntimeContext`，删除 fake client 中无关的 runtime-config/auth 方法（验证：focused 11 tests 无 React act warning、full frontend test/typecheck/build、owner context/docs gates）
  <!-- verified: 2026-07-10 method=generating-hook-test-runtime-isolation evidence="Focused red reproduced two AppRuntimeProvider act warnings after the synchronous missing-report assertion. Direct AppRuntimeContext injection removed unrelated runtime/auth effects and four fake methods. Focused 11/11 and generating/report 65/65 pass without warnings; frontend build and owner lint/context pass. Full frontend 137 files/829 tests pass and the generating hook file is absent from the remaining warning list." -->

## Phase 8: P0.059 browser evidence reconciliation

- [x] 8.1 Add an owner preflight that rejects theme-matrix and persistent-baseline claims and requires executable non-empty screenshot assertions in both pixel-parity specs.
  <!-- verified: 2026-07-10 method=p0059-browser-evidence-reconciliation evidence="Red failed on the first stale 8-theme owner claim. Green preflight passes 7/7 after scanning six owner documents, both Playwright sources and the P0.059 trigger." -->
- [x] 8.2 Add in-memory screenshot assertions to every covered generating/report browser state without creating baseline files or changing product UI.
  <!-- verified: 2026-07-10 method=p0059-browser-evidence-reconciliation evidence="Generating has three screenshot calls and Report has four; each uses page.screenshot() and asserts byteLength > 0. The report mobile overflow threshold was tightened from 420 to the 390px viewport and passes." -->
- [x] 8.3 Reconcile plan, test, BDD and P0.059 assets with the actual desktop/mobile DOM, state, geometry, overflow and screenshot-buffer evidence.
  <!-- verified: 2026-07-10 method=p0059-browser-evidence-reconciliation evidence="Removed unsupported theme-loop, computed-style, tab-switch and image-baseline claims. P0.059 README/seed/expected and trigger/verify now describe and enforce the current preflight, build, lint and browser evidence." -->
- [x] 8.4 Run focused Vitest, both Playwright specs, P0.059 setup/trigger/verify/cleanup, owner context, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0059-browser-evidence-reconciliation evidence="Focused preflight passes 7 tests; focused Playwright and scenario Playwright each pass 14 executions; P0.059 setup/trigger/verify/cleanup pass with 14 Vitest assertions, lint, pytest and build. Scenario output was cleaned; no environment restart or data cleanup occurred." -->

## Phase 9: P0.057 direct-start contract reconciliation

- [x] 9.1 Add an owner preflight that rejects workspace mount side-effect replay contracts and requires the current generated-client direct-start flow plus P0.057 wiring.
  <!-- verified: 2026-07-10 method=p0057-direct-start-contract-reconciliation evidence="Red preflight failed on the obsolete route-side-effect term set. Green passes 8 tests after checking the active spec, six plan artifacts, P0.057 assets, useReplayCtaHandlers, startPracticeFromParams and trigger wiring." -->
- [x] 9.2 Reconcile the active spec and plan/test/BDD documents with direct report-owner session creation/start and signed-out return to report.
  <!-- verified: 2026-07-10 method=p0057-direct-start-contract-reconciliation evidence="Spec v1.9 and plan v1.12 now match UI truth and frontend-workspace D-9: generated createPracticePlan/startPracticeSession, fresh session, direct practice navigation, and signed-out report return. Operation matrices include both write operations." -->
- [x] 9.3 Replace P0.057 expected/verify claims with executable `startPracticeFromParams -> practice` and pending-action `route=report` assertions; remove obsolete workspace checks.
  <!-- verified: 2026-07-10 method=p0057-direct-start-contract-reconciliation evidence="Initial scenario trigger passed 10 tests but verify failed on the obsolete route-side-effect marker. Green verify now checks direct-start test markers, shared helper use, practice navigation, report pending action, generated create/start calls and privacy negatives." -->
- [x] 9.4 Run focused tests, P0.057 setup/trigger/verify/cleanup, owner context, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0057-direct-start-contract-reconciliation evidence="Focused preflight passes 8 tests; replay/pending focused suite passes 10; P0.057 setup/trigger/verify/cleanup passes with 18 trigger tests; frontend typecheck passes. Scenario output was cleaned with no environment restart or data cleanup." -->

## Phase 10: P0.056 focused-runner evidence reconciliation

- [x] 10.1 Add a scenario/BDD preflight that rejects unsupported integrated-journey claims and requires all five trigger and verify markers.
  <!-- verified: 2026-07-10 method=p0056-focused-runner-evidence-reconciliation evidence="Red preflight failed on the expanded end-to-end claim. Green passes 9 tests and rejects integrated journey, transcript, pre-flatten Resume, fixed cross-file polling and theme claims while checking all five trigger/verify paths." -->
- [x] 10.2 Reconcile P0.056 README/seed/expected and owner BDD artifacts with the focused preflight/poller/screen/detail test evidence and flat Resume contract.
  <!-- verified: 2026-07-10 method=p0056-focused-runner-evidence-reconciliation evidence="README/seed/expected and BDD now describe preflight, poll hook, GeneratingScreen, ReportScreen and DetailSurface as independent deterministic gates. getResume replaces getResumeVersion, and fixed request counts/theme/transcript claims are removed." -->
- [x] 10.3 Add the missing `useReportGenerationPoll.test.tsx` verify marker and keep real-mode configuration proof distinct from fixture-backed owner tests.
  <!-- verified: 2026-07-10 method=p0056-focused-runner-evidence-reconciliation evidence="verify.sh now requires every focused file marker before static testid, lint and listTargetJobReports checks. README explicitly separates the real-mode bootstrap contract from deterministic test clients." -->
- [x] 10.4 Run focused preflight, P0.056 setup/trigger/verify/cleanup, owner contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0056-focused-runner-evidence-reconciliation evidence="Baseline P0.056 passed 43 tests despite expanded prose. Green setup/trigger/verify/cleanup passes with 5 files and 44 tests, scoped lint OK, and scenario output removed; no environment restart or data cleanup occurred." -->

## Phase 11: P0.058 failure-contract evidence reconciliation

- [x] 11.1 Add a scenario/BDD preflight that rejects unsupported GeneratingScreen UI, repeated-timeout and broad privacy claims and requires all six runner markers.
  <!-- verified: 2026-07-10 method=p0058-focused-failure-evidence-reconciliation evidence="Red preflight failed on the unsupported GeneratingScreen timeout UI claim. Green passes 10 tests and rejects repeated-timeout, live-backend and broad URL/storage/telemetry claims while checking all six trigger/verify markers." -->
- [x] 11.2 Reconcile P0.058 README/seed/expected and owner BDD artifacts with hook/component/route-state focused evidence.
  <!-- verified: 2026-07-10 method=p0058-focused-failure-evidence-reconciliation evidence="Scenario and BDD now separate failure/missing components, report hook/route and poll-hook evidence. Timeout stops at hook state; typed copy and route-state rendering remain explicit." -->
- [x] 11.3 Add preflight, ReportScreen and poll-hook verify markers while preserving typed error-copy checks.
  <!-- verified: 2026-07-10 method=p0058-focused-failure-evidence-reconciliation evidence="trigger now runs six focused files; verify requires each marker plus AI_PROVIDER_TIMEOUT and failureState.notFound.title keys. The prior verify observed only three of five files." -->
- [x] 11.4 Run focused preflight, P0.058 setup/trigger/verify/cleanup, owner contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0058-focused-failure-evidence-reconciliation evidence="Baseline P0.058 passed 5 files/29 tests despite expanded claims. Green setup/trigger/verify/cleanup passes with 6 files/39 tests and removes scenario output; no environment restart or data cleanup occurred." -->

## Phase 12: active visual contract reconciliation

- [x] 12.1 Extend the browser-evidence preflight to scan the active spec plus all six plan artifacts and reject visual or responsive claims not executed by P0.059.
  <!-- verified: 2026-07-10 method=frontend-report-active-visual-contract-reconciliation evidence="Focused red ran 10 preflight tests and failed only the Phase 8 browser-evidence case on the active spec theme-switching claim, proving the new seven-artifact guard observes the current drift before documentation changes." -->
- [x] 12.2 Reconcile active spec C-12, owner coverage/risk rows, test/BDD artifacts and P0.059 assets with the seven current browser states and their exact evidence.
  <!-- verified: 2026-07-10 method=frontend-report-active-visual-contract-reconciliation evidence="Spec v1.10, plan/test/BDD artifacts and P0.059 claims now enumerate the current seven states, explicit DOM/root/TopBar/390px evidence and per-state in-memory screenshots. Unsupported visual/responsive pattern scan is empty and the focused preflight passes 10/10." -->
- [x] 12.3 Run focused preflight, both generating/report Playwright specs, P0.059 setup/trigger/verify/cleanup, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=frontend-report-active-visual-contract-reconciliation evidence="Focused preflight passes 10/10 and focused Playwright passes 14/14. P0.059 setup/trigger/verify/cleanup passes real-mode 1/1, owner/i18n/negative 17/17, lint, pytest 3/3, build and Playwright 14/14. Both contexts, sync-doc-index, docs-check, diff-check and pruning surface pass with real_residuals=0; scenario output was removed with no environment restart or data cleanup." -->

## Phase 13: unconsumed report error helper removal

- [x] 13.1 Add a scoped source-surface RED assertion for the report error predicate with zero repository consumers.
  <!-- verified: 2026-07-10 method=unconsumed-report-error-helper-source-red evidence="Focused report source survey failed with exactly two readiness.ts offenders: isAiErrorCode and FAILURE_AI_ERROR_KEYS; the existing out-of-scope survey remained green." -->
- [x] 13.2 Delete `isAiErrorCode` and `FAILURE_AI_ERROR_KEYS` without changing `failureErrorCodeKey` or failure UI behavior.
  <!-- verified: 2026-07-10 method=unconsumed-report-error-helper-removal evidence="Deleted both isolated symbols with no replacement. Source/failure/missing-session tests pass 3 files/8 tests and scoped non-test symbol inventory is empty; failureErrorCodeKey and FAILURE_LABEL_BY_CODE are unchanged." -->
- [x] 13.3 Run focused report failure tests, typecheck, symbol inventory, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=unconsumed-report-error-helper-removal evidence="Focused source/failure tests pass 3 files/8 tests; report/generating owner passes 13 files/70 tests; typecheck and scoped symbol inventory pass. Report/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 14: typed i18n key mappings

- [x] 14.1 Replace Report detail-tab and missing-state dynamic message-key construction plus `as never` casts with explicit typed `MessageKey` mappings; verify focused report/i18n tests, locale reachability, typecheck/build, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=report-typed-message-key-maps evidence="All 13 active Report tab/missing-state keys are explicit MessageKey map values; scoped dynamic message-key cast search is zero. Focused Report/i18n tests, owner directory and full frontend suites, typecheck/build, P0.059, Report/product contexts and docs/diff/pruning gates pass with unchanged visible copy and navigation." -->

## Phase 15: report detail prototype call-surface pruning

- [x] 15.1 Add a Report detail prop-consumption contract and prove RED while `ReportDetailSurface` and its caller still carry unread `nav`.
  <!-- verified: 2026-07-10 method=report-detail-call-surface-red evidence="UI contract ran 43 tests: the new report-detail dependency contract failed on the existing ReportDetailSurface.nav parameter while the prior 42 tests passed; retained assertions pin parent ReportDashboard navigation, tab state and question selection." -->
- [x] 15.2 Delete the unread child prop and caller argument; verify Babel inventory reports zero unread `ReportDetailSurface` props while parent navigation and detail state remain intact.
  <!-- verified: 2026-07-10 method=report-detail-call-surface-green evidence="Removed only ReportDetailSurface.nav and its matching child argument. UI contract passes 43/43; Babel binding inventory reports detailUnread=[] while source assertions retain ReportDashboard.nav, setDetail and setActiveQuestion." -->
- [x] 15.3 Run UI contract, focused Report, P0.056/P0.059, static-browser detail-tab/question smoke, full frontend, typecheck/build, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=report-detail-regression-closeout evidence="UI contract passes 43/43 and focused Report passes 3 files/28 tests. P0.056 setup/trigger/verify/cleanup passes real-mode 1 plus owner 45 and scoped lint. P0.059 passes real-mode 1, Vitest 18, pytest 3, lint, build and Playwright 14. Static browser switches all five detail tabs, selects Q1 and retains replay marker behavior with no errors and 200 requests. Full frontend passes 137 files/841 tests and typecheck passes. Both owner contexts and diff/pruning gates pass with real_residuals=0. No scenario environment restart or data cleanup occurred." -->
