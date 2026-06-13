# 001 — Report Screen and Generating Handoff BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-06-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.056 GeneratingScreen → ReportScreen happy path

- [x] 创建场景目录 `test/scenarios/e2e/p0-056-generating-to-report-happy-path/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x）
- [x] 准备 fixture：`getFeedbackReport` 配置前 2 次 `report-generating` + 第 3 次 `default`（ready 含 5 dimensions + 5 questions + 3 highlights + 2 issues + 3 next_actions + retryFocusTurnIds=['turn-1','turn-3','turn-5'] + preparednessLevel='basically_ready' + provenance 6 字段）；fixture-backed transport 通过 `EI_FIXTURE_SCENARIO_*` 环境变量切换
- [x] 准备 ContextStrip fixture：`getTargetJob=default` + `getResumeVersion=default`，返回 target title/companyName + resume displayName；同时准备单 operation 失败 fallback case
- [x] 实现 setup.sh：构造 frontend 启动命令 + fake timer 控制 + InterviewContext 13 字段 hydrate
- [x] 实现 trigger.sh：执行 vitest 套件 `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-056-generating-to-report-happy-path.test.tsx`（或对应 spec name）
- [x] 实现 verify.sh：
  - generating-screen testid 命中 ≥ 10
  - report-dashboard testid 命中 ≥ 20
  - generated client `getFeedbackReport` 调用次数 = 3（轮询）+ 1（ReportScreen mount） = 4 次
  - generated client `listTargetJobReports` 调用次数 = 0（dashboard-only D-7 反向断言）
  - request init 不含 `Idempotency-Key` header（grep 反查）
  - nav generating → report 调用次数 = 1（防抖）
  - route params 含 sessionId + reportId + 13 字段 InterviewContext（与 `buildPracticeHandoffParams` 输出一致）
  - route params 不含 raw text（grep `answerText` / `questionText` / `hint:` / `promptHash` / `modelId.*raw` 0 命中）
  - 5 detail tab 切换顺序：questions（默认）→ readiness → dimensions → evidence → next；ARIA tablist / tab / tabpanel role 正确
  - ContextStrip 7 字段显示；target/resume label 来自 `getTargetJob` / `getResumeVersion`，失败时回退 ID；不读取 raw resume/JD/body
  - 4 Summary Cards 渲染
  - 维度卡片行 DimRow × N（与 fixture dimensions.length 一致）
  - 题目回顾 perq cards × 5
  - issues × 2 + highlights × 3
  - 复练 CTA × 2 渲染
  - 准备度 tier 文案：「基本可面」/ "basically ready"（zh / en 各一）
  - 维度状态文案：strong / meets_bar / needs_work 三态映射正确
  - 切 zh ↔ en 关键文案重绘
  - 切 dark / customAccent 关键元素 computed background / color 可见变化
  - 负向：`window.EI_DATA` / `data.jsx` literal grep 0 命中；voice 组件 import 0 命中；`getPracticeSession` / `appendSessionEvent` / `completePracticeSession` / `createPracticePlan` / `startPracticeSession` / `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief` 调用 0 命中；旧 prototype testid 0 命中
- [x] 实现 cleanup.sh：按 [bdd-plan §6](./bdd-plan.md#6-数据隔离与污染恢复) 顺序清理 mockTransport spy buffer + InterviewContext + Playwright browser context（如有）
- [x] 执行 `bash test/scenarios/e2e/p0-056-generating-to-report-happy-path/scripts/setup.sh && bash .../trigger.sh && bash .../verify.sh && bash .../cleanup.sh` 全绿
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 row：`E2E.P0.056 | frontend-report-dashboard C-1 C-2 C-5 C-8 C-11 | p0-056-generating-to-report-happy-path/ | ... | automated | Ready`
- [x] 记录验证证据到 plan §3.6 L2 修订说明（如经过 L2 review）或本 checklist 收口段

## E2E.P0.057 复练 CTA 路径 A + 路径 B

- [x] 创建场景目录 `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/`，含完整资产
- [x] 准备 fixture：4 子场景各自的 `getFeedbackReport` variant + auth 状态切换 mock
- [x] 实现 setup.sh / trigger.sh：分别 mount 4 子场景；点击 CTA；模拟未登录 useRequestAuth 流程
- [x] 实现 verify.sh：
  - （A）已登录 + 路径 A → nav workspace auto-start 调用 1 次；payload = { sourceSessionId, replayItems:['turn-1','turn-3','turn-5'], evidenceGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:'assisted', practiceGoal:'retry_current_round', autoStartPractice:'1' }；workspace owner 调用 startPracticeSession 并进入 fresh practice session；payload 字段 grep；负向 grep `answerText` / `questionText` / `hint:` 不在 payload
  - （B）已登录 + 路径 B → nav workspace auto-start 调用 1 次；payload = { nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:'assisted', practiceGoal:'next_round', autoStartPractice:'1' }；workspace owner 调用 startPracticeSession 并进入 fresh practice session
  - （C）未登录 + 路径 A → useRequestAuth 调用 1 次 + 参数验证；模拟 auth 成功后 pendingAction.resume() 恢复 workspace auto-start with same payload
  - （D）数据未 ready 兜底 → CTA aria-disabled='true' + 点击不发 nav
  - 所有 nav payload + URL params + console.log + localStorage / sessionStorage / telemetry 不含 raw text
  - `getFeedbackReport` 在 CTA 触发后不重复调用
  - 负向：`createPracticeVoiceTurn` / `getCompanyIntel` 调用 0 命中
- [x] 实现 cleanup.sh
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## E2E.P0.058 GeneratingScreen failed handoff + ReportFailureState + ReportMissingSessionState + cross-user + privacy

- [x] 创建场景目录 `test/scenarios/e2e/p0-058-report-failure-and-missing-session/`，含完整资产
- [x] 准备 fixture：6 子场景 fixture / auth 状态切换 mock；`getFeedbackReport=report-failed`（`response.body.status='failed'` + errorCode='AI_PROVIDER_TIMEOUT'）；cross-user 404 mock；timeout 模拟 fixture（永久 generating）
- [x] 实现 setup.sh / trigger.sh
- [x] 实现 verify.sh：
  - （A0）`generating?sessionId=S&reportId=R&...` + fixture `report-failed` → 第一次轮询返回 `status='failed'` 后自动 nav `report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT&...passThrough`；不渲染 ReportDashboard；nav 防抖 1 次
  - （A）`report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT&...` → ReportFailureState testid 命中；errorCode 文案映射（zh: 「AI 服务超时，请重试」/ en: "AI service timeout, please retry"）；CTA「重新生成」→ nav generating；CTA「返回 workspace」→ nav workspace；`getFeedbackReport` 调用次数 = 0
  - （B）`report?reportId=R`（缺 sessionId）→ ReportMissingSessionState testid 命中；CTA → nav workspace；`getFeedbackReport` 调用次数 = 0
  - （C）用户 B 调 `getFeedbackReport(R_A)` 返回 404 `REPORT_NOT_FOUND` → ReportFailureState 使用独立的 `report.failureState.notFound.{title,desc}` / `report.failureState.errorCode.REPORT_NOT_FOUND` i18n key（与 AI_* enum 文案显式区分），文案为 zh "未找到该报告" / en "Report not found"；testid `report-failure-state-not-found-{title,desc}` 命中；不暴露 R_A 存在性；不与 AI errorCode 文案混淆；负向断言：errorCode 不映射到 `failureState.errorCode.UNKNOWN`
  - （D）`generating?sessionId=S`（缺 reportId）→ GeneratingErrorState testid 命中；不调 `getFeedbackReport`；CTA「返回 workspace」可用
  - （E）`generating?reportId=R` + fixture 永久 generating + max attempts 达到 → state='timeout' → 渲染 timeout ErrorState + retry CTA；retry → 重启轮询；3 次 timeout 后显示「返回 workspace」fallback
  - 所有失败路径 console.log / URL search params / localStorage / sessionStorage / telemetry 不含 raw text；errorCode 仅暴露 B1 enum 字符串
- [x] 实现 cleanup.sh
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## E2E.P0.059 Playwright pixel parity + i18n + 旧口径负向

- [x] 创建场景目录 `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative/`，含完整资产
- [x] 准备 fixture：`getFeedbackReport=default`（ready 完整 + 准备度 basically_ready + 完整字段）；8 主题 × dark / customAccent 切换 helper；zh / en locale 切换 helper
- [x] 实现 setup.sh：准备场景输出目录；Playwright webServer 由 frontend config 托管
- [x] 实现 trigger.sh：执行 i18n 测试 + scoped legacy grep + frontend build + Playwright 套件 `pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts`
- [x] 实现 verify.sh：
  - `trigger.log` 必须包含 frontend build 与 Playwright run marker
  - `trigger.log` 必须包含 `tests/pixel-parity/generating.spec.ts` / `tests/pixel-parity/report.spec.ts` 两个实际执行路径
  - `trigger.log` 必须在 Playwright run marker 之后包含 passed marker，不能只检查 spec 文件存在
  - Playwright generating.spec.ts（desktop 1440×900 + mobile 390×844 + 5 阶段进度 + ErrorState + 8 主题 × dark）全绿
  - Playwright report.spec.ts（desktop + mobile + ReportDashboard + 5 detail tab + ReportFailureState + ReportMissingSessionState + 8 主题 × dark）全绿
  - DOM anchor / computed style / bounding box / responsive geometry / non-empty screenshot smoke 全部通过；仅当稳定 baseline 已提交或本 phase 明确更新 baseline 时才追加 `toHaveScreenshot`
  - bounding box stays in viewport, no overlap
  - `generating` TopBar 隐藏不占位；`report` 默认 App chrome / TopBar 可见且不进入一级导航
  - mobile 三列折叠为单列 + Detail Surface Accordion + CTA sticky bottom
  - `TestReportNamespaceZhEnSync` 通过：`report.*` zh / en key 集合相等
  - `TestGeneratingNamespaceZhEnSync` 通过
  - `TestErrorCodeI18nCoversAllAIErrors` 通过：`report.failureState.errorCode.*` 覆盖 B1 `AI_*` enum 全集（用 generated B1 常量做 source of truth）
  - `TestI18nKeyCountAtLeast60` 通过（`report.*` + `generating.*` ≥ 60 keys）
  - `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过：在 `frontend/src/app/screens/{report,generating}/` 范围 grep 以下字面量零出现：
    - `reportLayout` / `report_layout`
    - 旧 5 档 readiness 字面量（`fully_prepared` 等）
    - `readinessScore` / `readiness_score` numeric
    - `mistakes_queue` / `mistakesQueue`
    - `drill_builder` / `drillBuilder`
    - `growth_center` / `growthCenter`
    - `report_timeline` / `reportTimeline`
    - `report_form` / `reportForm`
    - 旧独立 `mistakes` route entry
    - `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief`
    - `VoiceSessionSurface` / `PracticeWaveformBars` 等 voice 组件 import
    - `window.EI_DATA` / `ui-design/src/data.jsx` import
  - 本 plan / BDD / test docs / spec §D-12 prohibition / `scripts/lint/frontend_report_dashboard_legacy.py` 自身允许枚举字面量作为禁止性断言（不属于实现 / runtime 范围）
  - `legacyNegative.test.ts`（report / generating 各一份）通过
  - 跨 owner regression：scenario `p0-044-047`（frontend-workspace-and-practice/002）重跑通过；backend-review/001 real handler regression `cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1` 在真实 handler 落地后作为回归证据
- [x] 实现 cleanup.sh
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## 收口

- [x] 4 个 scenario 目录 setup / trigger / verify / cleanup 全部执行通过
- [x] `test/scenarios/e2e/INDEX.md` P0 表追加 4 行（E2E.P0.056 / 057 / 058 / 059）；状态 Ready，automated
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` 全绿
- [x] `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all` 通过
- [x] `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q` 通过

## 2026-05-16 L2 review scenario evidence

- `E2E.P0.056` / `057` / `058` / `059` setup → trigger → verify → cleanup 均通过，覆盖 generating ready handoff、复练 CTA A/B、失败/缺参/cross-user、pixel parity + i18n + 旧口径负向。
- `E2E.P0.044` / `045` / `046` / `047` setup → trigger → verify → cleanup 均通过，覆盖 frontend-workspace-and-practice handoff regression。
- 2026-05-23 L2 update: P0.056-P0.059 trigger scripts now run `frontendOwners.realApiMode.test.ts` before fixture-backed report UI subcases, and verify scripts reject missing real-mode marker / default backend base URL / test-file marker; focused real-mode Vitest PASS.
