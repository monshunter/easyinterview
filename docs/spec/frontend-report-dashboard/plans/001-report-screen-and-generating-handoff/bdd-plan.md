# 001 — Report Screen and Generating Handoff BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 的 4 个 BDD 场景保留 `E2E.P0.xxx` 编号与 Given / When / Then 语义；001 本次代码交付的可执行入口落在 vitest + jsdom 主套件 + Playwright pixel parity，同步派生 4 个 `test/scenarios/e2e/p0-NNN-*` 目录，trigger 跑对应 Vitest / Playwright 套件，verify 反查 testid / nav payload / 负向 grep / i18n 完整性。

- 套件: `e2e`
- 阶段: `P0`
- 已占用编号现状（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md)）：`001-006`, `010-047`；backend-practice/003 已通过 Go HTTP scenario 预留 `048-051`（hint 四场景）；backend-review/001 预留 `052-055`（report 四场景，已在 backend-review/001 plan 内固化）。本 plan 在空闲号段 `056-059` 中分配 4 个场景
- 编号分配: `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059`
- 执行入口: vitest + jsdom 主套件 `pnpm --filter @easyinterview/frontend test` + Playwright `pnpm --filter @easyinterview/frontend test:pixel-parity` + 4 个 scenario 目录的 verify.sh
- 外部 Kind / shell 场景资产: 001 派生 `test/scenarios/e2e/p0-{056,057,058,059}-*` 目录与 backend-practice/002 + frontend-workspace-and-practice/002 同模式

每个场景的执行证据在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC / D |
|---------|------|------|----------------|-------------------|
| `E2E.P0.056` | GeneratingScreen 轮询 → ReportScreen happy path（5 阶段进度 + 轮询 ready → nav report → 5 detail tab + summary cards + 维度卡片 + 题目回顾 + 风险亮点） | primary | Phase 1 + Phase 2 + Phase 3 + Phase 4 | C-1, C-2, C-5, C-8, C-11 |
| `E2E.P0.057` | 复练 CTA 路径 A retry_current_round + 路径 B next_round nav practice + 未登录走 useRequestAuth + 已登录直接 nav + payload 完整性 | alternate | Phase 4 | C-9, C-10, D-5 |
| `E2E.P0.058` | GeneratingScreen failed handoff + ReportFailureState + ReportMissingSessionState + 跨用户 404 + 隐私 route params（不含 raw text） | failure + privacy | Phase 1 + Phase 2 + Phase 4 + Phase 5 | C-3, C-6, C-7, C-15, D-6 |
| `E2E.P0.059` | Playwright pixel parity desktop + mobile + 8 主题 × dark + i18n zh/en 完整性 + 旧 reportLayout / 5 档 readiness / 报告时间线 / mistakes / drill / growth_center 负向 grep | regression + UX | Phase 5 | C-12, C-13, D-7, D-10, D-11, D-12, D-13 |

## 2 Phase 1 + 2 + 3 + 4 — GeneratingScreen → ReportScreen happy path

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.056` | GeneratingScreen → ReportScreen happy path | 用户已登录；frontend-workspace-and-practice/002 已完成 completePracticeSession 模拟，handoff 到 `generating?planId=&targetJobId=&jdId=&resumeVersionId=&roundId=&sessionId=S&reportId=R&mode=text&modality=text&practiceMode=assisted&practiceGoal=baseline&hintUsed=true&hintCount=2`（13 字段 buildPracticeHandoffParams 输出）；fixture `getFeedbackReport` 配置为前 2 次轮询返回 `report-generating`（status='generating'），第 3 次返回 `default`（status='ready' + 完整内容）；`getTargetJob` / `getResumeVersion` fixture 返回 target title/companyName + resume displayName；InterviewContext 完整 hydrate；i18n locale=zh | （A）进入 `generating` route 触发 GeneratingScreen mount；（B）等待 ~3 次轮询（fake timer 推进 1.5s + 2.25s + 3.375s = 7.125s）；（C）status='ready' 触发 onReady callback nav report；（D）进入 `report` route 触发 ReportScreen mount + useFeedbackReport 拉取；（E）默认渲染 `questions` tab；（F）用户依次点击 readiness / dimensions / questions / evidence / next 5 个 tab；（G）切换 zh ↔ en；（H）切换 dark + customAccent | （1）GeneratingScreen 渲染：≥ 10 个 `generating-*` testid 命中；5 阶段列表 done/active/pending 状态正确切换；progress 进度条 % 增长；live evidence stream fade-in；底部 P95 SLA 提示文案与 ui-design 一致；（2）useReportGenerationPoll 调用 `getFeedbackReport(R)` 3 次；request init 不含 `Idempotency-Key`；（3）status='ready' → handoffNavigatedRef 防抖；nav report 调用 1 次；route params 含 sessionId + reportId + 13 字段 InterviewContext + 不含 raw text；（4）ReportScreen 路由命中 ReportDashboard（非 FailureState 非 MissingSession）；（5）≥ 20 个 `report-*` testid 命中：返回按钮 + Header + ContextStrip + 4 Summary Cards + DetailSurface tablist + DimRow × N + perq cards × 5 + issues × N + highlights × N + 复练 CTA × 2；（6）默认 `questions` tab panel active；readiness 通过显式点击激活；其他 panel 不渲染（display:none 或不在 DOM）；（7）依次点击 readiness / dimensions / questions / evidence / next tab，对应 panel 渲染；ARIA tablist / tab / tabpanel role 正确；（8）ContextStrip 通过 `getTargetJob` / `getResumeVersion` 显示 sessionId / targetJob.title + companyName / round（由 roundId i18n 本地推导）/ resumeVersion.displayName / modality.text / practiceMode.assisted / hints.2，任一 label 请求失败时回退 ID；（9）SummaryCards 4 张：准备度（拨号盘缩略 + 档位文案 zh/en）/ 维度（数量计数）/ 题目（5/5 计数）/ 下一步（next_action 文案）；（10）维度卡片行：DimRow × N（数量 = dimensions.length），每行 strong/meets_bar/needs_work 三态色调正确；（11）准备度 4 档（not_ready / needs_practice / basically_ready / well_prepared）文案 zh/en 矩阵正确；不出现旧 5 档 numeric / 旧 `readiness_score` 字面量；（12）切 zh ↔ en 关键文案重绘；切 dark / customAccent 关键元素 computed background / color 可见变化；（13）负向：`window.EI_DATA` / `data.jsx` literal grep 0 命中；voice 组件 import 0 命中；`getPracticeSession` / `appendSessionEvent` / `completePracticeSession` / `createPracticePlan` / `startPracticeSession` / `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief` / `listTargetJobReports` 调用 0 命中；旧 prototype testid（`mistakes-*` / `drill-*` / `growth-*` / `report-timeline-*` / `reportLayout-*`）0 命中 | `test/scenarios/e2e/p0-056-generating-to-report-happy-path/` |

## 3 Phase 4 — 复练 CTA 路径 A + 路径 B

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.057` | 复练 CTA 路径 A retry_current_round + 路径 B next_round | 分四子场景：（A）已登录用户 + ReportDashboard 已渲染（fixture `getFeedbackReport=default` 返回 ready report 含 retryFocusTurnIds=['turn-1','turn-3','turn-5'] + 准备度=needs_practice + next_action 第一行 type='retry_current_round'）；（B）已登录用户 + ReportDashboard 渲染（fixture 返回 ready report 准备度=basically_ready + next_action 第一行 type='next_round'）；（C）未登录用户 + ReportDashboard 同 A；（D）已登录用户 + 数据未 ready（fixture 持续返回 generating） | （A）点击 ReportHeader 的「复练当前轮」CTA 或 NextTab 的 `report-next-cta-a`；（B）点击「进入下一轮」CTA 或 `report-next-cta-b`；（C）点击「复练当前轮」CTA → useRequestAuth 触发 → 模拟 auth 成功 → pendingAction 恢复到 report → 自动 nav practice；（D）尝试点击 CTA | （A）`nav("practice", { sourceSessionId:S, replayItems:['turn-1','turn-3','turn-5'], evidenceGaps:..., planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:'assisted', practiceGoal:'retry_current_round' })`；不含 raw text（answerText / questionText / hint / promptHash / modelId raw）；`getFeedbackReport` 不被重复调用；（B）`nav("practice", { nextRoundId:R+1, roundName:'Round 2', roundId:R+1, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:'assisted', practiceGoal:'next_round' })`；（C）useRequestAuth 调用 `{type:'replay_practice', route:'report', params:{...sameParams, autoReplay:'1'}}`；auth 成功后 pendingAction.resume() 触发 nav practice with same payload；（D）CTA disabled（aria-disabled='true'）+ 点击不发 nav；（隐私）所有 nav payload + URL params + console.log + localStorage / sessionStorage / telemetry 不含 raw text / hint / question / answer / prompt body / response body / provider raw model id / promptVersion raw（promptVersion 是 generated provenance 6 字段之一，可暴露 version 字符串，但不暴露 prompt body）；（负向）`createPracticeVoiceTurn` / `getCompanyIntel` 调用 0 命中；旧 prototype CTA testid 0 命中 | `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/` |

## 4 Phase 2 + 4 + 5 — ReportFailureState + ReportMissingSession + cross-user + privacy

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.058` | GeneratingScreen failed handoff + ReportFailureState + ReportMissingSessionState + cross-user 404 + 隐私 route params | 分六子场景：（A0）`generating?sessionId=S&reportId=R&...passThrough` + fixture `getFeedbackReport=report-failed`（status='failed' + errorCode='AI_PROVIDER_TIMEOUT'）；（A）`report?sessionId=S&reportId=R&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT&...passThrough`；（B）`report?reportId=R`（缺 sessionId）；（C）用户 A 调 `getFeedbackReport(R_A)` 后 nav report；用户 B 切换 auth 后调 `getFeedbackReport(R_A)` → 404；（D）`generating?sessionId=S` 缺 reportId；（E）fixture `getFeedbackReport` 永久返回 generating（模拟超时）| （A0）进入 `generating` 路由并等待一次轮询；（A）进入 `report` 路由；（B）进入 `report` 路由；（C）用户 B 通过 URL 直接访问 `report?sessionId=S_A&reportId=R_A`；（D）进入 `generating` 路由；（E）进入 `generating` 路由 + 等待 max attempts | （A0）`getFeedbackReport(R)` 返回 `status='failed'` 后自动 `nav("report", {sessionId, reportId, reportStatus:'failed', errorCode:'AI_PROVIDER_TIMEOUT', ...passThrough})`；不渲染 ReportDashboard；nav 防抖 1 次；（A）渲染 `report-failure-state-{title,desc,error-code,retry-cta,back-to-workspace}` testid；errorCode 文案映射 zh/en 正确（`AI_PROVIDER_TIMEOUT` → 「AI 服务超时，请重试」/ "AI service timeout, please retry"）；不调 `getFeedbackReport`；CTA「重新生成」→ nav generating；CTA「返回 workspace」→ nav workspace；（B）渲染 `report-missing-session-{title,desc,cta}` testid；不调 `getFeedbackReport`；CTA → nav workspace；（C）用户 B 调 `getFeedbackReport(R_A)` 返回 404 `REPORT_NOT_FOUND` → ReportFailureState 但使用独立的 `report.failureState.notFound.{title,desc}` / `report.failureState.errorCode.REPORT_NOT_FOUND` i18n key（与 AI_* enum 文案显式区分），文案为 zh "未找到该报告" / en "Report not found"；不暴露 R_A 存在性；errorCode 显示 generic 文案而非 AI errorCode 文案；testid `report-failure-state-not-found-{title,desc}` 命中；（D）GeneratingScreen 渲染 generating-error-state；不调 `getFeedbackReport`；CTA「返回 workspace」可用；（E）max attempts 达到 → state='timeout' → 渲染 timeout ErrorState + retry CTA；retry → 重启轮询；3 次 timeout 后显示「返回 workspace」fallback；（隐私）所有失败路径 console.log / URL search params / localStorage / sessionStorage / telemetry 不含 raw text；errorCode 仅暴露 B1 enum 字符串（含 `REPORT_NOT_FOUND`），不暴露 raw provider error message；`listTargetJobReports` 调用次数 = 0 | `test/scenarios/e2e/p0-058-report-failure-and-missing-session/` |

## 5 Phase 5 — Pixel parity + i18n + legacy negative

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.059` | Playwright pixel parity + i18n + 旧口径负向 | Playwright runner + frontend dev server + 8 主题 × dark + customAccent；i18n locale zh + en；fixture `getFeedbackReport=default`（ready 完整 + 准备度 basically_ready + 5 dimensions + 5 questions + 3 highlights + 2 issues + 3 nextActions）；clean-checkout parity 不依赖未提交 screenshot baseline | （A）Playwright 加载 `/generating?...` desktop 1440×900；（B）desktop 5 阶段进度推进截图 smoke；（C）mobile 390×844 加载 `/generating`；（D）desktop 加载 `/report?...` 默认 `questions` tab；（E）切换 readiness / dimensions / questions / evidence / next 5 tab 各自断言；（F）desktop ReportFailureState 截图 smoke；（G）desktop ReportMissingSessionState 截图 smoke；（H）mobile 加载 `/report` 默认 `questions` tab；（I）mobile 切换 detail tab；（J）8 主题 × dark / customAccent 循环对每屏做 computed style + non-empty screenshot smoke；（K）切换 locale zh ↔ en 对关键屏做文案断言；（L）运行 i18n 完整性测试与 scoped legacy grep | （A-K）DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke 全部通过；仅当稳定 baseline 已提交或本 phase 明确更新 baseline 时才追加 `toHaveScreenshot` maxDiffPixels 阈值；`generating` TopBar 隐藏不占位，`report` 默认 App chrome / TopBar 可见且不进入一级导航；bounding box stays in viewport，no overlap；底部 CTA sticky 不被遮挡；mobile 三列折叠为单列 + Detail Surface Accordion；（L1）`TestReportNamespaceZhEnSync` 通过：`report.*` zh / en key 集合相等；（L2）`TestGeneratingNamespaceZhEnSync` 通过；（L3）`TestErrorCodeI18nCoversAllAIErrors` 通过：`report.failureState.errorCode.*` 覆盖 B1 `AI_*` enum 全集（用 generated B1 常量做 source of truth）；（L4）`python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过：在 `frontend/src/app/screens/{report,generating}/` 范围 grep 以下字面量零出现（除负向 test / docs allowlist）：`reportLayout` / `report_layout` / 旧 5 档 readiness（`fully_prepared` 等）/ `readinessScore` / `readiness_score` numeric / `mistakes_queue` / `mistakesQueue` / `drill_builder` / `drillBuilder` / `growth_center` / `growthCenter` / `report_timeline` / `reportTimeline` / `report_form` / `reportForm` / 旧独立 `mistakes` route entry / `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief` / `VoiceSessionSurface` 等 voice 组件 import / `window.EI_DATA` / `data.jsx` import；（L5）`legacyNegative.test.ts`（在 report / generating __tests__ 各一份）覆盖同上集合 + 不调 Practice operation；（L6）跨 owner regression 通过：frontend-workspace-and-practice plan 002 BDD `E2E.P0.044-047` Vitest 套件全绿；backend-review/001 BDD `E2E.P0.052-055` 如已 implement 则通过 | `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative/` |

## 6 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个 scenario 使用独立的 `user_id` / `session_id` / `report_id` / `target_job_id` 命名空间；不复用 `E2E.P0.018 ~ E2E.P0.055` 已占用的资源（含 workspace + frontend-workspace-and-practice + backend-practice + backend-review 范围）
- Vitest 套件：每个 test 文件用 fresh `setupTests.ts` mockTransport + InterviewContext 独立 hydrate；不污染其他文件状态
- Playwright：每个 spec 用独立 user / report fixture variant；non-empty screenshot smoke artifacts 与已有 workspace / practice 截图目录隔离；`toHaveScreenshot` baseline 仅在稳定 baseline 已提交或本 phase 明确更新 baseline 时启用
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（mockTransport spy buffer、Playwright browser context）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 7 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API 调用 + UI 渲染 + nav payload + 隐私红线 + pixel parity + scoped legacy grep）；不重复 useReportGenerationPoll / useFeedbackReport 内部 7 态 4 态、ReadinessTier 算法、retry_focus 选择、next_action 决策等单元测试覆盖（详见 [test-plan](./test-plan.md)）。001 阶段不存在 runtime AI 调用（前端不直连 LLM），"AI 失败测试" 通过 fixture variant 模拟，不在 BDD plan 直接测试 backend AI 失败。

## 8 与 spec AC 映射

| spec AC | 覆盖场景 |
|---------|----------|
| C-1（两条 owner route 接管） | `E2E.P0.056`（generating + report 都渲染正式 Screen） |
| C-2（GeneratingScreen 轮询 happy path） | `E2E.P0.056` 子断言（3 次轮询 + nav report） |
| C-3（GeneratingScreen 失败处理） | `E2E.P0.058` A0 子断言（`status='failed'` → nav failed report） |
| C-4（GeneratingScreen 超时） | `E2E.P0.058` ⑤ |
| C-5（ReportDashboard ready 渲染） | `E2E.P0.056` 子断言（≥ 20 testid + 4 Summary Cards + 维度卡片 + 题目回顾 + 风险亮点） |
| C-6（ReportFailureState） | `E2E.P0.058` ① |
| C-7（ReportMissingSessionState） | `E2E.P0.058` ② |
| C-8（5 detail tab 切换） | `E2E.P0.056` 子断言（5 tab + ARIA tablist） |
| C-9（复练 CTA 路径 A） | `E2E.P0.057` （A） + （C） |
| C-10（复练 CTA 路径 B） | `E2E.P0.057` （B） |
| C-11（UI source structure parity） | `E2E.P0.056` 子断言（testid 命中）+ `E2E.P0.059` |
| C-12（UI visual geometry parity） | `E2E.P0.059` |
| C-13（UI stale-contract negative） | `E2E.P0.059` 子断言（scoped legacy grep） |
| C-14（BDD 主流程 + 关键分支） | `E2E.P0.056` + `E2E.P0.057` + `E2E.P0.058` + `E2E.P0.059` |
| C-15（Privacy 红线） | `E2E.P0.056` 子断言（route params + console + localStorage + telemetry）+ `E2E.P0.057` 隐私分支 + `E2E.P0.058` 隐私分支 |
| D-3（GeneratingScreen 轮询节奏） | `E2E.P0.056` + `E2E.P0.058` A0 / ⑤ |
| D-4（状态分支三态） | `E2E.P0.056` + `E2E.P0.058` A0 / ① / ② |
| D-5（复练 CTA payload） | `E2E.P0.057` |
| D-6（报告失败状态语义） | `E2E.P0.058` ① |
| D-7 / D-13（i18n + 隐私红线） | `E2E.P0.059` + `E2E.P0.056-058` 子断言 |
| D-10 / D-11（4 档 readiness + 维度三态映射） | `E2E.P0.056` 子断言 + `E2E.P0.059` |
| D-12（retired 术语） | `E2E.P0.059` 子断言（scoped legacy grep） |
