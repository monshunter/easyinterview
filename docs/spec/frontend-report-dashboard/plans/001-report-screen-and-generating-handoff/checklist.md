# 001 — Report Screen and Generating Handoff Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)

## Phase 0: 跨 owner 前置 preflight

- [x] 0.1 新增 `frontend/src/app/screens/report/__tests__/preflight.test.ts`：断言 `openapi/openapi.yaml` 中 `FeedbackReport` schema 含 `errorCode` 字段（`oneOf: [ApiErrorCode, null]`）+ generated TS `FeedbackReport` interface 含可选 `errorCode` 属性；缺失 fail message 指向 backend-review/001 Phase 0.2 + 0.4
- [x] 0.2 在 `preflight.test.ts` 追加：断言 `openapi/fixtures/Reports/getFeedbackReport.json` 的 `scenarios.report-failed.response.body.status === 'failed'` + `scenarios.report-failed.response.body.errorCode` 非 null；同时断言 `scenarios.default` + `scenarios.report-generating` 已存在；缺失 fail message 指向 backend-review/001 Phase 0.4
- [x] 0.3 在 `preflight.test.ts` 追加：断言 `openapi/fixtures/Reports/listTargetJobReports.json` 的 `scenarios.empty.response.body.items === []` + `scenarios.empty.response.body.pageInfo.hasMore === false` + `scenarios.empty.response.body.pageInfo.nextCursor === null`；缺失 fail message 指向 backend-review/001 Phase 0.4
- [x] 0.4 在 `preflight.test.ts` 追加：断言 `shared/conventions.yaml#errors` 含 `REPORT_NOT_FOUND` + `httpStatus: 404` + `retryable: false` + generated TS 等价常量；缺失 fail message 指向 backend-review/001 Phase 0.1
- [x] 0.5 收口 gate：`pnpm --filter @easyinterview/frontend test src/app/screens/report/__tests__/preflight.test.ts` 全绿；如任一断言 fail，Phase 1 不启动，通过 bug-report / retrospective 通知 backend-review/001 owner

## Phase 1: GeneratingScreen 源级复刻 + useReportGenerationPoll hook + 状态分支

- [x] 1.1 新增 `frontend/src/app/screens/generating/GeneratingScreen.tsx` 按 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` lines 269-399 源级复刻：页头（标题 + 副文案）+ 进度条（百分比 + phase indicator）+ 5 阶段列表（done/active/pending 圆圈 + 标签）+ 实时观察流（fade-in evidence snippets）+ 底部提示（P95 SLA + 「通知我」UI-only 按钮）；reportId 缺失立即渲染 `GeneratingErrorState`，不发请求
- [x] 1.2 新增 `frontend/src/app/screens/generating/hooks/useReportGenerationPoll.ts`：7 态（idle/polling/ready/failed/timeout/error/paused）+ 指数退避（初始 1.5s × 1.5 上限 8s）+ max attempts 30 + visibility/focus 暂停-恢复 + onReady/onFailed callback + request init 不含 `Idempotency-Key` header
- [x] 1.3 新增 `frontend/src/app/screens/generating/components/`：`HeaderHero.tsx` / `ProgressBar.tsx` / `PhaseList.tsx` / `LiveEvidenceStream.tsx` / `SlaHint.tsx` / `GeneratingErrorState.tsx`；每个从 `ui-design/src/screens-p0-complete.jsx` 同名片段复刻 DOM；testid `generating-{header,progress,phase-${idx},live-stream,evidence-${idx},sla-hint,notify-cta,error-{title,desc,retry,back-to-workspace}}`
- [x] 1.4 路由壳替换：在 `frontend/src/app/App.tsx::renderRouteScreen` 中绑定 `generating` → `<GeneratingScreen route={route} />`（替换 D1 `PlaceholderScreen`）；保持 `generating` 在 `NO_CHROME_ROUTES` 中隐藏 TopBar；`report` 仍渲染 `PlaceholderScreen` 等待 Phase 2
- [x] 1.5 扩展 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `generating.*` 命名空间（≥ 20 keys：header.title / header.subtitle / phase.1 / phase.2 / phase.3 / phase.4 / phase.5 / progress.phaseN / evidence.streamLabel / sla.target / sla.notifyCta / errors.missingReportId / errors.timeout / errors.retry / errors.backToWorkspace 等）；`messages.ts` 类型聚合补齐；zh/en 同步无缺漏
- [x] 1.6 实现 `generating/__tests__/GeneratingScreen.test.tsx`：i18n zh/en 切换重绘 + ≥ 10 个 `generating-*` testid 存在 + reportId 缺失 → 不发请求 + 渲染 ErrorState + timeout 状态渲染 retry CTA + 负向断言不出现 `mistakesQueue` / 旧 `report-timeline` testid
- [x] 1.7 实现 `generating/__tests__/useReportGenerationPoll.test.ts`：7 态 + 指数退避节奏（fake timer 验证 1.5s / 2.25s / 3.375s / ... 上限 8s）+ max attempts 30 后 → state='timeout' + visibility/focus 暂停-恢复 + status='ready' callback + status='failed' callback + unmount 取消 inflight + 不含 `Idempotency-Key` header + cross-user 404 → state='notFound' callback
- [x] 1.8 兼容契约测试：扩展 `App.test.tsx` 添加 `generating-screen` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` route alias 处理（如必要）
- [x] 1.9 BDD-Gate: 验证 `E2E.P0.056` GeneratingScreen 部分通过（mount → 进度动画 → 轮询 → status='ready' nav report；Vitest scenario `frontend/src/app/scenarios/p0-056-generating-to-report-happy-path.test.tsx` 或对应 Playwright runner）

## Phase 2: ReportScreen 静态壳源级复刻 + 三态分支 + ContextStrip + Summary Cards

- [x] 2.1 新增 `frontend/src/app/screens/report/ReportScreen.tsx` 按 `ui-design/src/screen-report.jsx::ReportScreen` lines 1-44 源级复刻三态分发：`params.reportStatus === 'failed'` → `ReportFailureState`；缺 sessionId → `ReportMissingSessionState`；其他 → `ReportDashboard`
- [x] 2.2 新增 `frontend/src/app/screens/report/components/`：`ReportHeader.tsx`（标题 + 副标题 + 双 CTA `复练当前轮` + `进入下一轮`）/ `ReportContextStrip.tsx`（sessionId / targetJob / round / resume / modality / practiceMode / hints 显示条）/ `SummaryCards.tsx`（4 张 ReportStatButton：准备度 / 维度 / 题目 / 下一步）/ `ReportFailureState.tsx`（lines 61-77 复刻，含 errorCode 文案映射 + CTA「重新生成」+ CTA「返回 workspace」）/ `ReportMissingSessionState.tsx`（lines 46-59 复刻，含 CTA「返回 workspace」）/ `ReportDashboard.tsx`（顶层组件，调用 `useFeedbackReport` + 渲染 Header + ContextStrip + 4 Summary Cards + Detail Surface 骨架占位）
- [x] 2.3 新增 report 数据 hooks：`useFeedbackReport.ts` 单次拉取 `getFeedbackReport(reportId)`，4 态（loading/data/error/notFound），404 → state='notFound'，5xx → retry，request init 不含 `Idempotency-Key`；`useReportContextData.ts` 通过 generated `getTargetJob(targetJobId)` + `getResumeVersion(resumeVersionId)` 只读 ContextStrip label，失败时回退 ID，不读取 raw resume/JD/body 字段
- [x] 2.4 路由壳替换：`App.tsx::renderRouteScreen` 中绑定 `report` → `<ReportScreen route={route} />`（替换 D1 `PlaceholderScreen`）；保持 `report` 不在 `NO_CHROME_ROUTES` 中，默认 App chrome / TopBar 可见，同时不加入一级导航
- [x] 2.5 扩展 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `report.*` 命名空间（≥ 40 keys：header.title / header.subtitle / cta.replay / cta.nextRound / context.session / context.job / context.round / context.resume / context.modality.text / context.modality.voice / context.practiceMode.assisted / context.practiceMode.strict / context.hints / summary.readiness / summary.dimensions / summary.questions / summary.next / readiness.tier.notReady / readiness.tier.needsPractice / readiness.tier.basicallyReady / readiness.tier.wellPrepared / failureState.title / failureState.desc / failureState.errorCode.AI_PROVIDER_TIMEOUT / failureState.errorCode.AI_PROVIDER_SECRET_MISSING / failureState.errorCode.AI_PROVIDER_CONFIG_INVALID / failureState.errorCode.AI_OUTPUT_INVALID / failureState.errorCode.REPORT_NOT_FOUND / failureState.errorCode.UNKNOWN / failureState.notFound.title / failureState.notFound.desc / failureState.retry / failureState.backToWorkspace / missingSession.title / missingSession.desc / missingSession.cta / loading / errors.network / errors.retry / errors.backToWorkspace 等；其中 `failureState.errorCode.REPORT_NOT_FOUND` 与 `failureState.notFound.*` 用于 cross-user 404 通用失败态，不混入 AI_* 通用文案）；`messages.ts` 类型聚合补齐
- [x] 2.6 实现 `report/__tests__/ReportScreen.test.tsx`：三态切换（reportStatus='failed' / 缺 sessionId / 正常）+ loading / data / error / notFound 四态 + ≥ 10 个 `report-*` testid 存在
- [x] 2.7 实现 `report/__tests__/ReportFailureState.test.tsx` / `ReportMissingSessionState.test.tsx`：分别测各自渲染 + errorCode 文案映射（B1 `AI_*` enum 全覆盖）+ CTA 行为（点击「重新生成」/「返回 workspace」分别触发 nav）+ `TestReportFailureStateRendersNotFoundCopy`：404 / `REPORT_NOT_FOUND` 路径使用 `failureState.notFound.*` i18n key 与 AI_* enum 文案显式区分
- [x] 2.8 实现 `report/__tests__/useFeedbackReport.test.ts`：4 态 + 404 cross-user → ReportFailureState + 5xx + retry + 不含 `Idempotency-Key`
- [x] 2.9 实现 `report/__tests__/useReportContextData.test.ts`：`getTargetJob` + `getResumeVersion` 成功显示人类可读 label；单 operation 失败回退对应 ID；不读取 raw resume/JD/body 字段；request init 不含写操作 header
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
- [x] 3.9 实现 `report/__tests__/tabs/{ReadinessTab,DimensionsTab,QuestionsTab,EvidenceTab,NextTab}.test.tsx`：每个测对应 testid + 数据驱动 + 边界（空 dimensions / 空 questions / 空 issues / 空 highlights 各自 EmptyHint）+ 4 档 readiness 文案矩阵 + 三档维度状态矩阵 + 负向断言不出现旧 5 档 readiness numeric / rubric score_levels label
- [x] 3.10 BDD-Gate: 验证 `E2E.P0.056` ReportDashboard 渲染部分通过（含 5 detail tab 切换 + 维度卡片 + 题目回顾 + 风险亮点）

## Phase 4: 复练 CTA 行为 + ReportFailureState 完整 + GeneratingScreen handoff 完整

- [x] 4.0 在 `frontend/src/app/auth/pendingAction.ts` 注册新 `replay_practice` PendingAction type：加入 allowlist（如有 string union / validator）+ `encodePendingAction` / `decodePendingActionRoute` 支持 round-trip + 路由恢复到 `workspace` 并保留 `autoStartPractice=1`；新增 `frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts` 覆盖 `TestPendingActionEncodeDecodeReplayPractice` + `TestPendingActionReplayPracticeTypeAllowed` + 负向断言 URL params / localStorage 不含 raw text
- [x] 4.1 实现复练 CTA `goReplay()` 路径 A：在 `ReportHeader.tsx` 与 `tabs/NextTab.tsx` 的 `report-next-cta-a` 按钮上绑定 `goReplay()`；组装 payload `{ sourceSessionId, replayItems:retryFocusTurnIds, evidenceGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:InterviewContext.practiceMode, practiceGoal:'retry_current_round', autoStartPractice:'1' }`；未登录 → `useRequestAuth({type:'replay_practice', route:'workspace', params:{...sameParams}})`；已登录 → `nav("workspace", payload)`，由 workspace owner 创建 fresh practice session 后进入 practice
- [x] 4.2 实现复练 CTA `goNextRound()` 路径 B：同上但 payload 为 `{ nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:InterviewContext.practiceMode, practiceGoal:'next_round', autoStartPractice:'1' }`；nextRoundId 默认从 InterviewContext.roundId 推断；workspace owner 创建 fresh session 后进入 practice
- [x] 4.3 完整 ReportFailureState handoff：`ReportFailureState.tsx` CTA「重新生成」点击 → `nav("generating", { sessionId, reportId, ...passThroughContext })`；CTA「返回 workspace」点击 → `nav("workspace", { targetJobId, jdId, planId, resumeVersionId })`
- [x] 4.4 完整 GeneratingScreen handoff：`useReportGenerationPoll` 的 `onReady(report)` callback → `nav("report", { sessionId, reportId, ...passThrough })`；`onFailed(errorCode)` callback → `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode, ...passThrough })`；timeout state → 不自动 nav，用户点 retry 重启轮询；nav 调用必须防抖（handoffNavigatedRef）
- [x] 4.5 复练 CTA 数据未 ready 时禁用：report status='generating' 兜底（虽然不应进入但仍兜底）→ CTA disabled；不发 nav
- [x] 4.6 实现 `report/__tests__/ReplayCta.test.tsx`：路径 A 已登录经 workspace auto-start 创建 fresh session + 未登录 useRequestAuth 恢复到 workspace；payload 字段完整；负向断言 raw text（`answerText` / `questionText` / `hint` / `promptHash` / `modelId raw`）不在 payload；负向断言 `getFeedbackReport` / `appendSessionEvent` 在 ReplayCta 上下文中不被调用
- [x] 4.7 在 `report/__tests__/ReplayCta.test.tsx` 覆盖路径 B 同上；nextRoundId 推断逻辑测试
- [x] 4.8 实现 `report/__tests__/ReportFailureHandoff.test.tsx`：「重新生成」nav generating + 「返回 workspace」nav workspace；errorCode 不在 generating route params 中暴露 raw provider error
- [x] 4.9 扩展 `generating/__tests__/GeneratingScreen.test.tsx`：ready / failed / timeout 三态分别 nav + 防抖（多次 ready callback 只 nav 一次；fake timer 验证 1 次 nav）
- [x] 4.10 BDD-Gate: 验证 `E2E.P0.057` 通过（复练 CTA 路径 A + 路径 B 经 workspace auto-start 进入 fresh practice session）
- [x] 4.11 BDD-Gate: 验证 `E2E.P0.058` 通过（GeneratingScreen 轮询命中 `status='failed'` → nav failed report + ReportFailureState + ReportMissingSessionState + 跨用户 + 隐私 route params）
- [x] 4.12 BDD-Gate: 验证 `E2E.P0.056` 整链完整通过（含 GeneratingScreen mount → 进度动画 → 轮询 ready → nav report → ReportDashboard 渲染 → 5 detail tab 切换 → CTA wire 完整；Phase 1 + Phase 3 仅作局部断言，Phase 4 复练 CTA wire 完成后整链通过）

## Phase 5: 完整状态机集成 + Playwright pixel parity + scenario 加挂 + 旧口径负向

- [x] 5.1 `pnpm vitest run`（全 frontend 测试）+ `pnpm typecheck` 全绿；扩展 `App.test.tsx` 添加 `generating-screen` 与 `report-dashboard` testid 命中断言；扩展 `AppNormalize.test.tsx` 添加 `generating` / `report` route alias 处理；扩展 `pendingActionReplayPractice.test.ts` 添加 `replay_practice` pendingAction 恢复到 workspace auto-start 的 round-trip；扩展 `scenarios/p0-002-auth-pending-action-resume.test.tsx` 添加 `replay_practice` resume path 验证
- [x] 5.2 新增 `frontend/tests/pixel-parity/generating.spec.ts`：desktop 1440×900 + mobile 390×844；测 generating 主屏 + ErrorState + 5 阶段 + 8 主题 × dark 切换；clean-checkout 硬 gate 为 DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke，只有稳定 baseline 已提交或本 phase 明确更新 baseline 时才追加 `toHaveScreenshot`
- [x] 5.3 新增 `frontend/tests/pixel-parity/report.spec.ts`：desktop + mobile；测 report 主屏 + 5 detail tab + 三态（dashboard/failure/missing-session）+ 8 主题 × dark 切换；clean-checkout 硬 gate 同 5.2，且验证 `report` 默认 App chrome / TopBar 可见、不进入一级导航
- [x] 5.4 派生 4 个 scenario 目录 `test/scenarios/e2e/p0-056-generating-to-report-happy-path/` / `p0-057-replay-cta-paths-a-and-b/` / `p0-058-report-failure-and-missing-session/` / `p0-059-report-pixel-parity-i18n-and-legacy-negative/`，每个含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x 可执行）；trigger 跑对应 Vitest 套件 / Playwright spec；verify 反查 testid / nav payload / 负向 grep / i18n 完整性
- [x] 5.5 更新 `test/scenarios/e2e/INDEX.md` 在 P0 表追加 4 行（E2E.P0.056 / 057 / 058 / 059）；状态 Ready，automated
- [x] 5.6 新增 `scripts/lint/frontend_report_dashboard_legacy.py`：scoped grep 在 `frontend/src/app/screens/{report,generating}/`：`reportLayout` / `report_layout` / 旧 5 档 readiness（`fully_prepared` 等旧字面量） / `readinessScore` / `readiness_score` numeric / `mistakes_queue` / `mistakesQueue` / `drill_builder` / `drillBuilder` / `growth_center` / `growthCenter` / `report_timeline` / `reportTimeline` / `report_form` / `reportForm` / 旧独立 `mistakes` route entry / `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief` / `listTargetJobReports` 在实现代码中零出现；本 plan / BDD / test docs / spec §D-12 prohibition / preflight.test.ts 不属于实现范围；新增 `scripts/lint/frontend_report_dashboard_legacy_test.py` 覆盖 `test_frontend_report_dashboard_legacy_includes_terms` / `test_frontend_report_dashboard_legacy_allows_negative_docs`
- [x] 5.7 新增 `frontend/src/app/screens/report/__tests__/legacyNegative.test.ts` + `frontend/src/app/screens/generating/__tests__/legacyNegative.test.ts`：grep negative 同上集合 + 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / prototype helper + 不 import `VoiceSessionSurface` 等 voice 组件 + 不调 Practice operation + `TestListTargetJobReportsNotInvokedInReportOrGenerating`（mockTransport spy 断言 `listTargetJobReports` 调用次数 = 0）
- [x] 5.8 新增 `frontend/src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts`：断言 `report.*` 与 `generating.*` 命名空间 zh / en 同步无缺漏（key 集合相等）+ 新增 key ≥ 60 + 切换 locale 时所有 testid 文案重绘 + `report.failureState.errorCode.*` 覆盖 B1 `AI_*` enum 当前全部值（用 generated B1 常量做 source of truth）+ `TestReportFailureStateErrorCodeCoversReportNotFound` 显式断言 `report.failureState.errorCode.REPORT_NOT_FOUND` 与 `report.failureState.notFound.*` key 存在且 zh / en 同步
- [x] 5.9 跨 owner regression：在 Phase 5 收口阶段重跑：frontend-workspace-and-practice/002 BDD `E2E.P0.044-047`（保证未被破坏）；如 backend-review/001 Phase 5 已实施则跑 `E2E.P0.052-055` backend BDD；backend-practice/002 BDD `E2E.P0.038-043` 必要时通过 cmd/api 重跑
- [x] 5.10 BDD-Gate: 验证 `E2E.P0.059` 通过（Playwright pixel parity + i18n + 旧口径负向）
- [x] 5.11 新增 `frontend/src/app/screens/report/README.md` + `frontend/src/app/screens/generating/README.md`：简明 handoff 段落，记录 001 新增 component / hook / nav 边界 / handoff 给 backend-review 与 frontend-workspace-and-practice 的边界；引用 D-1 ~ D-14 决策
- [x] 5.12 收口 gate：`pnpm vitest run` 全绿 + `pnpm typecheck` 全绿 + `pnpm test:pixel-parity` 全绿 + `pnpm build` 全绿 + `make codegen-check` 通过 + `make validate-fixtures` 通过（依赖 backend-review/001 Phase 0 新增 fixture variants）+ `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过 + `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q` 通过 + `make docs-check` 通过 + `git diff --check` 通过
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
- `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q`：3 passed。
- `test/scenarios/e2e/p0-056-generating-to-report-happy-path/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-058-report-failure-and-missing-session/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative/scripts/{setup,trigger,verify,cleanup}.sh`
- 2026-05-16 cross-owner regression: `E2E.P0.044` / `E2E.P0.045` / `E2E.P0.046` / `E2E.P0.047` setup → trigger → verify → cleanup 全部通过。
- `make docs-check`
- `git diff --check`
