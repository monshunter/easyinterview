# 001 — Report Screen and Generating Handoff BDD Checklist

> **版本**: 1.14
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.056 GeneratingScreen → ReportScreen happy path

- [x] 创建场景目录 `test/scenarios/e2e/p0-056-generating-to-report-happy-path/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + `scripts/{setup,trigger,verify,cleanup}.sh`（chmod +x）
- [x] 准备 deterministic test clients：五个 focused 文件各自控制 `getFeedbackReport` / `getTargetJob` / `getResume` 响应，不声明跨文件固定请求序列
- [x] ContextStrip focused fixture 使用 flat `getResume(resumeId)` 与 `getTargetJob(targetJobId)`，覆盖成功 label、单 operation fallback 与 raw body avoidance
- [x] 实现 setup.sh：创建场景输出目录并写入 setup metadata
- [x] 实现 trigger.sh：先跑 real-mode bootstrap contract，再执行 preflight、poll hook、GeneratingScreen、ReportScreen、DetailSurface 五个 focused Vitest 文件
- [x] 实现 verify.sh：
  - trigger.log 包含五个 focused test-file pass marker，不能只依赖汇总行
  - poll hook 覆盖状态/backoff/ready/failed/cancel/read-only header；不锁跨文件请求次数
  - GeneratingScreen 与 ReportScreen 分别覆盖 route handoff 和 dashboard/error/missing states；不宣称单一 browser/live-backend journey
  - DetailSurface 覆盖五个 tab 与本地 replay marker
  - implementation `generating-*` / `report-*` testid inventory ≥ 30；out-of-scope lint 通过
  - `listTargetJobReports` 在 report/generating runtime 源码 0 命中
- [x] 实现 cleanup.sh：删除本场景 `.test-output` 目录
- [x] 执行 `bash test/scenarios/e2e/p0-056-generating-to-report-happy-path/scripts/setup.sh && bash .../trigger.sh && bash .../verify.sh && bash .../cleanup.sh` 全绿
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 row：`E2E.P0.056 | frontend-report-dashboard C-1 C-2 C-5 C-8 C-11 | p0-056-generating-to-report-happy-path/ | ... | automated | Ready`
- [x] 记录验证证据到 plan §3.6 L2 修订说明（如经过 L2 review）或本 checklist 收口段

## E2E.P0.057 复练 CTA 路径 A + 路径 B

- [x] 创建场景目录 `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/`，含完整资产
- [x] 准备 fixture：ready report（retryFocusTurnIds/issues）+ authenticated/unauthenticated runtime + generated create-plan/start-session responses + report-route pendingAction
- [x] 实现 setup.sh / trigger.sh：运行 owner preflight、`pendingActionReplayPractice.test.ts` 与 `ReplayCta.test.tsx`
- [x] 实现 verify.sh：
  - （A）已登录 + 路径 A → generated createPracticePlan + startPracticeSession 各 1 次，goal=retry_current_round，fresh session 后直接 nav practice；payload 保留 sourceReportId/sourceSessionId/replayItems/evidenceGaps
  - （B）已登录 + 路径 B → generated createPracticePlan + startPracticeSession 各 1 次，goal=next_round，roundId 按 canonical ladder 轮转，fresh session 后直接 nav practice
  - （C）未登录 report route → CTA 挂载前进入 auth_login；不读 report、不创建 plan、不启动 session
  - （D）`replay_practice` pendingAction encode/decode 保留 `route=report` 与安全 params，不泄漏 reserved keys
  - payload / URL params / localStorage 不含 raw answer / question / hint / prompt / model 内容
  - CTA 触发后 `getFeedbackReport` 不重复调用，`listTargetJobReports` 调用 0 次
  - 负向：`createPracticeVoiceTurn` / `getCompanyIntel` 调用 0 命中
- [x] 实现 cleanup.sh
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## E2E.P0.058 Focused failure contracts

- [x] 创建场景目录 `test/scenarios/e2e/p0-058-report-failure-and-missing-session/`，含完整资产
- [x] 准备 deterministic component/hook clients：failed status、missing ids、HTTP 404、HTTP 5xx/refresh 与 persistent generating；不声明 live backend sequence
- [x] 实现 setup.sh / trigger.sh：real-mode bootstrap 后执行 owner preflight、两个 report state component、两个 report hook/route 与 poll hook 六个 focused 文件
- [x] 实现 verify.sh：
  - trigger.log 包含六个 focused file marker，不能只依赖汇总行
  - ReportFailureState 覆盖 AI_* matrix、REPORT_NOT_FOUND 独立 copy、retry/back handlers 与 UNKNOWN fallback
  - ReportMissingSessionState / ReportScreen 覆盖 missing session/report、failed route、404 rendered state 与 no-fetch 分支
  - useFeedbackReport 覆盖 ready/404/missing/5xx-refresh；useReportGenerationPoll 覆盖 failed/404/timeout/backoff/cancel
  - typed i18n key `AI_PROVIDER_TIMEOUT` 与 `failureState.notFound.title` 存在
  - runner 不挂载 GeneratingScreen，不声明 timeout UI、重复 timeout fallback 或宽泛 URL/storage/telemetry 隐私证据
- [x] 实现 cleanup.sh：删除本场景 `.test-output` 目录
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## E2E.P0.059 Playwright pixel parity + i18n + 范围外输入负向

- [x] 创建场景目录 `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/`，含完整资产
- [x] 准备 fixture：`getFeedbackReport=default` / `report-generating` 与既有 target job、resume、runtime/auth fixtures，分别驱动 dashboard、generating 与失败/缺参浏览器状态
- [x] 实现 setup.sh：准备场景输出目录；Playwright webServer 由 frontend config 托管
- [x] 实现 trigger.sh：执行 owner/browser preflight + i18n 测试 + scoped out-of-scope grep + frontend build + Playwright 套件 `pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts`
- [x] 实现 verify.sh：
  - `trigger.log` 必须包含 frontend build 与 Playwright run marker
  - `trigger.log` 必须包含 `tests/pixel-parity/generating.spec.ts` / `tests/pixel-parity/report.spec.ts` 两个实际执行路径
  - `trigger.log` 必须在 Playwright run marker 之后包含 passed marker，不能只检查 spec 文件存在
  - owner/browser preflight 通过：active spec 与六份 plan artifact 只声明七个已执行浏览器状态的显式证据；两份 Playwright 源码各有真实截图调用与逐状态字节非空断言；P0.059 trigger 包含该 preflight
  - Playwright generating.spec.ts 全绿：desktop 主屏关键 DOM + bounding box、缺 reportId 错误态、mobile 390×844 overflow，三个状态均取得非空内存截图
  - Playwright report.spec.ts 全绿：desktop ReportDashboard + TopBar、ReportMissingSessionState、ReportFailureState、mobile 390×844 overflow，四个状态均取得非空内存截图
  - `TestReportNamespaceZhEnSync` 通过：`report.*` zh / en key 集合相等
  - `TestGeneratingNamespaceZhEnSync` 通过
  - `TestErrorCodeI18nCoversAllAIErrors` 通过：`report.failureState.errorCode.*` 覆盖 B1 `AI_*` enum 全集（用 generated B1 常量做 source of truth）
  - `TestI18nKeyCountAtLeast60` 通过（`report.*` + `generating.*` ≥ 60 keys）
  - `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all` 通过：在 `frontend/src/app/screens/{report,generating}/` 范围 grep 以下字面量零出现：
    - `reportLayout` / `report_layout`
    - 范围外 5 档 readiness 字面量（`fully_prepared` 等）
    - `readinessScore` / `readiness_score` numeric
    - `mistakes_queue` / `mistakesQueue`
    - `drill_builder` / `drillBuilder`
    - `growth_center` / `growthCenter`
    - `report_timeline` / `reportTimeline`
    - `report_form` / `reportForm`
    - 范围外独立 `mistakes` route entry
    - `createPracticeVoiceTurn` / `getCompanyIntel` / `getDebrief`
    - `ui-design/src/screen-practice` practice DOM import
    - `window.EI_DATA` / `ui-design/src/data.jsx` import
  - 本 plan / BDD / test docs / spec §D-12 prohibition / `scripts/lint/frontend_report_dashboard_out_of_scope.py` 自身允许枚举字面量作为禁止性断言（不属于实现 / runtime 范围）
  - `outOfScopeNegative.test.ts`（report / generating 各一份）通过
  - 跨 owner regression：scenario `p0-044-047`（frontend-workspace-and-practice/002）重跑通过；backend-review/001 real handler regression `cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1` 在真实 handler 落地后作为回归证据
- [x] 实现 cleanup.sh
- [x] 执行场景验证全绿
- [x] 在 INDEX 追加 row
- [x] 记录验证证据

## 收口

- [x] 4 个 scenario 目录 setup / trigger / verify / cleanup 全部执行通过
- [x] `test/scenarios/e2e/INDEX.md` P0 表追加 4 行（E2E.P0.056 / 057 / 058 / 059）；状态 Ready，automated
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` 全绿
- [x] `python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase all` 通过
- [x] `python3 -m pytest scripts/lint/frontend_report_dashboard_out_of_scope_test.py -q` 通过

## 2026-05-16 L2 review scenario evidence

- `E2E.P0.056` / `057` / `058` / `059` setup → trigger → verify → cleanup 均通过，覆盖 generating ready handoff、复练 CTA A/B、失败/缺参/cross-user、pixel parity + i18n + 范围外输入负向。
- `E2E.P0.044` / `045` / `046` / `047` setup → trigger → verify → cleanup 均通过，覆盖 frontend-workspace-and-practice handoff regression。
- 2026-05-23 L2 update: P0.056-P0.059 trigger scripts now run `frontendOwners.realApiMode.test.ts` before fixture-backed report UI subcases, and verify scripts reject missing real-mode marker / default backend base URL / test-file marker; focused real-mode Vitest PASS.
